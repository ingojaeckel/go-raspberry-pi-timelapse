#include <gtest/gtest.h>
#include "scene_graph/RelPredictor.h"

using namespace scene_graph;

TEST(RelPredictorTest, InitializeGeometric) {
    RelPredictor predictor;
    bool result = predictor.initialize("", "cpu");
    EXPECT_TRUE(result);
    EXPECT_TRUE(predictor.isInitialized());
}

TEST(RelPredictorTest, GetPredicateLabel) {
    EXPECT_EQ(RelPredictor::getPredicateLabel(SpatialPredicate::LEFT_OF), "left_of");
    EXPECT_EQ(RelPredictor::getPredicateLabel(SpatialPredicate::RIGHT_OF), "right_of");
    EXPECT_EQ(RelPredictor::getPredicateLabel(SpatialPredicate::OVERLAPS), "overlaps");
    EXPECT_EQ(RelPredictor::getPredicateLabel(SpatialPredicate::CONTAINS), "contains");
    EXPECT_EQ(RelPredictor::getPredicateLabel(SpatialPredicate::ON), "on");
    EXPECT_EQ(RelPredictor::getPredicateLabel(SpatialPredicate::UNDER), "under");
    EXPECT_EQ(RelPredictor::getPredicateLabel(SpatialPredicate::NEXT_TO), "next_to");
}

TEST(RelPredictorTest, PredictGeometric) {
    RelPredictor predictor;
    predictor.initialize("", "cpu");
    
    // Create two detections side by side
    std::vector<Detection> detections;
    
    Detection det1;
    det1.label = "object1";
    det1.bbox = BBox(0.3f, 0.5f, 0.2f, 0.2f);
    detections.push_back(det1);
    
    Detection det2;
    det2.label = "object2";
    det2.bbox = BBox(0.7f, 0.5f, 0.2f, 0.2f);
    detections.push_back(det2);
    
    cv::Mat dummy_image(480, 640, CV_8UC3);
    std::vector<RelationPrediction> relations = predictor.predict(detections, dummy_image, 0.5f);
    
    // Should find some relations (at least left_of/right_of)
    EXPECT_GT(relations.size(), 0);
}
