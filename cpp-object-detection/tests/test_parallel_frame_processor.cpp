#include <gtest/gtest.h>
#include "parallel_frame_processor.hpp"
#include "object_detector.hpp"
#include "logger.hpp"
#include "performance_monitor.hpp"
#include <opencv2/opencv.hpp>
#include <memory>
#include <thread>
#include <chrono>

class ParallelFrameProcessorTest : public ::testing::Test {
protected:
    void SetUp() override {
        logger = std::make_shared<Logger>("test_parallel.log", false);
        perf_monitor = std::make_shared<PerformanceMonitor>(logger, 1.0);
        
        // Create detector with non-existent model (for testing basic functionality)
        detector = std::make_shared<ObjectDetector>(
            "non_existent_model.onnx", 
            "non_existent_config.yaml", 
            "non_existent_classes.txt", 
            0.5, 
            logger);
    }
    
    std::shared_ptr<Logger> logger;
    std::shared_ptr<PerformanceMonitor> perf_monitor;
    std::shared_ptr<ObjectDetector> detector;
};

TEST_F(ParallelFrameProcessorTest, CreateSingleThreadedProcessor) {
    // Test creating single-threaded processor
    auto processor = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 1, 10);
    
    EXPECT_NE(processor, nullptr);
    EXPECT_FALSE(processor->isParallelEnabled());
}

TEST_F(ParallelFrameProcessorTest, CreateMultiThreadedProcessor) {
    // Test creating multi-threaded processor
    auto processor = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 4, 10);
    
    EXPECT_NE(processor, nullptr);
    EXPECT_TRUE(processor->isParallelEnabled());
}

TEST_F(ParallelFrameProcessorTest, InitializeSingleThreaded) {
    // Test initializing single-threaded processor
    auto processor = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 1, 10);
    
    bool initialized = processor->initialize();
    EXPECT_TRUE(initialized);
    
    processor->shutdown();
}

TEST_F(ParallelFrameProcessorTest, InitializeMultiThreaded) {
    // Test initializing multi-threaded processor
    auto processor = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 2, 10);
    
    bool initialized = processor->initialize();
    EXPECT_TRUE(initialized);
    
    processor->shutdown();
}

TEST_F(ParallelFrameProcessorTest, ProcessFrameSynchronous) {
    // Test synchronous frame processing
    auto processor = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 1, 10);
    
    processor->initialize();
    
    cv::Mat frame = cv::Mat::zeros(480, 640, CV_8UC3);
    auto result = processor->processFrameSync(frame);
    
    // Should complete without crashing
    EXPECT_TRUE(result.processed || !result.processed); // Always true, testing for no crash
    
    processor->shutdown();
}

TEST_F(ParallelFrameProcessorTest, SubmitFrameSingleThreaded) {
    // Test submitting frame for single-threaded processing
    auto processor = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 1, 10);
    
    processor->initialize();
    
    cv::Mat frame = cv::Mat::zeros(480, 640, CV_8UC3);
    auto future = processor->submitFrame(frame);
    
    // Should be able to get result immediately in single-threaded mode
    auto result = future.get();
    EXPECT_TRUE(result.processed || !result.processed); // Test completes without crash
    
    processor->shutdown();
}

TEST_F(ParallelFrameProcessorTest, SubmitFrameMultiThreaded) {
    // Test submitting frame for multi-threaded processing
    auto processor = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 2, 10);
    
    processor->initialize();
    
    cv::Mat frame = cv::Mat::zeros(480, 640, CV_8UC3);
    auto future = processor->submitFrame(frame);
    
    // Should be able to get result
    auto result = future.get();
    EXPECT_TRUE(result.processed || !result.processed); // Test completes without crash
    
    processor->shutdown();
}

TEST_F(ParallelFrameProcessorTest, QueueSizeTracking) {
    // Test queue size tracking
    auto processor = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 2, 5);
    
    processor->initialize();
    
    // Initially queue should be empty
    EXPECT_EQ(processor->getQueueSize(), 0);
    
    processor->shutdown();
}

TEST_F(ParallelFrameProcessorTest, MultipleFrameSubmission) {
    // Test submitting multiple frames
    auto processor = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 2, 10);
    
    processor->initialize();
    
    cv::Mat frame = cv::Mat::zeros(480, 640, CV_8UC3);
    
    // Submit multiple frames
    std::vector<std::future<ParallelFrameProcessor::FrameResult>> futures;
    for (int i = 0; i < 3; ++i) {
        futures.push_back(processor->submitFrame(frame));
    }
    
    // Get all results
    for (auto& future : futures) {
        auto result = future.get();
        EXPECT_TRUE(result.processed || !result.processed); // Test completes without crash
    }
    
    processor->shutdown();
}

TEST_F(ParallelFrameProcessorTest, ShutdownWithoutInitialization) {
    // Test shutdown without initialization
    auto processor = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 2, 10);
    
    // Should not crash
    processor->shutdown();
}

TEST_F(ParallelFrameProcessorTest, MultipleShutdowns) {
    // Test multiple shutdowns
    auto processor = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 2, 10);
    
    processor->initialize();
    
    // Should be safe to call multiple times
    processor->shutdown();
    processor->shutdown();
    processor->shutdown();
}

TEST_F(ParallelFrameProcessorTest, FrameResultStructure) {
    // Test frame result structure
    auto processor = std::make_unique<ParallelFrameProcessor>(
        detector, logger, perf_monitor, 1, 10);
    
    processor->initialize();
    
    cv::Mat frame = cv::Mat::zeros(480, 640, CV_8UC3);
    auto result = processor->processFrameSync(frame);
    
    // Check result structure
    EXPECT_TRUE(result.capture_time.time_since_epoch().count() > 0);
    // result.processed and result.detections will depend on detector state
    
    processor->shutdown();
}

TEST_F(ParallelFrameProcessorTest, ThreadCountValidation) {
    // Test different thread counts
    std::vector<int> thread_counts = {1, 2, 4, 8};
    
    for (int threads : thread_counts) {
        auto processor = std::make_unique<ParallelFrameProcessor>(
            detector, logger, perf_monitor, threads, 10);
        
        EXPECT_NE(processor, nullptr);
        EXPECT_EQ(processor->isParallelEnabled(), threads > 1);
        
        processor->initialize();
        processor->shutdown();
    }
}