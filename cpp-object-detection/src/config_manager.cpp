#include "config_manager.hpp"
#include "webcam_interface.hpp"
#include <iostream>
#include <sstream>
#include <algorithm>

ConfigManager::ConfigManager() : config_(std::make_unique<Config>()) {
    setDefaults();
}

ConfigManager::~ConfigManager() = default;

void ConfigManager::setDefaults() {
    // All defaults are already set in the struct definition
    // This method exists for potential future customizations
}

ConfigManager::ParseResult ConfigManager::parseArgs(int argc, char* argv[]) {
    for (int i = 1; i < argc; ++i) {
        std::string arg = argv[i];
        
        if (arg == "--help" || arg == "-h") {
            printUsage(argv[0]);
            return ParseResult::HELP_REQUESTED;
        }
        
        // Handle arguments with values
        if (i + 1 < argc) {
            std::string value = argv[i + 1];
            if (parseArgument(arg, value)) {
                ++i; // Skip the value argument
                continue;
            }
        }
        
        // Handle boolean flags
        if (arg == "--verbose" || arg == "-v") {
            config_->verbose = true;
        } else if (arg == "--enable-gpu") {
            config_->enable_gpu = true;
        } else if (arg == "--enable-parallel") {
            config_->enable_parallel_processing = true;
        } else if (arg == "--no-headless") {
            config_->headless = false;
        } else if (arg == "--show-preview") {
            config_->show_preview = true;
        } else if (arg == "--enable-streaming") {
            config_->enable_streaming = true;
        } else if (arg == "--enable-brightness-filter") {
            config_->enable_brightness_filter = true;
        } else if (arg == "--enable-burst-mode") {
            config_->enable_burst_mode = true;
        } else if (arg == "--list-cameras" || arg == "--list") {
            // List cameras and exit
            listCameras();
            return ParseResult::LIST_REQUESTED;
        } else {
            std::cerr << "Unknown argument: " << arg << std::endl;
            std::cerr.flush();
            return ParseResult::PARSE_ERROR;
        }
    }
    
    return ParseResult::SUCCESS;
}

bool ConfigManager::parseArgument(const std::string& arg, const std::string& value) {
    try {
        if (arg == "--max-fps") {
            config_->max_fps = std::stoi(value);
        } else if (arg == "--min-confidence") {
            config_->min_confidence = std::stod(value);
        } else if (arg == "--min-fps-warning") {
            config_->min_fps_warning_threshold = std::stoi(value);
        } else if (arg == "--log-file") {
            config_->log_file = value;
        } else if (arg == "--heartbeat-interval") {
            config_->heartbeat_interval_minutes = std::stoi(value);
        } else if (arg == "--summary-interval") {
            config_->summary_interval_minutes = std::stoi(value);
        } else if (arg == "--camera-id") {
            config_->camera_id = std::stoi(value);
        } else if (arg == "--frame-width") {
            config_->frame_width = std::stoi(value);
        } else if (arg == "--frame-height") {
            config_->frame_height = std::stoi(value);
        } else if (arg == "--model-path") {
            config_->model_path = value;
        } else if (arg == "--config-path") {
            config_->config_path = value;
        } else if (arg == "--classes-path") {
            config_->classes_path = value;
        } else if (arg == "--model-type") {
            config_->model_type = value;
        } else if (arg == "--detection-scale") {
            config_->detection_scale_factor = std::stod(value);
        } else if (arg == "--processing-threads") {
            config_->processing_threads = std::stoi(value);
        } else if (arg == "--max-frame-queue") {
            config_->max_frame_queue_size = std::stoi(value);
        } else if (arg == "--output-dir") {
            config_->output_dir = value;
        } else if (arg == "--analysis-rate-limit") {
            config_->analysis_rate_limit = std::stod(value);
        } else if (arg == "--streaming-port") {
            config_->streaming_port = std::stoi(value);
        } else if (arg == "--stationary-timeout") {
            config_->stationary_timeout_seconds = std::stoi(value);
        } else {
            return false;
        }
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Invalid value for " << arg << ": " << value << std::endl;
        std::cerr.flush();
        return false;
    }
}

const ConfigManager::Config& ConfigManager::getConfig() const {
    return *config_;
}

void ConfigManager::printUsage(const std::string& program_name) const {
    std::cout << "Object Detection Application\n"
              << "Usage: " << program_name << " [OPTIONS]\n\n"
              << "Real-time object detection from webcam data (720p)\n"
              << "Detects people, vehicles, and small animals (cat/dog/fox)\n\n"
              << "OPTIONS:\n"
              << "  -h, --help                     Show this help message\n"
              << "  -v, --verbose                  Enable verbose logging\n"
              << "  --list-cameras, --list         List all available cameras and exit\n"
              << "  --max-fps N                    Maximum frames per second to process (default: 5)\n"
              << "  --min-confidence N             Minimum confidence threshold (0.0-1.0, default: 0.5)\n"
              << "  --min-fps-warning N            FPS threshold for performance warnings (default: 1)\n"
              << "  --log-file FILE                Log file path (default: object_detection.log)\n"
              << "  --heartbeat-interval N         Heartbeat log interval in minutes (default: 10)\n"
              << "  --summary-interval N           Detection summary interval in minutes (default: 60)\n"
              << "  --camera-id N                  Camera device ID (default: 0)\n"
              << "  --frame-width N                Frame width in pixels (default: 1280)\n"
              << "  --frame-height N               Frame height in pixels (default: 720)\n"
              << "  --model-path FILE              Path to ONNX model file (default: models/yolov5s.onnx)\n"
              << "  --config-path FILE             Path to model config file (default: models/yolov5s.yaml)\n"
              << "  --classes-path FILE            Path to class names file (default: models/coco.names)\n"
              << "  --model-type TYPE              Detection model type (default: yolov5s)\n"
              << "                                 Available: yolov5s (fast), yolov5l (accurate), yolov8n, yolov8m\n"
              << "  --detection-scale N            Scale factor for detection (0.1-1.0, default: 0.5)\n"
              << "                                 Lower values = faster but may reduce accuracy\n"
              << "                                 0.5 = 50% reduction (1280x720 -> 640x360, 75% fewer pixels)\n"
              << "  --output-dir DIR               Directory to save detection photos (default: detections)\n"
              << "  --processing-threads N         Number of processing threads (default: 1)\n"
              << "  --enable-parallel              Enable parallel frame processing\n"
              << "  --max-frame-queue N            Maximum frames in processing queue (default: 10)\n"
              << "  --analysis-rate-limit N        Maximum images to analyze per second (default: 1.0)\n"
              << "                                 Lower values reduce CPU usage by adding sleep between analyses\n"
              << "  --enable-gpu                   Enable GPU acceleration (default: disabled)\n"
              << "                                 Linux: Uses CUDA backend if available\n"
              << "                                 macOS: Uses OpenCL backend for Intel integrated/discrete GPUs\n"
              << "  --no-headless                  Disable headless mode (show GUI windows)\n"
              << "  --show-preview                 Show real-time viewfinder with detection bounding boxes\n"
              << "  --enable-streaming             Enable MJPEG HTTP streaming over network (default: disabled)\n"
              << "  --streaming-port N             Port for HTTP streaming server (default: 8080)\n"
              << "  --enable-brightness-filter     Enable high brightness filter to reduce glass reflections (default: disabled)\n"
              << "  --stationary-timeout N         Seconds before stopping photos of stationary objects (default: 120)\n"
              << "  --enable-burst-mode            Enable burst mode to max out FPS when new objects enter (default: disabled)\n\n"
              << "MODEL TYPES:\n"
              << "  yolov5s    Fast model optimized for real-time detection (~65ms, 75% accuracy)\n"
              << "  yolov5l    High-accuracy model for better precision (~120ms, 85% accuracy)\n"
              << "  yolov8n    Ultra-fast nano model for embedded systems (~35ms, 70% accuracy)\n"
              << "  yolov8m    Maximum accuracy model (~150ms, 88% accuracy)\n\n"
              << "EXAMPLES:\n"
              << "  " << program_name << " --list-cameras\n"
              << "  " << program_name << " --max-fps 3 --min-confidence 0.7\n"
              << "  " << program_name << " --camera-id 1 --verbose --log-file /tmp/detection.log\n"
              << "  " << program_name << " --model-type yolov5l --max-fps 2  # High accuracy mode\n"
              << "  " << program_name << " --model-type yolov5s --processing-threads 4  # Fast parallel mode\n"
              << "  " << program_name << " --show-preview  # Development mode with real-time viewfinder\n"
              << "  " << program_name << " --max-fps 1 --frame-width 640 --frame-height 480  # Low-resource mode (32-bit)\n"
              << "  " << program_name << " --enable-streaming --streaming-port 8080  # Network streaming mode\n"
              << "SUPPORTED PLATFORMS:\n"
              << "  - Linux x86_64 (Intel Core i7, AMD Ryzen 5 3600)\n"
              << "  - Linux 386 (Intel Pentium M with 1.5GB RAM)\n"
              << "  - macOS x86_64 (Intel-based Macs)\n"
              << "  - Headless operation (no X11 required on Linux)\n"
              << "  - USB webcams (Logitech C920 recommended)\n\n"
              << "32-BIT LINUX RECOMMENDATIONS:\n"
              << "  For older hardware (Intel Pentium M, 1.5GB RAM):\n"
              << "  " << program_name << " --max-fps 1 --min-confidence 0.8 --frame-width 640 --frame-height 480 --analysis-rate-limit 0.5\n"
              << "  Consider using --detection-scale 0.5 for additional 2x speedup\n\n";
    
    // Explicit flush for macOS compatibility
    std::cout.flush();
}

bool ConfigManager::validateConfig() const {
    if (config_->max_fps <= 0 || config_->max_fps > 60) {
        std::cerr << "Invalid max_fps: " << config_->max_fps << " (must be 1-60)" << std::endl;
        return false;
    }
    
    if (config_->min_confidence < 0.0 || config_->min_confidence > 1.0) {
        std::cerr << "Invalid min_confidence: " << config_->min_confidence << " (must be 0.0-1.0)" << std::endl;
        return false;
    }
    
    if (config_->detection_scale_factor <= 0.0 || config_->detection_scale_factor > 1.0) {
        std::cerr << "Invalid detection_scale_factor: " << config_->detection_scale_factor << " (must be 0.0-1.0)" << std::endl;
        return false;
    }
    
    if (config_->min_fps_warning_threshold <= 0) {
        std::cerr << "Invalid min_fps_warning_threshold: " << config_->min_fps_warning_threshold << std::endl;
        return false;
    }
    
    if (config_->heartbeat_interval_minutes <= 0) {
        std::cerr << "Invalid heartbeat_interval_minutes: " << config_->heartbeat_interval_minutes << std::endl;
        return false;
    }
    
    if (config_->summary_interval_minutes <= 0) {
        std::cerr << "Invalid summary_interval_minutes: " << config_->summary_interval_minutes << std::endl;
        return false;
    }
    
    if (config_->camera_id < 0) {
        std::cerr << "Invalid camera_id: " << config_->camera_id << std::endl;
        return false;
    }
    
    if (config_->frame_width <= 0 || config_->frame_height <= 0) {
        std::cerr << "Invalid frame dimensions: " << config_->frame_width << "x" << config_->frame_height << std::endl;
        return false;
    }
    
    if (config_->processing_threads <= 0 || config_->processing_threads > 16) {
        std::cerr << "Invalid processing_threads: " << config_->processing_threads << " (must be 1-16)" << std::endl;
        return false;
    }
    
    if (config_->max_frame_queue_size <= 0 || config_->max_frame_queue_size > 100) {
        std::cerr << "Invalid max_frame_queue_size: " << config_->max_frame_queue_size << " (must be 1-100)" << std::endl;
        return false;
    }
    
    if (config_->analysis_rate_limit <= 0.0 || config_->analysis_rate_limit > 100.0) {
        std::cerr << "Invalid analysis_rate_limit: " << config_->analysis_rate_limit << " (must be 0.01-100)" << std::endl;
        return false;
    }
    
    if (config_->streaming_port <= 0 || config_->streaming_port > 65535) {
        std::cerr << "Invalid streaming_port: " << config_->streaming_port << " (must be 1-65535)" << std::endl;
        return false;
    }
    
    // Enable parallel processing automatically if more than 1 thread is specified
    if (config_->processing_threads > 1) {
        config_->enable_parallel_processing = true;
    }
    
    return true;
}

void ConfigManager::listCameras() const {
    std::cout << "Scanning for available cameras...\n" << std::endl;
    std::cout.flush();
    
    auto cameras = WebcamInterface::listAvailableCameras();
    
    if (cameras.empty()) {
        std::cout << "No cameras found." << std::endl;
        std::cout << "\nTroubleshooting tips:" << std::endl;
        std::cout << "- Check that your camera is connected via USB" << std::endl;
        std::cout << "- Verify camera permissions: sudo usermod -a -G video $USER" << std::endl;
        std::cout << "- Check for device files: ls -la /dev/video*" << std::endl;
        std::cout << "- Test manually: v4l2-ctl --list-devices" << std::endl;
    } else {
        std::cout << "Found " << cameras.size() << " camera(s):" << std::endl;
        std::cout << std::endl;
        
        for (const auto& camera : cameras) {
            std::cout << "  " << camera << std::endl;
        }
        
        std::cout << std::endl;
        std::cout << "To use a specific camera, use: --camera-id <ID>" << std::endl;
        std::cout << "Example: ./object_detection --camera-id 0" << std::endl;
    }
    
    // Explicit flush for macOS compatibility
    std::cout.flush();
}