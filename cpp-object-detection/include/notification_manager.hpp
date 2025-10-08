#pragma once

#include <string>
#include <memory>
#include <vector>
#include <functional>
#include <thread>
#include <atomic>
#include <mutex>
#include <opencv2/opencv.hpp>
#include "logger.hpp"
#include "detection_model_interface.hpp"

/**
 * Notification manager for real-time alerts when new objects are detected
 * Supports multiple notification mechanisms:
 * - Webhook/Callback URL
 * - Server-Sent Events (SSE) for HTTP push
 * - File-based notifications
 * - Stdio notifications
 */
class NotificationManager {
public:
    struct NotificationConfig {
        // Webhook configuration
        bool enable_webhook = false;
        std::string webhook_url = "";
        
        // SSE configuration
        bool enable_sse = false;
        int sse_port = 8081;
        
        // File-based notification configuration
        bool enable_file_notification = false;
        std::string notification_file_path = "/tmp/object_notifications.json";
        
        // Stdio notification configuration
        bool enable_stdio_notification = false;
    };

    struct NotificationData {
        std::string object_type;
        float x;
        float y;
        double confidence;
        std::chrono::system_clock::time_point timestamp;
        cv::Mat frame_with_boxes;  // Frame with bounding boxes
        std::vector<Detection> all_detections;  // All current detections
        
        // Status information
        double current_fps;
        double avg_processing_time_ms;
        int total_objects_detected;
        int total_images_saved;
        std::vector<std::pair<std::string, int>> top_objects;
        bool brightness_filter_active;
        bool gpu_enabled;
        bool burst_mode_enabled;
    };

    NotificationManager(std::shared_ptr<Logger> logger, const NotificationConfig& config);
    ~NotificationManager();

    /**
     * Initialize notification systems
     */
    bool initialize();
    
    /**
     * Send notification about a newly detected object
     */
    void notifyNewObject(const NotificationData& data);
    
    /**
     * Stop notification systems
     */
    void stop();
    
    /**
     * Check if any notification mechanism is enabled
     */
    bool isEnabled() const;

private:
    std::shared_ptr<Logger> logger_;
    NotificationConfig config_;
    std::atomic<bool> running_;
    std::atomic<bool> initialized_;
    
    // SSE server components
    int sse_server_socket_;
    std::thread sse_server_thread_;
    std::vector<int> sse_clients_;
    std::mutex sse_clients_mutex_;
    
    // Webhook notification
    void sendWebhookNotification(const NotificationData& data);
    
    // SSE notification
    void sendSSENotification(const NotificationData& data);
    void startSSEServer();
    void handleSSEClient(int client_socket);
    void broadcastSSEMessage(const std::string& message);
    
    // File notification
    void sendFileNotification(const NotificationData& data);
    
    // Stdio notification
    void sendStdioNotification(const NotificationData& data);
    
    // Helper functions
    std::string createNotificationJSON(const NotificationData& data);
    std::string encodeImageToBase64(const cv::Mat& image);
};
