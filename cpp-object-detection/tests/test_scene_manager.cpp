#include <gtest/gtest.h>
#include "scene_manager.hpp"
#include "logger.hpp"
#include <filesystem>

class SceneManagerTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Use a temporary database file for testing
        test_db_path_ = "/tmp/test_scenes.db";
        
        // Remove any existing test database
        std::filesystem::remove(test_db_path_);
        
        // Create logger for testing (with verbose output to /tmp)
        logger_ = std::make_shared<Logger>("/tmp/scene_manager_test.log", true);
        
        // Default scene match config
        config_ = SceneMatchConfig();
        config_.min_observation_seconds = 2;  // Short observation time for testing
        
        // Create scene manager
        scene_manager_ = std::make_unique<SceneManager>(test_db_path_, logger_, config_);
    }
    
    void TearDown() override {
        scene_manager_.reset();
        
        // Clean up test database
        std::filesystem::remove(test_db_path_);
        std::filesystem::remove("/tmp/scene_manager_test.log");
    }
    
    std::string test_db_path_;
    std::shared_ptr<Logger> logger_;
    SceneMatchConfig config_;
    std::unique_ptr<SceneManager> scene_manager_;
};

TEST_F(SceneManagerTest, Initialization) {
    EXPECT_TRUE(scene_manager_->initialize());
    
    // Verify database file was created
    EXPECT_TRUE(std::filesystem::exists(test_db_path_));
}

TEST_F(SceneManagerTest, CreateAndPersistScene) {
    ASSERT_TRUE(scene_manager_->initialize());
    
    // Create mock tracked objects (stationary objects)
    std::vector<ObjectDetector::ObjectTracker> tracked_objects;
    
    ObjectDetector::ObjectTracker obj1;
    obj1.object_type = "person";
    obj1.center = cv::Point2f(100, 100);
    obj1.is_stationary = true;
    obj1.was_present_last_frame = true;
    tracked_objects.push_back(obj1);
    
    ObjectDetector::ObjectTracker obj2;
    obj2.object_type = "car";
    obj2.center = cv::Point2f(300, 200);
    obj2.is_stationary = true;
    obj2.was_present_last_frame = true;
    tracked_objects.push_back(obj2);
    
    // Create a dummy frame
    cv::Mat frame(480, 640, CV_8UC3, cv::Scalar(128, 128, 128));
    
    // Update observation
    scene_manager_->updateObservation(tracked_objects, frame);
    
    // Initially should not be ready (need to wait for observation time)
    EXPECT_FALSE(scene_manager_->isReadyToAnalyzeScene());
    
    // Wait for observation time
    std::this_thread::sleep_for(std::chrono::seconds(3));
    
    // Update again to refresh timestamp
    scene_manager_->updateObservation(tracked_objects, frame);
    
    // Now should be ready
    EXPECT_TRUE(scene_manager_->isReadyToAnalyzeScene());
    
    // Analyze and persist scene
    int scene_id = scene_manager_->analyzeAndPersistScene();
    EXPECT_GT(scene_id, 0);
    
    // Verify scene was persisted
    auto scenes = scene_manager_->getAllScenes();
    EXPECT_EQ(scenes.size(), 1);
    EXPECT_EQ(scenes[0].id, scene_id);
    EXPECT_EQ(scenes[0].objects.size(), 2);
}

TEST_F(SceneManagerTest, SceneRecognition) {
    ASSERT_TRUE(scene_manager_->initialize());
    
    // Create first scene
    std::vector<ObjectDetector::ObjectTracker> tracked_objects1;
    
    ObjectDetector::ObjectTracker obj1;
    obj1.object_type = "person";
    obj1.center = cv::Point2f(100, 100);
    obj1.is_stationary = true;
    obj1.was_present_last_frame = true;
    tracked_objects1.push_back(obj1);
    
    ObjectDetector::ObjectTracker obj2;
    obj2.object_type = "car";
    obj2.center = cv::Point2f(300, 200);
    obj2.is_stationary = true;
    obj2.was_present_last_frame = true;
    tracked_objects1.push_back(obj2);
    
    cv::Mat frame(480, 640, CV_8UC3, cv::Scalar(128, 128, 128));
    
    scene_manager_->updateObservation(tracked_objects1, frame);
    std::this_thread::sleep_for(std::chrono::seconds(3));
    scene_manager_->updateObservation(tracked_objects1, frame);
    
    int scene_id1 = scene_manager_->analyzeAndPersistScene();
    EXPECT_GT(scene_id1, 0);
    
    // Reset observation
    scene_manager_->resetObservation();
    
    // Create similar scene (slightly different positions)
    std::vector<ObjectDetector::ObjectTracker> tracked_objects2;
    
    ObjectDetector::ObjectTracker obj3;
    obj3.object_type = "person";
    obj3.center = cv::Point2f(105, 105);  // Slightly different position
    obj3.is_stationary = true;
    obj3.was_present_last_frame = true;
    tracked_objects2.push_back(obj3);
    
    ObjectDetector::ObjectTracker obj4;
    obj4.object_type = "car";
    obj4.center = cv::Point2f(305, 205);  // Slightly different position
    obj4.is_stationary = true;
    obj4.was_present_last_frame = true;
    tracked_objects2.push_back(obj4);
    
    scene_manager_->updateObservation(tracked_objects2, frame);
    std::this_thread::sleep_for(std::chrono::seconds(3));
    scene_manager_->updateObservation(tracked_objects2, frame);
    
    // Should recognize as the same scene
    int scene_id2 = scene_manager_->analyzeAndPersistScene();
    EXPECT_EQ(scene_id2, scene_id1);  // Should match the first scene
    
    // Should still have only one scene in database
    auto scenes = scene_manager_->getAllScenes();
    EXPECT_EQ(scenes.size(), 1);
}

TEST_F(SceneManagerTest, DifferentScenesNotMatching) {
    ASSERT_TRUE(scene_manager_->initialize());
    
    // Create first scene with person and car
    std::vector<ObjectDetector::ObjectTracker> tracked_objects1;
    
    ObjectDetector::ObjectTracker obj1;
    obj1.object_type = "person";
    obj1.center = cv::Point2f(100, 100);
    obj1.is_stationary = true;
    obj1.was_present_last_frame = true;
    tracked_objects1.push_back(obj1);
    
    ObjectDetector::ObjectTracker obj2;
    obj2.object_type = "car";
    obj2.center = cv::Point2f(300, 200);
    obj2.is_stationary = true;
    obj2.was_present_last_frame = true;
    tracked_objects1.push_back(obj2);
    
    cv::Mat frame(480, 640, CV_8UC3, cv::Scalar(128, 128, 128));
    
    scene_manager_->updateObservation(tracked_objects1, frame);
    std::this_thread::sleep_for(std::chrono::seconds(3));
    scene_manager_->updateObservation(tracked_objects1, frame);
    
    int scene_id1 = scene_manager_->analyzeAndPersistScene();
    EXPECT_GT(scene_id1, 0);
    
    // Reset observation
    scene_manager_->resetObservation();
    
    // Create different scene with only person (different object count)
    std::vector<ObjectDetector::ObjectTracker> tracked_objects2;
    
    ObjectDetector::ObjectTracker obj3;
    obj3.object_type = "person";
    obj3.center = cv::Point2f(100, 100);
    obj3.is_stationary = true;
    obj3.was_present_last_frame = true;
    tracked_objects2.push_back(obj3);
    
    scene_manager_->updateObservation(tracked_objects2, frame);
    std::this_thread::sleep_for(std::chrono::seconds(3));
    scene_manager_->updateObservation(tracked_objects2, frame);
    
    // Should create a new scene (different object count)
    int scene_id2 = scene_manager_->analyzeAndPersistScene();
    EXPECT_NE(scene_id2, scene_id1);  // Should be different
    EXPECT_GT(scene_id2, scene_id1);  // Should have higher ID
    
    // Should have two scenes in database
    auto scenes = scene_manager_->getAllScenes();
    EXPECT_EQ(scenes.size(), 2);
}

TEST_F(SceneManagerTest, ResetObservation) {
    ASSERT_TRUE(scene_manager_->initialize());
    
    std::vector<ObjectDetector::ObjectTracker> tracked_objects;
    ObjectDetector::ObjectTracker obj;
    obj.object_type = "person";
    obj.center = cv::Point2f(100, 100);
    obj.is_stationary = true;
    obj.was_present_last_frame = true;
    tracked_objects.push_back(obj);
    
    cv::Mat frame(480, 640, CV_8UC3, cv::Scalar(128, 128, 128));
    
    scene_manager_->updateObservation(tracked_objects, frame);
    
    // Reset should clear observation
    scene_manager_->resetObservation();
    
    // Should not be ready after reset
    EXPECT_FALSE(scene_manager_->isReadyToAnalyzeScene());
}

TEST_F(SceneManagerTest, ConfigurationDefaults) {
    ASSERT_TRUE(scene_manager_->initialize());
    
    // Test default configuration values
    EXPECT_EQ(config_.min_observation_seconds, 2);  // We set this in SetUp
    EXPECT_DOUBLE_EQ(config_.position_tolerance, 0.15);
    EXPECT_DOUBLE_EQ(config_.min_match_score, 0.7);
}
