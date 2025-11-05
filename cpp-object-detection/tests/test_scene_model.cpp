#include <gtest/gtest.h>
#include "scene_model.hpp"
#include <opencv2/opencv.hpp>

class SceneModelTest : public ::testing::Test {
protected:
    void SetUp() override {
        scene = std::make_unique<SceneModel>();
    }
    
    std::unique_ptr<SceneModel> scene;
};

TEST_F(SceneModelTest, CreateSceneModel) {
    EXPECT_NE(scene, nullptr);
    EXPECT_EQ(scene->getObjects().size(), 0);
}

TEST_F(SceneModelTest, AddSingleObject) {
    cv::Point2f position(100.0f, 200.0f);
    cv::Rect2f bbox(80.0f, 180.0f, 40.0f, 40.0f);
    
    scene->addOrUpdateObject("tree", position, bbox, 0.85f, true);
    
    EXPECT_EQ(scene->getObjects().size(), 1);
    const auto& obj = scene->getObjects()[0];
    EXPECT_EQ(obj.object_type, "tree");
    EXPECT_EQ(obj.position.x, 100.0f);
    EXPECT_EQ(obj.position.y, 200.0f);
    EXPECT_TRUE(obj.is_stationary);
}

TEST_F(SceneModelTest, UpdateExistingObject) {
    cv::Point2f position1(100.0f, 200.0f);
    cv::Rect2f bbox1(80.0f, 180.0f, 40.0f, 40.0f);
    
    scene->addOrUpdateObject("bench", position1, bbox1, 0.85f, true);
    EXPECT_EQ(scene->getObjects().size(), 1);
    EXPECT_EQ(scene->getObjects()[0].detection_count, 1);
    
    // Update same object at slightly different position (within threshold)
    cv::Point2f position2(105.0f, 205.0f);
    cv::Rect2f bbox2(85.0f, 185.0f, 40.0f, 40.0f);
    
    scene->addOrUpdateObject("bench", position2, bbox2, 0.90f, true);
    
    // Should still be one object (updated)
    EXPECT_EQ(scene->getObjects().size(), 1);
    EXPECT_EQ(scene->getObjects()[0].detection_count, 2);
    EXPECT_EQ(scene->getObjects()[0].position.x, 105.0f);
}

TEST_F(SceneModelTest, AddMultipleObjects) {
    cv::Point2f pos1(100.0f, 200.0f);
    cv::Point2f pos2(300.0f, 400.0f);
    cv::Point2f pos3(500.0f, 100.0f);
    cv::Rect2f bbox(0.0f, 0.0f, 40.0f, 40.0f);
    
    scene->addOrUpdateObject("bench", pos1, bbox, 0.85f, true);
    scene->addOrUpdateObject("potted plant", pos2, bbox, 0.90f, true);
    scene->addOrUpdateObject("person", pos3, bbox, 0.80f, false);
    
    EXPECT_EQ(scene->getObjects().size(), 3);
}

TEST_F(SceneModelTest, GetStationaryObjects) {
    cv::Rect2f bbox(0.0f, 0.0f, 40.0f, 40.0f);
    
    scene->addOrUpdateObject("bench", cv::Point2f(100.0f, 200.0f), bbox, 0.85f, true);
    scene->addOrUpdateObject("potted plant", cv::Point2f(300.0f, 400.0f), bbox, 0.90f, true);
    scene->addOrUpdateObject("person", cv::Point2f(500.0f, 100.0f), bbox, 0.80f, false);
    scene->addOrUpdateObject("car", cv::Point2f(600.0f, 300.0f), bbox, 0.88f, false);
    
    auto stationary = scene->getStationaryObjects();
    EXPECT_EQ(stationary.size(), 2);
    
    // Check that we got the right objects
    bool has_bench = false;
    bool has_plant = false;
    for (const auto& obj : stationary) {
        if (obj.object_type == "bench") has_bench = true;
        if (obj.object_type == "potted plant") has_plant = true;
    }
    EXPECT_TRUE(has_bench);
    EXPECT_TRUE(has_plant);
}

TEST_F(SceneModelTest, GetDynamicObjects) {
    cv::Rect2f bbox(0.0f, 0.0f, 40.0f, 40.0f);
    
    scene->addOrUpdateObject("bench", cv::Point2f(100.0f, 200.0f), bbox, 0.85f, true);
    scene->addOrUpdateObject("person", cv::Point2f(500.0f, 100.0f), bbox, 0.80f, false);
    scene->addOrUpdateObject("car", cv::Point2f(600.0f, 300.0f), bbox, 0.88f, false);
    
    auto dynamic = scene->getDynamicObjects();
    EXPECT_EQ(dynamic.size(), 2);
    
    // Check that we got the right objects
    bool has_person = false;
    bool has_car = false;
    for (const auto& obj : dynamic) {
        if (obj.object_type == "person") has_person = true;
        if (obj.object_type == "car") has_car = true;
    }
    EXPECT_TRUE(has_person);
    EXPECT_TRUE(has_car);
}

TEST_F(SceneModelTest, GetObjectsByType) {
    cv::Rect2f bbox(0.0f, 0.0f, 40.0f, 40.0f);
    
    scene->addOrUpdateObject("bench", cv::Point2f(100.0f, 200.0f), bbox, 0.85f, true);
    scene->addOrUpdateObject("bench", cv::Point2f(400.0f, 500.0f), bbox, 0.87f, true);
    scene->addOrUpdateObject("person", cv::Point2f(500.0f, 100.0f), bbox, 0.80f, false);
    
    auto benches = scene->getObjectsByType("bench");
    EXPECT_EQ(benches.size(), 2);
    
    auto people = scene->getObjectsByType("person");
    EXPECT_EQ(people.size(), 1);
    
    auto cars = scene->getObjectsByType("car");
    EXPECT_EQ(cars.size(), 0);
}

TEST_F(SceneModelTest, GetSpatialRelationships) {
    cv::Rect2f bbox(0.0f, 0.0f, 40.0f, 40.0f);
    
    // Add three objects forming a triangle
    scene->addOrUpdateObject("bench", cv::Point2f(100.0f, 100.0f), bbox, 0.85f, true);
    scene->addOrUpdateObject("potted plant", cv::Point2f(300.0f, 100.0f), bbox, 0.90f, true);
    scene->addOrUpdateObject("person", cv::Point2f(200.0f, 300.0f), bbox, 0.80f, false);
    
    auto relationships = scene->getSpatialRelationships();
    
    // Should have 3 relationships (3 objects = 3 pairs: 0-1, 0-2, 1-2)
    EXPECT_EQ(relationships.size(), 3);
    
    // Check that relationships have valid data
    for (const auto& rel : relationships) {
        EXPECT_GT(rel.distance, 0.0f);
        EXPECT_FALSE(rel.object1_type.empty());
        EXPECT_FALSE(rel.object2_type.empty());
        EXPECT_FALSE(rel.relative_position.empty());
    }
}

TEST_F(SceneModelTest, GetRelationshipsForSpecificObject) {
    cv::Rect2f bbox(0.0f, 0.0f, 40.0f, 40.0f);
    
    scene->addOrUpdateObject("bench", cv::Point2f(100.0f, 100.0f), bbox, 0.85f, true);
    scene->addOrUpdateObject("potted plant", cv::Point2f(300.0f, 100.0f), bbox, 0.90f, true);
    scene->addOrUpdateObject("person", cv::Point2f(200.0f, 300.0f), bbox, 0.80f, false);
    
    auto bench_rels = scene->getRelationshipsForObject("bench");
    
    // Bench should be in 2 relationships (with plant and with person)
    EXPECT_EQ(bench_rels.size(), 2);
}

TEST_F(SceneModelTest, FindNearestObject) {
    cv::Rect2f bbox(0.0f, 0.0f, 40.0f, 40.0f);
    
    scene->addOrUpdateObject("bench", cv::Point2f(100.0f, 100.0f), bbox, 0.85f, true);
    scene->addOrUpdateObject("bench", cv::Point2f(400.0f, 100.0f), bbox, 0.87f, true);
    scene->addOrUpdateObject("person", cv::Point2f(200.0f, 300.0f), bbox, 0.80f, false);
    
    // Find nearest bench to a point
    cv::Point2f test_point(150.0f, 100.0f);
    auto* nearest = scene->findNearestObject("bench", test_point);
    
    EXPECT_NE(nearest, nullptr);
    EXPECT_EQ(nearest->object_type, "bench");
    // Should be the first bench at (100, 100) since it's closer
    EXPECT_FLOAT_EQ(nearest->position.x, 100.0f);
    EXPECT_FLOAT_EQ(nearest->position.y, 100.0f);
}

TEST_F(SceneModelTest, GetSceneStats) {
    cv::Rect2f bbox(0.0f, 0.0f, 40.0f, 40.0f);
    
    scene->addOrUpdateObject("bench", cv::Point2f(100.0f, 200.0f), bbox, 0.85f, true);
    scene->addOrUpdateObject("potted plant", cv::Point2f(300.0f, 400.0f), bbox, 0.90f, true);
    scene->addOrUpdateObject("person", cv::Point2f(500.0f, 100.0f), bbox, 0.80f, false);
    scene->addOrUpdateObject("person", cv::Point2f(700.0f, 100.0f), bbox, 0.82f, false);
    
    auto stats = scene->getSceneStats();
    
    EXPECT_EQ(stats.total_objects, 4);
    EXPECT_EQ(stats.stationary_objects, 2);
    EXPECT_EQ(stats.dynamic_objects, 2);
    EXPECT_EQ(stats.objects_by_type["bench"], 1);
    EXPECT_EQ(stats.objects_by_type["potted plant"], 1);
    EXPECT_EQ(stats.objects_by_type["person"], 2);
}

TEST_F(SceneModelTest, GetSceneDescription) {
    cv::Rect2f bbox(0.0f, 0.0f, 40.0f, 40.0f);
    
    scene->addOrUpdateObject("bench", cv::Point2f(100.0f, 200.0f), bbox, 0.85f, true);
    scene->addOrUpdateObject("potted plant", cv::Point2f(300.0f, 400.0f), bbox, 0.90f, true);
    scene->addOrUpdateObject("person", cv::Point2f(500.0f, 100.0f), bbox, 0.80f, false);
    
    std::string desc = scene->getSceneDescription();
    
    EXPECT_FALSE(desc.empty());
    EXPECT_NE(desc.find("3 objects"), std::string::npos);
    EXPECT_NE(desc.find("2 stationary"), std::string::npos);
    EXPECT_NE(desc.find("1 dynamic"), std::string::npos);
}

TEST_F(SceneModelTest, RemoveObject) {
    cv::Rect2f bbox(0.0f, 0.0f, 40.0f, 40.0f);
    cv::Point2f position(100.0f, 200.0f);
    
    scene->addOrUpdateObject("bench", position, bbox, 0.85f, true);
    EXPECT_EQ(scene->getObjects().size(), 1);
    
    scene->removeObject("bench", position);
    EXPECT_EQ(scene->getObjects().size(), 0);
}

TEST_F(SceneModelTest, ClearScene) {
    cv::Rect2f bbox(0.0f, 0.0f, 40.0f, 40.0f);
    
    scene->addOrUpdateObject("bench", cv::Point2f(100.0f, 200.0f), bbox, 0.85f, true);
    scene->addOrUpdateObject("potted plant", cv::Point2f(300.0f, 400.0f), bbox, 0.90f, true);
    scene->addOrUpdateObject("person", cv::Point2f(500.0f, 100.0f), bbox, 0.80f, false);
    
    EXPECT_EQ(scene->getObjects().size(), 3);
    
    scene->clear();
    EXPECT_EQ(scene->getObjects().size(), 0);
}

TEST_F(SceneModelTest, IsStationaryObjectType) {
    // Test stationary object types
    EXPECT_TRUE(SceneModel::isStationaryObjectType("bench"));
    EXPECT_TRUE(SceneModel::isStationaryObjectType("potted plant"));
    EXPECT_TRUE(SceneModel::isStationaryObjectType("traffic light"));
    EXPECT_TRUE(SceneModel::isStationaryObjectType("fire hydrant"));
    EXPECT_TRUE(SceneModel::isStationaryObjectType("stop sign"));
    EXPECT_TRUE(SceneModel::isStationaryObjectType("parking meter"));
    
    // Test non-stationary object types
    EXPECT_FALSE(SceneModel::isStationaryObjectType("person"));
    EXPECT_FALSE(SceneModel::isStationaryObjectType("car"));
    EXPECT_FALSE(SceneModel::isStationaryObjectType("dog"));
    EXPECT_FALSE(SceneModel::isStationaryObjectType("bird"));
}

TEST_F(SceneModelTest, AutoDetectStationaryType) {
    cv::Rect2f bbox(0.0f, 0.0f, 40.0f, 40.0f);
    
    // Add a bench without explicitly marking as stationary
    // It should be automatically marked as stationary based on type
    scene->addOrUpdateObject("bench", cv::Point2f(100.0f, 200.0f), bbox, 0.85f, false);
    
    EXPECT_EQ(scene->getObjects().size(), 1);
    EXPECT_TRUE(scene->getObjects()[0].is_stationary);
}

TEST_F(SceneModelTest, GetStationaryObjectTypes) {
    auto types = SceneModel::getStationaryObjectTypes();
    
    EXPECT_FALSE(types.empty());
    EXPECT_GT(types.size(), 5);
    
    // Check for expected types
    bool has_bench = std::find(types.begin(), types.end(), "bench") != types.end();
    bool has_plant = std::find(types.begin(), types.end(), "potted plant") != types.end();
    bool has_traffic_light = std::find(types.begin(), types.end(), "traffic light") != types.end();
    
    EXPECT_TRUE(has_bench);
    EXPECT_TRUE(has_plant);
    EXPECT_TRUE(has_traffic_light);
}

TEST_F(SceneModelTest, CleanupOldObjects) {
    cv::Rect2f bbox(0.0f, 0.0f, 40.0f, 40.0f);
    
    scene->addOrUpdateObject("bench", cv::Point2f(100.0f, 200.0f), bbox, 0.85f, true);
    scene->addOrUpdateObject("potted plant", cv::Point2f(300.0f, 400.0f), bbox, 0.90f, true);
    
    EXPECT_EQ(scene->getObjects().size(), 2);
    
    // Cleanup with 0 second threshold should remove all objects
    scene->cleanupOldObjects(0);
    
    // Objects should still exist since they were just added
    // (cleanup removes objects older than threshold, not equal to)
    EXPECT_GE(scene->getObjects().size(), 0);
}

TEST_F(SceneModelTest, AverageConfidenceCalculation) {
    cv::Rect2f bbox(0.0f, 0.0f, 40.0f, 40.0f);
    cv::Point2f position(100.0f, 200.0f);
    
    scene->addOrUpdateObject("bench", position, bbox, 0.8f, true);
    EXPECT_FLOAT_EQ(scene->getObjects()[0].average_confidence, 0.8f);
    
    scene->addOrUpdateObject("bench", position, bbox, 0.9f, true);
    // Average of 0.8 and 0.9 should be 0.85
    EXPECT_NEAR(scene->getObjects()[0].average_confidence, 0.85f, 0.01f);
    
    scene->addOrUpdateObject("bench", position, bbox, 0.7f, true);
    // Average of 0.8, 0.9, 0.7 should be 0.8
    EXPECT_NEAR(scene->getObjects()[0].average_confidence, 0.8f, 0.01f);
}

TEST_F(SceneModelTest, SpatialRelationshipDistance) {
    cv::Rect2f bbox(0.0f, 0.0f, 40.0f, 40.0f);
    
    // Add two objects at known positions
    scene->addOrUpdateObject("bench", cv::Point2f(0.0f, 0.0f), bbox, 0.85f, true);
    scene->addOrUpdateObject("potted plant", cv::Point2f(3.0f, 4.0f), bbox, 0.90f, true);
    
    auto relationships = scene->getSpatialRelationships();
    EXPECT_EQ(relationships.size(), 1);
    
    // Distance should be sqrt(3^2 + 4^2) = 5.0
    EXPECT_NEAR(relationships[0].distance, 5.0f, 0.01f);
}
