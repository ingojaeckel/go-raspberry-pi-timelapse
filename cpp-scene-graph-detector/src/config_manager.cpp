#include "scene_graph/config_manager.h"
#include "scene_graph/logger.h"
#include <iostream>
#include <cstring>

namespace scene_graph {

ConfigManager::ConfigManager() {
    // Set defaults
    runner_config_.backend = Backend::CPU;
    runner_config_.obj_threshold = 0.25f;
    runner_config_.rel_threshold = 0.5f;
    runner_config_.max_objects = 128;
    runner_config_.use_geometric_relations = true;
}

ConfigManager::~ConfigManager() {
}

ConfigManager::ParseResult ConfigManager::parseArgs(int argc, char** argv) {
    for (int i = 1; i < argc; ++i) {
        std::string arg = argv[i];
        
        if (arg == "-h" || arg == "--help") {
            return ParseResult::HELP_REQUESTED;
        } else if (arg == "--input" && i + 1 < argc) {
            io_config_.input_path = argv[++i];
        } else if (arg == "--model.detector" && i + 1 < argc) {
            runner_config_.detector_model_path = argv[++i];
        } else if (arg == "--model.relations" && i + 1 < argc) {
            runner_config_.relations_model_path = argv[++i];
        } else if (arg == "--labels" && i + 1 < argc) {
            runner_config_.labels_path = argv[++i];
        } else if (arg == "--backend" && i + 1 < argc) {
            std::string backend_str = argv[++i];
            if (backend_str == "cpu") {
                runner_config_.backend = Backend::CPU;
            } else if (backend_str == "opencl") {
                runner_config_.backend = Backend::OPENCL;
            } else if (backend_str == "auto") {
                runner_config_.backend = Backend::AUTO;
            } else {
                Logger::getInstance().error("Unknown backend: " + backend_str);
                return ParseResult::ERROR;
            }
        } else if (arg == "--threshold.obj" && i + 1 < argc) {
            runner_config_.obj_threshold = std::stof(argv[++i]);
        } else if (arg == "--threshold.rel" && i + 1 < argc) {
            runner_config_.rel_threshold = std::stof(argv[++i]);
        } else if (arg == "--max-objects" && i + 1 < argc) {
            runner_config_.max_objects = std::stoi(argv[++i]);
        } else if (arg == "--out.json" && i + 1 < argc) {
            io_config_.output_json = argv[++i];
        } else if (arg == "--out.dot" && i + 1 < argc) {
            io_config_.output_dot = argv[++i];
        } else if (arg == "--visualize" && i + 1 < argc) {
            io_config_.visualize_output = argv[++i];
        } else if (arg == "--video-fps" && i + 1 < argc) {
            io_config_.video_fps = std::stoi(argv[++i]);
        } else if (arg == "-v" || arg == "--verbose") {
            Logger::getInstance().setVerbose(true);
        } else {
            Logger::getInstance().error("Unknown argument: " + arg);
            return ParseResult::ERROR;
        }
    }
    
    return ParseResult::SUCCESS;
}

bool ConfigManager::validateConfig() const {
    auto& logger = Logger::getInstance();
    
    if (io_config_.input_path.empty()) {
        logger.error("Input path is required (--input)");
        return false;
    }
    
    if (runner_config_.detector_model_path.empty()) {
        logger.error("Detector model path is required (--model.detector)");
        return false;
    }
    
    if (runner_config_.labels_path.empty()) {
        logger.error("Labels file path is required (--labels)");
        return false;
    }
    
    if (runner_config_.obj_threshold < 0.0f || runner_config_.obj_threshold > 1.0f) {
        logger.error("Object threshold must be between 0.0 and 1.0");
        return false;
    }
    
    if (runner_config_.rel_threshold < 0.0f || runner_config_.rel_threshold > 1.0f) {
        logger.error("Relation threshold must be between 0.0 and 1.0");
        return false;
    }
    
    if (runner_config_.max_objects <= 0) {
        logger.error("Max objects must be greater than 0");
        return false;
    }
    
    return true;
}

void ConfigManager::displayHelp(const char* program_name) const {
    std::cout << "Usage: " << program_name << " [OPTIONS]\n\n";
    std::cout << "C++ Scene Graph Detector - Detects objects and spatial relations\n\n";
    std::cout << "Required Options:\n";
    std::cout << "  --input PATH                   Input image or video file\n";
    std::cout << "  --model.detector PATH          Path to object detector ONNX model\n";
    std::cout << "  --labels PATH                  Path to class labels file\n\n";
    std::cout << "Optional Model:\n";
    std::cout << "  --model.relations PATH         Path to relations ONNX model (optional)\n\n";
    std::cout << "Backend Options:\n";
    std::cout << "  --backend [cpu|opencl|auto]    Inference backend (default: cpu)\n\n";
    std::cout << "Detection Thresholds:\n";
    std::cout << "  --threshold.obj FLOAT          Object confidence threshold (default: 0.25)\n";
    std::cout << "  --threshold.rel FLOAT          Relation confidence threshold (default: 0.5)\n";
    std::cout << "  --max-objects INT              Maximum objects to detect (default: 128)\n\n";
    std::cout << "Output Options:\n";
    std::cout << "  --out.json PATH                Output JSON file path\n";
    std::cout << "  --out.dot PATH                 Output Graphviz DOT file path\n";
    std::cout << "  --visualize PATH               Output visualization image path\n\n";
    std::cout << "Video Options:\n";
    std::cout << "  --video-fps INT                Frames per second to process (default: 1)\n\n";
    std::cout << "Other Options:\n";
    std::cout << "  -v, --verbose                  Enable verbose logging\n";
    std::cout << "  -h, --help                     Show this help message\n\n";
    std::cout << "Examples:\n";
    std::cout << "  " << program_name << " \\\n";
    std::cout << "    --input image.jpg \\\n";
    std::cout << "    --model.detector models/yolov5s.onnx \\\n";
    std::cout << "    --labels models/coco.names \\\n";
    std::cout << "    --out.json output.json\n\n";
    std::cout << "  " << program_name << " \\\n";
    std::cout << "    --input video.mp4 \\\n";
    std::cout << "    --model.detector models/yolov5s.onnx \\\n";
    std::cout << "    --labels models/coco.names \\\n";
    std::cout << "    --backend opencl \\\n";
    std::cout << "    --video-fps 2 \\\n";
    std::cout << "    --out.json output.json \\\n";
    std::cout << "    --out.dot output.dot\n";
}

} // namespace scene_graph
