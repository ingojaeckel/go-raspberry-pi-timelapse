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
        int summary_interval_minutes = 60;  // Hourly summary interval
        
        // Video capture
        int camera_id = 0;
        int frame_width = 1280;   // 720p width
        int frame_height = 720;   // 720p height
        
        // Object detection
        std::string model_path = "models/yolov5s.onnx";
        std::string config_path = "models/yolov5s.yaml";
        std::string classes_path = "models/coco.names";
        std::string model_type = "yolov5s";  // Model type: yolov5s, yolov5l, yolov8n, yolov8m
        std::string output_dir = "detections";  // Directory to store detection photos
        double detection_scale_factor = 0.5;  // Scale factor for detection (0.5 = 50% reduction, 75% fewer pixels)
        
        // Performance
        bool enable_gpu = false;
        int processing_threads = 1;
        bool enable_parallel_processing = false;
        int max_frame_queue_size = 10;
        double analysis_rate_limit = 1.0;  // Maximum images to analyze per second (default: 1)
        
        // Debug
        bool verbose = false;
        bool headless = true;
        bool show_preview = false;  // Real-time viewfinder with bounding boxes
        
        // Network streaming
        bool enable_streaming = false;  // Enable MJPEG HTTP streaming
        int streaming_port = 8080;      // Port for HTTP streaming server
        
        // Image preprocessing
        bool enable_brightness_filter = false;  // Enable high brightness filter for glass reflections
        
        // Stationary object detection
        int stationary_timeout_seconds = 120;  // Stop taking photos after objects are stationary for this many seconds
    };

    enum class ParseResult {
        SUCCESS,
        HELP_REQUESTED,
        LIST_REQUESTED,
        PARSE_ERROR
    };

    ConfigManager();
    ~ConfigManager();

    /**
     * Parse command line arguments and populate configuration
     */
    ParseResult parseArgs(int argc, char* argv[]);
    
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
    
    /**
     * List available cameras and exit
     */
    void listCameras() const;

private:
    std::unique_ptr<Config> config_;
    
    void setDefaults();
    bool parseArgument(const std::string& arg, const std::string& value);
};