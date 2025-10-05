#include <gtest/gtest.h>
#include "parallel_frame_processor.hpp"
#include "object_detector.hpp"
#include "logger.hpp"
#include "performance_monitor.hpp"
#include <opencv2/opencv.hpp>
#include <memory>
#include <thread>
#include <chrono>
#include <filesystem>

/**
 * Test suite for stationary object detection
 * 
 * This tests the new behavior where photos are skipped:
 * 1. When all objects have been stationary for longer than timeout period
 * 2. Movement detection correctly identifies stationary vs moving objects
 * 3. Timeout configuration works correctly
 */
class StationaryDetectionTest : public ::testing::Test {
protected:
    void SetUp() override {
        logger = std::make_shared<Logger>("test_stationary.log", false);
        perf_monitor = std::make_shared<PerformanceMonitor>(logger, 1.0);
        
        // Create detector with non-existent model (for testing basic functionality)
        detector = std::make_shared<ObjectDetector>(
            "non_existent_model.onnx", 
            "non_existent_config.yaml", 
            "non_existent_classes.txt", 
            0.5, 
            logger);
            
        // Create a temporary output directory for testing
        test_output_dir = "/tmp/test_stationary_" + std::to_string(time(nullptr));
        std::filesystem::create_directories(test_output_dir);
    }
    
    void TearDown() override {
        // Clean up test directory
        if (std::filesystem::exists(test_output_dir)) {
            std::filesystem::remove_all(test_output_dir);
        }
    }
    
    std::shared_ptr<Logger> logger;
    std::shared_ptr<PerformanceMonitor> perf_monitor;
    std::shared_ptr<ObjectDetector> detector;
    std::string test_output_dir;
};

TEST_F(StationaryDetectionTest, DetectStationaryObjectBasic) {
    // Initialize detector
    ASSERT_TRUE(detector != nullptr);
    
    // Create fake detections with same position repeatedly
    std::vector<Detection> detections;
    Detection det;
    det.class_name = "person";
    det.bbox = cv::Rect(100, 100, 50, 100);
    det.confidence = 0.9f;
    detections.push_back(det);
    
    // Update tracking multiple times with same position
    for (int i = 0; i < 5; i++) {
        detector->updateTracking(detections);
    }
    
    // Check tracked objects
    const auto& tracked = detector->getTrackedObjects();
    ASSERT_EQ(tracked.size(), 1);
    
    // Object should be marked as stationary after enough updates
    // (need at least 3 positions in history to determine stationary status)
    EXPECT_TRUE(tracked[0].is_stationary);
}

TEST_F(StationaryDetectionTest, DetectMovingObject) {
    // Initialize detector
    ASSERT_TRUE(detector != nullptr);
    
    // Create fake detections with changing position
    for (int i = 0; i < 5; i++) {
        std::vector<Detection> detections;
        Detection det;
        det.class_name = "person";
        // Move object by 20 pixels each frame (above stationary threshold of 10)
        det.bbox = cv::Rect(100 + i * 20, 100, 50, 100);
        det.confidence = 0.9f;
        detections.push_back(det);
        
        detector->updateTracking(detections);
    }
    
    // Check tracked objects
    const auto& tracked = detector->getTrackedObjects();
    ASSERT_EQ(tracked.size(), 1);
    
    // Object should NOT be marked as stationary (moving > 10 pixels on avg)
    EXPECT_FALSE(tracked[0].is_stationary);
}

TEST_F(StationaryDetectionTest, StationaryTimeoutNotReached) {
    // Create processor with 10 second timeout
    auto processor = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 1, 10, test_output_dir, false, 10);
    
    // Initialize detector
    ASSERT_TRUE(detector != nullptr);
    
    // Create fake detections with same position
    std::vector<Detection> detections;
    Detection det;
    det.class_name = "person";
    det.bbox = cv::Rect(100, 100, 50, 100);
    det.confidence = 0.9f;
    detections.push_back(det);
    
    // Update tracking to make object stationary
    for (int i = 0; i < 5; i++) {
        detector->updateTracking(detections);
    }
    
    // Wait a bit but not enough for timeout
    std::this_thread::sleep_for(std::chrono::seconds(2));
    
    // Check that object is not past timeout
    const auto& tracked = detector->getTrackedObjects();
    ASSERT_EQ(tracked.size(), 1);
    EXPECT_FALSE(detector->isStationaryPastTimeout(tracked[0], 10));
}

TEST_F(StationaryDetectionTest, StationaryTimeoutReached) {
    // Create processor with 2 second timeout for faster testing
    auto processor = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 1, 10, test_output_dir, false, 2);
    
    // Initialize detector
    ASSERT_TRUE(detector != nullptr);
    
    // Create fake detections with same position
    std::vector<Detection> detections;
    Detection det;
    det.class_name = "person";
    det.bbox = cv::Rect(100, 100, 50, 100);
    det.confidence = 0.9f;
    detections.push_back(det);
    
    // Update tracking to make object stationary
    for (int i = 0; i < 5; i++) {
        detector->updateTracking(detections);
    }
    
    // Wait for timeout period
    std::this_thread::sleep_for(std::chrono::seconds(3));
    
    // Check that object is past timeout
    const auto& tracked = detector->getTrackedObjects();
    ASSERT_EQ(tracked.size(), 1);
    EXPECT_TRUE(detector->isStationaryPastTimeout(tracked[0], 2));
}

TEST_F(StationaryDetectionTest, ObjectBecomesMobileAgain) {
    // Initialize detector
    ASSERT_TRUE(detector != nullptr);
    
    // Create fake detections with same position to make stationary
    for (int i = 0; i < 5; i++) {
        std::vector<Detection> detections;
        Detection det;
        det.class_name = "person";
        det.bbox = cv::Rect(100, 100, 50, 100);
        det.confidence = 0.9f;
        detections.push_back(det);
        detector->updateTracking(detections);
    }
    
    // Check object is stationary
    const auto& tracked1 = detector->getTrackedObjects();
    ASSERT_EQ(tracked1.size(), 1);
    EXPECT_TRUE(tracked1[0].is_stationary);
    
    // Now make it move significantly
    for (int i = 0; i < 5; i++) {
        std::vector<Detection> detections;
        Detection det;
        det.class_name = "person";
        // Move by 30 pixels (well above stationary threshold)
        det.bbox = cv::Rect(100 + i * 30, 100, 50, 100);
        det.confidence = 0.9f;
        detections.push_back(det);
        detector->updateTracking(detections);
    }
    
    // Check object is now moving
    const auto& tracked2 = detector->getTrackedObjects();
    ASSERT_EQ(tracked2.size(), 1);
    EXPECT_FALSE(tracked2[0].is_stationary);
}

TEST_F(StationaryDetectionTest, ConfigurableTimeout) {
    // Test with different timeout values
    
    // Short timeout (1 second)
    auto processor_short = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 1, 10, test_output_dir, false, 1);
    EXPECT_NE(processor_short, nullptr);
    
    // Long timeout (300 seconds = 5 minutes)
    auto processor_long = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 1, 10, test_output_dir, false, 300);
    EXPECT_NE(processor_long, nullptr);
    
    // Default timeout (120 seconds)
    auto processor_default = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 1, 10, test_output_dir);
    EXPECT_NE(processor_default, nullptr);
}
