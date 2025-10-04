#include <gtest/gtest.h>
#include "config_manager.hpp"

class ConfigManagerTest : public ::testing::Test {
protected:
    void SetUp() override {
        config_manager = std::make_unique<ConfigManager>();
    }

    std::unique_ptr<ConfigManager> config_manager;
};

TEST_F(ConfigManagerTest, DefaultConfiguration) {
    const auto& config = config_manager->getConfig();
    
    EXPECT_EQ(config.max_fps, 5);
    EXPECT_DOUBLE_EQ(config.min_confidence, 0.5);
    EXPECT_EQ(config.min_fps_warning_threshold, 1);
    EXPECT_EQ(config.heartbeat_interval_minutes, 10);
    EXPECT_EQ(config.camera_id, 0);
    EXPECT_EQ(config.frame_width, 1280);
    EXPECT_EQ(config.frame_height, 720);
    EXPECT_TRUE(config.headless);
    EXPECT_FALSE(config.verbose);
    EXPECT_FALSE(config.show_preview);  // Viewfinder disabled by default
}

TEST_F(ConfigManagerTest, ValidConfiguration) {
    EXPECT_TRUE(config_manager->validateConfig());
}

TEST_F(ConfigManagerTest, ParseValidArguments) {
    const char* argv[] = {
        "program",
        "--max-fps", "3",
        "--min-confidence", "0.7",
        "--verbose"
    };
    int argc = sizeof(argv) / sizeof(argv[0]);
    
    EXPECT_EQ(config_manager->parseArgs(argc, const_cast<char**>(argv)), ConfigManager::ParseResult::SUCCESS);
    
    const auto& config = config_manager->getConfig();
    EXPECT_EQ(config.max_fps, 3);
    EXPECT_DOUBLE_EQ(config.min_confidence, 0.7);
    EXPECT_TRUE(config.verbose);
}

TEST_F(ConfigManagerTest, InvalidMaxFps) {
    ConfigManager::Config config;
    config.max_fps = 0;
    
    // Manually create invalid config manager for testing
    // (This test verifies validation logic)
    const char* argv[] = {"program", "--max-fps", "0"};
    int argc = sizeof(argv) / sizeof(argv[0]);
    
    EXPECT_EQ(config_manager->parseArgs(argc, const_cast<char**>(argv)), ConfigManager::ParseResult::SUCCESS);
    EXPECT_FALSE(config_manager->validateConfig());
}

TEST_F(ConfigManagerTest, InvalidConfidence) {
    const char* argv[] = {"program", "--min-confidence", "1.5"};
    int argc = sizeof(argv) / sizeof(argv[0]);
    
    EXPECT_EQ(config_manager->parseArgs(argc, const_cast<char**>(argv)), ConfigManager::ParseResult::SUCCESS);
    EXPECT_FALSE(config_manager->validateConfig());
}

TEST_F(ConfigManagerTest, HelpArgument) {
    const char* argv[] = {"program", "--help"};
    int argc = sizeof(argv) / sizeof(argv[0]);
    
    EXPECT_EQ(config_manager->parseArgs(argc, const_cast<char**>(argv)), ConfigManager::ParseResult::HELP_REQUESTED);
}

TEST_F(ConfigManagerTest, ShowPreviewArgument) {
    const char* argv[] = {"program", "--show-preview"};
    int argc = sizeof(argv) / sizeof(argv[0]);
    
    EXPECT_EQ(config_manager->parseArgs(argc, const_cast<char**>(argv)), ConfigManager::ParseResult::SUCCESS);
    
    const auto& config = config_manager->getConfig();
    EXPECT_TRUE(config.show_preview);
}