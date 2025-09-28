#include <iostream>
#include <memory>
#include <thread>
#include <chrono>
#include <csignal>
#include <atomic>
#include <cstdio>  // For std::setvbuf

#include "config_manager.hpp"
#include "logger.hpp"
#include "webcam_interface.hpp"
#include "object_detector.hpp"
#include "performance_monitor.hpp"
#include "parallel_frame_processor.hpp"

// Global flag for clean shutdown
std::atomic<bool> running{true};

void signalHandler(int signal) {
    std::cout << "\nReceived signal " << signal << ". Shutting down gracefully..." << std::endl;
    running = false;
}

int main(int argc, char* argv[]) {
    // Force line buffering for stdout (macOS compatibility)
    std::setvbuf(stdout, nullptr, _IOLBF, 0);
    std::setvbuf(stderr, nullptr, _IOLBF, 0);
    
    // Setup signal handlers for graceful shutdown
    std::signal(SIGINT, signalHandler);
    std::signal(SIGTERM, signalHandler);

    try {
        // Parse configuration
        ConfigManager config_manager;
        auto parse_result = config_manager.parseArgs(argc, argv);
        
        if (parse_result == ConfigManager::ParseResult::HELP_REQUESTED ||
            parse_result == ConfigManager::ParseResult::LIST_REQUESTED) {
            // Ensure output is flushed before exiting (macOS compatibility)
            std::cout.flush();
            std::cerr.flush();
            return 0;  // Success exit for help/list
        }
        
        if (parse_result == ConfigManager::ParseResult::PARSE_ERROR) {
            std::cerr << "Error parsing arguments. Use --help for usage information." << std::endl;
            std::cerr.flush();
            return 1;
        }

        if (!config_manager.validateConfig()) {
            std::cerr << "Invalid configuration. Exiting." << std::endl;
            std::cerr.flush();
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

        // Initialize object detector with model type selection
        DetectionModelFactory::ModelType model_type;
        try {
            model_type = DetectionModelFactory::parseModelType(config.model_type);
        } catch (const std::exception& e) {
            logger->error("Invalid model type: " + config.model_type);
            logger->error("Available models: yolov5s, yolov5l, yolov8n, yolov8m");
            return 1;
        }

        auto detector = std::make_shared<ObjectDetector>(
            config.model_path, config.config_path, config.classes_path,
            config.min_confidence, logger, model_type);

        if (!detector->initialize()) {
            logger->error("Failed to initialize object detector");
            return 1;
        }

        logger->info("Object detector initialized successfully");
        
        // Log model performance characteristics
        auto model_metrics = detector->getModelMetrics();
        logger->info("Using model: " + model_metrics.model_name + " (" + model_metrics.model_type + ")");
        logger->info("Model accuracy: " + std::to_string(static_cast<int>(model_metrics.accuracy_score * 100)) + "%");
        logger->info("Expected inference time: ~" + std::to_string(model_metrics.avg_inference_time_ms) + "ms");
        logger->info("Model description: " + model_metrics.description);
        
        logger->info("Target objects: person, vehicle, small animals (cat/dog/fox)");
        logger->info("Minimum confidence threshold: " + std::to_string(config.min_confidence));
        logger->info("Maximum processing rate: " + std::to_string(config.max_fps) + " fps");
        logger->info("Performance warning threshold: " + std::to_string(config.min_fps_warning_threshold) + " fps");

        // Initialize parallel frame processor
        int effective_threads = config.enable_parallel_processing ? config.processing_threads : 1;
        auto frame_processor = std::make_shared<ParallelFrameProcessor>(
            detector, logger, perf_monitor, effective_threads, config.max_frame_queue_size);

        if (!frame_processor->initialize()) {
            logger->error("Failed to initialize parallel frame processor");
            return 1;
        }

        if (config.enable_parallel_processing) {
            logger->info("Parallel processing enabled with " + std::to_string(config.processing_threads) + " threads");
        } else {
            logger->info("Sequential processing enabled (single-threaded)");
        }

        // Main processing loop
        cv::Mat frame;
        auto last_heartbeat = std::chrono::steady_clock::now();
        auto heartbeat_interval = std::chrono::minutes(config.heartbeat_interval_minutes);
        auto frame_interval = std::chrono::milliseconds(1000 / config.max_fps);
        auto last_frame_time = std::chrono::steady_clock::now();
        
        // Queue to track pending frame processing futures
        std::queue<std::future<ParallelFrameProcessor::FrameResult>> pending_frames;

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

            // Submit frame for processing (parallel or sequential)
            auto future = frame_processor->submitFrame(frame);
            pending_frames.push(std::move(future));

            // Process completed frames
            while (!pending_frames.empty()) {
                auto& front_future = pending_frames.front();
                
                // Check if the result is ready (non-blocking for parallel, immediate for sequential)
                if (frame_processor->isParallelEnabled()) {
                    // For parallel processing, only process if ready
                    if (front_future.wait_for(std::chrono::milliseconds(0)) != std::future_status::ready) {
                        break;
                    }
                }
                
                try {
                    auto result = front_future.get();
                    if (result.processed) {
                        // Process detection results here if needed
                        // Results are already logged by the frame processor
                    }
                } catch (const std::exception& e) {
                    logger->error("Error processing frame result: " + std::string(e.what()));
                }
                
                pending_frames.pop();
            }

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
        
        // Shutdown frame processor first
        frame_processor->shutdown();
        
        // Process any remaining frames
        while (!pending_frames.empty()) {
            try {
                auto result = pending_frames.front().get();
                // Process final results if needed
            } catch (...) {
                // Ignore errors during shutdown
            }
            pending_frames.pop();
        }
        
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