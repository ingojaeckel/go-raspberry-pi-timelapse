#include "logger.hpp"
#include <iostream>
#include <iomanip>
#include <sstream>

Logger::Logger(const std::string& log_file, bool verbose) 
    : verbose_(verbose) {
    file_stream_ = std::make_unique<std::ofstream>(log_file, std::ios::app);
    if (!file_stream_->is_open()) {
        std::cerr << "Warning: Could not open log file " << log_file 
                  << ". Logging to console only." << std::endl;
    }
}

Logger::~Logger() {
    if (file_stream_ && file_stream_->is_open()) {
        file_stream_->close();
    }
}

void Logger::logObjectDetection(const std::string& object_type, 
                               const std::string& action, 
                               double confidence) {
    std::stringstream ss;
    ss << object_type << " " << action << " frame (" 
       << std::fixed << std::setprecision(0) << (confidence * 100) << "% confidence)";
    log(Level::INFO, ss.str());
}

void Logger::logHeartbeat() {
    log(Level::INFO, "Detection system operational - heartbeat");
}

void Logger::logPerformance(double fps, int processed_frames, int total_frames) {
    std::stringstream ss;
    ss << "Performance: " << std::fixed << std::setprecision(2) << fps << " fps, "
       << "processed " << processed_frames << "/" << total_frames << " frames";
    log(Level::INFO, ss.str());
}

void Logger::logPerformanceWarning(double fps, double threshold) {
    std::stringstream ss;
    ss << "Performance warning: processing rate " << std::fixed << std::setprecision(2) 
       << fps << " fps is below threshold of " << threshold << " fps";
    log(Level::WARNING, ss.str());
}

void Logger::log(Level level, const std::string& message) {
    writeLog(level, message);
}

void Logger::debug(const std::string& message) {
    if (verbose_) {
        log(Level::DEBUG, message);
    }
}

void Logger::info(const std::string& message) {
    log(Level::INFO, message);
}

void Logger::warning(const std::string& message) {
    log(Level::WARNING, message);
}

void Logger::error(const std::string& message) {
    log(Level::ERROR, message);
}

std::string Logger::getCurrentTimestamp() const {
    auto now = std::chrono::system_clock::now();
    auto time_t = std::chrono::system_clock::to_time_t(now);
    auto ms = std::chrono::duration_cast<std::chrono::milliseconds>(
        now.time_since_epoch()) % 1000;

    std::stringstream ss;
    ss << std::put_time(std::localtime(&time_t), "%a %d %b at %I:%M:%S%p")
       << "." << std::setfill('0') << std::setw(3) << ms.count();
    
    return ss.str();
}

std::string Logger::levelToString(Level level) const {
    switch (level) {
        case Level::DEBUG:   return "DEBUG";
        case Level::INFO:    return "INFO";
        case Level::WARNING: return "WARNING";
        case Level::ERROR:   return "ERROR";
        default:             return "UNKNOWN";
    }
}

void Logger::writeLog(Level level, const std::string& message) {
    std::lock_guard<std::mutex> lock(log_mutex_);
    
    std::string timestamp = getCurrentTimestamp();
    std::string level_str = levelToString(level);
    std::string log_line = "On " + timestamp + " PT, " + message;
    
    // Write to file if available
    if (file_stream_ && file_stream_->is_open()) {
        *file_stream_ << "[" << level_str << "] " << log_line << std::endl;
        file_stream_->flush();
    }
    
    // Also write to console if verbose or if it's a warning/error
    if (verbose_ || level >= Level::WARNING) {
        if (level >= Level::WARNING) {
            std::cerr << "[" << level_str << "] " << log_line << std::endl;
        } else {
            std::cout << "[" << level_str << "] " << log_line << std::endl;
        }
    }
}