#pragma once

#include <memory>
#include <queue>
#include <future>
#include <chrono>
#include <opencv2/opencv.hpp>

#include "config_manager.hpp"
#include "logger.hpp"
#include "webcam_interface.hpp"
#include "object_detector.hpp"
#include "performance_monitor.hpp"
#include "parallel_frame_processor.hpp"
#include "detection_model_interface.hpp"

/**
 * Context structure to hold shared application state
 */
struct ApplicationContext {
    // Configuration
    ConfigManager config_manager;
    ConfigManager::Config config;
    
    // Core components
    std::shared_ptr<Logger> logger;
    std::shared_ptr<PerformanceMonitor> perf_monitor;
    std::shared_ptr<WebcamInterface> webcam;
    std::shared_ptr<ObjectDetector> detector;
    std::shared_ptr<ParallelFrameProcessor> frame_processor;
    
    // Processing state
    std::queue<std::future<ParallelFrameProcessor::FrameResult>> pending_frames;
    cv::Mat frame;
    std::chrono::steady_clock::time_point last_heartbeat;
    std::chrono::steady_clock::time_point last_frame_time;
    std::chrono::milliseconds heartbeat_interval;
    std::chrono::milliseconds frame_interval;
};

// Function declarations for main.cpp helper functions
void setupSignalHandlers();
bool parseAndValidateConfig(ApplicationContext& ctx, int argc, char* argv[]);
bool initializeComponents(ApplicationContext& ctx);
void runMainProcessingLoop(ApplicationContext& ctx);
void performGracefulShutdown(ApplicationContext& ctx);