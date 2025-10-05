#include "performance_monitor.hpp"
#include <sstream>
#include <iomanip>

PerformanceMonitor::PerformanceMonitor(std::shared_ptr<Logger> logger, 
                                      double min_fps_threshold)
    : logger_(logger), min_fps_threshold_(min_fps_threshold),
      total_frames_processed_(0), total_frames_captured_(0),
      total_processing_time_ms_(0.0), last_processing_time_ms_(0.0), current_fps_(0.0) {
    
    last_frame_time_ = std::chrono::high_resolution_clock::now();
    last_warning_time_ = std::chrono::high_resolution_clock::now();
    last_report_time_ = std::chrono::high_resolution_clock::now();
}

PerformanceMonitor::~PerformanceMonitor() = default;

void PerformanceMonitor::startFrameProcessing() {
    frame_start_time_ = std::chrono::high_resolution_clock::now();
    total_frames_captured_++;
}

void PerformanceMonitor::endFrameProcessing() {
    auto end_time = std::chrono::high_resolution_clock::now();
    auto processing_time = std::chrono::duration_cast<std::chrono::microseconds>(
        end_time - frame_start_time_);
    
    double processing_time_ms = processing_time.count() / 1000.0;
    total_processing_time_ms_ += processing_time_ms;
    last_processing_time_ms_ = processing_time_ms;
    total_frames_processed_++;
    
    updateFPS();
    checkForCounterOverflow();
    
    if (total_frames_processed_ % 100 == 0) {
        // only print every 100 frames to reduce log spam
        logger_->debug("Frame processed in " + std::to_string(processing_time_ms) + " ms");
    }
}

double PerformanceMonitor::getCurrentFPS() const {
    return current_fps_;
}

double PerformanceMonitor::getAverageProcessingTime() const {
    if (total_frames_processed_ == 0) {
        return 0.0;
    }
    return total_processing_time_ms_ / total_frames_processed_;
}

double PerformanceMonitor::getLastProcessingTime() const {
    return last_processing_time_ms_;
}

void PerformanceMonitor::checkPerformanceThreshold() {
    if (current_fps_ < min_fps_threshold_ && shouldLogWarning()) {
        logger_->logPerformanceWarning(current_fps_, min_fps_threshold_);
        last_warning_time_ = std::chrono::high_resolution_clock::now();
    }
}

void PerformanceMonitor::reset() {
    total_frames_processed_ = 0;
    total_frames_captured_ = 0;
    total_processing_time_ms_ = 0.0;
    last_processing_time_ms_ = 0.0;
    current_fps_ = 0.0;
    last_frame_time_ = std::chrono::high_resolution_clock::now();
    last_warning_time_ = std::chrono::high_resolution_clock::now();
    last_report_time_ = std::chrono::high_resolution_clock::now();
}

std::string PerformanceMonitor::getStatsSummary() const {
    std::stringstream ss;
    ss << std::fixed << std::setprecision(2);
    ss << "FPS: " << current_fps_;
    ss << ", Avg processing time: " << getAverageProcessingTime() << " ms";
    ss << ", Frames processed/captured: " << total_frames_processed_ 
       << "/" << total_frames_captured_;
    
    if (total_frames_captured_ > 0) {
        double processing_ratio = static_cast<double>(total_frames_processed_) / 
                                 total_frames_captured_ * 100.0;
        ss << " (" << std::setprecision(1) << processing_ratio << "%)";
    }
    
    return ss.str();
}

void PerformanceMonitor::logPerformanceReport() {
    if (shouldLogReport()) {
        logger_->info("Performance report: " + getStatsSummary());
        last_report_time_ = std::chrono::high_resolution_clock::now();
    }
}

void PerformanceMonitor::updateFPS() {
    auto current_time = std::chrono::high_resolution_clock::now();
    auto time_diff = std::chrono::duration_cast<std::chrono::milliseconds>(
        current_time - last_frame_time_);
    
    if (time_diff.count() > 0) {
        current_fps_ = 1000.0 / time_diff.count();
    }
    
    last_frame_time_ = current_time;
}

bool PerformanceMonitor::shouldLogWarning() const {
    auto now = std::chrono::high_resolution_clock::now();
    auto time_since_last_warning = std::chrono::duration_cast<std::chrono::seconds>(
        now - last_warning_time_);
    
    return time_since_last_warning.count() >= PERFORMANCE_WARNING_INTERVAL_SECONDS;
}

bool PerformanceMonitor::shouldLogReport() const {
    auto now = std::chrono::high_resolution_clock::now();
    auto time_since_last_report = std::chrono::duration_cast<std::chrono::seconds>(
        now - last_report_time_);
    
    return time_since_last_report.count() >= PERFORMANCE_REPORT_INTERVAL_SECONDS;
}

void PerformanceMonitor::checkForCounterOverflow() {
    // Reset counters when approaching overflow to prevent issues during long-term operation
    // With 1 fps, 1M frames = ~11.5 days, so this is a reasonable safety check
    if (total_frames_processed_ >= MAX_FRAME_COUNT || total_frames_captured_ >= MAX_FRAME_COUNT) {
        logger_->info("Performance counters reset after processing " + 
                     std::to_string(total_frames_processed_) + " frames (overflow prevention)");
        
        // Keep the average processing time but reset counters
        double avg_time = getAverageProcessingTime();
        total_frames_processed_ = 100;  // Start with 100 to maintain average
        total_frames_captured_ = 100;
        total_processing_time_ms_ = avg_time * 100;
    }
}