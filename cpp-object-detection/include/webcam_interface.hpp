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
     * Perform camera health check and attempt recovery if needed
     * Returns true if camera is healthy or was successfully recovered
     */
    bool healthCheck();
    
    /**
     * Attempt to reconnect to the camera
     * Returns true if reconnection was successful
     */
    bool reconnect();
    
    /**
     * Keep camera active to prevent USB standby
     * Should be called periodically during long-running operation
     */
    void keepAlive();
    
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
    
    // Keep-alive tracking to prevent USB camera standby
    std::chrono::steady_clock::time_point last_keepalive_time_;
    static constexpr int KEEPALIVE_INTERVAL_SECONDS = 30;
    
    // Health check tracking
    int consecutive_failures_;
    static constexpr int MAX_CONSECUTIVE_FAILURES = 5;
    
    bool testCameraCapabilities();
    void setCameraProperties();
};