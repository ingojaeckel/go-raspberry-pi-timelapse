#include <iostream>
#include <string>
#include <opencv2/opencv.hpp>
#include "scene_graph/runner.h"

void printUsage(const char* program_name) {
    std::cout << "Usage: " << program_name << " [OPTIONS]\n\n";
    std::cout << "C++ Scene Graph Detector - Real-time webcam object detection and spatial relations\n\n";
    std::cout << "OPTIONS:\n";
    std::cout << "  --camera-id N             Webcam device ID (default: 0)\n";
    std::cout << "  --model.detector PATH     Path to object detector ONNX model (required)\n";
    std::cout << "  --model.relations PATH    Path to relations ONNX model (optional)\n";
    std::cout << "  --labels PATH             Path to class labels file (required)\n";
    std::cout << "  --backend TYPE            Backend: cpu, opencl, auto (default: cpu)\n";
    std::cout << "  --threshold.obj N         Object detection threshold 0-1 (default: 0.25)\n";
    std::cout << "  --threshold.rel N         Relation threshold 0-1 (default: 0.5)\n";
    std::cout << "  --max-objects N           Maximum objects to detect (default: 128)\n";
    std::cout << "  --output-dir PATH         Directory for saving scene change photos (default: output/)\n";
    std::cout << "  --fps N                   Frames per second for processing (default: 1)\n";
    std::cout << "  --show-preview            Show real-time preview with bounding boxes and scene graph (default: on)\n";
    std::cout << "  --no-preview              Disable real-time preview window\n";
    std::cout << "  --verbose                 Show detailed analysis timing information\n";
    std::cout << "  -h, --help                Show this help message\n";
    std::cout << "\nEXAMPLE:\n";
    std::cout << "  " << program_name << " \\\n";
    std::cout << "    --camera-id 0 \\\n";
    std::cout << "    --model.detector models/detector.onnx \\\n";
    std::cout << "    --labels assets/labels/coco.txt \\\n";
    std::cout << "    --show-preview --verbose\n";
}

int main(int argc, char** argv) {
    // Parse command line arguments
    std::string detector_model;
    std::string relations_model;
    std::string labels_path;
    std::string backend = "cpu";
    std::string output_dir = "output";
    float obj_threshold = 0.25f;
    float rel_threshold = 0.5f;
    int max_objects = 128;
    int camera_id = 0;  // Default to camera 0
    int fps = 1;
    bool show_preview = true;  // Default to showing preview
    bool verbose = false;  // Default to non-verbose
    
    for (int i = 1; i < argc; ++i) {
        std::string arg = argv[i];
        
        if (arg == "-h" || arg == "--help") {
            printUsage(argv[0]);
            return 0;
        } else if (arg == "--camera-id" && i + 1 < argc) {
            camera_id = std::stoi(argv[++i]);
        } else if (arg == "--model.detector" && i + 1 < argc) {
            detector_model = argv[++i];
        } else if (arg == "--model.relations" && i + 1 < argc) {
            relations_model = argv[++i];
        } else if (arg == "--labels" && i + 1 < argc) {
            labels_path = argv[++i];
        } else if (arg == "--backend" && i + 1 < argc) {
            backend = argv[++i];
        } else if (arg == "--threshold.obj" && i + 1 < argc) {
            obj_threshold = std::stof(argv[++i]);
        } else if (arg == "--threshold.rel" && i + 1 < argc) {
            rel_threshold = std::stof(argv[++i]);
        } else if (arg == "--max-objects" && i + 1 < argc) {
            max_objects = std::stoi(argv[++i]);
        } else if (arg == "--output-dir" && i + 1 < argc) {
            output_dir = argv[++i];
        } else if (arg == "--fps" && i + 1 < argc) {
            fps = std::stoi(argv[++i]);
        } else if (arg == "--show-preview") {
            show_preview = true;
        } else if (arg == "--no-preview") {
            show_preview = false;
        } else if (arg == "--verbose") {
            verbose = true;
        }
    }
    
    // Validate required arguments
    if (detector_model.empty()) {
        std::cerr << "Error: --model.detector is required\n";
        printUsage(argv[0]);
        return 1;
    }
    
    if (labels_path.empty()) {
        std::cerr << "Error: --labels is required\n";
        printUsage(argv[0]);
        return 1;
    }
    
    // Initialize runner
    scene_graph::RunnerConfig config;
    config.detector_model_path = detector_model;
    config.relation_model_path = relations_model;
    config.labels_path = labels_path;
    config.backend = backend;
    config.object_threshold = obj_threshold;
    config.relation_threshold = rel_threshold;
    config.max_objects = max_objects;
    config.show_preview = show_preview;
    config.verbose = verbose;
    
    scene_graph::Runner runner;
    if (!runner.initialize(config)) {
        std::cerr << "Failed to initialize runner\n";
        return 1;
    }
    
    std::cout << "Scene Graph Detector - Real-time Webcam Mode\n";
    std::cout << "Camera: " << camera_id << "\n";
    std::cout << "Backend: " << backend << "\n";
    std::cout << "Object threshold: " << obj_threshold << "\n";
    std::cout << "Relation threshold: " << rel_threshold << "\n";
    std::cout << "Output directory: " << output_dir << "\n";
    std::cout << "Verbose mode: " << (verbose ? "enabled" : "disabled") << "\n";
    if (show_preview) {
        std::cout << "Preview: enabled (Press ESC or 'q' to quit)\n\n";
    } else {
        std::cout << "Preview: disabled\n\n";
    }
    
    // Create output directory if it doesn't exist
    std::system(("mkdir -p " + output_dir).c_str());
    
    // Process webcam
    if (!runner.processWebcam(camera_id, output_dir, fps)) {
        std::cerr << "Failed to process webcam\n";
        return 1;
    }
    
    std::cout << "Done!\n";
    return 0;
}
