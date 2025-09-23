#pragma once

#include <opencv2/opencv.hpp>
#include <opencv2/dnn.hpp>
#include <vector>
#include <string>
#include <memory>
#include "logger.hpp"

/**
 * Object detection using YOLO or similar deep learning model
 */
class ObjectDetector {
public:
    struct Detection {
        std::string class_name;
        double confidence;
        cv::Rect bbox;
        int class_id;
    };

    struct ObjectTracker {
        std::string object_type;
        cv::Point2f center;
        bool was_present_last_frame;
        int frames_since_detection;
    };

    ObjectDetector(const std::string& model_path,
                  const std::string& config_path,
                  const std::string& classes_path,
                  double confidence_threshold,
                  std::shared_ptr<Logger> logger);
    
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

private:
    std::string model_path_;
    std::string config_path_;
    std::string classes_path_;
    double confidence_threshold_;
    std::shared_ptr<Logger> logger_;
    
    cv::dnn::Net net_;
    std::vector<std::string> class_names_;
    std::vector<ObjectTracker> tracked_objects_;
    
    bool initialized_;
    
    // Model parameters
    static constexpr int INPUT_WIDTH = 640;
    static constexpr int INPUT_HEIGHT = 640;
    static constexpr float SCALE_FACTOR = 1.0 / 255.0;
    static const cv::Scalar MEAN;
    
    bool loadClassNames();
    bool loadModel();
    std::vector<Detection> postProcess(const cv::Mat& frame, 
                                     const std::vector<cv::Mat>& outputs);
    void updateTrackedObjects(const std::vector<Detection>& detections);
    bool isTargetClass(const std::string& class_name) const;
    void logObjectEvents(const std::vector<Detection>& current_detections);
};