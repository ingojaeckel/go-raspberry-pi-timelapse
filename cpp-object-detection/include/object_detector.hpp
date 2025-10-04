#pragma once

#include <opencv2/opencv.hpp>
#include <vector>
#include <string>
#include <memory>
#include <deque>
#include "logger.hpp"
#include "detection_model_interface.hpp"

/**
 * Object detection orchestrator using pluggable detection models
 * This class manages object tracking and event logging while delegating
 * the actual detection to interchangeable model implementations
 */
class ObjectDetector {
public:
    struct ObjectTracker {
        std::string object_type;
        cv::Point2f center;
        cv::Point2f previous_center;  // Track previous position for movement detection
        std::deque<cv::Point2f> position_history;  // Track path of object movement
        bool was_present_last_frame;
        int frames_since_detection;
        bool is_new;  // Flag to indicate if this is a newly entered object
        
        // Constants for movement history tracking
        static constexpr size_t MAX_POSITION_HISTORY = 10;  // Keep last 10 positions
    };

    ObjectDetector(const std::string& model_path,
                  const std::string& config_path,
                  const std::string& classes_path,
                  double confidence_threshold,
                  std::shared_ptr<Logger> logger,
                  DetectionModelFactory::ModelType model_type = DetectionModelFactory::ModelType::YOLO_V5_SMALL,
                  double detection_scale_factor = 1.0);
    
    ~ObjectDetector();

    /**
     * Initialize the object detection model
     */
    bool initialize();
    
    /**
     * Detect objects in a frame
     */
    std::vector<Detection> detectObjects(const cv::Mat& frame);
    
    /**
     * Process frame and track object enter/exit events
     */
    void processFrame(const cv::Mat& frame);
    
    /**
     * Get list of target object classes we're interested in
     */
    static std::vector<std::string> getTargetClasses();
    
    /**
     * Check if a class name is a target class we're interested in
     */
    bool isTargetClass(const std::string& class_name) const;
    
    /**
     * Get current model information and performance metrics
     */
    ModelMetrics getModelMetrics() const;
    
    /**
     * Switch to a different detection model
     */
    bool switchModel(DetectionModelFactory::ModelType new_model_type);
    
    /**
     * Get available model types with their characteristics
     */
    static std::vector<ModelMetrics> getAvailableModels();

private:
    std::string model_path_;
    std::string config_path_;
    std::string classes_path_;
    double confidence_threshold_;
    double detection_scale_factor_;
    std::shared_ptr<Logger> logger_;
    DetectionModelFactory::ModelType model_type_;
    
    std::unique_ptr<IDetectionModel> detection_model_;
    std::vector<ObjectTracker> tracked_objects_;
    
    bool initialized_;
    
    void updateTrackedObjects(const std::vector<Detection>& detections);
    void logObjectEvents(const std::vector<Detection>& current_detections);
};