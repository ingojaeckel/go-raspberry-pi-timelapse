#include "application_context.hpp"
#include <iostream>
#include <csignal>
#include <thread>

// External reference to global running flag
extern std::atomic<bool> running;

void signalHandler(int signal) {
    std::cout << "\nReceived signal " << signal << ". Shutting down gracefully..." << std::endl;
    running = false;
}

void setupSignalHandlers() {
    std::signal(SIGINT, signalHandler);
    std::signal(SIGTERM, signalHandler);
}

bool parseAndValidateConfig(ApplicationContext& ctx, int argc, char* argv[]) {
    auto parse_result = ctx.config_manager.parseArgs(argc, argv);
    
    if (parse_result == ConfigManager::ParseResult::HELP_REQUESTED ||
        parse_result == ConfigManager::ParseResult::LIST_REQUESTED) {
        std::exit(0);  // Success exit for help/list
    }
    
    if (parse_result == ConfigManager::ParseResult::PARSE_ERROR) {
        std::cerr << "Error parsing arguments. Use --help for usage information." << std::endl;
        return false;
    }

    if (!ctx.config_manager.validateConfig()) {
        std::cerr << "Invalid configuration. Exiting." << std::endl;
        return false;
    }

    ctx.config = ctx.config_manager.getConfig();
    return true;
}

bool initializeComponents(ApplicationContext& ctx) {
    // Initialize logger
    ctx.logger = std::make_shared<Logger>(ctx.config.log_file, ctx.config.verbose);
    ctx.logger->info("Object Detection Application Starting");
    ctx.logger->info("Version: 1.0.0");
    ctx.logger->info("Target: Real-time object detection from webcam data");

    // Initialize performance monitor
    ctx.perf_monitor = std::make_shared<PerformanceMonitor>(
        ctx.logger, ctx.config.min_fps_warning_threshold);

    // Initialize webcam interface
    ctx.webcam = std::make_shared<WebcamInterface>(
        ctx.config.camera_id, ctx.config.frame_width, ctx.config.frame_height, ctx.logger);
    
    if (!ctx.webcam->initialize()) {
        ctx.logger->error("Failed to initialize webcam interface");
        return false;
    }

    ctx.logger->info("Webcam initialized: " + ctx.webcam->getCameraInfo());

    // Initialize object detector with model type selection
    DetectionModelFactory::ModelType model_type;
    try {
        model_type = DetectionModelFactory::parseModelType(ctx.config.model_type);
    } catch (const std::exception& e) {
        ctx.logger->error("Invalid model type: " + ctx.config.model_type);
        ctx.logger->error("Available models: yolov5s, yolov5l, yolov8n, yolov8m");
        return false;
    }

    ctx.detector = std::make_shared<ObjectDetector>(
        ctx.config.model_path, ctx.config.config_path, ctx.config.classes_path,
        ctx.config.min_confidence, ctx.logger, model_type);

    if (!ctx.detector->initialize()) {
        ctx.logger->error("Failed to initialize object detector");
        return false;
    }

    ctx.logger->info("Object detector initialized successfully");
    
    // Log model performance characteristics
    auto model_metrics = ctx.detector->getModelMetrics();
    ctx.logger->info("Using model: " + model_metrics.model_name + " (" + model_metrics.model_type + ")");
    ctx.logger->info("Model accuracy: " + std::to_string(static_cast<int>(model_metrics.accuracy_score * 100)) + "%");
    ctx.logger->info("Expected inference time: ~" + std::to_string(model_metrics.avg_inference_time_ms) + "ms");
    ctx.logger->info("Model description: " + model_metrics.description);
    
    ctx.logger->info("Target objects: person, vehicle, small animals (cat/dog/fox)");
    ctx.logger->info("Minimum confidence threshold: " + std::to_string(ctx.config.min_confidence));
    ctx.logger->info("Maximum processing rate: " + std::to_string(ctx.config.max_fps) + " fps");
    ctx.logger->info("Performance warning threshold: " + std::to_string(ctx.config.min_fps_warning_threshold) + " fps");
    ctx.logger->info("Detection photos will be saved to: " + config.output_dir);

    // Initialize parallel frame processor
    int effective_threads = ctx.config.enable_parallel_processing ? ctx.config.processing_threads : 1;
    ctx.frame_processor = std::make_shared<ParallelFrameProcessor>(
        ctx.detector, ctx.logger, ctx.perf_monitor, effective_threads, ctx.config.max_frame_queue_size, config.output_dir);

    if (!ctx.frame_processor->initialize()) {
        ctx.logger->error("Failed to initialize parallel frame processor");
        return false;
    }

    if (ctx.config.enable_parallel_processing) {
        ctx.logger->info("Parallel processing enabled with " + std::to_string(ctx.config.processing_threads) + " threads");
    } else {
        ctx.logger->info("Sequential processing enabled (single-threaded)");
    }

    // Initialize timing variables
    ctx.last_heartbeat = std::chrono::steady_clock::now();
    ctx.heartbeat_interval = std::chrono::minutes(ctx.config.heartbeat_interval_minutes);
    ctx.frame_interval = std::chrono::milliseconds(1000 / ctx.config.max_fps);
    ctx.last_frame_time = std::chrono::steady_clock::now();

    return true;
}

void runMainProcessingLoop(ApplicationContext& ctx) {
    ctx.logger->info("Starting main processing loop...");

    while (running) {
        auto loop_start = std::chrono::steady_clock::now();

        // Check if enough time has passed for next frame
        if (loop_start - ctx.last_frame_time < ctx.frame_interval) {
            std::this_thread::sleep_for(std::chrono::milliseconds(10));
            continue;
        }

        // Capture frame from webcam
        if (!ctx.webcam->captureFrame(ctx.frame)) {
            ctx.logger->warning("Failed to capture frame from webcam");
            std::this_thread::sleep_for(std::chrono::milliseconds(100));
            continue;
        }

        // Start performance monitoring
        ctx.perf_monitor->startFrameProcessing();

        // Submit frame for processing (parallel or sequential)
        auto future = ctx.frame_processor->submitFrame(ctx.frame);
        ctx.pending_frames.push(std::move(future));

        // Process completed frames
        while (!ctx.pending_frames.empty()) {
            auto& front_future = ctx.pending_frames.front();
            
            // Check if the result is ready (non-blocking for parallel, immediate for sequential)
            if (ctx.frame_processor->isParallelEnabled()) {
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
                ctx.logger->error("Error processing frame result: " + std::string(e.what()));
            }
            
            ctx.pending_frames.pop();
        }

        // End performance monitoring
        ctx.perf_monitor->endFrameProcessing();

        // Check performance and log warnings if needed
        ctx.perf_monitor->checkPerformanceThreshold();

        ctx.last_frame_time = loop_start;

        // Periodic heartbeat logging
        auto now = std::chrono::steady_clock::now();
        if (now - ctx.last_heartbeat >= ctx.heartbeat_interval) {
            ctx.logger->logHeartbeat();
            ctx.perf_monitor->logPerformanceReport();
            ctx.last_heartbeat = now;
        }

        // Small delay to prevent excessive CPU usage
        std::this_thread::sleep_for(std::chrono::milliseconds(1));
    }
}

void performGracefulShutdown(ApplicationContext& ctx) {
    ctx.logger->info("Shutting down gracefully...");
    
    // Shutdown frame processor first
    ctx.frame_processor->shutdown();
    
    // Process any remaining frames
    while (!ctx.pending_frames.empty()) {
        try {
            auto result = ctx.pending_frames.front().get();
            // Process final results if needed
        } catch (...) {
            // Ignore errors during shutdown
        }
        ctx.pending_frames.pop();
    }
    
    ctx.webcam->release();
    ctx.logger->info("Object Detection Application stopped");
}