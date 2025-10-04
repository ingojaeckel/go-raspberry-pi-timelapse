#include "system_monitor.hpp"
#include <sys/statvfs.h>
#include <fstream>
#include <sstream>
#include <iomanip>
#include <vector>
#include <algorithm>
#include <dirent.h>
#include <sys/stat.h>
#include <unistd.h>

SystemMonitor::SystemMonitor(std::shared_ptr<Logger> logger, 
                            const std::string& output_dir)
    : logger_(logger), output_dir_(output_dir) {
    last_check_time_ = std::chrono::steady_clock::now();
    last_cleanup_time_ = std::chrono::steady_clock::now();
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
    
    // Perform cleanup if needed and interval has passed
    if (shouldPerformCleanup() && isDiskSpaceCritical()) {
        int deleted = cleanupOldDetections();
        if (deleted > 0) {
            logger_->info("Cleaned up " + std::to_string(deleted) + " old detection photos due to low disk space");
        }
        last_cleanup_time_ = std::chrono::steady_clock::now();
    }
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

int SystemMonitor::cleanupOldDetections() {
    // Get list of all detection image files
    struct dirent* entry;
    DIR* dir = opendir(output_dir_.c_str());
    
    if (!dir) {
        logger_->warning("Cannot open output directory for cleanup: " + output_dir_);
        return 0;
    }
    
    // Collect all .jpg files with their modification times
    std::vector<std::pair<std::string, time_t>> files;
    
    while ((entry = readdir(dir)) != nullptr) {
        std::string filename = entry->d_name;
        
        // Skip . and .. and non-jpg files
        if (filename == "." || filename == ".." || 
            filename.substr(filename.find_last_of(".") + 1) != "jpg") {
            continue;
        }
        
        std::string filepath = output_dir_ + "/" + filename;
        struct stat file_stat;
        
        if (stat(filepath.c_str(), &file_stat) == 0) {
            files.push_back({filepath, file_stat.st_mtime});
        }
    }
    
    closedir(dir);
    
    if (files.empty()) {
        return 0;
    }
    
    // Sort by modification time (oldest first)
    std::sort(files.begin(), files.end(),
              [](const auto& a, const auto& b) {
                  return a.second < b.second;
              });
    
    // Delete oldest 20% of files or until we have enough space
    int to_delete = std::max(1, static_cast<int>(files.size() * 0.2));
    int deleted = 0;
    
    for (int i = 0; i < to_delete && i < static_cast<int>(files.size()); ++i) {
        if (unlink(files[i].first.c_str()) == 0) {
            deleted++;
            logger_->debug("Deleted old detection photo: " + files[i].first);
        }
    }
    
    return deleted;
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

bool SystemMonitor::shouldPerformCleanup() const {
    auto now = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(now - last_cleanup_time_);
    return elapsed.count() >= CLEANUP_INTERVAL_SECONDS;
}
