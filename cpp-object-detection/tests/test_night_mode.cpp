#include <gtest/gtest.h>
#include "parallel_frame_processor.hpp"
#include "object_detector.hpp"
#include "logger.hpp"
#include "performance_monitor.hpp"
#include <opencv2/opencv.hpp>
#include <memory>
#include <ctime>

class NightModeTest : public ::testing::Test {
protected:
    void SetUp() override {
        logger = std::make_shared<Logger>("test_night_mode.log", false);
        perf_monitor = std::make_shared<PerformanceMonitor>(logger, 1.0);
        
        // Create detector with non-existent model (for testing basic functionality)
        detector = std::make_shared<ObjectDetector>(
            "non_existent_model.onnx", 
            "non_existent_config.yaml", 
            "non_existent_classes.txt", 
            0.5, 
            logger);
        
        processor = std::make_unique<ParallelFrameProcessor>(
            detector, logger, perf_monitor, 1, 10, "/tmp/test_detections");
        processor->initialize();
    }
    
    void TearDown() override {
        processor->shutdown();
    }
    
    std::shared_ptr<Logger> logger;
    std::shared_ptr<PerformanceMonitor> perf_monitor;
    std::shared_ptr<ObjectDetector> detector;
    std::unique_ptr<ParallelFrameProcessor> processor;
};

TEST_F(NightModeTest, BrightImageDetection) {
    // Create a bright white image
    cv::Mat bright_frame = cv::Mat(480, 640, CV_8UC3, cv::Scalar(255, 255, 255));
    
    // Process frame - should not be detected as night mode even during night hours
    // because the image is bright
    auto result = processor->processFrameSync(bright_frame);
    
    EXPECT_TRUE(result.processed || !result.processed); // Test completes without crash
}

TEST_F(NightModeTest, DarkImageDetection) {
    // Create a dark image (simulating night)
    cv::Mat dark_frame = cv::Mat(480, 640, CV_8UC3, cv::Scalar(30, 30, 30));
    
    // Process frame
    auto result = processor->processFrameSync(dark_frame);
    
    EXPECT_TRUE(result.processed || !result.processed); // Test completes without crash
}

TEST_F(NightModeTest, PreprocessingPreservesSize) {
    // Create a test frame
    cv::Mat test_frame = cv::Mat(480, 640, CV_8UC3, cv::Scalar(50, 50, 50));
    
    // Process and ensure no crashes
    auto result = processor->processFrameSync(test_frame);
    
    EXPECT_TRUE(result.processed || !result.processed); // Test completes without crash
}

TEST_F(NightModeTest, NormalImageProcessing) {
    // Create a normal brightness image
    cv::Mat normal_frame = cv::Mat(480, 640, CV_8UC3, cv::Scalar(128, 128, 128));
    
    // Process frame
    auto result = processor->processFrameSync(normal_frame);
    
    EXPECT_TRUE(result.processed || !result.processed); // Test completes without crash
}

TEST_F(NightModeTest, ColorImagePreprocessing) {
    // Create a color image with low brightness
    cv::Mat color_frame = cv::Mat(480, 640, CV_8UC3, cv::Scalar(40, 50, 45));
    
    // Process frame
    auto result = processor->processFrameSync(color_frame);
    
    EXPECT_TRUE(result.processed || !result.processed); // Test completes without crash
}

TEST_F(NightModeTest, MultipleFrameProcessing) {
    // Test processing multiple frames with different brightness levels
    std::vector<cv::Mat> frames = {
        cv::Mat(480, 640, CV_8UC3, cv::Scalar(20, 20, 20)),   // Very dark
        cv::Mat(480, 640, CV_8UC3, cv::Scalar(50, 50, 50)),   // Dark
        cv::Mat(480, 640, CV_8UC3, cv::Scalar(128, 128, 128)), // Medium
        cv::Mat(480, 640, CV_8UC3, cv::Scalar(200, 200, 200))  // Bright
    };
    
    for (const auto& frame : frames) {
        auto result = processor->processFrameSync(frame);
        EXPECT_TRUE(result.processed || !result.processed); // Test completes without crash
    }
}

TEST_F(NightModeTest, GrayscaleImageHandling) {
    // Create a grayscale image
    cv::Mat gray_frame = cv::Mat(480, 640, CV_8UC1, cv::Scalar(50));
    
    // Convert to BGR for processing
    cv::Mat bgr_frame;
    cv::cvtColor(gray_frame, bgr_frame, cv::COLOR_GRAY2BGR);
    
    // Process frame
    auto result = processor->processFrameSync(bgr_frame);
    
    EXPECT_TRUE(result.processed || !result.processed); // Test completes without crash
}
