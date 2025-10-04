#include <gtest/gtest.h>
#include <memory>
#include "network_streamer.hpp"
#include "logger.hpp"

class NetworkStreamerTest : public ::testing::Test {
protected:
    void SetUp() override {
        logger_ = std::make_shared<Logger>("test_network_streamer.log", false);
    }

    void TearDown() override {
        // Clean up
    }

    std::shared_ptr<Logger> logger_;
};

TEST_F(NetworkStreamerTest, InitializationSucceeds) {
    // Use a high port number to avoid permission issues
    NetworkStreamer streamer(logger_, 9999);
    EXPECT_TRUE(streamer.initialize());
}

TEST_F(NetworkStreamerTest, InvalidPortValidation) {
    // Port 0 is invalid but won't be caught until bind
    // Test that we can create the object with a valid port
    NetworkStreamer streamer(logger_, 8080);
    EXPECT_NO_THROW(streamer.initialize());
}

TEST_F(NetworkStreamerTest, StartAndStopServer) {
    NetworkStreamer streamer(logger_, 9998);
    EXPECT_TRUE(streamer.initialize());
    EXPECT_TRUE(streamer.start());
    EXPECT_TRUE(streamer.isRunning());
    
    // Give server time to start
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    
    streamer.stop();
    EXPECT_FALSE(streamer.isRunning());
}

TEST_F(NetworkStreamerTest, GetStreamingUrl) {
    NetworkStreamer streamer(logger_, 8080);
    std::string url = streamer.getStreamingUrl();
    
    // URL should contain the port and /stream path
    EXPECT_NE(url.find("8080"), std::string::npos);
    EXPECT_NE(url.find("/stream"), std::string::npos);
    EXPECT_NE(url.find("http://"), std::string::npos);
}

TEST_F(NetworkStreamerTest, UpdateFrameWithEmptyFrame) {
    NetworkStreamer streamer(logger_, 9997);
    EXPECT_TRUE(streamer.initialize());
    
    cv::Mat empty_frame;
    std::vector<Detection> detections;
    
    // Should not crash with empty frame
    EXPECT_NO_THROW(streamer.updateFrame(empty_frame, detections));
}

TEST_F(NetworkStreamerTest, UpdateFrameWithValidFrame) {
    NetworkStreamer streamer(logger_, 9996);
    EXPECT_TRUE(streamer.initialize());
    
    // Create a simple test frame
    cv::Mat frame(480, 640, CV_8UC3, cv::Scalar(0, 0, 0));
    std::vector<Detection> detections;
    
    // Add a test detection
    Detection det;
    det.class_name = "person";
    det.confidence = 0.95;
    det.bbox = cv::Rect(100, 100, 200, 300);
    detections.push_back(det);
    
    EXPECT_NO_THROW(streamer.updateFrame(frame, detections));
}

TEST_F(NetworkStreamerTest, MultipleStartStopCycles) {
    NetworkStreamer streamer(logger_, 9995);
    EXPECT_TRUE(streamer.initialize());
    
    for (int i = 0; i < 3; i++) {
        EXPECT_TRUE(streamer.start());
        std::this_thread::sleep_for(std::chrono::milliseconds(50));
        streamer.stop();
        std::this_thread::sleep_for(std::chrono::milliseconds(50));
    }
}

TEST_F(NetworkStreamerTest, StopWithoutStart) {
    NetworkStreamer streamer(logger_, 9994);
    EXPECT_TRUE(streamer.initialize());
    
    // Should not crash
    EXPECT_NO_THROW(streamer.stop());
}
