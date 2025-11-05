#include <gtest/gtest.h>
#include "scene_graph/RelPredictor.h"

using namespace scene_graph;

class RelPredictorTest : public ::testing::Test {
protected:
    void SetUp() override {
        predictor = std::make_unique<RelPredictor>();
    }

    std::unique_ptr<RelPredictor> predictor;
};

TEST_F(RelPredictorTest, InitialState) {
    EXPECT_FALSE(predictor->isModelLoaded());
}

TEST_F(RelPredictorTest, EmptyDetections) {
    std::vector<Detection> detections;
    auto relations = predictor->infer(detections);
    
    EXPECT_EQ(relations.size(), 0);
}

TEST_F(RelPredictorTest, GeometricLeftOf) {
    std::vector<Detection> detections;
    
    // Object on the left
    Detection det1(0, "person", 0.9f, cv::Rect(50, 100, 80, 150));
    // Object on the right
    Detection det2(1, "car", 0.85f, cv::Rect(200, 100, 100, 80));
    
    detections.push_back(det1);
    detections.push_back(det2);
    
    auto relations = predictor->infer(detections, 0.5f);
    
    // Should find at least one left_of relation
    bool found_left_of = false;
    for (const auto& rel : relations) {
        if (rel.predicate_label == "left_of" && 
            rel.subject_id == 0 && rel.object_id == 1) {
            found_left_of = true;
            break;
        }
    }
    
    EXPECT_TRUE(found_left_of);
}

TEST_F(RelPredictorTest, GeometricOverlaps) {
    std::vector<Detection> detections;
    
    // Two overlapping objects
    Detection det1(0, "person", 0.9f, cv::Rect(100, 100, 100, 100));
    Detection det2(1, "backpack", 0.75f, cv::Rect(150, 120, 80, 80));
    
    detections.push_back(det1);
    detections.push_back(det2);
    
    auto relations = predictor->infer(detections, 0.0f);
    
    // Should find overlaps relation
    bool found_overlaps = false;
    for (const auto& rel : relations) {
        if (rel.predicate_label == "overlaps") {
            found_overlaps = true;
            break;
        }
    }
    
    EXPECT_TRUE(found_overlaps);
}

TEST_F(RelPredictorTest, PredicateLabels) {
    EXPECT_EQ(RelPredictor::getPredicateLabel(SpatialPredicate::LEFT_OF), "left_of");
    EXPECT_EQ(RelPredictor::getPredicateLabel(SpatialPredicate::RIGHT_OF), "right_of");
    EXPECT_EQ(RelPredictor::getPredicateLabel(SpatialPredicate::OVERLAPS), "overlaps");
    EXPECT_EQ(RelPredictor::getPredicateLabel(SpatialPredicate::CONTAINS), "contains");
}

TEST_F(RelPredictorTest, PredicateIds) {
    EXPECT_EQ(RelPredictor::getPredicateId(SpatialPredicate::LEFT_OF), 0);
    EXPECT_EQ(RelPredictor::getPredicateId(SpatialPredicate::RIGHT_OF), 1);
    EXPECT_EQ(RelPredictor::getPredicateId(SpatialPredicate::OVERLAPS), 4);
}
