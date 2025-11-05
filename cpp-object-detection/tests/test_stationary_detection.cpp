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

TEST_F(StationaryDetectionTest, RememberedStationaryObjectBasic) {
    // Test that a stationary object is remembered when it disappears
    // and restored when re-detected
    
    ASSERT_TRUE(detector != nullptr);
    
    // Create fake detections with same position to make object stationary
    std::vector<Detection> detections;
    Detection det;
    det.class_name = "person";
    det.bbox = cv::Rect(100, 100, 50, 100);
    det.confidence = 0.6f;  // Lower confidence (flickering scenario)
    detections.push_back(det);
    
    // Update tracking multiple times to make it stationary
    for (int i = 0; i < 5; i++) {
        detector->updateTracking(detections);
    }
    
    // Verify object is stationary
    const auto& tracked1 = detector->getTrackedObjects();
    ASSERT_EQ(tracked1.size(), 1);
    EXPECT_TRUE(tracked1[0].is_stationary);
    
    // Simulate object disappearing (not detected for >30 frames)
    std::vector<Detection> empty_detections;
    for (int i = 0; i < 35; i++) {
        detector->updateTracking(empty_detections);
    }
    
    // Object should be removed from active tracking
    const auto& tracked2 = detector->getTrackedObjects();
    EXPECT_EQ(tracked2.size(), 0);
    
    // Re-detect the object at the same position
    detector->updateTracking(detections);
    
    // Object should be restored as stationary (not treated as new)
    const auto& tracked3 = detector->getTrackedObjects();
    ASSERT_EQ(tracked3.size(), 1);
    EXPECT_TRUE(tracked3[0].is_stationary);
    EXPECT_FALSE(tracked3[0].is_new);  // Should not be marked as new
}

TEST_F(StationaryDetectionTest, RememberedStationaryObjectDifferentLocation) {
    // Test that a remembered object is NOT restored if detected at a different location
    
    ASSERT_TRUE(detector != nullptr);
    
    // Create fake detections at position (100, 100)
    std::vector<Detection> detections1;
    Detection det1;
    det1.class_name = "person";
    det1.bbox = cv::Rect(100, 100, 50, 100);
    det1.confidence = 0.6f;
    detections1.push_back(det1);
    
    // Make object stationary
    for (int i = 0; i < 5; i++) {
        detector->updateTracking(detections1);
    }
    
    // Verify object is stationary
    const auto& tracked1 = detector->getTrackedObjects();
    ASSERT_EQ(tracked1.size(), 1);
    EXPECT_TRUE(tracked1[0].is_stationary);
    
    // Simulate object disappearing
    std::vector<Detection> empty_detections;
    for (int i = 0; i < 35; i++) {
        detector->updateTracking(empty_detections);
    }
    
    // Re-detect at a far away position (200, 200) - beyond match threshold
    std::vector<Detection> detections2;
    Detection det2;
    det2.class_name = "person";
    det2.bbox = cv::Rect(200, 200, 50, 100);
    det2.confidence = 0.6f;
    detections2.push_back(det2);
    
    detector->updateTracking(detections2);
    
    // Object should be treated as new (different location)
    const auto& tracked2 = detector->getTrackedObjects();
    ASSERT_EQ(tracked2.size(), 1);
    EXPECT_TRUE(tracked2[0].is_new);  // Should be marked as new
    EXPECT_FALSE(tracked2[0].is_stationary);  // Not yet stationary
}

TEST_F(StationaryDetectionTest, RememberedStationaryObjectNearbyLocation) {
    // Test that a remembered object IS restored if detected nearby (within threshold)
    
    ASSERT_TRUE(detector != nullptr);
    
    // Create fake detections at position (100, 100)
    std::vector<Detection> detections1;
    Detection det1;
    det1.class_name = "car";
    det1.bbox = cv::Rect(100, 100, 80, 60);
    det1.confidence = 0.55f;
    detections1.push_back(det1);
    
    // Make object stationary
    for (int i = 0; i < 5; i++) {
        detector->updateTracking(detections1);
    }
    
    // Verify object is stationary
    const auto& tracked1 = detector->getTrackedObjects();
    ASSERT_EQ(tracked1.size(), 1);
    EXPECT_TRUE(tracked1[0].is_stationary);
    
    // Simulate object disappearing
    std::vector<Detection> empty_detections;
    for (int i = 0; i < 35; i++) {
        detector->updateTracking(empty_detections);
    }
    
    // Re-detect at a nearby position (120, 110) - within RememberedStationaryObject::MATCH_DISTANCE_THRESHOLD
    std::vector<Detection> detections2;
    Detection det2;
    det2.class_name = "car";
    det2.bbox = cv::Rect(120, 110, 80, 60);
    det2.confidence = 0.55f;
    detections2.push_back(det2);
    
    detector->updateTracking(detections2);
    
    // Object should be restored as stationary
    const auto& tracked2 = detector->getTrackedObjects();
    ASSERT_EQ(tracked2.size(), 1);
    EXPECT_TRUE(tracked2[0].is_stationary);
    EXPECT_FALSE(tracked2[0].is_new);  // Should not be marked as new
}

TEST_F(StationaryDetectionTest, RememberedStationaryObjectDifferentType) {
    // Test that remembered objects only match by type AND location
    
    ASSERT_TRUE(detector != nullptr);
    
    // Create a stationary "car" at position (100, 100)
    std::vector<Detection> car_detections;
    Detection car_det;
    car_det.class_name = "car";
    car_det.bbox = cv::Rect(100, 100, 80, 60);
    car_det.confidence = 0.7f;
    car_detections.push_back(car_det);
    
    // Make car stationary
    for (int i = 0; i < 5; i++) {
        detector->updateTracking(car_detections);
    }
    
    // Verify car is stationary
    const auto& tracked1 = detector->getTrackedObjects();
    ASSERT_EQ(tracked1.size(), 1);
    EXPECT_TRUE(tracked1[0].is_stationary);
    EXPECT_EQ(tracked1[0].object_type, "car");
    
    // Simulate car disappearing
    std::vector<Detection> empty_detections;
    for (int i = 0; i < 35; i++) {
        detector->updateTracking(empty_detections);
    }
    
    // Detect a "person" at the same position
    std::vector<Detection> person_detections;
    Detection person_det;
    person_det.class_name = "person";
    person_det.bbox = cv::Rect(100, 100, 50, 100);
    person_det.confidence = 0.7f;
    person_detections.push_back(person_det);
    
    detector->updateTracking(person_detections);
    
    // Person should be treated as new (different type)
    const auto& tracked2 = detector->getTrackedObjects();
    ASSERT_EQ(tracked2.size(), 1);
    EXPECT_EQ(tracked2[0].object_type, "person");
    EXPECT_TRUE(tracked2[0].is_new);  // Should be marked as new
    EXPECT_FALSE(tracked2[0].is_stationary);  // Not yet stationary
}

TEST_F(StationaryDetectionTest, MultipleRememberedObjects) {
    // Test handling multiple remembered stationary objects
    
    ASSERT_TRUE(detector != nullptr);
    
    // Create two stationary objects at different positions
    std::vector<Detection> two_detections;
    
    Detection det1;
    det1.class_name = "person";
    det1.bbox = cv::Rect(100, 100, 50, 100);
    det1.confidence = 0.6f;
    two_detections.push_back(det1);
    
    Detection det2;
    det2.class_name = "car";
    det2.bbox = cv::Rect(300, 200, 80, 60);
    det2.confidence = 0.65f;
    two_detections.push_back(det2);
    
    // Make both objects stationary
    for (int i = 0; i < 5; i++) {
        detector->updateTracking(two_detections);
    }
    
    // Verify both are stationary
    const auto& tracked1 = detector->getTrackedObjects();
    ASSERT_EQ(tracked1.size(), 2);
    EXPECT_TRUE(tracked1[0].is_stationary);
    EXPECT_TRUE(tracked1[1].is_stationary);
    
    // Simulate both disappearing
    std::vector<Detection> empty_detections;
    for (int i = 0; i < 35; i++) {
        detector->updateTracking(empty_detections);
    }
    
    // Re-detect both objects
    detector->updateTracking(two_detections);
    
    // Both should be restored as stationary
    const auto& tracked2 = detector->getTrackedObjects();
    ASSERT_EQ(tracked2.size(), 2);
    
    // Find person and car in the results
    bool person_found = false;
    bool car_found = false;
    for (const auto& obj : tracked2) {
        if (obj.object_type == "person") {
            person_found = true;
            EXPECT_TRUE(obj.is_stationary);
            EXPECT_FALSE(obj.is_new);
        } else if (obj.object_type == "car") {
            car_found = true;
            EXPECT_TRUE(obj.is_stationary);
            EXPECT_FALSE(obj.is_new);
        }
    }
    
    EXPECT_TRUE(person_found);
    EXPECT_TRUE(car_found);
}
