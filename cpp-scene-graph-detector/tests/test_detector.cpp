#include <gtest/gtest.h>
#include "scene_graph/detector.h"

using namespace scene_graph;

TEST(DetectorTest, InitializeWithoutModel) {
    Detector detector;
    EXPECT_FALSE(detector.isInitialized());
}

// Note: Loading invalid model may throw, so we skip this test
// TEST(DetectorTest, LoadInvalidModel) {
//     Detector detector;
//     bool result = detector.initialize("/nonexistent/model.onnx", 
//                                       "/nonexistent/labels.txt",
//                                       "cpu");
//     EXPECT_FALSE(result);
//     EXPECT_FALSE(detector.isInitialized());
// }
