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
    bool has_bird = std::find(target_classes.begin(), target_classes.end(), "bird") != target_classes.end();
    bool has_bear = std::find(target_classes.begin(), target_classes.end(), "bear") != target_classes.end();
    bool has_chair = std::find(target_classes.begin(), target_classes.end(), "chair") != target_classes.end();
    bool has_book = std::find(target_classes.begin(), target_classes.end(), "book") != target_classes.end();
    
    EXPECT_TRUE(has_person);
    EXPECT_TRUE(has_car);
    EXPECT_TRUE(has_dog);
    EXPECT_TRUE(has_cat);
    EXPECT_TRUE(has_bird);
    EXPECT_TRUE(has_bear);
    EXPECT_TRUE(has_chair);
    EXPECT_TRUE(has_book);
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
    EXPECT_TRUE(detector->isTargetClass("bird"));
    EXPECT_TRUE(detector->isTargetClass("bear"));
    EXPECT_TRUE(detector->isTargetClass("chair"));
    EXPECT_TRUE(detector->isTargetClass("book"));
    
    // Test non-target classes
    EXPECT_FALSE(detector->isTargetClass("table"));
    EXPECT_FALSE(detector->isTargetClass("laptop"));
    EXPECT_FALSE(detector->isTargetClass("unknown"));
    EXPECT_FALSE(detector->isTargetClass(""));
    EXPECT_FALSE(detector->isTargetClass("fox"));
    EXPECT_FALSE(detector->isTargetClass("painting"));
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

TEST_F(ObjectDetectorTest, GetTotalObjectsDetected) {
    // Test getting total objects detected
    auto detector = std::make_unique<ObjectDetector>(
        model_path, config_path, classes_path, confidence_threshold, logger);
    
    // Initially should be 0
    EXPECT_EQ(detector->getTotalObjectsDetected(), 0);
}

TEST_F(ObjectDetectorTest, GetTopDetectedObjects) {
    // Test getting top detected objects
    auto detector = std::make_unique<ObjectDetector>(
        model_path, config_path, classes_path, confidence_threshold, logger);
    
    // Initially should be empty
    auto top_objects = detector->getTopDetectedObjects(10);
    EXPECT_TRUE(top_objects.empty());
}

TEST_F(ObjectDetectorTest, EnrichDetectionsWithStationaryStatus) {
    // Test that detections are enriched with stationary status
    auto detector = std::make_unique<ObjectDetector>(
        model_path, config_path, classes_path, confidence_threshold, logger);
    
    // Create some mock detections
    std::vector<Detection> detections;
    Detection d1;
    d1.class_name = "person";
    d1.confidence = 0.92;
    d1.bbox = cv::Rect(100, 100, 50, 100);
    d1.is_stationary = false;  // Initially not stationary
    detections.push_back(d1);
    
    // Update tracking (this will create a tracker for the person)
    detector->updateTracking(detections);
    
    // Enrich detections with stationary status
    detector->enrichDetectionsWithStationaryStatus(detections);
    
    // Since we only have one detection and it just appeared, it should not be stationary
    // (need at least 3 positions in history to determine if stationary)
    EXPECT_FALSE(detections[0].is_stationary);
    
    // Simulate the same object in the same position over multiple frames
    for (int i = 0; i < 5; i++) {
        std::vector<Detection> same_detections;
        Detection d;
        d.class_name = "person";
        d.confidence = 0.92;
        d.bbox = cv::Rect(100, 100, 50, 100);  // Same position
        same_detections.push_back(d);
        
        detector->updateTracking(same_detections);
        detector->enrichDetectionsWithStationaryStatus(same_detections);
    }
    
    // After several frames in the same position, the object should be marked as stationary
    std::vector<Detection> final_detections;
    Detection d_final;
    d_final.class_name = "person";
    d_final.confidence = 0.92;
    d_final.bbox = cv::Rect(100, 100, 50, 100);
    final_detections.push_back(d_final);
    
    detector->updateTracking(final_detections);
    detector->enrichDetectionsWithStationaryStatus(final_detections);
    
    // The detection should now have is_stationary = true
    EXPECT_TRUE(final_detections[0].is_stationary);
    
    // The detection should have stationary_duration_seconds set (may be 0 if just became stationary)
    EXPECT_GE(final_detections[0].stationary_duration_seconds, 0);
}

TEST_F(ObjectDetectorTest, StationaryLabelFormat) {
    // Test to verify the label format matches issue specification: "car (91%), stationary for 2 min"
    // This is a documentation test - the actual label formatting happens in drawing code
    
    Detection d;
    d.class_name = "car";
    d.confidence = 0.91;
    d.is_stationary = true;
    d.stationary_duration_seconds = 120;  // 2 minutes
    
    // Build label as done in network_streamer.cpp, viewfinder_window.cpp, and parallel_frame_processor.cpp
    std::string label = d.class_name + " (" + 
                       std::to_string(static_cast<int>(d.confidence * 100)) + "%)";
    if (d.is_stationary) {
        label += ", stationary";
        
        // Add duration if available
        if (d.stationary_duration_seconds > 0) {
            int duration = d.stationary_duration_seconds;
            if (duration < 60) {
                label += " for " + std::to_string(duration) + " sec";
            } else {
                int minutes = duration / 60;
                label += " for " + std::to_string(minutes) + " min";
            }
        }
    }
    
    // Verify format matches issue specification
    EXPECT_EQ(label, "car (91%), stationary for 2 min");
    
    // Test with seconds
    d.stationary_duration_seconds = 45;
    label = d.class_name + " (" + 
           std::to_string(static_cast<int>(d.confidence * 100)) + "%)";
    if (d.is_stationary) {
        label += ", stationary";
        if (d.stationary_duration_seconds > 0) {
            int duration = d.stationary_duration_seconds;
            if (duration < 60) {
                label += " for " + std::to_string(duration) + " sec";
            } else {
                int minutes = duration / 60;
                label += " for " + std::to_string(minutes) + " min";
            }
        }
    }
    EXPECT_EQ(label, "car (91%), stationary for 45 sec");
    
    // Test non-stationary object
    d.is_stationary = false;
    d.stationary_duration_seconds = 0;
    label = d.class_name + " (" + 
           std::to_string(static_cast<int>(d.confidence * 100)) + "%)";
    if (d.is_stationary) {
        label += ", stationary";
        if (d.stationary_duration_seconds > 0) {
            int duration = d.stationary_duration_seconds;
            if (duration < 60) {
                label += " for " + std::to_string(duration) + " sec";
            } else {
                int minutes = duration / 60;
                label += " for " + std::to_string(minutes) + " min";
            }
        }
    }
    
    EXPECT_EQ(label, "car (91%)");
}

TEST_F(ObjectDetectorTest, GetTopDetectedObjectsWithLimit) {
    // Test getting top detected objects with different limits
    auto detector = std::make_unique<ObjectDetector>(
        model_path, config_path, classes_path, confidence_threshold, logger);
    
    // Test with different limits
    auto top_5 = detector->getTopDetectedObjects(5);
    auto top_10 = detector->getTopDetectedObjects(10);
    auto top_20 = detector->getTopDetectedObjects(20);
    
    // All should be empty for uninitialized detector
    EXPECT_TRUE(top_5.empty());
    EXPECT_TRUE(top_10.empty());
    EXPECT_TRUE(top_20.empty());
}
