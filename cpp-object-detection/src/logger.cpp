#include "logger.hpp"
#include <iostream>
#include <iomanip>
#include <sstream>
#include <map>
#include <algorithm>

Logger::Logger(const std::string& log_file, bool verbose) 
    : verbose_(verbose) {
    file_stream_ = std::make_unique<std::ofstream>(log_file, std::ios::app);
    if (!file_stream_->is_open()) {
        std::cerr << "Warning: Could not open log file " << log_file 
                  << ". Logging to console only." << std::endl;
    }
    summary_period_start_ = std::chrono::system_clock::now();
    program_start_time_ = summary_period_start_;
}

Logger::~Logger() {
    if (file_stream_ && file_stream_->is_open()) {
        file_stream_->close();
    }
}

void Logger::logObjectEntry(const std::string& object_type,
                           float x, float y,
                           double confidence) {
    std::stringstream ss;
    ss << "new " << object_type << " entered frame at (" 
       << std::fixed << std::setprecision(0) << x << ", " << y << ") ("
       << std::setprecision(0) << (confidence * 100) << "% confidence)";
    log(Level::INFO, ss.str());
}

void Logger::logObjectMovement(const std::string& object_type,
                              float old_x, float old_y,
                              float new_x, float new_y,
                              double confidence) {
    std::stringstream ss;
    ss << object_type << " seen earlier moved from (" 
       << std::fixed << std::setprecision(0) << old_x << ", " << old_y << ") -> ("
       << new_x << ", " << new_y << ") ("
       << std::setprecision(0) << (confidence * 100) << "% confidence)";
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

void Logger::recordDetection(const std::string& object_type, bool is_stationary, bool is_exit) {
    std::lock_guard<std::mutex> lock(summary_mutex_);
    DetectionEvent event;
    event.object_type = object_type;
    event.timestamp = std::chrono::system_clock::now();
    event.is_stationary = is_stationary;
    event.is_exit = is_exit;
    detection_events_.push_back(event);
    all_detection_events_.push_back(event);  // Also track for final summary
}

std::string Logger::formatTime(const std::chrono::system_clock::time_point& time) const {
    auto time_t = std::chrono::system_clock::to_time_t(time);
    std::stringstream ss;
    ss << std::put_time(std::localtime(&time_t), "%H:%M");
    return ss.str();
}

void Logger::generateTimeline(std::stringstream& summary, const std::vector<DetectionEvent>& events) {
    // Generate timeline showing entry and exit events
    for (size_t i = 0; i < events.size(); ++i) {
        const auto& event = events[i];
        
        if (event.is_exit) {
            // Object left the scene
            summary << "at " << formatTime(event.timestamp) << ", ";
            summary << event.object_type << " left\n";
        } else if (event.is_stationary) {
            // Skip individual stationary events - we'll summarize them as presence periods
            // Look ahead to find when this stationary period ends
            size_t j = i;
            while (j + 1 < events.size() && 
                   events[j + 1].object_type == event.object_type &&
                   events[j + 1].is_stationary && 
                   !events[j + 1].is_exit) {
                j++;
            }
            
            // Only show stationary period if it's longer than a single event
            if (j > i) {
                summary << "from " << formatTime(event.timestamp) 
                       << "-" << formatTime(events[j].timestamp) 
                       << " " << event.object_type << " was present\n";
                i = j;  // Skip processed events
            }
        } else {
            // Dynamic object detection (entry)
            // Count consecutive detections of the same dynamic object at similar times
            int same_type_count = 1;
            size_t end_idx = i;
            while (end_idx + 1 < events.size() && 
                   events[end_idx + 1].object_type == event.object_type &&
                   !events[end_idx + 1].is_stationary &&
                   !events[end_idx + 1].is_exit &&
                   std::chrono::duration_cast<std::chrono::seconds>(
                       events[end_idx + 1].timestamp - event.timestamp).count() < 10) {
                same_type_count++;
                end_idx++;
            }
            
            summary << "at " << formatTime(event.timestamp) << ", ";
            if (same_type_count == 1) {
                summary << "a " << event.object_type << " was detected";
            } else if (same_type_count == 2) {
                summary << "two " << (event.object_type == "person" ? "people" : event.object_type + "s") << " were detected";
            } else {
                summary << same_type_count << " " << (event.object_type == "person" ? "people" : event.object_type + "s") << " were detected";
            }
            summary << "\n";
            i = end_idx;  // Skip processed events
        }
    }
}

void Logger::printHourlySummary() {
    std::lock_guard<std::mutex> lock(summary_mutex_);
    
    if (detection_events_.empty()) {
        return;
    }
    
    auto period_end = std::chrono::system_clock::now();
    
    // Count total objects by type
    std::map<std::string, int> object_counts;
    for (const auto& event : detection_events_) {
        object_counts[event.object_type]++;
    }
    
    // Build summary header
    std::stringstream summary;
    summary << "\n========================================\n";
    summary << "Detection Summary: " 
            << formatTime(summary_period_start_) << "-" 
            << formatTime(period_end) << "\n";
    summary << "========================================\n";
    
    // Print counts with proper pluralization
    bool first = true;
    for (const auto& [type, count] : object_counts) {
        if (!first) summary << ", ";
        summary << count << "x ";
        if (type == "person") {
            summary << (count > 1 ? "people" : "person");
        } else {
            summary << type << (count > 1 ? "s" : "");
        }
        first = false;
    }
    summary << " were detected.\n\nTimeline:\n";
    
    // Generate timeline using helper function
    generateTimeline(summary, detection_events_);
    
    summary << "========================================\n";
    
    // Print to stdout
    std::cout << summary.str() << std::flush;
    
    // Clear events for next period
    detection_events_.clear();
    summary_period_start_ = period_end;
}

void Logger::checkAndPrintSummary(int interval_minutes) {
    auto now = std::chrono::system_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::minutes>(now - summary_period_start_);
    
    if (elapsed.count() >= interval_minutes) {
        printHourlySummary();
    }
}

void Logger::printFinalSummary() {
    std::lock_guard<std::mutex> lock(summary_mutex_);
    
    if (all_detection_events_.empty()) {
        std::cout << "\n========================================\n";
        std::cout << "Final Detection Summary\n";
        std::cout << "========================================\n";
        std::cout << "No objects were detected during program runtime.\n";
        std::cout << "========================================\n" << std::flush;
        return;
    }
    
    auto period_end = std::chrono::system_clock::now();
    
    // Count total objects by type
    std::map<std::string, int> object_counts;
    for (const auto& event : all_detection_events_) {
        object_counts[event.object_type]++;
    }
    
    // Build summary header
    std::stringstream summary;
    summary << "\n========================================\n";
    summary << "Final Detection Summary: " 
            << formatTime(program_start_time_) << "-" 
            << formatTime(period_end) << "\n";
    summary << "Program Runtime: ";
    
    // Calculate and display runtime duration
    auto runtime_seconds = std::chrono::duration_cast<std::chrono::seconds>(period_end - program_start_time_).count();
    int hours = runtime_seconds / 3600;
    int minutes = (runtime_seconds % 3600) / 60;
    int seconds = runtime_seconds % 60;
    
    if (hours > 0) {
        summary << hours << "h " << minutes << "m " << seconds << "s";
    } else if (minutes > 0) {
        summary << minutes << "m " << seconds << "s";
    } else {
        summary << seconds << "s";
    }
    summary << "\n";
    summary << "========================================\n";
    
    // Print counts with proper pluralization
    bool first = true;
    for (const auto& [type, count] : object_counts) {
        if (!first) summary << ", ";
        summary << count << "x ";
        if (type == "person") {
            summary << (count > 1 ? "people" : "person");
        } else {
            summary << type << (count > 1 ? "s" : "");
        }
        first = false;
    }
    summary << " were detected.\n\nTimeline:\n";
    
    // Generate timeline using helper function
    generateTimeline(summary, all_detection_events_);
    
    summary << "========================================\n";
    
    // Print to stdout
    std::cout << summary.str() << std::flush;
}