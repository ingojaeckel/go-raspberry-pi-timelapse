#include <gtest/gtest.h>
#include "viewfinder_window.hpp"
#include "logger.hpp"
#include <opencv2/opencv.hpp>

class ViewfinderWindowTest : public ::testing::Test {
protected:
    void SetUp() override {
        logger = std::make_shared<Logger>("test_viewfinder.log", false);
        viewfinder = std::make_unique<ViewfinderWindow>(logger, "Test Viewfinder");
    }

    void TearDown() override {
        if (viewfinder) {
            viewfinder->close();
        }
    }

    std::shared_ptr<Logger> logger;
    std::unique_ptr<ViewfinderWindow> viewfinder;
};

TEST_F(ViewfinderWindowTest, Initialization) {
    // Note: This test may fail in headless environments without X11
    // In CI/CD, it should pass as we're not actually calling initialize()
    EXPECT_NE(viewfinder, nullptr);
}

TEST_F(ViewfinderWindowTest, ShowFrameWithoutInitialization) {
    // Create a test frame
    cv::Mat test_frame(480, 640, CV_8UC3, cv::Scalar(0, 0, 255));
    
    // Create empty detection vector
    std::vector<Detection> detections;
    
    // Should not crash when showing frame without initialization
    EXPECT_NO_THROW(viewfinder->showFrame(test_frame, detections));
}

TEST_F(ViewfinderWindowTest, ShowFrameWithDetections) {
    // Create a test frame
    cv::Mat test_frame(480, 640, CV_8UC3, cv::Scalar(0, 0, 255));
    
    // Create test detections
    std::vector<Detection> detections;
    Detection det;
    det.class_name = "person";
    det.confidence = 0.85;
    det.bbox = cv::Rect(100, 100, 200, 300);
    detections.push_back(det);
    
    // Should not crash when showing frame with detections
    EXPECT_NO_THROW(viewfinder->showFrame(test_frame, detections));
}

TEST_F(ViewfinderWindowTest, CloseWithoutInitialization) {
    // Should not crash when closing without initialization
    EXPECT_NO_THROW(viewfinder->close());
}

TEST_F(ViewfinderWindowTest, ShouldCloseReturnsTrue) {
    // Without initialization, should return true
    EXPECT_TRUE(viewfinder->shouldClose());
}

TEST_F(ViewfinderWindowTest, ShowFrameWithStatsWithoutInitialization) {
    // Create a test frame
    cv::Mat test_frame(480, 640, CV_8UC3, cv::Scalar(0, 0, 255));
    
    // Create test detections
    std::vector<Detection> detections;
    Detection det;
    det.class_name = "cat";
    det.confidence = 0.92;
    det.bbox = cv::Rect(50, 50, 150, 150);
    detections.push_back(det);
    
    // Create top objects list
    std::vector<std::pair<std::string, int>> top_objects = {
        {"cat", 5},
        {"person", 3},
        {"dog", 2}
    };
    
    auto start_time = std::chrono::steady_clock::now();
    
    // Should not crash when showing frame with stats without initialization
    EXPECT_NO_THROW(viewfinder->showFrameWithStats(
        test_frame, 
        detections,
        15.5,  // FPS
        45.2,  // avg processing time
        10,    // total objects detected
        3,     // total images saved
        start_time,
        top_objects,
        640,   // camera width
        480,   // camera height
        0,     // camera ID
        "Test Camera",
        320,   // detection width
        240,   // detection height
        false, // brightness filter active
        false, // GPU enabled
        false, // burst mode enabled
        75.5,  // disk usage percent
        62.3   // CPU temp celsius
    ));
}
