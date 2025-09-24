#pragma once

#include <opencv2/opencv.hpp>
#include <memory>
#include <vector>
#include "logger.hpp"

/**
 * Webcam interface for capturing frames from USB cameras
 */
class WebcamInterface {
public:
    WebcamInterface(int camera_id, int width, int height, 
                   std::shared_ptr<Logger> logger);
    ~WebcamInterface();

    /**
     * Initialize the camera connection
     */
    bool initialize();
    
    /**
     * Capture a frame from the camera
     * Returns true if frame was captured successfully
     */
    bool captureFrame(cv::Mat& frame);
    
    /**
     * Check if camera is initialized and ready
     */
    bool isReady() const;
    
    /**
     * Get camera information
     */
    std::string getCameraInfo() const;
    
    /**
     * Release camera resources
     */
    void release();
    
    /**
     * List all available cameras
     * Returns a vector of camera information strings
     */
    static std::vector<std::string> listAvailableCameras();

private:
    int camera_id_;
    int width_;
    int height_;
    std::shared_ptr<Logger> logger_;
    std::unique_ptr<cv::VideoCapture> capture_;
    bool initialized_;
    
    bool testCameraCapabilities();
    void setCameraProperties();
};