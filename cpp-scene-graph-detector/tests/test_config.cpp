#include <gtest/gtest.h>
#include "scene_graph/config_manager.h"

using namespace scene_graph;

class ConfigManagerTest : public ::testing::Test {
protected:
    void SetUp() override {
        config = std::make_unique<ConfigManager>();
    }

    std::unique_ptr<ConfigManager> config;
};

TEST_F(ConfigManagerTest, DefaultConfig) {
    const auto& runner_config = config->getRunnerConfig();
    
    EXPECT_EQ(runner_config.backend, Backend::CPU);
    EXPECT_FLOAT_EQ(runner_config.obj_threshold, 0.25f);
    EXPECT_FLOAT_EQ(runner_config.rel_threshold, 0.5f);
    EXPECT_EQ(runner_config.max_objects, 128);
    EXPECT_TRUE(runner_config.use_geometric_relations);
}

TEST_F(ConfigManagerTest, ParseValidArgs) {
    const char* argv[] = {
        "program",
        "--input", "test.jpg",
        "--model.detector", "model.onnx",
        "--labels", "labels.txt",
        "--threshold.obj", "0.3",
        "--threshold.rel", "0.6",
        "--max-objects", "64"
    };
    int argc = sizeof(argv) / sizeof(argv[0]);
    
    auto result = config->parseArgs(argc, const_cast<char**>(argv));
    
    EXPECT_EQ(result, ConfigManager::ParseResult::SUCCESS);
    
    const auto& io_config = config->getIOConfig();
    const auto& runner_config = config->getRunnerConfig();
    
    EXPECT_EQ(io_config.input_path, "test.jpg");
    EXPECT_EQ(runner_config.detector_model_path, "model.onnx");
    EXPECT_EQ(runner_config.labels_path, "labels.txt");
    EXPECT_FLOAT_EQ(runner_config.obj_threshold, 0.3f);
    EXPECT_FLOAT_EQ(runner_config.rel_threshold, 0.6f);
    EXPECT_EQ(runner_config.max_objects, 64);
}

TEST_F(ConfigManagerTest, ParseBackends) {
    {
        const char* argv[] = {"program", "--backend", "cpu"};
        config->parseArgs(3, const_cast<char**>(argv));
        EXPECT_EQ(config->getRunnerConfig().backend, Backend::CPU);
    }
    
    {
        auto config2 = std::make_unique<ConfigManager>();
        const char* argv[] = {"program", "--backend", "opencl"};
        config2->parseArgs(3, const_cast<char**>(argv));
        EXPECT_EQ(config2->getRunnerConfig().backend, Backend::OPENCL);
    }
    
    {
        auto config3 = std::make_unique<ConfigManager>();
        const char* argv[] = {"program", "--backend", "auto"};
        config3->parseArgs(3, const_cast<char**>(argv));
        EXPECT_EQ(config3->getRunnerConfig().backend, Backend::AUTO);
    }
}

TEST_F(ConfigManagerTest, ParseHelpRequest) {
    const char* argv[] = {"program", "--help"};
    int argc = sizeof(argv) / sizeof(argv[0]);
    
    auto result = config->parseArgs(argc, const_cast<char**>(argv));
    
    EXPECT_EQ(result, ConfigManager::ParseResult::HELP_REQUESTED);
}

TEST_F(ConfigManagerTest, ValidationRequiresInput) {
    EXPECT_FALSE(config->validateConfig());
}

TEST_F(ConfigManagerTest, ValidationRequiresDetector) {
    const char* argv[] = {
        "program",
        "--input", "test.jpg"
    };
    config->parseArgs(3, const_cast<char**>(argv));
    
    EXPECT_FALSE(config->validateConfig());
}

TEST_F(ConfigManagerTest, ValidationRequiresLabels) {
    const char* argv[] = {
        "program",
        "--input", "test.jpg",
        "--model.detector", "model.onnx"
    };
    config->parseArgs(5, const_cast<char**>(argv));
    
    EXPECT_FALSE(config->validateConfig());
}

TEST_F(ConfigManagerTest, ValidConfigPasses) {
    const char* argv[] = {
        "program",
        "--input", "test.jpg",
        "--model.detector", "model.onnx",
        "--labels", "labels.txt"
    };
    config->parseArgs(7, const_cast<char**>(argv));
    
    EXPECT_TRUE(config->validateConfig());
}
