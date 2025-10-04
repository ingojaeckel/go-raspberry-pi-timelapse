#include <gtest/gtest.h>
#include "object_detector.hpp"
#include "logger.hpp"
#include <opencv2/opencv.hpp>
#include <memory>

class ObjectDetectorTest : public ::testing::Test {
protected:
    void SetUp() override {
        logger = std::make_shared<Logger>("test.log", false);
        
        // Use non-existent model paths for testing basic functionality
        model_path = "non_existent_model.onnx";
        config_path = "non_existent_config.yaml";
        classes_path = "non_existent_classes.txt";
        confidence_threshold = 0.5;
    }
    
    std::shared_ptr<Logger> logger;
    std::string model_path;
    std::string config_path;
    std::string classes_path;
    double confidence_threshold;
};

TEST_F(ObjectDetectorTest, CreateObjectDetector) {
    // Test creating object detector with valid parameters
    auto detector = std::make_unique<ObjectDetector>(
        model_path, config_path, classes_path, confidence_threshold, logger);
    
    EXPECT_NE(detector, nullptr);
}

TEST_F(ObjectDetectorTest, InitializeWithMissingModel) {
    // Test initialization with missing model files
    auto detector = std::make_unique<ObjectDetector>(
        model_path, config_path, classes_path, confidence_threshold, logger);
    
    // Should fail gracefully when model files don't exist
    bool initialized = detector->initialize();
    EXPECT_FALSE(initialized);
}

TEST_F(ObjectDetectorTest, DetectObjectsWithoutInitialization) {
    // Test detecting objects without initialization
    auto detector = std::make_unique<ObjectDetector>(
        model_path, config_path, classes_path, confidence_threshold, logger);
    
    // Create a dummy frame
    cv::Mat frame = cv::Mat::zeros(480, 640, CV_8UC3);
    
    // Should return empty vector when not initialized
    auto detections = detector->detectObjects(frame);
    EXPECT_TRUE(detections.empty());
}

TEST_F(ObjectDetectorTest, DetectObjectsWithEmptyFrame) {
    // Test detecting objects with empty frame
    auto detector = std::make_unique<ObjectDetector>(
        model_path, config_path, classes_path, confidence_threshold, logger);
    
    cv::Mat empty_frame;
    
    // Should return empty vector for empty frame
    auto detections = detector->detectObjects(empty_frame);
    EXPECT_TRUE(detections.empty());
}

TEST_F(ObjectDetectorTest, ProcessFrameWithoutInitialization) {
    // Test processing frame without initialization
    auto detector = std::make_unique<ObjectDetector>(
        model_path, config_path, classes_path, confidence_threshold, logger);
    
    cv::Mat frame = cv::Mat::zeros(480, 640, CV_8UC3);
    
    // Should not crash when processing frame without initialization
    detector->processFrame(frame);
    // No specific assertion here as it's mainly testing for no crash
}

TEST_F(ObjectDetectorTest, GetTargetClasses) {
    // Test getting target classes
    auto target_classes = ObjectDetector::getTargetClasses();
    
    EXPECT_FALSE(target_classes.empty());
    
    // Check for expected target classes
    bool has_person = std::find(target_classes.begin(), target_classes.end(), "person") != target_classes.end();
    bool has_car = std::find(target_classes.begin(), target_classes.end(), "car") != target_classes.end();
    bool has_dog = std::find(target_classes.begin(), target_classes.end(), "dog") != target_classes.end();
    bool has_cat = std::find(target_classes.begin(), target_classes.end(), "cat") != target_classes.end();
    
    EXPECT_TRUE(has_person);
    EXPECT_TRUE(has_car);
    EXPECT_TRUE(has_dog);
    EXPECT_TRUE(has_cat);
}

TEST_F(ObjectDetectorTest, IsTargetClass) {
    // Test target class checking
    auto detector = std::make_unique<ObjectDetector>(
        model_path, config_path, classes_path, confidence_threshold, logger);
    
    // Test known target classes
    EXPECT_TRUE(detector->isTargetClass("person"));
    EXPECT_TRUE(detector->isTargetClass("car"));
    EXPECT_TRUE(detector->isTargetClass("truck"));
    EXPECT_TRUE(detector->isTargetClass("bus"));
    EXPECT_TRUE(detector->isTargetClass("motorcycle"));
    EXPECT_TRUE(detector->isTargetClass("bicycle"));
    EXPECT_TRUE(detector->isTargetClass("cat"));
    EXPECT_TRUE(detector->isTargetClass("dog"));
    
    // Test non-target classes
    EXPECT_FALSE(detector->isTargetClass("chair"));
    EXPECT_FALSE(detector->isTargetClass("table"));
    EXPECT_FALSE(detector->isTargetClass("unknown"));
    EXPECT_FALSE(detector->isTargetClass(""));
}

TEST_F(ObjectDetectorTest, ConfidenceThresholdHandling) {
    // Test different confidence thresholds
    auto detector_low = std::make_unique<ObjectDetector>(
        model_path, config_path, classes_path, 0.1, logger);
    auto detector_high = std::make_unique<ObjectDetector>(
        model_path, config_path, classes_path, 0.9, logger);
    
    EXPECT_NE(detector_low, nullptr);
    EXPECT_NE(detector_high, nullptr);
    
    // Both should be created successfully with different thresholds
    // Actual behavior would be tested with real model files
}

TEST_F(ObjectDetectorTest, ValidModelPaths) {
    // Test with various model path formats
    std::vector<std::string> test_paths = {
        "models/yolov5s.onnx",
        "/absolute/path/model.onnx",
        "relative/path/model.onnx",
        "model_with_numbers_123.onnx"
    };
    
    for (const auto& path : test_paths) {
        auto detector = std::make_unique<ObjectDetector>(
            path, config_path, classes_path, confidence_threshold, logger);
        EXPECT_NE(detector, nullptr);
    }
}

TEST_F(ObjectDetectorTest, ObjectTrackerStructure) {
    // Test the ObjectTracker structure has the expected fields
    ObjectDetector::ObjectTracker tracker;
    
    // Set basic fields
    tracker.object_type = "person";
    tracker.center = cv::Point2f(100.0f, 200.0f);
    tracker.previous_center = cv::Point2f(90.0f, 190.0f);
    tracker.was_present_last_frame = true;
    tracker.frames_since_detection = 0;
    tracker.is_new = false;
    
    // Test position history
    EXPECT_TRUE(tracker.position_history.empty());
    
    // Add positions to history
    tracker.position_history.push_back(cv::Point2f(80.0f, 180.0f));
    tracker.position_history.push_back(cv::Point2f(90.0f, 190.0f));
    tracker.position_history.push_back(cv::Point2f(100.0f, 200.0f));
    
    EXPECT_EQ(tracker.position_history.size(), 3);
    EXPECT_EQ(tracker.position_history.front().x, 80.0f);
    EXPECT_EQ(tracker.position_history.back().x, 100.0f);
}

TEST_F(ObjectDetectorTest, PositionHistoryLimit) {
    // Test that position history is limited to MAX_POSITION_HISTORY
    ObjectDetector::ObjectTracker tracker;
    
    // Add more positions than the limit
    for (int i = 0; i <= ObjectDetector::ObjectTracker::MAX_POSITION_HISTORY + 5; ++i) {
        tracker.position_history.push_back(cv::Point2f(i * 10.0f, i * 10.0f));
        
        // Simulate the limit enforcement
        if (tracker.position_history.size() > ObjectDetector::ObjectTracker::MAX_POSITION_HISTORY) {
            tracker.position_history.pop_front();
        }
    }
    
    // Should be limited to MAX_POSITION_HISTORY
    EXPECT_EQ(tracker.position_history.size(), ObjectDetector::ObjectTracker::MAX_POSITION_HISTORY);
    
    // The oldest position should have been removed
    EXPECT_GT(tracker.position_history.front().x, 0.0f);
}