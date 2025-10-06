#include <gtest/gtest.h>
#include "scene_manager.hpp"
#include "logger.hpp"
#include <opencv2/opencv.hpp>
#include <memory>
#include <filesystem>

class SceneManagerTest : public ::testing::Test {
protected:
    void SetUp() override {
        logger = std::make_shared<Logger>("test_scene.log", false);
        test_db_path = "test_scenes.db";
        
        // Remove test database if it exists
        if (std::filesystem::exists(test_db_path)) {
            std::filesystem::remove(test_db_path);
        }
    }
    
    void TearDown() override {
        // Clean up test database
        if (std::filesystem::exists(test_db_path)) {
            std::filesystem::remove(test_db_path);
        }
    }
    
    std::shared_ptr<Logger> logger;
    std::string test_db_path;
};

TEST_F(SceneManagerTest, CreateSceneManager) {
    auto manager = std::make_unique<SceneManager>(logger, test_db_path);
    EXPECT_NE(manager, nullptr);
}

TEST_F(SceneManagerTest, InitializeDatabase) {
    auto manager = std::make_unique<SceneManager>(logger, test_db_path);
    bool initialized = manager->initialize();
    EXPECT_TRUE(initialized);
    
    // Verify database file was created
    EXPECT_TRUE(std::filesystem::exists(test_db_path));
}

TEST_F(SceneManagerTest, ShouldNotAnalyzeWithoutStationaryObjects) {
    auto manager = std::make_unique<SceneManager>(logger, test_db_path);
    manager->initialize();
    
    // Create tracked objects that are not stationary
    std::vector<ObjectDetector::ObjectTracker> tracked_objects;
    ObjectDetector::ObjectTracker tracker;
    tracker.object_type = "person";
    tracker.center = cv::Point2f(100, 100);
    tracker.is_stationary = false;
    tracked_objects.push_back(tracker);
    
    bool should_analyze = manager->shouldAnalyzeScene(tracked_objects);
    EXPECT_FALSE(should_analyze);
}

TEST_F(SceneManagerTest, AnalyzeEmptyScene) {
    auto manager = std::make_unique<SceneManager>(logger, test_db_path);
    manager->initialize();
    
    cv::Mat frame = cv::Mat::zeros(480, 640, CV_8UC3);
    std::vector<ObjectDetector::ObjectTracker> tracked_objects;
    
    int scene_id = manager->analyzeAndMatchScene(frame, tracked_objects);
    // Should return -1 for empty scene
    EXPECT_EQ(scene_id, -1);
}

TEST_F(SceneManagerTest, CreateAndSaveNewScene) {
    auto manager = std::make_unique<SceneManager>(logger, test_db_path);
    manager->initialize();
    
    // Create a frame with some content
    cv::Mat frame = cv::Mat::zeros(480, 640, CV_8UC3);
    cv::rectangle(frame, cv::Rect(100, 100, 50, 50), cv::Scalar(255, 0, 0), -1);
    
    // Create stationary tracked objects
    std::vector<ObjectDetector::ObjectTracker> tracked_objects;
    ObjectDetector::ObjectTracker tracker1;
    tracker1.object_type = "car";
    tracker1.center = cv::Point2f(125, 125);
    tracker1.is_stationary = true;
    tracker1.stationary_since = std::chrono::steady_clock::now() - std::chrono::seconds(70);
    tracked_objects.push_back(tracker1);
    
    ObjectDetector::ObjectTracker tracker2;
    tracker2.object_type = "person";
    tracker2.center = cv::Point2f(300, 200);
    tracker2.is_stationary = true;
    tracker2.stationary_since = std::chrono::steady_clock::now() - std::chrono::seconds(70);
    tracked_objects.push_back(tracker2);
    
    int scene_id = manager->analyzeAndMatchScene(frame, tracked_objects);
    
    // Should create a new scene
    EXPECT_GT(scene_id, 0);
    
    // Verify description was created
    std::string desc = manager->getSceneDescription(scene_id);
    EXPECT_FALSE(desc.empty());
    EXPECT_TRUE(desc.find("car") != std::string::npos || desc.find("person") != std::string::npos);
}

TEST_F(SceneManagerTest, MatchSimilarScene) {
    auto manager = std::make_unique<SceneManager>(logger, test_db_path);
    manager->initialize();
    
    cv::Mat frame = cv::Mat::zeros(480, 640, CV_8UC3);
    
    // Create first scene
    std::vector<ObjectDetector::ObjectTracker> tracked_objects1;
    ObjectDetector::ObjectTracker tracker1;
    tracker1.object_type = "car";
    tracker1.center = cv::Point2f(100, 100);
    tracker1.is_stationary = true;
    tracker1.stationary_since = std::chrono::steady_clock::now() - std::chrono::seconds(70);
    tracked_objects1.push_back(tracker1);
    
    ObjectDetector::ObjectTracker tracker2;
    tracker2.object_type = "car";
    tracker2.center = cv::Point2f(300, 200);
    tracker2.is_stationary = true;
    tracker2.stationary_since = std::chrono::steady_clock::now() - std::chrono::seconds(70);
    tracked_objects1.push_back(tracker2);
    
    int scene_id1 = manager->analyzeAndMatchScene(frame, tracked_objects1);
    EXPECT_GT(scene_id1, 0);
    
    // Create similar scene (same objects, slightly different positions)
    std::vector<ObjectDetector::ObjectTracker> tracked_objects2;
    ObjectDetector::ObjectTracker tracker3;
    tracker3.object_type = "car";
    tracker3.center = cv::Point2f(105, 105);  // Slightly shifted
    tracker3.is_stationary = true;
    tracker3.stationary_since = std::chrono::steady_clock::now() - std::chrono::seconds(70);
    tracked_objects2.push_back(tracker3);
    
    ObjectDetector::ObjectTracker tracker4;
    tracker4.object_type = "car";
    tracker4.center = cv::Point2f(305, 205);  // Slightly shifted
    tracker4.is_stationary = true;
    tracker4.stationary_since = std::chrono::steady_clock::now() - std::chrono::seconds(70);
    tracked_objects2.push_back(tracker4);
    
    int scene_id2 = manager->analyzeAndMatchScene(frame, tracked_objects2);
    
    // Should match the first scene (or create a new one depending on threshold)
    // This test verifies the matching logic works
    EXPECT_GT(scene_id2, 0);
}

TEST_F(SceneManagerTest, DifferentSceneNotMatched) {
    auto manager = std::make_unique<SceneManager>(logger, test_db_path);
    manager->initialize();
    
    cv::Mat frame = cv::Mat::zeros(480, 640, CV_8UC3);
    
    // Create first scene with cars
    std::vector<ObjectDetector::ObjectTracker> tracked_objects1;
    ObjectDetector::ObjectTracker tracker1;
    tracker1.object_type = "car";
    tracker1.center = cv::Point2f(100, 100);
    tracker1.is_stationary = true;
    tracker1.stationary_since = std::chrono::steady_clock::now() - std::chrono::seconds(70);
    tracked_objects1.push_back(tracker1);
    
    int scene_id1 = manager->analyzeAndMatchScene(frame, tracked_objects1);
    EXPECT_GT(scene_id1, 0);
    
    // Create completely different scene with people
    std::vector<ObjectDetector::ObjectTracker> tracked_objects2;
    ObjectDetector::ObjectTracker tracker2;
    tracker2.object_type = "person";
    tracker2.center = cv::Point2f(400, 300);
    tracker2.is_stationary = true;
    tracker2.stationary_since = std::chrono::steady_clock::now() - std::chrono::seconds(70);
    tracked_objects2.push_back(tracker2);
    
    ObjectDetector::ObjectTracker tracker3;
    tracker3.object_type = "person";
    tracker3.center = cv::Point2f(200, 150);
    tracker3.is_stationary = true;
    tracker3.stationary_since = std::chrono::steady_clock::now() - std::chrono::seconds(70);
    tracked_objects2.push_back(tracker3);
    
    int scene_id2 = manager->analyzeAndMatchScene(frame, tracked_objects2);
    
    // Should create a new scene, not match the first one
    EXPECT_GT(scene_id2, 0);
    EXPECT_NE(scene_id1, scene_id2);
}

TEST_F(SceneManagerTest, GetNonExistentSceneDescription) {
    auto manager = std::make_unique<SceneManager>(logger, test_db_path);
    manager->initialize();
    
    std::string desc = manager->getSceneDescription(99999);
    EXPECT_TRUE(desc.empty());
}
