#include "system_monitor.hpp"
#include <sys/statvfs.h>
#include <fstream>
#include <sstream>
#include <iomanip>

SystemMonitor::SystemMonitor(std::shared_ptr<Logger> logger, 
                            const std::string& output_dir)
    : logger_(logger), output_dir_(output_dir) {
    last_check_time_ = std::chrono::steady_clock::now();
}

SystemMonitor::~SystemMonitor() = default;

void SystemMonitor::performPeriodicCheck() {
    if (!shouldPerformCheck()) {
        return;
    }
    
    checkDiskSpace();
    checkCPUTemperature();
    logSystemStats();
    
    last_check_time_ = std::chrono::steady_clock::now();
}

unsigned long long SystemMonitor::getAvailableDiskSpace() const {
    struct statvfs stat;
    
    if (statvfs(output_dir_.c_str(), &stat) != 0) {
        // If output_dir doesn't exist yet, try current directory
        if (statvfs(".", &stat) != 0) {
            return 0;
        }
    }
    
    // Available space = available blocks * block size
    return static_cast<unsigned long long>(stat.f_bavail) * stat.f_frsize;
}

double SystemMonitor::getDiskUsagePercent() const {
    struct statvfs stat;
    
    if (statvfs(output_dir_.c_str(), &stat) != 0) {
        if (statvfs(".", &stat) != 0) {
            return 0.0;
        }
    }
    
    unsigned long long total_space = static_cast<unsigned long long>(stat.f_blocks) * stat.f_frsize;
    unsigned long long free_space = static_cast<unsigned long long>(stat.f_bfree) * stat.f_frsize;
    
    if (total_space == 0) {
        return 0.0;
    }
    
    return (static_cast<double>(total_space - free_space) / total_space) * 100.0;
}

double SystemMonitor::getCPUTemperature() const {
    // Try common Linux thermal zone path
    std::ifstream temp_file("/sys/class/thermal/thermal_zone0/temp");
    if (temp_file.is_open()) {
        int temp_millicelsius;
        temp_file >> temp_millicelsius;
        temp_file.close();
        return temp_millicelsius / 1000.0;
    }
    
    // Try Raspberry Pi specific path
    std::ifstream rpi_temp("/sys/devices/virtual/thermal/thermal_zone0/temp");
    if (rpi_temp.is_open()) {
        int temp_millicelsius;
        rpi_temp >> temp_millicelsius;
        rpi_temp.close();
        return temp_millicelsius / 1000.0;
    }
    
    return -1.0;  // Temperature unavailable
}

bool SystemMonitor::isDiskSpaceCritical() const {
    unsigned long long available = getAvailableDiskSpace();
    double usage_percent = getDiskUsagePercent();
    
    return (available < MIN_FREE_SPACE_BYTES) || (usage_percent > DISK_SPACE_CRITICAL_PERCENT);
}

void SystemMonitor::logSystemStats() {
    std::ostringstream stats;
    stats << "System statistics: ";
    
    // Disk space
    unsigned long long available_bytes = getAvailableDiskSpace();
    double usage_percent = getDiskUsagePercent();
    double available_mb = available_bytes / (1024.0 * 1024.0);
    
    stats << "Disk: " << std::fixed << std::setprecision(1) 
          << usage_percent << "% used, " 
          << available_mb << " MB free";
    
    // CPU temperature
    double cpu_temp = getCPUTemperature();
    if (cpu_temp > 0) {
        stats << " | CPU temp: " << std::setprecision(1) << cpu_temp << "°C";
    }
    
    logger_->info(stats.str());
}

void SystemMonitor::checkDiskSpace() {
    unsigned long long available = getAvailableDiskSpace();
    double usage_percent = getDiskUsagePercent();
    
    if (available < MIN_FREE_SPACE_BYTES || usage_percent > DISK_SPACE_CRITICAL_PERCENT) {
        double available_mb = available / (1024.0 * 1024.0);
        logger_->error("Critical disk space: " + std::to_string(static_cast<int>(usage_percent)) + 
                      "% used, only " + std::to_string(static_cast<int>(available_mb)) + " MB free");
    } else if (usage_percent > DISK_SPACE_WARNING_PERCENT) {
        logger_->warning("Low disk space: " + std::to_string(static_cast<int>(usage_percent)) + "% used");
    }
}

void SystemMonitor::checkCPUTemperature() {
    double cpu_temp = getCPUTemperature();
    
    if (cpu_temp < 0) {
        return;  // Temperature not available
    }
    
    if (cpu_temp > CPU_TEMP_CRITICAL_CELSIUS) {
        logger_->error("Critical CPU temperature: " + std::to_string(static_cast<int>(cpu_temp)) + "°C");
    } else if (cpu_temp > CPU_TEMP_WARNING_CELSIUS) {
        logger_->warning("High CPU temperature: " + std::to_string(static_cast<int>(cpu_temp)) + "°C");
    }
}

bool SystemMonitor::shouldPerformCheck() const {
    auto now = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(now - last_check_time_);
    return elapsed.count() >= CHECK_INTERVAL_SECONDS;
}
