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

SystemStats gatherSystemStats(ApplicationContext& ctx) {
    SystemStats stats;
    stats.top_objects = ctx.detector->getTopDetectedObjects(10);
    stats.total_objects_detected = ctx.detector->getTotalObjectsDetected();
    stats.total_images_saved = ctx.frame_processor->getTotalImagesSaved();
    stats.brightness_filter_active = ctx.frame_processor->isBrightnessFilterActive();
    stats.current_fps = ctx.perf_monitor->getCurrentFPS();
    stats.avg_processing_time_ms = ctx.perf_monitor->getAverageProcessingTime();
    return stats;
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
        ctx.config.min_confidence, ctx.logger, model_type, ctx.config.detection_scale_factor,
        ctx.config.enable_gpu);

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
    ctx.logger->info("Detection photos will be saved to: " + ctx.config.output_dir);

    // Initialize parallel frame processor
    int effective_threads = ctx.config.enable_parallel_processing ? ctx.config.processing_threads : 1;
    ctx.frame_processor = std::make_shared<ParallelFrameProcessor>(
        ctx.detector, ctx.logger, ctx.perf_monitor, effective_threads, ctx.config.max_frame_queue_size, 
        ctx.config.output_dir, ctx.config.enable_brightness_filter, ctx.config.stationary_timeout_seconds);

    if (!ctx.frame_processor->initialize()) {
        ctx.logger->error("Failed to initialize parallel frame processor");
        return false;
    }

    if (ctx.config.enable_parallel_processing) {
        ctx.logger->info("Parallel processing enabled with " + std::to_string(ctx.config.processing_threads) + " threads");
    } else {
        ctx.logger->info("Sequential processing enabled (single-threaded)");
    }
    
    if (ctx.config.enable_brightness_filter) {
        ctx.logger->info("High brightness filter enabled - will reduce glass reflections in bright conditions");
    }

    // Initialize viewfinder if preview is enabled
    if (ctx.config.show_preview) {
        ctx.viewfinder = std::make_shared<ViewfinderWindow>(ctx.logger);
        if (!ctx.viewfinder->initialize()) {
            ctx.logger->error("Failed to initialize viewfinder window");
            return false;
        }
        ctx.logger->info("Real-time viewfinder enabled - press 'q' or ESC to stop");
    }

    // Initialize network streamer if streaming is enabled
    if (ctx.config.enable_streaming) {
        ctx.network_streamer = std::make_shared<NetworkStreamer>(ctx.logger, ctx.config.streaming_port);
        if (!ctx.network_streamer->initialize()) {
            ctx.logger->error("Failed to initialize network streamer");
            return false;
        }
        if (!ctx.network_streamer->start()) {
            ctx.logger->error("Failed to start network streamer");
            return false;
        }
    }

    // Initialize system monitor for long-term operation
    ctx.system_monitor = std::make_shared<SystemMonitor>(ctx.logger, ctx.config.output_dir);
    ctx.logger->info("System monitor initialized for resource tracking");

    // Initialize Google Sheets client if enabled
    if (ctx.config.enable_google_sheets) {
        GoogleSheetsClient::Config sheets_config;
        sheets_config.enabled = true;
        sheets_config.spreadsheet_id = ctx.config.google_sheets_id;
        sheets_config.api_key = ctx.config.google_sheets_api_key;
        sheets_config.sheet_name = ctx.config.google_sheets_name;
        
        ctx.google_sheets_client = std::make_shared<GoogleSheetsClient>(sheets_config, ctx.logger);
        if (!ctx.google_sheets_client->initialize()) {
            ctx.logger->error("Failed to initialize Google Sheets client");
            return false;
        }
        ctx.logger->info("Google Sheets integration enabled");
        
        // Pass Google Sheets client to object detector for event logging
        ctx.detector->setGoogleSheetsClient(ctx.google_sheets_client);
    }

    // Initialize notification manager if enabled
    if (ctx.config.enable_notifications) {
        NotificationManager::NotificationConfig notif_config;
        notif_config.enable_webhook = ctx.config.enable_webhook;
        notif_config.webhook_url = ctx.config.webhook_url;
        notif_config.enable_sse = ctx.config.enable_sse;
        notif_config.sse_port = ctx.config.sse_port;
        notif_config.enable_file_notification = ctx.config.enable_file_notification;
        notif_config.notification_file_path = ctx.config.notification_file_path;
        notif_config.enable_stdio_notification = ctx.config.enable_stdio_notification;
        
        ctx.notification_manager = std::make_shared<NotificationManager>(ctx.logger, notif_config);
        if (!ctx.notification_manager->initialize()) {
            ctx.logger->error("Failed to initialize notification manager");
            return false;
        }
        ctx.logger->info("Notification system initialized");
    }

    // Initialize timing variables
    ctx.last_heartbeat = std::chrono::steady_clock::now();
    ctx.start_time = std::chrono::steady_clock::now();
    ctx.heartbeat_interval = std::chrono::minutes(ctx.config.heartbeat_interval_minutes);
    ctx.frame_interval = std::chrono::milliseconds(1000 / ctx.config.max_fps);
    ctx.last_frame_time = std::chrono::steady_clock::now();
    
    // Store detection resolution (scaled from camera resolution)
    ctx.detection_width = static_cast<int>(ctx.config.frame_width * ctx.config.detection_scale_factor);
    ctx.detection_height = static_cast<int>(ctx.config.frame_height * ctx.config.detection_scale_factor);

    return true;
}

void runMainProcessingLoop(ApplicationContext& ctx) {
    ctx.logger->info("Starting main processing loop...");
    ctx.logger->info("Analysis rate limit: " + std::to_string(ctx.config.analysis_rate_limit) + " images/second");
    
    if (ctx.config.enable_burst_mode) {
        ctx.logger->info("Burst mode: ENABLED - will max out FPS when new objects enter the scene");
    } else {
        ctx.logger->info("Burst mode: DISABLED");
    }

    // Track time for periodic camera health checks
    auto last_health_check = std::chrono::steady_clock::now();
    constexpr int HEALTH_CHECK_INTERVAL_SECONDS = 60;  // Check every minute

    while (running) {
        auto loop_start = std::chrono::steady_clock::now();

        // Periodic camera health check
        auto health_check_elapsed = std::chrono::duration_cast<std::chrono::seconds>(
            loop_start - last_health_check);
        if (health_check_elapsed.count() >= HEALTH_CHECK_INTERVAL_SECONDS) {
            if (!ctx.webcam->healthCheck()) {
                ctx.logger->error("Camera health check failed - stopping application");
                running = false;
                break;
            }
            last_health_check = loop_start;
        }

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
                    
                    // Display in viewfinder if enabled
                    if (ctx.config.show_preview && ctx.viewfinder) {
                        // Get statistics for display
                        auto stats = gatherSystemStats(ctx);
                        
                        // Get camera name (empty string if not available)
                        std::string camera_name = "";  // Could be extended to get actual camera name
                        
                        ctx.viewfinder->showFrameWithStats(
                            ctx.frame, 
                            result.detections,
                            stats.current_fps,
                            stats.avg_processing_time_ms,
                            stats.total_objects_detected,
                            stats.total_images_saved,
                            ctx.start_time,
                            stats.top_objects,
                            ctx.config.frame_width,
                            ctx.config.frame_height,
                            ctx.config.camera_id,
                            camera_name,
                            ctx.detection_width,
                            ctx.detection_height,
                            stats.brightness_filter_active,
                            ctx.config.enable_gpu,
                            ctx.config.enable_burst_mode
                        );
                        
                        // Check if user wants to close the viewfinder
                        if (ctx.viewfinder->shouldClose()) {
                            ctx.logger->info("Viewfinder closed by user - stopping application");
                            running = false;
                        }
                    }
                    
                    // Update network streamer if enabled
                    if (ctx.config.enable_streaming && ctx.network_streamer) {
                        // Get statistics for display (same as viewfinder)
                        auto stats = gatherSystemStats(ctx);
                        std::string camera_name = "";
                        
                        ctx.network_streamer->updateFrameWithStats(
                            ctx.frame,
                            result.detections,
                            stats.current_fps,
                            stats.avg_processing_time_ms,
                            stats.total_objects_detected,
                            stats.total_images_saved,
                            ctx.start_time,
                            stats.top_objects,
                            ctx.config.frame_width,
                            ctx.config.frame_height,
                            ctx.config.camera_id,
                            camera_name,
                            ctx.detection_width,
                            ctx.detection_height,
                            stats.brightness_filter_active,
                            ctx.config.enable_gpu,
                            ctx.config.enable_burst_mode
                        );
                    }
                    
                    // Send notifications for newly detected objects
                    if (ctx.config.enable_notifications && ctx.notification_manager) {
                        const auto& tracked = ctx.detector->getTrackedObjects();
                        
                        for (const auto& obj : tracked) {
                            // Only notify for newly entered objects in current frame
                            if (obj.is_new && obj.was_present_last_frame && obj.frames_since_detection == 0) {
                                // Create frame with bounding boxes for notification
                                cv::Mat frame_with_boxes = ctx.frame.clone();
                                
                                // Draw all current detections on the frame
                                for (const auto& det : result.detections) {
                                    cv::rectangle(frame_with_boxes, det.bbox, cv::Scalar(0, 255, 0), 2);
                                    std::string label = det.class_name + " " + 
                                        std::to_string(static_cast<int>(det.confidence * 100)) + "%";
                                    cv::putText(frame_with_boxes, label, 
                                        cv::Point(det.bbox.x, det.bbox.y - 10),
                                        cv::FONT_HERSHEY_SIMPLEX, 0.5, cv::Scalar(0, 255, 0), 2);
                                }
                                
                                // Gather status information
                                auto stats = gatherSystemStats(ctx);
                                
                                // Create notification data
                                NotificationManager::NotificationData notif_data;
                                notif_data.object_type = obj.object_type;
                                notif_data.x = obj.center.x;
                                notif_data.y = obj.center.y;
                                notif_data.confidence = 0.0;  // Will be set from detection if available
                                
                                // Find corresponding detection to get confidence
                                for (const auto& det : result.detections) {
                                    if (det.class_name == obj.object_type) {
                                        notif_data.confidence = det.confidence;
                                        break;
                                    }
                                }
                                
                                notif_data.timestamp = std::chrono::system_clock::now();
                                notif_data.frame_with_boxes = frame_with_boxes;
                                notif_data.all_detections = result.detections;
                                notif_data.current_fps = stats.current_fps;
                                notif_data.avg_processing_time_ms = stats.avg_processing_time_ms;
                                notif_data.total_objects_detected = stats.total_objects_detected;
                                notif_data.total_images_saved = stats.total_images_saved;
                                notif_data.top_objects = stats.top_objects;
                                notif_data.brightness_filter_active = stats.brightness_filter_active;
                                notif_data.gpu_enabled = ctx.config.enable_gpu;
                                notif_data.burst_mode_enabled = ctx.config.enable_burst_mode;
                                
                                // Send notification
                                ctx.notification_manager->notifyNewObject(notif_data);
                            }
                        }
                    }
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
        
        // Check and print hourly summary
        ctx.logger->checkAndPrintSummary(ctx.config.summary_interval_minutes);

        // Perform periodic system resource checks
        if (ctx.system_monitor) {
            ctx.system_monitor->performPeriodicCheck();
        }

        // Burst mode logic: detect new object types and activate/deactivate burst mode
        if (ctx.config.enable_burst_mode) {
            // Get current object types from tracked objects
            std::set<std::string> current_object_types;
            const auto& tracked = ctx.detector->getTrackedObjects();
            
            bool has_new_object_type = false;
            bool all_objects_stationary = true;
            
            for (const auto& obj : tracked) {
                // Only consider objects present in current frame
                if (obj.was_present_last_frame && obj.frames_since_detection == 0) {
                    current_object_types.insert(obj.object_type);
                    
                    // Check if this is a new object type not seen in previous frame
                    if (ctx.previous_object_types.find(obj.object_type) == ctx.previous_object_types.end()) {
                        has_new_object_type = true;
                    }
                    
                    // Check if this object is newly entered (not just a new type)
                    if (obj.is_new) {
                        has_new_object_type = true;
                    }
                    
                    // Check if any object is not stationary
                    if (!obj.is_stationary) {
                        all_objects_stationary = false;
                    }
                }
            }
            
            // Update burst mode state
            bool previous_burst_state = ctx.burst_mode_active;
            
            if (has_new_object_type) {
                // Activate burst mode when new object type enters
                ctx.burst_mode_active = true;
                if (!previous_burst_state) {
                    ctx.logger->info("Burst mode ACTIVATED - new object type detected");
                }
            } else if (all_objects_stationary && !current_object_types.empty()) {
                // Deactivate burst mode when all objects are stationary
                if (ctx.burst_mode_active) {
                    ctx.burst_mode_active = false;
                    ctx.logger->info("Burst mode DEACTIVATED - all objects stationary");
                }
            } else if (current_object_types.empty()) {
                // Deactivate burst mode when no objects are present
                if (ctx.burst_mode_active) {
                    ctx.burst_mode_active = false;
                    ctx.logger->info("Burst mode DEACTIVATED - no objects detected");
                }
            }
            
            // Update previous object types for next iteration
            ctx.previous_object_types = current_object_types;
        }

        // Apply rate limiting with evenly distributed sleep time
        // Calculate required sleep time based on analysis rate limit and actual processing time
        double target_interval_ms = 1000.0 / ctx.config.analysis_rate_limit;
        double actual_processing_time_ms = ctx.perf_monitor->getLastProcessingTime();
        double sleep_time_ms = target_interval_ms - actual_processing_time_ms;
        
        // Skip sleep if burst mode is active
        if (ctx.config.enable_burst_mode && ctx.burst_mode_active) {
            // In burst mode, we process frames as fast as possible
            // Only add minimal delay to prevent excessive CPU usage
            std::this_thread::sleep_for(std::chrono::milliseconds(1));
            ctx.logger->debug("Burst mode active: skipping normal rate limiting (minimal 1ms delay)");
        } else if (sleep_time_ms > 0) {
            auto sleep_duration = std::chrono::milliseconds(static_cast<long>(sleep_time_ms));
            ctx.logger->debug("Rate limiting: sleeping for " + std::to_string(sleep_time_ms) + " ms (processing took " + std::to_string(actual_processing_time_ms) + " ms, target interval: " + std::to_string(target_interval_ms) + " ms)");
            std::this_thread::sleep_for(sleep_duration);
        } else {
            // Processing took longer than target interval - small delay to prevent excessive CPU usage
            std::this_thread::sleep_for(std::chrono::milliseconds(1));
        }
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
    
    // Close viewfinder if it was open
    if (ctx.viewfinder) {
        ctx.viewfinder->close();
    }
    
    // Stop network streamer if it was running
    if (ctx.network_streamer) {
        ctx.network_streamer->stop();
    }
    
    ctx.webcam->release();
    
    // Print final summary covering entire program runtime
    ctx.logger->printFinalSummary();
    
    ctx.logger->info("Object Detection Application stopped");
}