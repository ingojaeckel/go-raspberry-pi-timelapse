#include <gtest/gtest.h>
#include "scene_graph/Detector.h"

using namespace scene_graph;

TEST(DetectorTest, InitializeWithoutModel) {
    Detector detector;
    EXPECT_FALSE(detector.isInitialized());
}

TEST(DetectorTest, LoadInvalidModel) {
    Detector detector;
    bool result = detector.initialize("/nonexistent/model.onnx", 
                                      "/nonexistent/labels.txt",
                                      "cpu");
    EXPECT_FALSE(result);
    EXPECT_FALSE(detector.isInitialized());
}

int main(int argc, char** argv) {
    ::testing::InitGoogleTest(&argc, argv);
    return RUN_ALL_TESTS();
}
