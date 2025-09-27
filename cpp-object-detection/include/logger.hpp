#pragma once

#include <string>
#include <fstream>
#include <memory>
#include <chrono>
#include <mutex>

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

    Logger(const std::string& log_file, bool verbose = false);
    ~Logger();

    /**
     * Log an object detection event
     */
    void logObjectDetection(const std::string& object_type, 
                          const std::string& action, // "entered" or "exited"
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

private:
    std::unique_ptr<std::ofstream> file_stream_;
    bool verbose_;
    std::mutex log_mutex_;
    
    std::string getCurrentTimestamp() const;
    std::string levelToString(Level level) const;
    void writeLog(Level level, const std::string& message);
};