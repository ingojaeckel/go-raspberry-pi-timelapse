#pragma once

#include <string>
#include <memory>

/**
 * Configuration manager for the object detection application.
 * Handles command-line arguments and provides default values.
 */
class ConfigManager {
public:
    struct Config {
        // Frame processing
        int max_fps = 5;
        double min_confidence = 0.5;
        int min_fps_warning_threshold = 1;
        
        // Logging
        std::string log_file = "object_detection.log";
        int heartbeat_interval_minutes = 10;
        
        // Video capture
        int camera_id = 0;
        int frame_width = 1280;   // 720p width
        int frame_height = 720;   // 720p height
        
        // Object detection
        std::string model_path = "models/yolov5s.onnx";
        std::string config_path = "models/yolov5s.yaml";
        std::string classes_path = "models/coco.names";
        
        // Performance
        bool enable_gpu = false;
        int processing_threads = 1;
        
        // Debug
        bool verbose = false;
        bool headless = true;
    };

    ConfigManager();
    ~ConfigManager();

    /**
     * Parse command line arguments and populate configuration
     */
    bool parseArgs(int argc, char* argv[]);
    
    /**
     * Get the current configuration
     */
    const Config& getConfig() const;
    
    /**
     * Print usage information
     */
    void printUsage(const std::string& program_name) const;
    
    /**
     * Validate configuration values
     */
    bool validateConfig() const;

private:
    std::unique_ptr<Config> config_;
    
    void setDefaults();
    bool parseArgument(const std::string& arg, const std::string& value);
};