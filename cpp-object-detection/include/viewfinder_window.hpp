#pragma once

#include <opencv2/opencv.hpp>
#include <string>
#include <memory>
#include "logger.hpp"
#include "detection_model_interface.hpp"

/**
 * Viewfinder window for real-time preview of detection with bounding boxes
 * This provides instant visual feedback during development
 */
class ViewfinderWindow {
public:
    ViewfinderWindow(std::shared_ptr<Logger> logger, 
                     const std::string& window_name = "Object Detection - Live Preview");
    ~ViewfinderWindow();

    /**
     * Initialize the window
     */
    bool initialize();
    
    /**
     * Display a frame with bounding boxes for detected objects
     */
    void showFrame(const cv::Mat& frame, const std::vector<Detection>& detections);
    
    /**
     * Check if the window should be closed (user pressed a key)
     */
    bool shouldClose();
    
    /**
     * Close and cleanup the window
     */
    void close();

private:
    std::shared_ptr<Logger> logger_;
    std::string window_name_;
    bool initialized_;
    
    // Draw bounding boxes on frame
    cv::Mat drawBoundingBoxes(const cv::Mat& frame, const std::vector<Detection>& detections);
    
    // Get color for different object classes
    cv::Scalar getColorForClass(const std::string& class_name) const;
};
