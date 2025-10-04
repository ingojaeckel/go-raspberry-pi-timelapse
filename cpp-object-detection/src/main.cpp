#include <iostream>
#include <atomic>

#include "application_context.hpp"

// Global flag for clean shutdown - must be defined here as it's used by signal handler
std::atomic<bool> running{true};

int main(int argc, char* argv[]) {
    setupSignalHandlers();

    try {
        ApplicationContext ctx;
        
        if (!parseAndValidateConfig(ctx, argc, argv)) {
            return 1;
        }
        
        if (!initializeComponents(ctx)) {
            return 1;
        }
        
        runMainProcessingLoop(ctx);
        performGracefulShutdown(ctx);

    } catch (const std::exception& e) {
        std::cerr << "Fatal error: " << e.what() << std::endl;
        return 1;
    } catch (...) {
        std::cerr << "Unknown fatal error occurred" << std::endl;
        return 1;
    }

    return 0;
}