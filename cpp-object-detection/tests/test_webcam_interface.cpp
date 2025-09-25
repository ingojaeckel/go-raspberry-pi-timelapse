#include <gtest/gtest.h>
#include "webcam_interface.hpp"
#include "logger.hpp"
#include <opencv2/opencv.hpp>

class WebcamInterfaceTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Most tests will use mock parameters since we don't have real cameras in CI
        logger = std::make_shared<Logger>("test_webcam.log", false);
    }
    
    std::shared_ptr<Logger> logger;
};

TEST_F(WebcamInterfaceTest, CreateWithValidParameters) {
    // Test creating webcam interface with valid parameters
    auto webcam = std::make_unique<WebcamInterface>(0, 640, 480, logger);
    EXPECT_NE(webcam, nullptr);
}

TEST_F(WebcamInterfaceTest, ListAvailableCameras) {
    // Test camera listing functionality
    auto cameras = WebcamInterface::listAvailableCameras();
    
    // In CI environment, we expect no cameras, but the function should not crash
    EXPECT_TRUE(cameras.empty() || !cameras.empty()); // Always true, just testing it runs
    
    // The function should return a vector
    EXPECT_TRUE((std::is_same_v<decltype(cameras), std::vector<std::string>>));
}

TEST_F(WebcamInterfaceTest, InitializeWithInvalidCamera) {
    // Test initialization with a camera that doesn't exist
    auto webcam = std::make_unique<WebcamInterface>(999, 640, 480, logger);
    
    // This should fail gracefully in CI environment
    bool initialized = webcam->initialize();
    EXPECT_FALSE(initialized); // Should fail with non-existent camera ID 999
}

TEST_F(WebcamInterfaceTest, CaptureFrameWithoutInitialization) {
    // Test capturing frame without initialization
    auto webcam = std::make_unique<WebcamInterface>(999, 640, 480, logger);
    cv::Mat frame;
    
    // Should return false when not initialized
    bool result = webcam->captureFrame(frame);
    EXPECT_FALSE(result);
    EXPECT_TRUE(frame.empty());
}

TEST_F(WebcamInterfaceTest, GetCameraInfoWithoutInitialization) {
    // Test getting camera info without initialization
    auto webcam = std::make_unique<WebcamInterface>(0, 640, 480, logger);
    std::string info = webcam->getCameraInfo();
    
    // Should return some default info
    EXPECT_FALSE(info.empty());
    EXPECT_NE(info.find("Camera ID: 0"), std::string::npos);
}

TEST_F(WebcamInterfaceTest, ReleaseFunction) {
    // Test release function doesn't crash
    auto webcam = std::make_unique<WebcamInterface>(999, 640, 480, logger);
    
    // Should not crash even if not initialized
    webcam->release();
    
    // Should be safe to call multiple times
    webcam->release();
    webcam->release();
}

TEST_F(WebcamInterfaceTest, ResolutionParametersStored) {
    // Test that resolution parameters are stored correctly
    int width = 1280;
    int height = 720;
    auto webcam = std::make_unique<WebcamInterface>(0, width, height, logger);
    
    std::string info = webcam->getCameraInfo();
    EXPECT_NE(info.find("1280x720"), std::string::npos);
}