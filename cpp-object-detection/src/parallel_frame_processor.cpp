#include "parallel_frame_processor.hpp"
#include <chrono>

ParallelFrameProcessor::ParallelFrameProcessor(std::shared_ptr<ObjectDetector> detector,
                                             std::shared_ptr<Logger> logger,
                                             std::shared_ptr<PerformanceMonitor> perf_monitor,
                                             int num_threads,
                                             size_t max_queue_size)
    : detector_(detector), logger_(logger), perf_monitor_(perf_monitor),
      num_threads_(num_threads), max_queue_size_(max_queue_size),
      shutdown_requested_(false), frames_in_progress_(0) {
}

ParallelFrameProcessor::~ParallelFrameProcessor() {
    shutdown();
}

bool ParallelFrameProcessor::initialize() {
    if (num_threads_ <= 1) {
        logger_->info("Parallel processing disabled - using sequential processing");
        return true;
    }
    
    logger_->info("Initializing parallel frame processor with " + std::to_string(num_threads_) + " threads");
    
    // Start worker threads
    worker_threads_.reserve(num_threads_);
    for (int i = 0; i < num_threads_; ++i) {
        worker_threads_.emplace_back(&ParallelFrameProcessor::workerThread, this);
    }
    
    logger_->info("Parallel frame processor initialized successfully");
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

ParallelFrameProcessor::FrameResult ParallelFrameProcessor::processFrameInternal(const cv::Mat& frame) {
    auto start_time = std::chrono::high_resolution_clock::now();
    
    FrameResult result;
    result.capture_time = start_time;
    result.processed = true;
    
    try {
        // Perform object detection
        result.detections = detector_->detectObjects(frame);
        
        // Log detection events (note: this might need synchronization for logging)
        for (const auto& detection : result.detections) {
            if (detector_->isTargetClass(detection.class_name)) {
                logger_->logObjectDetection(detection.class_name, "detected", detection.confidence);
            }
        }
        
    } catch (const std::exception& e) {
        logger_->error("Error processing frame: " + std::string(e.what()));
        result.processed = false;
    }
    
    return result;
}