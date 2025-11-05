#ifndef SCENE_GRAPH_CONFIG_MANAGER_HPP
#define SCENE_GRAPH_CONFIG_MANAGER_HPP

#include <string>
#include "scene_graph/Runner.h"

namespace scene_graph {

class ConfigManager {
public:
    ConfigManager();
    ~ConfigManager();

    // Parse command-line arguments
    enum class ParseResult {
        SUCCESS,
        HELP_REQUESTED,
        ERROR
    };

    ParseResult parseArgs(int argc, char** argv);

    // Validate configuration
    bool validateConfig() const;

    // Get configuration
    const RunnerConfig& getRunnerConfig() const { return runner_config_; }

    // Input/output configuration
    struct IOConfig {
        std::string input_path;
        std::string output_json;
        std::string output_dot;
        std::string visualize_output;
        bool is_video;
        int video_fps;

        IOConfig() : input_path(""), output_json(""), output_dot(""), 
                     visualize_output(""), is_video(false), video_fps(1) {}
    };

    const IOConfig& getIOConfig() const { return io_config_; }

    // Display help
    void displayHelp(const char* program_name) const;

private:
    RunnerConfig runner_config_;
    IOConfig io_config_;

    std::string backendFromString(const std::string& backend_str);
};

} // namespace scene_graph

#endif // SCENE_GRAPH_CONFIG_MANAGER_HPP
