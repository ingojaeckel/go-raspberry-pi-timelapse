#include "parallel_frame_processor.hpp"
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
                                             const std::string& output_dir)
    : detector_(detector), logger_(logger), perf_monitor_(perf_monitor),
      num_threads_(num_threads), max_queue_size_(max_queue_size), output_dir_(output_dir),
      shutdown_requested_(false), frames_in_progress_(0) {
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
    
    // Check if we're in night mode
    bool night_mode = isNightMode(frame);
    
    // Helper lambda to annotate a frame with bounding boxes
    auto annotateFrame = [&](const cv::Mat& input_frame) -> cv::Mat {
        cv::Mat annotated = input_frame.clone();
        
        // Draw bounding boxes for each detection
        for (const auto& detection : detections) {
            cv::Scalar color = getColorForClass(detection.class_name);
            
            // Draw rectangle around the object
            cv::rectangle(annotated, detection.bbox, color, 2);
            
            // Draw label with class name and confidence
            std::string label = detection.class_name + " " + 
                               std::to_string(static_cast<int>(detection.confidence * 100)) + "%";
            int baseline;
            cv::Size text_size = cv::getTextSize(label, cv::FONT_HERSHEY_SIMPLEX, 0.5, 1, &baseline);
            
            // Draw label background
            cv::Point text_origin(detection.bbox.x, detection.bbox.y - 5);
            cv::rectangle(annotated, 
                         cv::Point(text_origin.x, text_origin.y - text_size.height - 2),
                         cv::Point(text_origin.x + text_size.width, text_origin.y + 2),
                         color, cv::FILLED);
            
            // Draw label text
            cv::putText(annotated, label, text_origin, 
                       cv::FONT_HERSHEY_SIMPLEX, 0.5, cv::Scalar(0, 0, 0), 1);
        }
        
        return annotated;
    };
    
    // Generate base filename with timestamp and detected objects
    std::string base_filename = generateFilename(detections);
    
    // Save original frame with bounding boxes
    cv::Mat annotated_original = annotateFrame(frame);
    std::string original_filepath = output_dir_ + "/" + base_filename;
    
    if (cv::imwrite(original_filepath, annotated_original)) {
        logger_->info("Saved detection photo: " + original_filepath);
    } else {
        logger_->error("Failed to save detection photo: " + original_filepath);
    }
    
    // If in night mode, also save preprocessed version
    if (night_mode) {
        cv::Mat preprocessed_frame = preprocessForNight(frame);
        cv::Mat annotated_preprocessed = annotateFrame(preprocessed_frame);
        
        // Insert "night-enhanced" before the file extension
        std::string preprocessed_filename = base_filename;
        size_t ext_pos = preprocessed_filename.rfind(".jpg");
        if (ext_pos != std::string::npos) {
            preprocessed_filename.insert(ext_pos, " night-enhanced");
        }
        
        std::string preprocessed_filepath = output_dir_ + "/" + preprocessed_filename;
        
        if (cv::imwrite(preprocessed_filepath, annotated_preprocessed)) {
            logger_->info("Saved night-enhanced detection photo: " + preprocessed_filepath);
        } else {
            logger_->error("Failed to save night-enhanced detection photo: " + preprocessed_filepath);
        }
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

bool ParallelFrameProcessor::isNightTime() const {
    // Get current time
    auto now = std::chrono::system_clock::now();
    auto time_t_now = std::chrono::system_clock::to_time_t(now);
    std::tm tm_now;
    localtime_r(&time_t_now, &tm_now);
    
    int hour = tm_now.tm_hour;
    
    // Consider night time between 8 PM (20:00) and 6 AM (6:00)
    return hour >= 20 || hour < 6;
}

double ParallelFrameProcessor::calculateBrightness(const cv::Mat& frame) const {
    if (frame.empty()) {
        return 0.0;
    }
    
    // Convert to grayscale for brightness calculation
    cv::Mat gray;
    if (frame.channels() == 3) {
        cv::cvtColor(frame, gray, cv::COLOR_BGR2GRAY);
    } else {
        gray = frame;
    }
    
    // Calculate average brightness (0-255)
    cv::Scalar mean_scalar = cv::mean(gray);
    return mean_scalar[0];
}

bool ParallelFrameProcessor::isNightMode(const cv::Mat& frame) const {
    // Check both time of day and darkness level
    bool is_night_time = isNightTime();
    double brightness = calculateBrightness(frame);
    
    // Consider it night mode if:
    // 1. It's night time (20:00-6:00), OR
    // 2. Brightness is very low (< 50 out of 255, about 20%)
    bool is_dark = brightness < 50.0;
    
    if (is_night_time || is_dark) {
        logger_->debug("Night mode detected - time: " + std::string(is_night_time ? "yes" : "no") + 
                      ", brightness: " + std::to_string(static_cast<int>(brightness)));
        return true;
    }
    
    return false;
}

cv::Mat ParallelFrameProcessor::preprocessForNight(const cv::Mat& frame) const {
    if (frame.empty()) {
        return frame;
    }
    
    // Use CLAHE (Contrast Limited Adaptive Histogram Equalization) for better night-time enhancement
    // This is more effective than simple histogram equalization as it prevents over-amplification of noise
    
    cv::Mat lab_image;
    cv::cvtColor(frame, lab_image, cv::COLOR_BGR2Lab);
    
    // Split the LAB image into L, A, and B channels
    std::vector<cv::Mat> lab_planes(3);
    cv::split(lab_image, lab_planes);
    
    // Apply CLAHE to the L channel (lightness)
    cv::Ptr<cv::CLAHE> clahe = cv::createCLAHE();
    clahe->setClipLimit(2.0);  // Limit contrast enhancement to prevent over-amplification
    clahe->setTilesGridSize(cv::Size(8, 8));  // Grid size for local histogram equalization
    clahe->apply(lab_planes[0], lab_planes[0]);
    
    // Merge the channels back
    cv::Mat enhanced_lab;
    cv::merge(lab_planes, enhanced_lab);
    
    // Convert back to BGR
    cv::Mat enhanced_frame;
    cv::cvtColor(enhanced_lab, enhanced_frame, cv::COLOR_Lab2BGR);
    
    return enhanced_frame;
}

ParallelFrameProcessor::FrameResult ParallelFrameProcessor::processFrameInternal(const cv::Mat& frame) {
    auto start_time = std::chrono::high_resolution_clock::now();
    
    FrameResult result;
    result.capture_time = start_time;
    result.processed = true;
    
    try {
        // Check if we're in night mode
        bool night_mode = isNightMode(frame);
        
        // Preprocess frame for detection if in night mode
        cv::Mat detection_frame = frame;
        if (night_mode) {
            detection_frame = preprocessForNight(frame);
            logger_->debug("Applied night mode preprocessing for detection");
        }
        
        // Perform object detection on the preprocessed frame (if night) or original (if day)
        result.detections = detector_->detectObjects(detection_frame);
        
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