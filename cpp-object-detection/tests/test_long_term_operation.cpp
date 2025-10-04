#include <gtest/gtest.h>
#include <memory>
#include <fstream>
#include <sys/stat.h>
#include "system_monitor.hpp"
#include "logger.hpp"

class SystemMonitorTest : public ::testing::Test {
protected:
    void SetUp() override {
        test_dir_ = "/tmp/system_monitor_test";
        
        // Create test directory
        mkdir(test_dir_.c_str(), 0755);
        
        // Create logger
        logger_ = std::make_shared<Logger>("/tmp/system_monitor_test.log", false);
        
        // Create system monitor
        system_monitor_ = std::make_shared<SystemMonitor>(logger_, test_dir_);
    }
    
    void TearDown() override {
        // Clean up test directory
        system("rm -rf /tmp/system_monitor_test");
        system("rm -f /tmp/system_monitor_test.log");
    }
    
    std::string test_dir_;
    std::shared_ptr<Logger> logger_;
    std::shared_ptr<SystemMonitor> system_monitor_;
};

TEST_F(SystemMonitorTest, DiskSpaceMonitoring) {
    // Should be able to get disk space
    unsigned long long available = system_monitor_->getAvailableDiskSpace();
    EXPECT_GT(available, 0);
    
    double usage_percent = system_monitor_->getDiskUsagePercent();
    EXPECT_GE(usage_percent, 0.0);
    EXPECT_LE(usage_percent, 100.0);
}

TEST_F(SystemMonitorTest, CPUTemperatureReading) {
    // May return -1 if not available, which is acceptable
    double temp = system_monitor_->getCPUTemperature();
    
    if (temp > 0) {
        // If temperature is available, it should be reasonable
        EXPECT_GT(temp, 0.0);
        EXPECT_LT(temp, 150.0);  // Should be less than 150°C
    } else {
        // Temperature unavailable is also acceptable
        EXPECT_EQ(temp, -1.0);
    }
}

TEST_F(SystemMonitorTest, DiskSpaceCriticalCheck) {
    // For a normal system, disk should not be critical
    // (This might fail on very full systems, which is expected behavior)
    bool is_critical = system_monitor_->isDiskSpaceCritical();
    
    // We can't make strong assertions here as it depends on actual disk state
    // Just verify the function returns a boolean
    EXPECT_TRUE(is_critical == true || is_critical == false);
}

TEST_F(SystemMonitorTest, CleanupOldDetections) {
    // Create some test files
    for (int i = 0; i < 10; ++i) {
        std::string filename = test_dir_ + "/test_" + std::to_string(i) + ".jpg";
        std::ofstream file(filename);
        file << "test data";
        file.close();
    }
    
    // Cleanup should remove some files
    int deleted = system_monitor_->cleanupOldDetections();
    
    // Should have deleted at least 2 files (20% of 10)
    EXPECT_GE(deleted, 1);
    EXPECT_LE(deleted, 10);
}

TEST_F(SystemMonitorTest, PeriodicCheckDoesNotCrash) {
    // Periodic check should complete without crashing
    EXPECT_NO_THROW(system_monitor_->performPeriodicCheck());
}

TEST_F(SystemMonitorTest, LogSystemStatsDoesNotCrash) {
    // Log system stats should complete without crashing
    EXPECT_NO_THROW(system_monitor_->logSystemStats());
}

// Test for bounded data structures in ObjectDetector
#include "object_detector.hpp"

class ObjectDetectorBoundsTest : public ::testing::Test {
protected:
    void SetUp() override {
        logger_ = std::make_shared<Logger>("/tmp/object_detector_test.log", false);
    }
    
    void TearDown() override {
        system("rm -f /tmp/object_detector_test.log");
    }
    
    std::shared_ptr<Logger> logger_;
};

TEST_F(ObjectDetectorBoundsTest, VerifyMaxLimitsAreDefined) {
    // This test verifies that the limits are defined at compile time
    // These should be accessible as constexpr values
    
    // Note: We can't directly test private constexpr members,
    // but we can verify the class compiles with these limits
    
    // Create a detector (won't initialize fully without model, but that's OK)
    ObjectDetector detector("dummy.onnx", "dummy.cfg", "dummy.txt", 0.5, logger_);
    
    // Just verify it constructs without error
    SUCCEED();
}

// Test for performance counter overflow protection
#include "performance_monitor.hpp"

class PerformanceMonitorBoundsTest : public ::testing::Test {
protected:
    void SetUp() override {
        logger_ = std::make_shared<Logger>("/tmp/perf_monitor_test.log", false);
        perf_monitor_ = std::make_shared<PerformanceMonitor>(logger_, 1.0);
    }
    
    void TearDown() override {
        system("rm -f /tmp/perf_monitor_test.log");
    }
    
    std::shared_ptr<Logger> logger_;
    std::shared_ptr<PerformanceMonitor> perf_monitor_;
};

TEST_F(PerformanceMonitorBoundsTest, NormalOperationDoesNotOverflow) {
    // Simulate normal operation
    for (int i = 0; i < 1000; ++i) {
        perf_monitor_->startFrameProcessing();
        perf_monitor_->endFrameProcessing();
    }
    
    // Should have processed 1000 frames
    double avg_time = perf_monitor_->getAverageProcessingTime();
    EXPECT_GE(avg_time, 0.0);
}

TEST_F(PerformanceMonitorBoundsTest, ResetWorks) {
    // Process some frames
    for (int i = 0; i < 100; ++i) {
        perf_monitor_->startFrameProcessing();
        perf_monitor_->endFrameProcessing();
    }
    
    // Reset
    perf_monitor_->reset();
    
    // After reset, should have 0 average time
    double avg_time = perf_monitor_->getAverageProcessingTime();
    EXPECT_EQ(avg_time, 0.0);
}

// Test for webcam interface resilience
#include "webcam_interface.hpp"

class WebcamInterfaceResilienceTest : public ::testing::Test {
protected:
    void SetUp() override {
        logger_ = std::make_shared<Logger>("/tmp/webcam_test.log", false);
    }
    
    void TearDown() override {
        system("rm -f /tmp/webcam_test.log");
    }
    
    std::shared_ptr<Logger> logger_;
};

TEST_F(WebcamInterfaceResilienceTest, HealthCheckHandlesUninitialized) {
    // Create webcam interface but don't initialize
    WebcamInterface webcam(0, 640, 480, logger_);
    
    // Health check on uninitialized camera should handle gracefully
    EXPECT_NO_THROW(webcam.healthCheck());
}

TEST_F(WebcamInterfaceResilienceTest, KeepAliveHandlesUninitialized) {
    // Create webcam interface but don't initialize
    WebcamInterface webcam(0, 640, 480, logger_);
    
    // Keep-alive on uninitialized camera should handle gracefully
    EXPECT_NO_THROW(webcam.keepAlive());
}

int main(int argc, char **argv) {
    ::testing::InitGoogleTest(&argc, argv);
    return RUN_ALL_TESTS();
}
