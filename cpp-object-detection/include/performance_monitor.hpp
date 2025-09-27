#pragma once

#include <chrono>
#include <memory>
#include "logger.hpp"

/**
 * Performance monitoring for frame processing rates and timing
 */
class PerformanceMonitor {
public:
    PerformanceMonitor(std::shared_ptr<Logger> logger, 
                      double min_fps_threshold = 1.0);
    ~PerformanceMonitor();

    /**
     * Mark the start of frame processing
     */
    void startFrameProcessing();
    
    /**
     * Mark the end of frame processing and update statistics
     */
    void endFrameProcessing();
    
    /**
     * Get current frames per second
     */
    double getCurrentFPS() const;
    
    /**
     * Get average processing time per frame (in milliseconds)
     */
    double getAverageProcessingTime() const;
    
    /**
     * Check if performance is below threshold and log warning if needed
     */
    void checkPerformanceThreshold();
    
    /**
     * Reset statistics
     */
    void reset();
    
    /**
     * Get statistics summary
     */
    std::string getStatsSummary() const;

    /**
     * Log periodic performance report
     */
    void logPerformanceReport();

private:
    std::shared_ptr<Logger> logger_;
    double min_fps_threshold_;
    
    // Timing
    std::chrono::high_resolution_clock::time_point frame_start_time_;
    std::chrono::high_resolution_clock::time_point last_frame_time_;
    
    // Statistics
    int total_frames_processed_;
    int total_frames_captured_;
    double total_processing_time_ms_;
    double current_fps_;
    
    // Performance tracking
    std::chrono::high_resolution_clock::time_point last_warning_time_;
    std::chrono::high_resolution_clock::time_point last_report_time_;
    
    static constexpr int PERFORMANCE_WARNING_INTERVAL_SECONDS = 60;
    static constexpr int PERFORMANCE_REPORT_INTERVAL_SECONDS = 300; // 5 minutes
    
    void updateFPS();
    bool shouldLogWarning() const;
    bool shouldLogReport() const;
};