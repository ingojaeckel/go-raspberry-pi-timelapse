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
     * Display a frame with bounding boxes and debug information
     */
    void showFrameWithStats(const cv::Mat& frame, 
                           const std::vector<Detection>& detections,
                           double current_fps,
                           double avg_processing_time_ms,
                           int total_objects_detected,
                           int total_images_saved,
                           const std::chrono::steady_clock::time_point& start_time,
                           const std::vector<std::pair<std::string, int>>& top_objects,
                           int camera_width,
                           int camera_height,
                           int camera_id,
                           const std::string& camera_name,
                           int detection_width,
                           int detection_height,
                           bool brightness_filter_active = false,
                           bool gpu_enabled = false,
                           bool burst_mode_enabled = false,
                           bool night_mode_active = false,
                           const cv::Mat& preprocessed_frame = cv::Mat());
    
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
    bool show_debug_info_;
    
    // Draw bounding boxes on frame
    cv::Mat drawBoundingBoxes(const cv::Mat& frame, const std::vector<Detection>& detections);
    
    // Draw debug information overlay
    void drawDebugInfo(cv::Mat& frame,
                      double current_fps,
                      double avg_processing_time_ms,
                      int total_objects_detected,
                      int total_images_saved,
                      const std::chrono::steady_clock::time_point& start_time,
                      const std::vector<std::pair<std::string, int>>& top_objects,
                      int camera_width,
                      int camera_height,
                      int camera_id,
                      const std::string& camera_name,
                      int detection_width,
                      int detection_height,
                      bool brightness_filter_active = false,
                      bool gpu_enabled = false,
                      bool burst_mode_enabled = false,
                      bool night_mode_active = false);
    
    // Get color for different object classes
    cv::Scalar getColorForClass(const std::string& class_name) const;
};
