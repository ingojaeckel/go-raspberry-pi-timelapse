#include <gtest/gtest.h>
#include "scene_graph/Detector.h"
#include "scene_graph/logger.h"

using namespace scene_graph;

class DetectorTest : public ::testing::Test {
protected:
    void SetUp() override {
        Logger::getInstance().setVerbose(false);
        detector = std::make_unique<Detector>();
    }

    std::unique_ptr<Detector> detector;
};

TEST_F(DetectorTest, InitialState) {
    EXPECT_FALSE(detector->isLoaded());
}

TEST_F(DetectorTest, BackendInfo) {
    std::string info = detector->getBackendInfo();
    EXPECT_NE(info.find("Backend:"), std::string::npos);
}

TEST_F(DetectorTest, InferWithoutModel) {
    cv::Mat dummy_frame(640, 640, CV_8UC3, cv::Scalar(128, 128, 128));
    auto detections = detector->infer(dummy_frame);
    
    EXPECT_EQ(detections.size(), 0);
}

TEST_F(DetectorTest, InferWithEmptyFrame) {
    cv::Mat empty_frame;
    auto detections = detector->infer(empty_frame);
    
    EXPECT_EQ(detections.size(), 0);
}
