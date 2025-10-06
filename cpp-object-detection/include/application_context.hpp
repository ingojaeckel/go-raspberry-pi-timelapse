#pragma once

#include <memory>
#include <queue>
#include <future>
#include <chrono>
#include <set>
#include <opencv2/opencv.hpp>

#include "config_manager.hpp"
#include "logger.hpp"
#include "webcam_interface.hpp"
#include "object_detector.hpp"
#include "performance_monitor.hpp"
#include "parallel_frame_processor.hpp"
#include "detection_model_interface.hpp"
#include "viewfinder_window.hpp"
#include "network_streamer.hpp"
#include "system_monitor.hpp"
#include "google_sheets_client.hpp"

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
    std::shared_ptr<ViewfinderWindow> viewfinder;
    std::shared_ptr<NetworkStreamer> network_streamer;
    std::shared_ptr<SystemMonitor> system_monitor;
    std::shared_ptr<GoogleSheetsClient> google_sheets_client;
    
    // Processing state
    std::queue<std::future<ParallelFrameProcessor::FrameResult>> pending_frames;
    cv::Mat frame;
    std::chrono::steady_clock::time_point last_heartbeat;
    std::chrono::steady_clock::time_point last_frame_time;
    std::chrono::steady_clock::time_point start_time;
    std::chrono::milliseconds heartbeat_interval;
    std::chrono::milliseconds frame_interval;
    int detection_width;
    int detection_height;
    
    // Burst mode state
    bool burst_mode_active = false;
    std::set<std::string> previous_object_types;  // Track object types from previous frame
};

// Function declarations for main.cpp helper functions
void setupSignalHandlers();
bool parseAndValidateConfig(ApplicationContext& ctx, int argc, char* argv[]);
bool initializeComponents(ApplicationContext& ctx);
void runMainProcessingLoop(ApplicationContext& ctx);
void performGracefulShutdown(ApplicationContext& ctx);