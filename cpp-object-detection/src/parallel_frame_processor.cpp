#include "parallel_frame_processor.hpp"
#include "drawing_utils.hpp"
#include <chrono>
#include <iomanip>
#include <sstream>
#include <set>
#include <sys/stat.h>

ParallelFrameProcessor::ParallelFrameProcessor(std::shared_ptr<ObjectDetector> detector,
                                             std::shared_ptr<Logger> logger,
                                             std::shared_ptr<PerformanceMonitor> perf_monitor,
                                             int num_threads,
                                             size_t max_queue_size,
                                             const std::string& output_dir,
                                             bool enable_brightness_filter,
                                             int stationary_timeout_seconds)
    : detector_(detector), logger_(logger), perf_monitor_(perf_monitor),
      num_threads_(num_threads), max_queue_size_(max_queue_size), output_dir_(output_dir),
      enable_brightness_filter_(enable_brightness_filter),
      stationary_timeout_seconds_(stationary_timeout_seconds),
      total_images_saved_(0), shutdown_requested_(false), frames_in_progress_(0),
      brightness_filter_active_(false) {
    last_photo_time_ = std::chrono::steady_clock::now() - std::chrono::seconds(PHOTO_INTERVAL_SECONDS);
}

ParallelFrameProcessor::~ParallelFrameProcessor() {
    shutdown();
}

bool ParallelFrameProcessor::initialize() {
    if (num_threads_ <= 1) {
        logger_->info("Parallel processing disabled - using sequential processing");
    } else {
        logger_->info("Initializing parallel frame processor with " + std::to_string(num_threads_) + " threads");
        
        // Start worker threads
        worker_threads_.reserve(num_threads_);
        for (int i = 0; i < num_threads_; ++i) {
            worker_threads_.emplace_back(&ParallelFrameProcessor::workerThread, this);
        }
        
        logger_->info("Parallel frame processor initialized successfully");
    }
    
    // Create output directory if it doesn't exist
    struct stat st;
    if (stat(output_dir_.c_str(), &st) != 0) {
        if (mkdir(output_dir_.c_str(), 0755) != 0) {
            logger_->warning("Failed to create output directory: " + output_dir_);
        } else {
            logger_->info("Created output directory: " + output_dir_);
        }
    }
    
    return true;
}

std::future<ParallelFrameProcessor::FrameResult> ParallelFrameProcessor::submitFrame(const cv::Mat& frame) {
    if (num_threads_ <= 1) {
        // Single-threaded mode - process synchronously
        std::promise<FrameResult> promise;
        auto future = promise.get_future();
        try {
            auto result = processFrameSync(frame);
            promise.set_value(result);
        } catch (...) {
            promise.set_exception(std::current_exception());
        }
        return future;
    }
    
    // Multi-threaded mode - queue for processing
    std::unique_lock<std::mutex> lock(queue_mutex_);
    
    // Check if queue is full
    if (frame_queue_.size() >= max_queue_size_) {
        logger_->warning("Frame queue full, dropping frame");
        std::promise<FrameResult> promise;
        auto future = promise.get_future();
        
        FrameResult empty_result;
        empty_result.processed = false;
        empty_result.capture_time = std::chrono::high_resolution_clock::now();
        promise.set_value(empty_result);
        return future;
    }
    
    // Create promise/future pair for the result
    std::promise<FrameResult> promise;
    auto future = promise.get_future();
    
    // Add frame to queue
    frame_queue_.emplace(frame.clone(), std::move(promise));
    frames_in_progress_++;
    
    lock.unlock();
    queue_condition_.notify_one();
    
    return future;
}

ParallelFrameProcessor::FrameResult ParallelFrameProcessor::processFrameSync(const cv::Mat& frame) {
    return processFrameInternal(frame);
}

void ParallelFrameProcessor::shutdown() {
    if (num_threads_ <= 1 || shutdown_requested_.load()) {
        return;
    }
    
    logger_->info("Shutting down parallel frame processor...");
    
    shutdown_requested_ = true;
    queue_condition_.notify_all();
    
    // Wait for all worker threads to finish
    for (auto& thread : worker_threads_) {
        if (thread.joinable()) {
            thread.join();
        }
    }
    
    // Clear any remaining frames in queue
    std::unique_lock<std::mutex> lock(queue_mutex_);
    while (!frame_queue_.empty()) {
        auto& [frame, promise] = frame_queue_.front();
        FrameResult empty_result;
        empty_result.processed = false;
        empty_result.capture_time = std::chrono::high_resolution_clock::now();
        promise.set_value(empty_result);
        frame_queue_.pop();
    }
    
    logger_->info("Parallel frame processor shutdown complete");
}

size_t ParallelFrameProcessor::getQueueSize() const {
    std::unique_lock<std::mutex> lock(const_cast<std::mutex&>(queue_mutex_));
    return frame_queue_.size();
}

void ParallelFrameProcessor::workerThread() {
    logger_->debug("Worker thread started");
    
    while (!shutdown_requested_.load()) {
        std::unique_lock<std::mutex> lock(queue_mutex_);
        
        // Wait for work or shutdown
        queue_condition_.wait(lock, [this] {
            return !frame_queue_.empty() || shutdown_requested_.load();
        });
        
        if (shutdown_requested_.load() && frame_queue_.empty()) {
            break;
        }
        
        if (frame_queue_.empty()) {
            continue;
        }
        
        // Get next frame to process
        auto [frame, promise] = std::move(frame_queue_.front());
        frame_queue_.pop();
        lock.unlock();
        
        try {
            // Process the frame
            auto result = processFrameInternal(frame);
            promise.set_value(result);
        } catch (...) {
            promise.set_exception(std::current_exception());
        }
        
        frames_in_progress_--;
    }
    
    logger_->debug("Worker thread exiting");
}

void ParallelFrameProcessor::saveDetectionPhoto(const cv::Mat& frame, const std::vector<Detection>& detections, const std::shared_ptr<ObjectDetector>& detector) {
    std::lock_guard<std::mutex> lock(photo_mutex_);
    
    // Count current object types
    std::map<std::string, int> current_object_counts;
    for (const auto& detection : detections) {
        current_object_counts[detection.class_name]++;
    }
    
    // Check if there are new object types or new instances
    bool has_new_objects = false;
    bool has_new_types = false;
    
    // Check for new types
    for (const auto& [type, count] : current_object_counts) {
        if (last_saved_object_counts_.find(type) == last_saved_object_counts_.end()) {
            has_new_types = true;
            logger_->info("New object type detected: " + type);
            break;
        }
    }
    
    // Check for new instances of existing types (more objects of same type)
    if (!has_new_types) {
        for (const auto& [type, count] : current_object_counts) {
            auto it = last_saved_object_counts_.find(type);
            if (it != last_saved_object_counts_.end() && count > it->second) {
                has_new_objects = true;
                logger_->info("New instance of " + type + " detected (count: " + 
                             std::to_string(it->second) + " -> " + std::to_string(count) + ")");
                break;
            }
        }
    }
    
    // Also check if any tracked object is marked as new
    if (!has_new_types && !has_new_objects && detector) {
        const auto& tracked = detector->getTrackedObjects();
        for (const auto& obj : tracked) {
            if (obj.is_new && obj.frames_since_detection == 0) {
                has_new_objects = true;
                logger_->info("Newly entered " + obj.object_type + " detected by tracker");
                break;
            }
        }
    }
    
    auto now = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(now - last_photo_time_);
    
    // Check if all detected objects are stationary and past the timeout
    bool all_stationary_past_timeout = false;
    if (detector && !detections.empty()) {
        const auto& tracked = detector->getTrackedObjects();
        if (!tracked.empty()) {
            all_stationary_past_timeout = true;
            for (const auto& obj : tracked) {
                if (obj.was_present_last_frame) {
                    if (!detector->isStationaryPastTimeout(obj, stationary_timeout_seconds_)) {
                        all_stationary_past_timeout = false;
                        break;
                    }
                }
            }
        }
    }
    
    // Skip photo if all objects are stationary past timeout
    if (all_stationary_past_timeout) {
        logger_->debug("Skipping photo - all objects stationary for more than " + 
                      std::to_string(stationary_timeout_seconds_) + " seconds");
        return;
    }
    
    // Save photo if:
    // 1. New types or new instances detected (immediate save, bypass 10s limit)
    // 2. OR enough time has passed (10 second interval) for stationary objects
    bool should_save_immediately = has_new_types || has_new_objects;
    bool enough_time_passed = elapsed.count() >= PHOTO_INTERVAL_SECONDS;
    
    if (!should_save_immediately && !enough_time_passed) {
        return;
    }
    
    if (should_save_immediately) {
        logger_->info("Saving photo immediately due to new objects/types detected");
    }
    
    // Update last photo time and object counts
    last_photo_time_ = now;
    last_saved_object_counts_ = current_object_counts;
    
    // Create a copy of the frame to draw on
    cv::Mat annotated_frame = frame.clone();
    
    // Draw bounding boxes for each detection
    for (const auto& detection : detections) {
        cv::Scalar color = getColorForClass(detection.class_name);
        
        // Draw rectangle around the object
        cv::rectangle(annotated_frame, detection.bbox, color, 2);
        
        // Draw label with class name and confidence
        std::string label = detection.class_name + " (" + 
                           std::to_string(static_cast<int>(detection.confidence * 100)) + "%)";
        
        // Add stationary indicator if object is stationary
        if (detection.is_stationary) {
            label += ", stationary";
            
            // Add duration if available
            if (detection.stationary_duration_seconds > 0) {
                int duration = detection.stationary_duration_seconds;
                if (duration < 60) {
                    label += " for " + std::to_string(duration) + " sec";
                } else {
                    int minutes = duration / 60;
                    label += " for " + std::to_string(minutes) + " min";
                }
            }
        }
        
        // Draw label with auto-positioning to avoid cutoff at screen edges
        DrawingUtils::drawBoundingBoxLabel(annotated_frame, label, detection.bbox, color);
    }
    
    // Generate filename with timestamp and detected objects
    std::string filename = generateFilename(detections);
    std::string filepath = output_dir_ + "/" + filename;
    
    // Save the image
    if (cv::imwrite(filepath, annotated_frame)) {
        logger_->info("Saved detection photo: " + filepath);
        total_images_saved_++;
    } else {
        logger_->error("Failed to save detection photo: " + filepath);
    }
}

cv::Scalar ParallelFrameProcessor::getColorForClass(const std::string& class_name) const {
    // Define colors for different object types (BGR format)
    // Green for persons, red for cats, blue for dogs, yellow for vehicles, etc.
    if (class_name == "person") {
        return cv::Scalar(0, 255, 0);  // Green
    } else if (class_name == "cat") {
        return cv::Scalar(0, 0, 255);  // Red
    } else if (class_name == "dog") {
        return cv::Scalar(255, 0, 0);  // Blue
    } else if (class_name == "bird") {
        return cv::Scalar(255, 255, 0);  // Cyan
    } else if (class_name == "bear") {
        return cv::Scalar(0, 128, 128);  // Dark cyan/teal
    } else if (class_name == "car" || class_name == "truck" || class_name == "bus") {
        return cv::Scalar(0, 255, 255);  // Yellow
    } else if (class_name == "motorcycle" || class_name == "bicycle") {
        return cv::Scalar(255, 0, 255);  // Magenta
    } else if (class_name == "chair") {
        return cv::Scalar(128, 0, 128);  // Purple
    } else if (class_name == "book") {
        return cv::Scalar(255, 128, 0);  // Orange
    } else {
        return cv::Scalar(255, 255, 255);  // White for unknown
    }
}

std::string ParallelFrameProcessor::generateFilename(const std::vector<Detection>& detections) const {
    // Get current time
    auto now = std::chrono::system_clock::now();
    auto time_t_now = std::chrono::system_clock::to_time_t(now);
    std::tm tm_now;
    localtime_r(&time_t_now, &tm_now);
    
    // Format timestamp: "2025-10-04 010000"
    std::ostringstream timestamp;
    timestamp << std::put_time(&tm_now, "%Y-%m-%d %H%M%S");
    
    // Collect unique object types
    std::set<std::string> object_types;
    for (const auto& detection : detections) {
        object_types.insert(detection.class_name);
    }
    
    // Build object string (e.g., "person cat detected")
    std::ostringstream object_str;
    bool first = true;
    for (const auto& obj_type : object_types) {
        if (!first) object_str << " ";
        object_str << obj_type;
        first = false;
    }
    object_str << " detected";
    
    // Combine to create filename
    return timestamp.str() + " " + object_str.str() + ".jpg";
}

ParallelFrameProcessor::FrameResult ParallelFrameProcessor::processFrameInternal(const cv::Mat& frame) {
    auto start_time = std::chrono::high_resolution_clock::now();
    
    FrameResult result;
    result.capture_time = start_time;
    result.processed = true;
    
    try {
        // Apply brightness filter if enabled and high brightness is detected
        cv::Mat processed_frame = frame.clone();
        if (enable_brightness_filter_ && detectHighBrightness(frame)) {
            processed_frame = applyBrightnessFilter(frame);
            brightness_filter_active_ = true;
        } else {
            brightness_filter_active_ = false;
        }
        
        // Perform object detection on the (possibly filtered) frame
        result.detections = detector_->detectObjects(processed_frame);
        
        // Filter for target classes and log detections
        std::vector<Detection> target_detections;
        for (const auto& detection : result.detections) {
            if (detector_->isTargetClass(detection.class_name)) {
                target_detections.push_back(detection);
                
                // Log detection with center coordinates
                cv::Point2f center(
                    detection.bbox.x + detection.bbox.width / 2.0f,
                    detection.bbox.y + detection.bbox.height / 2.0f
                );
                logger_->info("detected " + detection.class_name + " at coordinates: (" + 
                             std::to_string(static_cast<int>(center.x)) + ", " + 
                             std::to_string(static_cast<int>(center.y)) + ") with confidence " + 
                             std::to_string(static_cast<int>(detection.confidence * 100)) + "%");
            }
        }
        
        // Update object tracking before saving photo
        if (!target_detections.empty()) {
            detector_->updateTracking(target_detections);
            // Enrich detections with stationary status from tracked objects
            detector_->enrichDetectionsWithStationaryStatus(target_detections);
        }
        
        // Save photo with bounding boxes if we have target detections
        if (!target_detections.empty()) {
            saveDetectionPhoto(frame, target_detections, detector_);
        }
        
    } catch (const std::exception& e) {
        logger_->error("Error processing frame: " + std::string(e.what()));
        result.processed = false;
    }
    
    return result;
}

int ParallelFrameProcessor::getTotalImagesSaved() const {
    return total_images_saved_;
}

bool ParallelFrameProcessor::detectHighBrightness(const cv::Mat& frame) {
    // Convert to grayscale for brightness analysis
    cv::Mat gray;
    if (frame.channels() == 3) {
        cv::cvtColor(frame, gray, cv::COLOR_BGR2GRAY);
    } else {
        gray = frame;
    }
    
    // Calculate mean brightness
    cv::Scalar mean_brightness = cv::mean(gray);
    double avg_brightness = mean_brightness[0];
    
    // High brightness threshold (on 0-255 scale, values above 180 indicate very bright conditions)
    const double HIGH_BRIGHTNESS_THRESHOLD = 180.0;
    
    if (avg_brightness > HIGH_BRIGHTNESS_THRESHOLD) {
        logger_->debug("High brightness detected: " + std::to_string(static_cast<int>(avg_brightness)) + "/255");
        return true;
    }
    
    return false;
}

cv::Mat ParallelFrameProcessor::applyBrightnessFilter(const cv::Mat& frame) {
    cv::Mat filtered = frame.clone();
    
    // Apply CLAHE (Contrast Limited Adaptive Histogram Equalization) to reduce glare
    // This helps reduce the impact of reflections while preserving details
    cv::Mat lab;
    cv::cvtColor(filtered, lab, cv::COLOR_BGR2Lab);
    
    // Split into L, a, b channels
    std::vector<cv::Mat> lab_channels;
    cv::split(lab, lab_channels);
    
    // Apply CLAHE to the L channel (lightness)
    cv::Ptr<cv::CLAHE> clahe = cv::createCLAHE(2.0, cv::Size(8, 8));
    clahe->apply(lab_channels[0], lab_channels[0]);
    
    // Merge channels back
    cv::merge(lab_channels, lab);
    cv::cvtColor(lab, filtered, cv::COLOR_Lab2BGR);
    
    // Apply gamma correction to reduce overexposure
    // Gamma < 1 darkens the image, reducing bright spots
    const double gamma = 0.7;
    cv::Mat lut(1, 256, CV_8U);
    uchar* p = lut.ptr();
    for (int i = 0; i < 256; ++i) {
        p[i] = cv::saturate_cast<uchar>(pow(i / 255.0, gamma) * 255.0);
    }
    cv::LUT(filtered, lut, filtered);
    
    logger_->debug("Applied brightness filter to reduce reflections");
    return filtered;
}