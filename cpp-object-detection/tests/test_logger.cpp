#include <gtest/gtest.h>
#include <fstream>
#include <filesystem>
#include "logger.hpp"

class LoggerTest : public ::testing::Test {
protected:
    void SetUp() override {
        test_log_file = "/tmp/test_object_detection.log";
        // Remove any existing test log file
        std::filesystem::remove(test_log_file);
    }

    void TearDown() override {
        // Clean up test log file
        std::filesystem::remove(test_log_file);
    }

    std::string test_log_file;
};

TEST_F(LoggerTest, CreateLogger) {
    auto logger = std::make_unique<Logger>(test_log_file, false);
    EXPECT_TRUE(std::filesystem::exists(test_log_file));
}

TEST_F(LoggerTest, LogBasicMessage) {
    auto logger = std::make_unique<Logger>(test_log_file, false);
    logger->info("Test message");
    
    // Ensure the file was written to
    std::ifstream file(test_log_file);
    EXPECT_TRUE(file.good());
    
    std::string line;
    std::getline(file, line);
    EXPECT_TRUE(line.find("Test message") != std::string::npos);
    EXPECT_TRUE(line.find("[INFO]") != std::string::npos);
}

TEST_F(LoggerTest, LogObjectDetection) {
    auto logger = std::make_unique<Logger>(test_log_file, false);
    logger->logObjectDetection("person", "entered", 0.85);
    
    std::ifstream file(test_log_file);
    std::string line;
    std::getline(file, line);
    
    EXPECT_TRUE(line.find("person entered frame") != std::string::npos);
    EXPECT_TRUE(line.find("85% confidence") != std::string::npos);
}

TEST_F(LoggerTest, LogPerformanceWarning) {
    auto logger = std::make_unique<Logger>(test_log_file, false);
    logger->logPerformanceWarning(0.5, 1.0);
    
    std::ifstream file(test_log_file);
    std::string line;
    std::getline(file, line);
    
    EXPECT_TRUE(line.find("Performance warning") != std::string::npos);
    EXPECT_TRUE(line.find("0.50 fps") != std::string::npos);
    EXPECT_TRUE(line.find("threshold of 1.00 fps") != std::string::npos);
}

TEST_F(LoggerTest, LogHeartbeat) {
    auto logger = std::make_unique<Logger>(test_log_file, false);
    logger->logHeartbeat();
    
    std::ifstream file(test_log_file);
    std::string line;
    std::getline(file, line);
    
    EXPECT_TRUE(line.find("Detection system operational - heartbeat") != std::string::npos);
}

TEST_F(LoggerTest, LogLevels) {
    auto logger = std::make_unique<Logger>(test_log_file, true); // verbose mode
    
    logger->debug("Debug message");
    logger->info("Info message");
    logger->warning("Warning message");
    logger->error("Error message");
    
    std::ifstream file(test_log_file);
    std::string line;
    
    // Check debug message
    std::getline(file, line);
    EXPECT_TRUE(line.find("[DEBUG]") != std::string::npos);
    EXPECT_TRUE(line.find("Debug message") != std::string::npos);
    
    // Check info message
    std::getline(file, line);
    EXPECT_TRUE(line.find("[INFO]") != std::string::npos);
    EXPECT_TRUE(line.find("Info message") != std::string::npos);
    
    // Check warning message
    std::getline(file, line);
    EXPECT_TRUE(line.find("[WARNING]") != std::string::npos);
    EXPECT_TRUE(line.find("Warning message") != std::string::npos);
    
    // Check error message
    std::getline(file, line);
    EXPECT_TRUE(line.find("[ERROR]") != std::string::npos);
    EXPECT_TRUE(line.find("Error message") != std::string::npos);
}