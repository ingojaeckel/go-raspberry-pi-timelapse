#include <iostream>
#include <string>
#include <iomanip>
#include <opencv2/opencv.hpp>
#include "scene_graph/runner.h"

void printUsage(const char* program_name) {
    std::cout << "Usage: " << program_name << " [OPTIONS]\n\n";
    std::cout << "C++ Scene Graph Detector - Detect objects and spatial relations\n\n";
    std::cout << "OPTIONS:\n";
    std::cout << "  --input PATH              Input image or video file (required)\n";
    std::cout << "  --camera-id N             Use webcam instead of file (default: 0)\n";
    std::cout << "  --model.detector PATH     Path to object detector ONNX model (required)\n";
    std::cout << "  --model.relations PATH    Path to relations ONNX model (optional)\n";
    std::cout << "  --labels PATH             Path to class labels file (required)\n";
    std::cout << "  --backend TYPE            Backend: cpu, opencl, auto (default: cpu)\n";
    std::cout << "  --threshold.obj N         Object detection threshold 0-1 (default: 0.25)\n";
    std::cout << "  --threshold.rel N         Relation threshold 0-1 (default: 0.5)\n";
    std::cout << "  --max-objects N           Maximum objects to detect (default: 128)\n";
    std::cout << "  --out.json PATH           Output JSON file path (required)\n";
    std::cout << "  --out.dot PATH            Output Graphviz DOT file path (optional)\n";
    std::cout << "  --visualize PATH          Save visualization image (optional)\n";
    std::cout << "  --show-preview            Show real-time preview window\n";
    std::cout << "  --fps N                   Frames per second for video (default: 1)\n";
    std::cout << "  -h, --help                Show this help message\n";
    std::cout << "\nEXAMPLES:\n";
    std::cout << "  # Process single image\n";
    std::cout << "  " << program_name << " --input image.jpg \\\n";
    std::cout << "    --model.detector models/detector.onnx \\\n";
    std::cout << "    --labels assets/labels.txt \\\n";
    std::cout << "    --out.json output.json\n\n";
    std::cout << "  # Process webcam with preview\n";
    std::cout << "  " << program_name << " --camera-id 0 \\\n";
    std::cout << "    --model.detector models/detector.onnx \\\n";
    std::cout << "    --labels assets/labels.txt \\\n";
    std::cout << "    --out.json output/scene.json \\\n";
    std::cout << "    --show-preview\n";
}

int main(int argc, char** argv) {
    // Parse command line arguments
    std::string input_path;
    std::string detector_model;
    std::string relations_model;
    std::string labels_path;
    std::string backend = "cpu";
    std::string output_json;
    std::string output_dot;
    std::string visualize_path;
    float obj_threshold = 0.25f;
    float rel_threshold = 0.5f;
    int max_objects = 128;
    int camera_id = -1;
    int fps = 1;
    bool show_preview = false;
    bool use_camera = false;
    
    for (int i = 1; i < argc; ++i) {
        std::string arg = argv[i];
        
        if (arg == "-h" || arg == "--help") {
            printUsage(argv[0]);
            return 0;
        } else if (arg == "--input" && i + 1 < argc) {
            input_path = argv[++i];
        } else if (arg == "--camera-id" && i + 1 < argc) {
            camera_id = std::stoi(argv[++i]);
            use_camera = true;
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
        } else if (arg == "--out.json" && i + 1 < argc) {
            output_json = argv[++i];
        } else if (arg == "--out.dot" && i + 1 < argc) {
            output_dot = argv[++i];
        } else if (arg == "--visualize" && i + 1 < argc) {
            visualize_path = argv[++i];
        } else if (arg == "--show-preview") {
            show_preview = true;
        } else if (arg == "--fps" && i + 1 < argc) {
            fps = std::stoi(argv[++i]);
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
    
    if (!use_camera && input_path.empty()) {
        std::cerr << "Error: Either --input or --camera-id is required\n";
        printUsage(argv[0]);
        return 1;
    }
    
    if (output_json.empty() && !use_camera) {
        std::cerr << "Error: --out.json is required for image/video processing\n";
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
    
    scene_graph::Runner runner;
    if (!runner.initialize(config)) {
        std::cerr << "Failed to initialize runner\n";
        return 1;
    }
    
    std::cout << "Scene Graph Detector initialized\n";
    std::cout << "Backend: " << backend << "\n";
    std::cout << "Object threshold: " << obj_threshold << "\n";
    std::cout << "Relation threshold: " << rel_threshold << "\n";
    
    // Process input
    if (use_camera) {
        // Webcam mode
        std::cout << "Processing webcam " << camera_id << "...\n";
        std::string output_dir = output_json.empty() ? "output" : output_json;
        
        if (!runner.processWebcam(camera_id, output_dir, fps)) {
            std::cerr << "Failed to process webcam\n";
            return 1;
        }
    } else {
        // Check if input is video or image
        cv::VideoCapture cap(input_path);
        bool is_video = cap.isOpened() && cap.get(cv::CAP_PROP_FRAME_COUNT) > 1;
        cap.release();
        
        if (is_video) {
            // Video mode
            std::cout << "Processing video: " << input_path << "\n";
            std::string output_dir = output_json.substr(0, output_json.find_last_of("/\\"));
            if (output_dir.empty()) output_dir = ".";
            
            if (!runner.processVideo(input_path, output_dir, fps)) {
                std::cerr << "Failed to process video\n";
                return 1;
            }
        } else {
            // Image mode
            std::cout << "Processing image: " << input_path << "\n";
            cv::Mat image = cv::imread(input_path);
            if (image.empty()) {
                std::cerr << "Failed to load image: " << input_path << "\n";
                return 1;
            }
            
            scene_graph::SceneGraph graph = runner.processImage(image);
            
            // Save JSON
            graph.toJSON(output_json);
            std::cout << "Saved scene graph to: " << output_json << "\n";
            
            // Save DOT if requested
            if (!output_dot.empty()) {
                graph.toDOT(output_dot);
                std::cout << "Saved Graphviz DOT to: " << output_dot << "\n";
            }
            
            // Save visualization if requested
            if (!visualize_path.empty()) {
                cv::Mat vis = runner.visualize(image, graph);
                cv::imwrite(visualize_path, vis);
                std::cout << "Saved visualization to: " << visualize_path << "\n";
            }
            
            // Print summary
            std::cout << "\nScene Graph Summary:\n";
            std::cout << "Objects detected: " << graph.getNodeCount() << "\n";
            std::cout << "Relations found: " << graph.getEdgeCount() << "\n";
            
            for (const auto& node : graph.getNodes()) {
                std::cout << "  - " << node.label << " (confidence: " 
                         << std::fixed << std::setprecision(2) << node.score << ")\n";
            }
        }
    }
    
    std::cout << "Done!\n";
    return 0;
}
