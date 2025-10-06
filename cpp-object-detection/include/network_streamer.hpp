#pragma once

#include <opencv2/opencv.hpp>
#include <string>
#include <memory>
#include <atomic>
#include <thread>
#include <mutex>
#include <vector>
#include "logger.hpp"
#include "detection_model_interface.hpp"

/**
 * Network streamer for broadcasting video feed with object detection over HTTP
 * Implements MJPEG streaming compatible with web browsers and VLC
 */
class NetworkStreamer {
public:
    NetworkStreamer(std::shared_ptr<Logger> logger, int port = 8080);
    ~NetworkStreamer();

    /**
     * Initialize the streaming server
     */
    bool initialize();
    
    /**
     * Start the streaming server in a background thread
     */
    bool start();
    
    /**
     * Update the current frame to be streamed
     * This should be called from the main processing loop
     */
    void updateFrame(const cv::Mat& frame, const std::vector<Detection>& detections);
    
    /**
     * Update the current frame with statistics overlay
     */
    void updateFrameWithStats(const cv::Mat& frame, 
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
     * Stop the streaming server
     */
    void stop();
    
    /**
     * Check if server is running
     */
    bool isRunning() const { return running_; }
    
    /**
     * Get the streaming URL
     */
    std::string getStreamingUrl() const;

private:
    std::shared_ptr<Logger> logger_;
    int port_;
    std::atomic<bool> running_;
    std::atomic<bool> initialized_;
    
    // Frame management
    cv::Mat current_frame_;
    std::mutex frame_mutex_;
    
    // Server thread
    std::thread server_thread_;
    int server_socket_;
    
    // Server functions
    void serverLoop();
    void handleClient(int client_socket);
    std::vector<uchar> encodeFrameAsJpeg(const cv::Mat& frame);
    cv::Mat drawBoundingBoxes(const cv::Mat& frame, const std::vector<Detection>& detections);
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
    cv::Scalar getColorForClass(const std::string& class_name) const;
    std::string getLocalIpAddress() const;
};
