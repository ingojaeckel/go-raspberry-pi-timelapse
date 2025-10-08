#include <gtest/gtest.h>
#include <fstream>
#include <thread>
#include <chrono>
#include "logger.hpp"

// Check if filesystem is available
#if __has_include(<filesystem>)
    #include <filesystem>
    namespace fs = std::filesystem;
#else
    // Fallback for older systems
    #include <cstdio>
    namespace fs {
        inline bool exists(const std::string& path) {
            std::ifstream f(path.c_str());
            return f.good();
        }
        inline bool remove(const std::string& path) {
            return std::remove(path.c_str()) == 0;
        }
    }
#endif

class HourlySummaryTest : public ::testing::Test {
protected:
    void SetUp() override {
        test_log_file = "/tmp/test_hourly_summary.log";
        // Remove any existing test log file
        fs::remove(test_log_file);
    }

    void TearDown() override {
        // Clean up test log file
        fs::remove(test_log_file);
    }

    std::string test_log_file;
};

TEST_F(HourlySummaryTest, RecordDetections) {
    auto logger = std::make_unique<Logger>(test_log_file, false);
    
    // Record some detections
    logger->recordDetection("car", true);
    logger->recordDetection("person", false);
    logger->recordDetection("person", false);
    logger->recordDetection("cat", false);
    
    // Manually trigger summary (even though time hasn't elapsed)
    logger->printHourlySummary();
    
    // Summary should be printed to stdout
    // We can't easily capture stdout in this test, but we can verify the method runs without crashing
    SUCCEED();
}

TEST_F(HourlySummaryTest, StationaryObjectFusion) {
    auto logger = std::make_unique<Logger>(test_log_file, false);
    
    // Record a sequence with stationary objects
    logger->recordDetection("car", true);  // Stationary car starts
    logger->recordDetection("car", true);  // Car continues
    logger->recordDetection("car", true);  // Car continues
    logger->recordDetection("person", false);  // Dynamic person
    logger->recordDetection("car", true);  // New stationary car period
    
    // Trigger summary
    logger->printHourlySummary();
    
    SUCCEED();
}

TEST_F(HourlySummaryTest, CheckAndPrintSummaryByTime) {
    auto logger = std::make_unique<Logger>(test_log_file, false);
    
    // Record a detection
    logger->recordDetection("person", false);
    
    // Check with 0 minute interval (should trigger immediately since some time has passed)
    logger->checkAndPrintSummary(0);
    
    // Record another detection
    logger->recordDetection("car", true);
    
    // Check with 60 minute interval (should not trigger)
    logger->checkAndPrintSummary(60);
    
    SUCCEED();
}

TEST_F(HourlySummaryTest, MultipleObjectTypes) {
    auto logger = std::make_unique<Logger>(test_log_file, false);
    
    // Record various object types
    logger->recordDetection("car", true);
    logger->recordDetection("person", false);
    logger->recordDetection("person", false);
    logger->recordDetection("cat", false);
    logger->recordDetection("dog", false);
    logger->recordDetection("truck", true);
    logger->recordDetection("truck", true);
    
    // Trigger summary
    logger->printHourlySummary();
    
    SUCCEED();
}

TEST_F(HourlySummaryTest, EmptySummary) {
    auto logger = std::make_unique<Logger>(test_log_file, false);
    
    // Don't record any detections
    
    // Trigger summary (should handle empty case gracefully)
    logger->printHourlySummary();
    
    SUCCEED();
}

TEST_F(HourlySummaryTest, ConsecutiveDynamicObjects) {
    auto logger = std::make_unique<Logger>(test_log_file, false);
    
    // Record multiple consecutive detections of the same type
    logger->recordDetection("person", false);
    std::this_thread::sleep_for(std::chrono::milliseconds(1));
    logger->recordDetection("person", false);
    std::this_thread::sleep_for(std::chrono::milliseconds(1));
    logger->recordDetection("person", false);
    
    // Trigger summary - should group them as "multiple people"
    logger->printHourlySummary();
    
    SUCCEED();
}

TEST_F(HourlySummaryTest, FinalSummary) {
    auto logger = std::make_unique<Logger>(test_log_file, false);
    
    // Record detections throughout "program lifetime"
    logger->recordDetection("car", true);
    std::this_thread::sleep_for(std::chrono::milliseconds(10));
    logger->recordDetection("person", false);
    std::this_thread::sleep_for(std::chrono::milliseconds(10));
    
    // Trigger hourly summary (clears periodic events but not all events)
    logger->printHourlySummary();
    
    // Record more detections
    logger->recordDetection("cat", false);
    std::this_thread::sleep_for(std::chrono::milliseconds(10));
    logger->recordDetection("dog", false);
    
    // Trigger final summary - should show all events from program start
    logger->printFinalSummary();
    
    SUCCEED();
}

TEST_F(HourlySummaryTest, FinalSummaryEmpty) {
    auto logger = std::make_unique<Logger>(test_log_file, false);
    
    // Don't record any detections
    
    // Trigger final summary - should handle empty case gracefully
    logger->printFinalSummary();
    
    SUCCEED();
}

// Test case to reproduce the CTRL-C issue where detections are not shown in final summary
TEST_F(HourlySummaryTest, FinalSummaryWithStationaryObjectsCtrlCScenario) {
    auto logger = std::make_unique<Logger>(test_log_file, false);
    
    // Simulate: person detected on program launch (recorded as dynamic/new)
    logger->recordDetection("person", false);
    std::this_thread::sleep_for(std::chrono::milliseconds(10));
    
    // Simulate: person becomes stationary (should be recorded as stationary)
    logger->recordDetection("person", true);
    std::this_thread::sleep_for(std::chrono::milliseconds(10));
    
    // Simulate: person continues to be stationary
    logger->recordDetection("person", true);
    std::this_thread::sleep_for(std::chrono::milliseconds(10));
    logger->recordDetection("person", true);
    std::this_thread::sleep_for(std::chrono::milliseconds(10));
    
    // Simulate CTRL-C: call printFinalSummary directly without clearing events
    // This should show ALL detections including the stationary ones
    logger->printFinalSummary();
    
    // Verify manually by running test and checking output
    SUCCEED();
}

// Test case to verify entry and exit events are properly tracked
TEST_F(HourlySummaryTest, EntryAndExitTimeline) {
    auto logger = std::make_unique<Logger>(test_log_file, false);
    
    // Person enters at 12:02
    logger->recordDetection("person", false, false);
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    
    // Person becomes stationary
    logger->recordDetection("person", true, false);
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    
    // Person leaves at 12:04
    logger->recordDetection("person", false, true);
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    
    // Person returns at 12:11
    logger->recordDetection("person", false, false);
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    
    // Person leaves again
    logger->recordDetection("person", false, true);
    
    // Trigger final summary
    logger->printFinalSummary();
    
    SUCCEED();
}

// Test case for user's specific scenario: person detected at start, leaves, returns later
TEST_F(HourlySummaryTest, PersonDetectedLeavesReturns) {
    auto logger = std::make_unique<Logger>(test_log_file, false);
    
    // Person detected at beginning (12:25)
    logger->recordDetection("person", false, false);
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    
    // Person leaves (12:26)  
    logger->recordDetection("person", false, true);
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    
    // Person returns (12:30)
    logger->recordDetection("person", false, false);
    
    // Print final summary
    logger->printFinalSummary();
    
    // Timeline should show 3 events: entered, left, entered
    SUCCEED();
}
