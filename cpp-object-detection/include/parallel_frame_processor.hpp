#pragma once

#include <opencv2/opencv.hpp>
#include <memory>
#include <queue>
#include <thread>
#include <mutex>
#include <condition_variable>
#include <atomic>
#include <future>
#include <chrono>
#include <map>
#include "object_detector.hpp"
#include "logger.hpp"
#include "performance_monitor.hpp"
#include "detection_model_interface.hpp"

/**
 * Parallel frame processor that can handle multiple frames concurrently
 * while maintaining sequential processing order
 */
class ParallelFrameProcessor {
public:
    struct FrameResult {
        std::chrono::high_resolution_clock::time_point capture_time;
        bool processed;
        std::vector<Detection> detections;
    };

    ParallelFrameProcessor(std::shared_ptr<ObjectDetector> detector,
                          std::shared_ptr<Logger> logger,
                          std::shared_ptr<PerformanceMonitor> perf_monitor,
                          int num_threads = 1,
                          size_t max_queue_size = 10,
                          const std::string& output_dir = "detections",
                          bool enable_brightness_filter = false);
    
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
    
    /**
     * Get total number of images saved since start
     */
    int getTotalImagesSaved() const;
    
    /**
     * Check if brightness filter is currently active
     */
    bool isBrightnessFilterActive() const { return brightness_filter_active_; }

private:
    std::shared_ptr<ObjectDetector> detector_;
    std::shared_ptr<Logger> logger_;
    std::shared_ptr<PerformanceMonitor> perf_monitor_;
    
    int num_threads_;
    size_t max_queue_size_;
    std::string output_dir_;
    bool enable_brightness_filter_;
    std::atomic<bool> brightness_filter_active_;
    
    // Photo storage rate limiting
    std::chrono::steady_clock::time_point last_photo_time_;
    std::mutex photo_mutex_;
    static constexpr int PHOTO_INTERVAL_SECONDS = 10;
    int total_images_saved_;
    
    // Track object state from last saved photo
    std::map<std::string, int> last_saved_object_counts_;
    
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
    
    // Helper methods for photo storage
    void saveDetectionPhoto(const cv::Mat& frame, const std::vector<Detection>& detections, const std::shared_ptr<ObjectDetector>& detector);
    cv::Scalar getColorForClass(const std::string& class_name) const;
    std::string generateFilename(const std::vector<Detection>& detections) const;
    
    // Brightness detection and filtering
    bool detectHighBrightness(const cv::Mat& frame);
    cv::Mat applyBrightnessFilter(const cv::Mat& frame);
};