#include "scene_graph/logger.h"
#include <iostream>
#include <iomanip>
#include <ctime>
#include <sstream>

namespace scene_graph {

Logger& Logger::getInstance() {
    static Logger instance;
    return instance;
}

Logger::Logger() : min_level_(LogLevel::INFO), verbose_(false) {
}

Logger::~Logger() {
}

void Logger::setLogLevel(LogLevel level) {
    min_level_ = level;
}

void Logger::setVerbose(bool verbose) {
    verbose_ = verbose;
    if (verbose) {
        min_level_ = LogLevel::DEBUG;
    }
}

void Logger::log(LogLevel level, const std::string& message) {
    if (level < min_level_) {
        return;
    }
    
    std::string timestamp = getTimestamp();
    std::string level_str = getLevelString(level);
    
    std::cout << "[" << timestamp << "] [" << level_str << "] " << message << std::endl;
}

void Logger::debug(const std::string& message) {
    log(LogLevel::DEBUG, message);
}

void Logger::info(const std::string& message) {
    log(LogLevel::INFO, message);
}

void Logger::warning(const std::string& message) {
    log(LogLevel::WARNING, message);
}

void Logger::error(const std::string& message) {
    log(LogLevel::ERROR, message);
}

std::string Logger::getLevelString(LogLevel level) const {
    switch (level) {
        case LogLevel::DEBUG: return "DEBUG";
        case LogLevel::INFO: return "INFO";
        case LogLevel::WARNING: return "WARNING";
        case LogLevel::ERROR: return "ERROR";
        default: return "UNKNOWN";
    }
}

std::string Logger::getTimestamp() const {
    auto now = std::time(nullptr);
    auto tm = *std::localtime(&now);
    
    std::ostringstream oss;
    oss << std::put_time(&tm, "%Y-%m-%d %H:%M:%S");
    return oss.str();
}

} // namespace scene_graph
