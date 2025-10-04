#pragma once

#include <string>
#include <memory>
#include <chrono>
#include "logger.hpp"

/**
 * System resource monitoring for long-term operation
 * Tracks CPU temperature, disk space, and memory usage
 */
class SystemMonitor {
public:
    SystemMonitor(std::shared_ptr<Logger> logger, 
                  const std::string& output_dir);
    ~SystemMonitor();

    /**
     * Perform periodic system checks (should be called regularly from main loop)
     */
    void performPeriodicCheck();
    
    /**
     * Get available disk space in bytes
     */
    unsigned long long getAvailableDiskSpace() const;
    
    /**
     * Get disk usage percentage for output directory
     */
    double getDiskUsagePercent() const;
    
    /**
     * Get CPU temperature in Celsius (returns -1 if unavailable)
     */
    double getCPUTemperature() const;
    
    /**
     * Check if disk space is critically low
     */
    bool isDiskSpaceCritical() const;
    
    /**
     * Log system statistics summary
     */
    void logSystemStats();

private:
    std::shared_ptr<Logger> logger_;
    std::string output_dir_;
    
    // Tracking for periodic checks
    std::chrono::steady_clock::time_point last_check_time_;
    static constexpr int CHECK_INTERVAL_SECONDS = 300;  // 5 minutes
    
    // Thresholds
    static constexpr double DISK_SPACE_WARNING_PERCENT = 90.0;
    static constexpr double DISK_SPACE_CRITICAL_PERCENT = 95.0;
    static constexpr double CPU_TEMP_WARNING_CELSIUS = 75.0;
    static constexpr double CPU_TEMP_CRITICAL_CELSIUS = 85.0;
    static constexpr unsigned long long MIN_FREE_SPACE_BYTES = 100 * 1024 * 1024;  // 100 MB
    
    bool shouldPerformCheck() const;
    void checkDiskSpace();
    void checkCPUTemperature();
};
