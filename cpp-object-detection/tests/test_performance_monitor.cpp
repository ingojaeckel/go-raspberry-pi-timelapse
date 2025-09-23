#include <gtest/gtest.h>
#include <thread>
#include <chrono>
#include "performance_monitor.hpp"
#include "logger.hpp"

// Check if filesystem is available
#if __has_include(<filesystem>)
    #include <filesystem>
    namespace fs = std::filesystem;
#else
    // Fallback for older systems
    #include <cstdio>
    namespace fs {
        inline bool remove(const std::string& path) {
            return std::remove(path.c_str()) == 0;
        }
    }
#endif

class PerformanceMonitorTest : public ::testing::Test {
protected:
    void SetUp() override {
        test_log_file = "/tmp/test_perf_monitor.log";
        fs::remove(test_log_file);
        logger = std::make_shared<Logger>(test_log_file, false);
        perf_monitor = std::make_unique<PerformanceMonitor>(logger, 1.0);
    }

    void TearDown() override {
        fs::remove(test_log_file);
    }

    std::string test_log_file;
    std::shared_ptr<Logger> logger;
    std::unique_ptr<PerformanceMonitor> perf_monitor;
};

TEST_F(PerformanceMonitorTest, InitialState) {
    EXPECT_DOUBLE_EQ(perf_monitor->getCurrentFPS(), 0.0);
    EXPECT_DOUBLE_EQ(perf_monitor->getAverageProcessingTime(), 0.0);
}

TEST_F(PerformanceMonitorTest, SingleFrameProcessing) {
    perf_monitor->startFrameProcessing();
    
    // Simulate some processing time
    std::this_thread::sleep_for(std::chrono::milliseconds(10));
    
    perf_monitor->endFrameProcessing();
    
    EXPECT_GT(perf_monitor->getAverageProcessingTime(), 0.0);
    EXPECT_GT(perf_monitor->getCurrentFPS(), 0.0);
}

TEST_F(PerformanceMonitorTest, MultipleFrameProcessing) {
    const int num_frames = 5;
    
    for (int i = 0; i < num_frames; ++i) {
        perf_monitor->startFrameProcessing();
        std::this_thread::sleep_for(std::chrono::milliseconds(5));
        perf_monitor->endFrameProcessing();
        
        // Small delay between frames
        std::this_thread::sleep_for(std::chrono::milliseconds(20));
    }
    
    EXPECT_GT(perf_monitor->getAverageProcessingTime(), 0.0);
    EXPECT_GT(perf_monitor->getCurrentFPS(), 0.0);
    
    std::string stats = perf_monitor->getStatsSummary();
    EXPECT_TRUE(stats.find("FPS:") != std::string::npos);
    EXPECT_TRUE(stats.find("Avg processing time:") != std::string::npos);
    EXPECT_TRUE(stats.find("Frames processed/captured:") != std::string::npos);
    EXPECT_TRUE(stats.find("5/5") != std::string::npos);
}

TEST_F(PerformanceMonitorTest, Reset) {
    perf_monitor->startFrameProcessing();
    std::this_thread::sleep_for(std::chrono::milliseconds(10));
    perf_monitor->endFrameProcessing();
    
    EXPECT_GT(perf_monitor->getAverageProcessingTime(), 0.0);
    
    perf_monitor->reset();
    
    EXPECT_DOUBLE_EQ(perf_monitor->getCurrentFPS(), 0.0);
    EXPECT_DOUBLE_EQ(perf_monitor->getAverageProcessingTime(), 0.0);
}

TEST_F(PerformanceMonitorTest, StatsFormat) {
    perf_monitor->startFrameProcessing();
    std::this_thread::sleep_for(std::chrono::milliseconds(10));
    perf_monitor->endFrameProcessing();
    
    std::string stats = perf_monitor->getStatsSummary();
    
    // Verify format contains expected elements
    EXPECT_TRUE(stats.find("FPS:") != std::string::npos);
    EXPECT_TRUE(stats.find("ms") != std::string::npos);
    EXPECT_TRUE(stats.find("1/1") != std::string::npos);
    EXPECT_TRUE(stats.find("(100%)") != std::string::npos);
}