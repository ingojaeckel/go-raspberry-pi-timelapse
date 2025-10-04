// Example demonstrating the hourly summary feature
// This shows how the feature would output summaries when detections are recorded

#include "logger.hpp"
#include <iostream>
#include <thread>
#include <chrono>

int main() {
    // Create logger instance
    Logger logger("demo_summary.log", true);
    
    std::cout << "=== Hourly Summary Demo ===" << std::endl;
    std::cout << "This demonstrates the hourly object detection summary feature.\n" << std::endl;
    
    // Simulate a series of detections over time
    std::cout << "Recording detections..." << std::endl;
    
    // Stationary car from 0:00-0:10
    logger.recordDetection("car", true);
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    logger.recordDetection("car", true);
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    logger.recordDetection("car", true);
    
    // Person at 0:10
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    logger.recordDetection("person", false);
    
    // Animal at 0:30
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    logger.recordDetection("cat", false);
    
    // Another animal at 0:31
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    logger.recordDetection("dog", false);
    
    // Two people at 0:50 (detected close together)
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    logger.recordDetection("person", false);
    std::this_thread::sleep_for(std::chrono::milliseconds(50));
    logger.recordDetection("person", false);
    
    // Stationary car from 0:50-1:00
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    logger.recordDetection("car", true);
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    logger.recordDetection("car", true);
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    logger.recordDetection("car", true);
    
    std::cout << "\n=== Printing Summary ===" << std::endl;
    
    // Manually trigger summary (normally this would happen automatically after 1 hour)
    logger.printHourlySummary();
    
    std::cout << "\n=== Demo Complete ===" << std::endl;
    std::cout << "In production, this summary would print automatically every hour (configurable)." << std::endl;
    std::cout << "The summary_interval_minutes config option controls the interval." << std::endl;
    
    return 0;
}
