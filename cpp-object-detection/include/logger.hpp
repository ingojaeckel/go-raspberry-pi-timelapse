#pragma once

#include <string>
#include <fstream>
#include <memory>
#include <chrono>
#include <mutex>
#include <vector>
#include <map>

/**
 * Logging system with structured output and timestamps
 */
class Logger {
public:
    enum class Level {
        DEBUG = 0,
        INFO = 1,
        WARNING = 2,
        ERROR = 3
    };

    struct DetectionEvent {
        std::string object_type;
        std::chrono::system_clock::time_point timestamp;
        bool is_stationary;  // True if object stayed in roughly same position
    };

    Logger(const std::string& log_file, bool verbose = false);
    ~Logger();

    /**
     * Log object entry with position
     */
    void logObjectEntry(const std::string& object_type,
                       float x, float y,
                       double confidence);
    
    /**
     * Log object movement with old and new positions
     */
    void logObjectMovement(const std::string& object_type,
                          float old_x, float old_y,
                          float new_x, float new_y,
                          double confidence);
    
    /**
     * Log heartbeat message indicating system is still running
     */
    void logHeartbeat();
    
    /**
     * Log performance metrics
     */
    void logPerformance(double fps, int processed_frames, int total_frames);
    
    /**
     * Log performance warning
     */
    void logPerformanceWarning(double fps, double threshold);
    
    /**
     * General logging methods
     */
    void log(Level level, const std::string& message);
    void debug(const std::string& message);
    void info(const std::string& message);
    void warning(const std::string& message);
    void error(const std::string& message);
    
    /**
     * Record a detection event for hourly summary
     */
    void recordDetection(const std::string& object_type, bool is_stationary = false);
    
    /**
     * Print hourly summary and reset tracking
     */
    void printHourlySummary();
    
    /**
     * Check if summary interval has elapsed and print if needed
     */
    void checkAndPrintSummary(int interval_minutes = 60);
    
    /**
     * Print final summary covering entire program runtime
     */
    void printFinalSummary();

private:
    std::unique_ptr<std::ofstream> file_stream_;
    bool verbose_;
    std::mutex log_mutex_;
    
    // Hourly summary tracking
    std::vector<DetectionEvent> detection_events_;
    std::chrono::system_clock::time_point summary_period_start_;
    std::mutex summary_mutex_;
    
    // Final summary tracking (all events throughout program lifetime)
    std::vector<DetectionEvent> all_detection_events_;
    std::chrono::system_clock::time_point program_start_time_;
    
    std::string getCurrentTimestamp() const;
    std::string levelToString(Level level) const;
    void writeLog(Level level, const std::string& message);
    std::string formatTime(const std::chrono::system_clock::time_point& time) const;
};