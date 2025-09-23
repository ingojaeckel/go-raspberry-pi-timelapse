#include "config_manager.hpp"
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

bool ConfigManager::parseArgs(int argc, char* argv[]) {
    for (int i = 1; i < argc; ++i) {
        std::string arg = argv[i];
        
        if (arg == "--help" || arg == "-h") {
            printUsage(argv[0]);
            return false;
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
        } else if (arg == "--no-headless") {
            config_->headless = false;
        } else {
            std::cerr << "Unknown argument: " << arg << std::endl;
            return false;
        }
    }
    
    return true;
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
        } else if (arg == "--processing-threads") {
            config_->processing_threads = std::stoi(value);
        } else {
            return false;
        }
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Invalid value for " << arg << ": " << value << std::endl;
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
              << "  --max-fps N                    Maximum frames per second to process (default: 5)\n"
              << "  --min-confidence N             Minimum confidence threshold (0.0-1.0, default: 0.5)\n"
              << "  --min-fps-warning N            FPS threshold for performance warnings (default: 1)\n"
              << "  --log-file FILE                Log file path (default: object_detection.log)\n"
              << "  --heartbeat-interval N         Heartbeat log interval in minutes (default: 10)\n"
              << "  --camera-id N                  Camera device ID (default: 0)\n"
              << "  --frame-width N                Frame width in pixels (default: 1280)\n"
              << "  --frame-height N               Frame height in pixels (default: 720)\n"
              << "  --model-path FILE              Path to ONNX model file (default: models/yolov5s.onnx)\n"
              << "  --config-path FILE             Path to model config file (default: models/yolov5s.yaml)\n"
              << "  --classes-path FILE            Path to class names file (default: models/coco.names)\n"
              << "  --processing-threads N         Number of processing threads (default: 1)\n"
              << "  --enable-gpu                   Enable GPU acceleration if available\n"
              << "  --no-headless                  Disable headless mode (show GUI windows)\n\n"
              << "EXAMPLES:\n"
              << "  " << program_name << " --max-fps 3 --min-confidence 0.7\n"
              << "  " << program_name << " --camera-id 1 --verbose --log-file /tmp/detection.log\n"
              << "  " << program_name << " --frame-width 640 --frame-height 480 --max-fps 10\n\n"
              << "SUPPORTED PLATFORMS:\n"
              << "  - Linux x86_64 (Intel Core i7, AMD Ryzen 5 3600)\n"
              << "  - Linux 386 (Intel Pentium M)\n"
              << "  - Headless operation (no X11 required)\n"
              << "  - USB webcams (Logitech C920 recommended)\n\n";
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
    
    if (config_->min_fps_warning_threshold <= 0) {
        std::cerr << "Invalid min_fps_warning_threshold: " << config_->min_fps_warning_threshold << std::endl;
        return false;
    }
    
    if (config_->heartbeat_interval_minutes <= 0) {
        std::cerr << "Invalid heartbeat_interval_minutes: " << config_->heartbeat_interval_minutes << std::endl;
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
    
    return true;
}