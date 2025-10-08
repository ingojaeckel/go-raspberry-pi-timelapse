#include <gtest/gtest.h>
#include <fstream>
#include <thread>
#include <chrono>
#include "notification_manager.hpp"
#include "logger.hpp"

// Check if filesystem is available
#if __has_include(<filesystem>)
    #include <filesystem>
    namespace fs = std::filesystem;
#else
    // Fallback for older systems
    #include <cstdio>
    namespace fs {
        inline bool exists(const std::string& path) {
            std::ifstream f(path.c_str());
            return f.good();
        }
        inline bool remove(const std::string& path) {
            return std::remove(path.c_str()) == 0;
        }
    }
#endif

class NotificationManagerTest : public ::testing::Test {
protected:
    void SetUp() override {
        test_log_file = "/tmp/test_notification.log";
        test_notification_file = "/tmp/test_notifications.json";
        
        // Clean up existing files
        fs::remove(test_log_file);
        fs::remove(test_notification_file);
        
        logger = std::make_shared<Logger>(test_log_file, false);
    }

    void TearDown() override {
        // Clean up test files
        fs::remove(test_log_file);
        fs::remove(test_notification_file);
    }

    std::string test_log_file;
    std::string test_notification_file;
    std::shared_ptr<Logger> logger;
};

TEST_F(NotificationManagerTest, CreateNotificationManager) {
    NotificationManager::NotificationConfig config;
    config.enable_stdio_notification = true;
    
    auto manager = std::make_unique<NotificationManager>(logger, config);
    EXPECT_TRUE(manager->initialize());
}

TEST_F(NotificationManagerTest, IsEnabledWhenNoNotificationsConfigured) {
    NotificationManager::NotificationConfig config;
    // All notifications disabled by default
    
    auto manager = std::make_unique<NotificationManager>(logger, config);
    EXPECT_FALSE(manager->isEnabled());
}

TEST_F(NotificationManagerTest, IsEnabledWhenStdioEnabled) {
    NotificationManager::NotificationConfig config;
    config.enable_stdio_notification = true;
    
    auto manager = std::make_unique<NotificationManager>(logger, config);
    EXPECT_TRUE(manager->isEnabled());
}

TEST_F(NotificationManagerTest, FileNotificationCreatesFile) {
    NotificationManager::NotificationConfig config;
    config.enable_file_notification = true;
    config.notification_file_path = test_notification_file;
    
    auto manager = std::make_unique<NotificationManager>(logger, config);
    EXPECT_TRUE(manager->initialize());
    
    // Create test notification data
    NotificationManager::NotificationData data;
    data.object_type = "person";
    data.x = 100.0f;
    data.y = 200.0f;
    data.confidence = 0.95;
    data.timestamp = std::chrono::system_clock::now();
    data.current_fps = 5.0;
    data.avg_processing_time_ms = 150.0;
    data.total_objects_detected = 1;
    data.total_images_saved = 0;
    data.brightness_filter_active = false;
    data.gpu_enabled = false;
    data.burst_mode_enabled = false;
    
    manager->notifyNewObject(data);
    
    // Wait a bit for file write
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    
    // Check that file was created
    EXPECT_TRUE(fs::exists(test_notification_file));
    
    // Read and verify content
    std::ifstream file(test_notification_file);
    EXPECT_TRUE(file.good());
    
    std::string content((std::istreambuf_iterator<char>(file)),
                        std::istreambuf_iterator<char>());
    
    // Check that JSON contains expected fields
    EXPECT_TRUE(content.find("\"event\":\"new_object_detected\"") != std::string::npos);
    EXPECT_TRUE(content.find("\"type\":\"person\"") != std::string::npos);
    EXPECT_TRUE(content.find("\"confidence\":0.95") != std::string::npos);
}

TEST_F(NotificationManagerTest, StdioNotificationDoesNotThrow) {
    NotificationManager::NotificationConfig config;
    config.enable_stdio_notification = true;
    
    auto manager = std::make_unique<NotificationManager>(logger, config);
    EXPECT_TRUE(manager->initialize());
    
    // Create test notification data
    NotificationManager::NotificationData data;
    data.object_type = "cat";
    data.x = 50.0f;
    data.y = 75.0f;
    data.confidence = 0.85;
    data.timestamp = std::chrono::system_clock::now();
    data.current_fps = 3.0;
    data.avg_processing_time_ms = 200.0;
    data.total_objects_detected = 5;
    data.total_images_saved = 2;
    data.brightness_filter_active = false;
    data.gpu_enabled = false;
    data.burst_mode_enabled = false;
    
    // Should not throw
    EXPECT_NO_THROW(manager->notifyNewObject(data));
}

TEST_F(NotificationManagerTest, NotificationWithDetections) {
    NotificationManager::NotificationConfig config;
    config.enable_file_notification = true;
    config.notification_file_path = test_notification_file;
    
    auto manager = std::make_unique<NotificationManager>(logger, config);
    EXPECT_TRUE(manager->initialize());
    
    // Create test notification data with detections
    NotificationManager::NotificationData data;
    data.object_type = "dog";
    data.x = 150.0f;
    data.y = 250.0f;
    data.confidence = 0.92;
    data.timestamp = std::chrono::system_clock::now();
    data.current_fps = 4.5;
    data.avg_processing_time_ms = 180.0;
    data.total_objects_detected = 10;
    data.total_images_saved = 5;
    data.brightness_filter_active = true;
    data.gpu_enabled = true;
    data.burst_mode_enabled = true;
    
    // Add some detections
    Detection det1;
    det1.class_name = "dog";
    det1.confidence = 0.92;
    det1.bbox = cv::Rect(100, 200, 50, 75);
    data.all_detections.push_back(det1);
    
    Detection det2;
    det2.class_name = "person";
    det2.confidence = 0.88;
    det2.bbox = cv::Rect(300, 150, 80, 120);
    data.all_detections.push_back(det2);
    
    // Add top objects
    data.top_objects.push_back({"dog", 5});
    data.top_objects.push_back({"person", 3});
    
    manager->notifyNewObject(data);
    
    // Wait a bit for file write
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    
    // Read and verify content
    std::ifstream file(test_notification_file);
    std::string content((std::istreambuf_iterator<char>(file)),
                        std::istreambuf_iterator<char>());
    
    // Check detections are included
    EXPECT_TRUE(content.find("\"all_detections\":[") != std::string::npos);
    EXPECT_TRUE(content.find("\"class\":\"dog\"") != std::string::npos);
    EXPECT_TRUE(content.find("\"class\":\"person\"") != std::string::npos);
    
    // Check status info
    EXPECT_TRUE(content.find("\"fps\":4.50") != std::string::npos);
    EXPECT_TRUE(content.find("\"brightness_filter_active\":true") != std::string::npos);
    EXPECT_TRUE(content.find("\"gpu_enabled\":true") != std::string::npos);
    EXPECT_TRUE(content.find("\"burst_mode_enabled\":true") != std::string::npos);
    
    // Check top objects
    EXPECT_TRUE(content.find("\"top_objects\":[") != std::string::npos);
}

TEST_F(NotificationManagerTest, StopNotificationManager) {
    NotificationManager::NotificationConfig config;
    config.enable_stdio_notification = true;
    
    auto manager = std::make_unique<NotificationManager>(logger, config);
    EXPECT_TRUE(manager->initialize());
    
    // Should stop without error
    EXPECT_NO_THROW(manager->stop());
}
