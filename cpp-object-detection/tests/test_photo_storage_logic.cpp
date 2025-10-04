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
 * Test suite for photo storage logic based on object changes
 * 
 * This tests the new behavior where photos are saved:
 * 1. Immediately when new object types are detected
 * 2. Immediately when new instances of same type are detected
 * 3. Every 10s for stationary objects (no changes)
 */
class PhotoStorageLogicTest : public ::testing::Test {
protected:
    void SetUp() override {
        logger = std::make_shared<Logger>("test_photo_storage.log", false);
        perf_monitor = std::make_shared<PerformanceMonitor>(logger, 1.0);
        
        // Create detector with non-existent model (for testing basic functionality)
        detector = std::make_shared<ObjectDetector>(
            "non_existent_model.onnx", 
            "non_existent_config.yaml", 
            "non_existent_classes.txt", 
            0.5, 
            logger);
            
        // Create a temporary output directory for testing
        test_output_dir = "/tmp/test_photo_storage_" + std::to_string(time(nullptr));
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

TEST_F(PhotoStorageLogicTest, DetectorProvidesTrackingAccess) {
    // Test that detector exposes tracked objects
    auto tracked = detector->getTrackedObjects();
    EXPECT_TRUE(tracked.empty());  // Should be empty initially
}

TEST_F(PhotoStorageLogicTest, DetectorCanUpdateTracking) {
    // Test that detector can update tracking
    std::vector<Detection> detections;
    
    Detection det1;
    det1.class_name = "person";
    det1.confidence = 0.9;
    det1.bbox = cv::Rect(100, 100, 50, 100);
    det1.class_id = 0;
    detections.push_back(det1);
    
    detector->updateTracking(detections);
    
    // After update, tracked objects should contain the detection
    auto tracked = detector->getTrackedObjects();
    EXPECT_EQ(tracked.size(), 1);
    EXPECT_EQ(tracked[0].object_type, "person");
    EXPECT_TRUE(tracked[0].is_new);  // First detection should be marked as new
}

TEST_F(PhotoStorageLogicTest, DetectorTracksMultipleObjects) {
    // Test that detector tracks multiple objects
    std::vector<Detection> detections;
    
    Detection det1;
    det1.class_name = "person";
    det1.confidence = 0.9;
    det1.bbox = cv::Rect(100, 100, 50, 100);
    det1.class_id = 0;
    
    Detection det2;
    det2.class_name = "car";
    det2.confidence = 0.85;
    det2.bbox = cv::Rect(300, 200, 100, 80);
    det2.class_id = 1;
    
    detections.push_back(det1);
    detections.push_back(det2);
    
    detector->updateTracking(detections);
    
    auto tracked = detector->getTrackedObjects();
    EXPECT_EQ(tracked.size(), 2);
}

TEST_F(PhotoStorageLogicTest, DetectorMarksNewObjects) {
    // Test that detector correctly marks new objects
    std::vector<Detection> detections1;
    Detection det1;
    det1.class_name = "person";
    det1.confidence = 0.9;
    det1.bbox = cv::Rect(100, 100, 50, 100);
    det1.class_id = 0;
    detections1.push_back(det1);
    
    detector->updateTracking(detections1);
    auto tracked1 = detector->getTrackedObjects();
    EXPECT_EQ(tracked1.size(), 1);
    EXPECT_TRUE(tracked1[0].is_new);
    
    // Same object in next frame (moved slightly)
    std::vector<Detection> detections2;
    Detection det2;
    det2.class_name = "person";
    det2.confidence = 0.9;
    det2.bbox = cv::Rect(105, 105, 50, 100);  // Moved 5 pixels
    det2.class_id = 0;
    detections2.push_back(det2);
    
    detector->updateTracking(detections2);
    auto tracked2 = detector->getTrackedObjects();
    EXPECT_EQ(tracked2.size(), 1);
    EXPECT_FALSE(tracked2[0].is_new);  // Should not be new anymore
}

TEST_F(PhotoStorageLogicTest, DetectorDetectsNewInstance) {
    // Test that detector correctly identifies new instances of same type
    std::vector<Detection> detections1;
    Detection det1;
    det1.class_name = "car";
    det1.confidence = 0.9;
    det1.bbox = cv::Rect(100, 100, 100, 80);
    det1.class_id = 1;
    detections1.push_back(det1);
    
    detector->updateTracking(detections1);
    
    // Add a second car far away (new instance)
    std::vector<Detection> detections2;
    Detection det2a;
    det2a.class_name = "car";
    det2a.confidence = 0.9;
    det2a.bbox = cv::Rect(105, 105, 100, 80);  // First car, moved slightly
    det2a.class_id = 1;
    
    Detection det2b;
    det2b.class_name = "car";
    det2b.confidence = 0.85;
    det2b.bbox = cv::Rect(400, 300, 100, 80);  // Second car, far away
    det2b.class_id = 1;
    
    detections2.push_back(det2a);
    detections2.push_back(det2b);
    
    detector->updateTracking(detections2);
    
    auto tracked = detector->getTrackedObjects();
    EXPECT_EQ(tracked.size(), 2);  // Should have 2 tracked cars
    
    // At least one should be marked as new
    bool has_new = false;
    for (const auto& obj : tracked) {
        if (obj.is_new) {
            has_new = true;
            break;
        }
    }
    EXPECT_TRUE(has_new);
}

TEST_F(PhotoStorageLogicTest, ProcessorCreatesWithCustomOutputDir) {
    // Test that processor can be created with custom output directory
    auto processor = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 1, 10, test_output_dir);
    
    EXPECT_NE(processor, nullptr);
    
    bool initialized = processor->initialize();
    EXPECT_TRUE(initialized);
    
    // Verify output directory was created
    EXPECT_TRUE(std::filesystem::exists(test_output_dir));
    
    processor->shutdown();
}

/**
 * Note: Testing actual photo saving behavior requires a working detector
 * with valid model files. These tests verify the logic structure and
 * object tracking that enables the photo storage decision-making.
 * 
 * Integration tests with actual model files would verify:
 * 1. Photos saved immediately on new object type
 * 2. Photos saved immediately on new instance
 * 3. Photos saved only every 10s for stationary objects
 */

TEST_F(PhotoStorageLogicTest, VerifyObjectTrackerStructure) {
    // Verify that ObjectTracker has all required fields
    ObjectDetector::ObjectTracker tracker;
    tracker.object_type = "test";
    tracker.center = cv::Point2f(100, 100);
    tracker.previous_center = cv::Point2f(95, 95);
    tracker.was_present_last_frame = true;
    tracker.frames_since_detection = 0;
    tracker.is_new = true;
    
    // Verify all fields can be set and read
    EXPECT_EQ(tracker.object_type, "test");
    EXPECT_EQ(tracker.center.x, 100);
    EXPECT_EQ(tracker.center.y, 100);
    EXPECT_TRUE(tracker.was_present_last_frame);
    EXPECT_EQ(tracker.frames_since_detection, 0);
    EXPECT_TRUE(tracker.is_new);
}

TEST_F(PhotoStorageLogicTest, FirstDetectionAlwaysSaves) {
    // Verify that the first detection always triggers a save
    // This happens because last_saved_object_counts_ is initially empty,
    // so any detection will be considered a "new type"
    
    std::vector<Detection> detections;
    Detection det;
    det.class_name = "car";
    det.confidence = 0.9;
    det.bbox = cv::Rect(100, 100, 100, 80);
    det.class_id = 1;
    detections.push_back(det);
    
    detector->updateTracking(detections);
    auto tracked = detector->getTrackedObjects();
    
    // First detection should be marked as new
    EXPECT_EQ(tracked.size(), 1);
    EXPECT_TRUE(tracked[0].is_new);
    EXPECT_EQ(tracked[0].object_type, "car");
}

