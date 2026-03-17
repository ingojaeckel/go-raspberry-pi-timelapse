#include "scene_graph/runner.h"
#include <iostream>
#include <chrono>
#include <sstream>
#include <iomanip>

namespace scene_graph {

Runner::Runner() : initialized_(false) {}

Runner::~Runner() {}

bool Runner::initialize(const RunnerConfig& config) {
    config_ = config;
    
    // Initialize detector
    detector_.reset(loadDetector(config.detector_model_path, 
                                  config.labels_path,
                                  config.backend));
    if (!detector_) {
        std::cerr << "Failed to initialize detector" << std::endl;
        return false;
    }
    
    // Check if backend matches requested backend and warn if fallback occurred
    std::string actual_backend = detector_->getActualBackend();
    if (config.backend == "opencl" && actual_backend == "cpu") {
#ifdef __APPLE__
        std::cerr << "WARNING: OpenCL backend requested but not available on macOS." << std::endl;
        std::cerr << "         OpenCL support on macOS may require additional configuration." << std::endl;
        std::cerr << "         Falling back to CPU backend. Performance will be reduced." << std::endl;
#else
        std::cerr << "WARNING: OpenCL backend requested but failed to initialize." << std::endl;
        std::cerr << "         Falling back to CPU backend. Check OpenCL drivers/runtime." << std::endl;
#endif
    }
    
    // Initialize relation predictor
    rel_predictor_.reset(loadRelPredictor(config.relation_model_path,
                                          config.backend));
    if (!rel_predictor_) {
        std::cerr << "Failed to initialize relation predictor" << std::endl;
        return false;
    }
    
    initialized_ = true;
    return true;
}

SceneGraph Runner::processImage(const cv::Mat& image, double& elapsed_ms) {
    SceneGraph graph;
    
    if (!initialized_) {
        std::cerr << "Runner not initialized" << std::endl;
        elapsed_ms = 0.0;
        return graph;
    }
    
    auto start_time = std::chrono::high_resolution_clock::now();
    
    // Detect objects
    std::vector<Detection> detections = detector_->detect(image, config_.object_threshold);
    
    // Limit to max objects
    if (detections.size() > static_cast<size_t>(config_.max_objects)) {
        detections.resize(config_.max_objects);
    }
    
    // Create nodes
    for (size_t i = 0; i < detections.size(); ++i) {
        Node node;
        node.id = static_cast<int>(i);
        node.class_id = detections[i].class_id;
        node.label = detections[i].label;
        node.bbox = detections[i].bbox;
        node.score = detections[i].score;
        graph.addNode(node);
    }
    
    // Predict relations
    std::vector<RelationPrediction> relations = rel_predictor_->predict(
        detections, image, config_.relation_threshold);
    
    // Create edges
    for (const auto& rel : relations) {
        Edge edge;
        edge.src_id = rel.subject_id;
        edge.dst_id = rel.object_id;
        edge.predicate_id = static_cast<int>(rel.predicate);
        edge.label = rel.predicate_label;
        edge.score = rel.score;
        graph.addEdge(edge);
    }
    
    // Calculate elapsed time
    auto end_time = std::chrono::high_resolution_clock::now();
    std::chrono::duration<double, std::milli> duration = end_time - start_time;
    elapsed_ms = duration.count();
    
    // Add metadata
    auto now = std::chrono::system_clock::now();
    auto time_t = std::chrono::system_clock::to_time_t(now);
    std::ostringstream oss;
    oss << std::put_time(std::localtime(&time_t), "%Y-%m-%d %H:%M:%S");
    graph.setMetadata("timestamp", oss.str());
    graph.setMetadata("image_width", std::to_string(image.cols));
    graph.setMetadata("image_height", std::to_string(image.rows));
    graph.setMetadata("analysis_time_ms", std::to_string(elapsed_ms));
    
    return graph;
}

bool Runner::processWebcam(int camera_id,
                          const std::string& output_dir,
                          int frames_per_second) {
    cv::VideoCapture cap(camera_id);
    if (!cap.isOpened()) {
        std::cerr << "Failed to open camera: " << camera_id << std::endl;
        return false;
    }
    
    // Set resolution
    cap.set(cv::CAP_PROP_FRAME_WIDTH, 1280);
    cap.set(cv::CAP_PROP_FRAME_HEIGHT, 720);
    
    int frame_delay = 1000 / frames_per_second;
    int frame_count = 0;
    cv::Mat frame;
    SceneGraph prev_graph;
    
    while (true) {
        if (!cap.read(frame)) {
            std::cerr << "Failed to read frame" << std::endl;
            break;
        }
        
        double elapsed_ms = 0.0;
        SceneGraph graph = processImage(frame, elapsed_ms);
        
        // Log timing if verbose
        if (config_.verbose) {
            std::cout << "Frame " << frame_count << ": Analysis took " 
                     << std::fixed << std::setprecision(2) << elapsed_ms 
                     << " ms | Objects: " << graph.getNodeCount() 
                     << " | Relations: " << graph.getEdgeCount() << std::endl;
        }
        
        // Show preview if requested
        if (config_.show_preview) {
            cv::Mat vis = visualize(frame, graph);
            drawAnalysisTime(vis, elapsed_ms);
            cv::imshow("Scene Graph Detector", vis);
            
            int key = cv::waitKey(frame_delay);
            if (key == 27 || key == 'q') { // ESC or 'q' to quit
                break;
            }
        } else {
            cv::waitKey(frame_delay);
        }
        
        // Check if scene changed
        if (hasSceneChanged(prev_graph, graph)) {
            // Save outputs
            std::ostringstream oss;
            oss << output_dir << "/scene_" << std::setw(6) << std::setfill('0') 
                << frame_count << ".json";
            graph.toJSON(oss.str());
            
            // Save visualization with timing
            cv::Mat vis = visualize(frame, graph);
            drawAnalysisTime(vis, elapsed_ms);
            std::ostringstream img_oss;
            img_oss << output_dir << "/scene_" << std::setw(6) << std::setfill('0') 
                    << frame_count << ".jpg";
            cv::imwrite(img_oss.str(), vis);
            
            std::cout << "Scene changed - saved outputs";
            if (config_.verbose) {
                std::cout << " (analysis: " << std::fixed << std::setprecision(2) 
                         << elapsed_ms << " ms)";
            }
            std::cout << std::endl;
            
            prev_graph = graph;
        }
        
        frame_count++;
    }
    
    cap.release();
    cv::destroyAllWindows();
    return true;
}

bool Runner::hasSceneChanged(const SceneGraph& prev, const SceneGraph& current) {
    // Simple scene change detection: different number or types of objects
    if (prev.getNodeCount() == 0) {
        return true; // First frame
    }
    
    if (prev.getNodeCount() != current.getNodeCount()) {
        return true;
    }
    
    // Check if object types have changed
    std::vector<std::string> prev_labels, curr_labels;
    for (const auto& node : prev.getNodes()) {
        prev_labels.push_back(node.label);
    }
    for (const auto& node : current.getNodes()) {
        curr_labels.push_back(node.label);
    }
    
    std::sort(prev_labels.begin(), prev_labels.end());
    std::sort(curr_labels.begin(), curr_labels.end());
    
    return (prev_labels != curr_labels);
}

cv::Mat Runner::visualize(const cv::Mat& image, const SceneGraph& graph) {
    cv::Mat vis = image.clone();
    
    drawBoundingBoxes(vis, graph);
    drawSceneDescription(vis, graph);
    
    return vis;
}

void Runner::drawBoundingBoxes(cv::Mat& image, const SceneGraph& graph) {
    for (const auto& node : graph.getNodes()) {
        // Convert normalized coordinates to pixel coordinates
        float cx = node.bbox.x * image.cols;
        float cy = node.bbox.y * image.rows;
        float w = node.bbox.width * image.cols;
        float h = node.bbox.height * image.rows;
        
        int x1 = static_cast<int>(cx - w/2);
        int y1 = static_cast<int>(cy - h/2);
        int x2 = static_cast<int>(cx + w/2);
        int y2 = static_cast<int>(cy + h/2);
        
        // Draw bounding box
        cv::rectangle(image, cv::Point(x1, y1), cv::Point(x2, y2), 
                     cv::Scalar(0, 255, 0), 2);
        
        // Draw label
        std::ostringstream label_oss;
        label_oss << node.label << " " << std::fixed << std::setprecision(2) << node.score;
        std::string label = label_oss.str();
        
        int baseline = 0;
        cv::Size text_size = cv::getTextSize(label, cv::FONT_HERSHEY_SIMPLEX, 
                                             0.5, 1, &baseline);
        
        cv::rectangle(image, 
                     cv::Point(x1, y1 - text_size.height - 5),
                     cv::Point(x1 + text_size.width, y1),
                     cv::Scalar(0, 255, 0), -1);
        
        cv::putText(image, label, cv::Point(x1, y1 - 5),
                   cv::FONT_HERSHEY_SIMPLEX, 0.5, cv::Scalar(0, 0, 0), 1);
    }
}

void Runner::drawSceneDescription(cv::Mat& image, const SceneGraph& graph) {
    // Draw scene graph text at the bottom
    std::ostringstream desc;
    desc << "Objects: " << graph.getNodeCount() << " | Relations: " << graph.getEdgeCount();
    
    // Add some relations
    int rel_count = 0;
    for (const auto& edge : graph.getEdges()) {
        if (rel_count >= 3) break; // Show max 3 relations
        
        const Node* src = graph.findNode(edge.src_id);
        const Node* dst = graph.findNode(edge.dst_id);
        
        if (src && dst) {
            desc << " | " << src->label << " " << edge.label << " " << dst->label;
            rel_count++;
        }
    }
    
    std::string text = desc.str();
    
    int baseline = 0;
    cv::Size text_size = cv::getTextSize(text, cv::FONT_HERSHEY_SIMPLEX, 
                                         0.6, 1, &baseline);
    
    int text_y = image.rows - 10;
    
    // Draw background
    cv::rectangle(image, 
                 cv::Point(0, text_y - text_size.height - 5),
                 cv::Point(image.cols, image.rows),
                 cv::Scalar(0, 0, 0), -1);
    
    // Draw text
    cv::putText(image, text, cv::Point(10, text_y),
               cv::FONT_HERSHEY_SIMPLEX, 0.6, cv::Scalar(255, 255, 255), 1);
}

void Runner::drawAnalysisTime(cv::Mat& image, double elapsed_ms) {
    // Draw analysis time in top-right corner
    std::ostringstream oss;
    oss << std::fixed << std::setprecision(1) << elapsed_ms << " ms";
    std::string text = oss.str();
    
    int baseline = 0;
    cv::Size text_size = cv::getTextSize(text, cv::FONT_HERSHEY_SIMPLEX, 
                                         0.7, 2, &baseline);
    
    int text_x = image.cols - text_size.width - 15;
    int text_y = 30;
    
    // Draw semi-transparent background
    cv::Mat roi = image(cv::Rect(text_x - 5, text_y - text_size.height - 5, 
                                 text_size.width + 10, text_size.height + 10));
    cv::Mat color(roi.size(), CV_8UC3, cv::Scalar(0, 0, 0));
    cv::addWeighted(color, 0.6, roi, 0.4, 0, roi);
    
    // Draw text
    cv::putText(image, text, cv::Point(text_x, text_y),
               cv::FONT_HERSHEY_SIMPLEX, 0.7, cv::Scalar(0, 255, 255), 2);
}

} // namespace scene_graph
