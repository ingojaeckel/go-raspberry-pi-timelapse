#pragma once

#include <opencv2/opencv.hpp>
#include <memory>
#include <queue>
#include <thread>
#include <mutex>
#include <condition_variable>
#include <atomic>
#include <future>
#include "object_detector.hpp"
#include "logger.hpp"
#include "performance_monitor.hpp"

/**
 * Parallel frame processor that can handle multiple frames concurrently
 * while maintaining sequential processing order
 */
class ParallelFrameProcessor {
public:
    struct FrameResult {
        std::chrono::high_resolution_clock::time_point capture_time;
        bool processed;
        std::vector<ObjectDetector::Detection> detections;
    };

    ParallelFrameProcessor(std::shared_ptr<ObjectDetector> detector,
                          std::shared_ptr<Logger> logger,
                          std::shared_ptr<PerformanceMonitor> perf_monitor,
                          int num_threads = 1,
                          size_t max_queue_size = 10);
    
    ~ParallelFrameProcessor();

    /**
     * Initialize the processor and start worker threads
     */
    bool initialize();
    
    /**
     * Submit a frame for processing
     * Returns future that will contain the detection results
     */
    std::future<FrameResult> submitFrame(const cv::Mat& frame);
    
    /**
     * Process a frame synchronously (for single-threaded mode)
     */
    FrameResult processFrameSync(const cv::Mat& frame);
    
    /**
     * Shutdown the processor and stop all threads
     */
    void shutdown();
    
    /**
     * Check if parallel processing is enabled
     */
    bool isParallelEnabled() const { return num_threads_ > 1; }
    
    /**
     * Get current queue size
     */
    size_t getQueueSize() const;

private:
    std::shared_ptr<ObjectDetector> detector_;
    std::shared_ptr<Logger> logger_;
    std::shared_ptr<PerformanceMonitor> perf_monitor_;
    
    int num_threads_;
    size_t max_queue_size_;
    
    // Threading infrastructure
    std::vector<std::thread> worker_threads_;
    std::queue<std::pair<cv::Mat, std::promise<FrameResult>>> frame_queue_;
    std::mutex queue_mutex_;
    std::condition_variable queue_condition_;
    std::atomic<bool> shutdown_requested_;
    std::atomic<size_t> frames_in_progress_;
    
    // Worker thread function
    void workerThread();
    
    // Process a single frame (thread-safe)
    FrameResult processFrameInternal(const cv::Mat& frame);
};