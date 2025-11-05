#ifndef SCENE_GRAPH_LOGGER_HPP
#define SCENE_GRAPH_LOGGER_HPP

#include <string>
#include <fstream>
#include <memory>

namespace scene_graph {

enum class LogLevel {
    DEBUG,
    INFO,
    WARNING,
    ERROR
};

class Logger {
public:
    static Logger& getInstance();

    // Set log level
    void setLogLevel(LogLevel level);

    // Enable/disable verbose mode
    void setVerbose(bool verbose);

    // Log messages
    void log(LogLevel level, const std::string& message);
    void debug(const std::string& message);
    void info(const std::string& message);
    void warning(const std::string& message);
    void error(const std::string& message);

private:
    Logger();
    ~Logger();
    Logger(const Logger&) = delete;
    Logger& operator=(const Logger&) = delete;

    LogLevel min_level_;
    bool verbose_;

    std::string getLevelString(LogLevel level) const;
    std::string getTimestamp() const;
};

} // namespace scene_graph

#endif // SCENE_GRAPH_LOGGER_HPP
