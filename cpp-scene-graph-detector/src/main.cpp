#include <iostream>
#include <opencv2/opencv.hpp>
#include "scene_graph/Runner.h"
#include "scene_graph/config_manager.h"
#include "scene_graph/logger.h"

int main(int argc, char** argv) {
    using namespace scene_graph;
    
    auto& logger = Logger::getInstance();
    
    // Parse command-line arguments
    ConfigManager config_manager;
    auto parse_result = config_manager.parseArgs(argc, argv);
    
    if (parse_result == ConfigManager::ParseResult::HELP_REQUESTED) {
        config_manager.displayHelp(argv[0]);
        return 0;
    }
    
    if (parse_result == ConfigManager::ParseResult::ERROR) {
        logger.error("Failed to parse command-line arguments");
        config_manager.displayHelp(argv[0]);
        return 1;
    }
    
    // Validate configuration
    if (!config_manager.validateConfig()) {
        logger.error("Invalid configuration");
        config_manager.displayHelp(argv[0]);
        return 1;
    }
    
    logger.info("Starting C++ Scene Graph Detector");
    
    // Initialize runner
    Runner runner;
    if (!runner.initialize(config_manager.getRunnerConfig())) {
        logger.error("Failed to initialize runner");
        return 1;
    }
    
    const auto& io_config = config_manager.getIOConfig();
    
    // Check if input is video or image
    cv::Mat test_frame = cv::imread(io_config.input_path);
    bool is_video = test_frame.empty();
    
    if (is_video) {
        // Process video
        logger.info("Processing video: " + io_config.input_path);
        auto graphs = runner.processVideo(io_config.input_path, io_config.video_fps);
        
        if (graphs.empty()) {
            logger.error("Failed to process video or no frames extracted");
            return 1;
        }
        
        logger.info("Processed " + std::to_string(graphs.size()) + " frames");
        
        // Save results from last frame (or implement rolling save)
        if (!graphs.empty()) {
            const auto& last_graph = graphs.back();
            
            if (!io_config.output_json.empty()) {
                if (last_graph.saveJSON(io_config.output_json)) {
                    logger.info("Saved JSON to: " + io_config.output_json);
                } else {
                    logger.error("Failed to save JSON");
                }
            }
            
            if (!io_config.output_dot.empty()) {
                if (last_graph.saveDOT(io_config.output_dot)) {
                    logger.info("Saved DOT to: " + io_config.output_dot);
                } else {
                    logger.error("Failed to save DOT");
                }
            }
            
            // Print summary
            std::cout << "\n" << last_graph.getSummary() << std::endl;
        }
        
    } else {
        // Process single image
        logger.info("Processing image: " + io_config.input_path);
        auto graph = runner.processImage(test_frame);
        
        // Save results
        if (!io_config.output_json.empty()) {
            if (graph.saveJSON(io_config.output_json)) {
                logger.info("Saved JSON to: " + io_config.output_json);
            } else {
                logger.error("Failed to save JSON");
            }
        }
        
        if (!io_config.output_dot.empty()) {
            if (graph.saveDOT(io_config.output_dot)) {
                logger.info("Saved DOT to: " + io_config.output_dot);
            } else {
                logger.error("Failed to save DOT");
            }
        }
        
        // Print summary
        std::cout << "\n" << graph.getSummary() << std::endl;
        
        // Visualize if requested
        if (!io_config.visualize_output.empty()) {
            cv::Mat vis_image = test_frame.clone();
            
            // Draw bounding boxes
            for (const auto& node : graph.getNodes()) {
                int x = static_cast<int>(node.bbox.x - node.bbox.width / 2);
                int y = static_cast<int>(node.bbox.y - node.bbox.height / 2);
                int w = static_cast<int>(node.bbox.width);
                int h = static_cast<int>(node.bbox.height);
                
                cv::Rect rect(x, y, w, h);
                cv::rectangle(vis_image, rect, cv::Scalar(0, 255, 0), 2);
                
                std::string label = node.label + " (" + 
                                   std::to_string(static_cast<int>(node.score * 100)) + "%)";
                cv::putText(vis_image, label, cv::Point(x, y - 5),
                           cv::FONT_HERSHEY_SIMPLEX, 0.5, cv::Scalar(0, 255, 0), 1);
            }
            
            if (cv::imwrite(io_config.visualize_output, vis_image)) {
                logger.info("Saved visualization to: " + io_config.visualize_output);
            } else {
                logger.error("Failed to save visualization");
            }
        }
    }
    
    // Print statistics
    auto stats = runner.getStats();
    logger.info("Statistics:");
    logger.info("  Frames processed: " + std::to_string(stats.total_frames_processed));
    logger.info("  Total objects detected: " + std::to_string(stats.total_objects_detected));
    logger.info("  Total relations found: " + std::to_string(stats.total_relations_found));
    logger.info("  Average processing time: " + 
               std::to_string(stats.avg_processing_time_ms) + " ms");
    
    logger.info("Scene graph detection completed successfully");
    
    return 0;
}
