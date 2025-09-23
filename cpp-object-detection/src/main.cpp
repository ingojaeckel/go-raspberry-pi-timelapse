#include <iostream>
#include <memory>
#include <thread>
#include <chrono>
#include <csignal>
#include <atomic>

#include "config_manager.hpp"
#include "logger.hpp"
#include "webcam_interface.hpp"
#include "object_detector.hpp"
#include "performance_monitor.hpp"

// Global flag for clean shutdown
std::atomic<bool> running{true};

void signalHandler(int signal) {
    std::cout << "\nReceived signal " << signal << ". Shutting down gracefully..." << std::endl;
    running = false;
}

int main(int argc, char* argv[]) {
    // Setup signal handlers for graceful shutdown
    std::signal(SIGINT, signalHandler);
    std::signal(SIGTERM, signalHandler);

    try {
        // Parse configuration
        ConfigManager config_manager;
        if (!config_manager.parseArgs(argc, argv)) {
            config_manager.printUsage(argv[0]);
            return 1;
        }

        if (!config_manager.validateConfig()) {
            std::cerr << "Invalid configuration. Exiting." << std::endl;
            return 1;
        }

        const auto& config = config_manager.getConfig();

        // Initialize logger
        auto logger = std::make_shared<Logger>(config.log_file, config.verbose);
        logger->info("Object Detection Application Starting");
        logger->info("Version: 1.0.0");
        logger->info("Target: Real-time object detection from webcam data");

        // Initialize performance monitor
        auto perf_monitor = std::make_shared<PerformanceMonitor>(
            logger, config.min_fps_warning_threshold);

        // Initialize webcam interface
        auto webcam = std::make_shared<WebcamInterface>(
            config.camera_id, config.frame_width, config.frame_height, logger);
        
        if (!webcam->initialize()) {
            logger->error("Failed to initialize webcam interface");
            return 1;
        }

        logger->info("Webcam initialized: " + webcam->getCameraInfo());

        // Initialize object detector
        auto detector = std::make_shared<ObjectDetector>(
            config.model_path, config.config_path, config.classes_path,
            config.min_confidence, logger);

        if (!detector->initialize()) {
            logger->error("Failed to initialize object detector");
            return 1;
        }

        logger->info("Object detector initialized successfully");
        logger->info("Target objects: person, vehicle, small animals (cat/dog/fox)");
        logger->info("Minimum confidence threshold: " + std::to_string(config.min_confidence));
        logger->info("Maximum processing rate: " + std::to_string(config.max_fps) + " fps");
        logger->info("Performance warning threshold: " + std::to_string(config.min_fps_warning_threshold) + " fps");

        // Main processing loop
        cv::Mat frame;
        auto last_heartbeat = std::chrono::steady_clock::now();
        auto heartbeat_interval = std::chrono::minutes(config.heartbeat_interval_minutes);
        auto frame_interval = std::chrono::milliseconds(1000 / config.max_fps);
        auto last_frame_time = std::chrono::steady_clock::now();

        logger->info("Starting main processing loop...");

        while (running) {
            auto loop_start = std::chrono::steady_clock::now();

            // Check if enough time has passed for next frame
            if (loop_start - last_frame_time < frame_interval) {
                std::this_thread::sleep_for(std::chrono::milliseconds(10));
                continue;
            }

            // Capture frame from webcam
            if (!webcam->captureFrame(frame)) {
                logger->warning("Failed to capture frame from webcam");
                std::this_thread::sleep_for(std::chrono::milliseconds(100));
                continue;
            }

            // Start performance monitoring
            perf_monitor->startFrameProcessing();

            // Process frame for object detection
            detector->processFrame(frame);

            // End performance monitoring
            perf_monitor->endFrameProcessing();

            // Check performance and log warnings if needed
            perf_monitor->checkPerformanceThreshold();

            last_frame_time = loop_start;

            // Periodic heartbeat logging
            auto now = std::chrono::steady_clock::now();
            if (now - last_heartbeat >= heartbeat_interval) {
                logger->logHeartbeat();
                perf_monitor->logPerformanceReport();
                last_heartbeat = now;
            }

            // Small delay to prevent excessive CPU usage
            std::this_thread::sleep_for(std::chrono::milliseconds(1));
        }

        // Cleanup
        logger->info("Shutting down gracefully...");
        webcam->release();
        logger->info("Object Detection Application stopped");

    } catch (const std::exception& e) {
        std::cerr << "Fatal error: " << e.what() << std::endl;
        return 1;
    } catch (...) {
        std::cerr << "Unknown fatal error occurred" << std::endl;
        return 1;
    }

    return 0;
}