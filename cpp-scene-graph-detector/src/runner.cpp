#include "scene_graph/Runner.h"
#include "scene_graph/logger.h"
#include <chrono>

namespace scene_graph {

Runner::Runner() : initialized_(false), next_node_id_(0) {
}

Runner::~Runner() {
}

bool Runner::initialize(const RunnerConfig& config) {
    auto& logger = Logger::getInstance();
    
    config_ = config;
    
    // Initialize detector
    detector_ = std::make_unique<Detector>();
    if (!detector_->loadDetector(config.detector_model_path, config.labels_path, config.backend)) {
        logger.error("Failed to initialize detector");
        return false;
    }
    
    // Initialize relation predictor
    rel_predictor_ = std::make_unique<RelPredictor>();
    if (!config.relations_model_path.empty()) {
        if (!rel_predictor_->loadRelPredictor(config.relations_model_path)) {
            if (config.use_geometric_relations) {
                logger.info("Using geometric relation inference");
            } else {
                logger.error("Failed to load relation predictor and geometric relations disabled");
                return false;
            }
        }
    } else {
        logger.info("No relation model specified, using geometric relation inference");
    }
    
    initialized_ = true;
    stats_ = Stats();
    next_node_id_ = 0;
    
    logger.info("Runner initialized successfully");
    return true;
}

SceneGraph Runner::processImage(const cv::Mat& image) {
    SceneGraph graph;
    
    if (!initialized_) {
        Logger::getInstance().error("Runner not initialized");
        return graph;
    }
    
    auto start_time = std::chrono::high_resolution_clock::now();
    
    // Detect objects
    auto detections = detector_->infer(image, config_.obj_threshold, config_.max_objects);
    
    // Convert detections to nodes
    auto nodes = detectionsToNodes(detections);
    for (const auto& node : nodes) {
        graph.addNode(node);
    }
    
    // Predict relations
    auto relations = rel_predictor_->infer(detections, config_.rel_threshold);
    
    // Convert relations to edges
    auto edges = relationsToEdges(relations);
    for (const auto& edge : edges) {
        graph.addEdge(edge);
    }
    
    auto end_time = std::chrono::high_resolution_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end_time - start_time);
    
    // Update stats
    stats_.total_frames_processed++;
    stats_.total_objects_detected += static_cast<int>(nodes.size());
    stats_.total_relations_found += static_cast<int>(edges.size());
    stats_.avg_processing_time_ms = 
        (stats_.avg_processing_time_ms * (stats_.total_frames_processed - 1) + duration.count()) 
        / stats_.total_frames_processed;
    
    Logger::getInstance().debug("Processed frame in " + std::to_string(duration.count()) + "ms");
    
    return graph;
}

std::vector<SceneGraph> Runner::processVideo(const std::string& video_path, int frames_per_second) {
    std::vector<SceneGraph> graphs;
    
    if (!initialized_) {
        Logger::getInstance().error("Runner not initialized");
        return graphs;
    }
    
    auto& logger = Logger::getInstance();
    
    cv::VideoCapture cap(video_path);
    if (!cap.isOpened()) {
        logger.error("Failed to open video file: " + video_path);
        return graphs;
    }
    
    double fps = cap.get(cv::CAP_PROP_FPS);
    int frame_interval = static_cast<int>(fps / frames_per_second);
    if (frame_interval < 1) frame_interval = 1;
    
    logger.info("Processing video at " + std::to_string(frames_per_second) + " FPS");
    logger.info("Original video FPS: " + std::to_string(fps) + ", processing every " 
               + std::to_string(frame_interval) + " frames");
    
    cv::Mat frame;
    int frame_count = 0;
    
    while (cap.read(frame)) {
        if (frame_count % frame_interval == 0) {
            auto graph = processImage(frame);
            graphs.push_back(graph);
            
            logger.debug("Processed frame " + std::to_string(frame_count) + ": " 
                        + std::to_string(graph.getNodes().size()) + " objects, "
                        + std::to_string(graph.getEdges().size()) + " relations");
        }
        frame_count++;
    }
    
    cap.release();
    
    logger.info("Processed " + std::to_string(graphs.size()) + " frames from video");
    
    return graphs;
}

std::vector<Node> Runner::detectionsToNodes(const std::vector<Detection>& detections) {
    std::vector<Node> nodes;
    
    for (const auto& det : detections) {
        Node node;
        node.id = next_node_id_++;
        node.class_id = det.class_id;
        node.label = det.label;
        node.score = det.confidence;
        
        // Convert cv::Rect to BBox (center format)
        node.bbox.x = det.box.x + det.box.width / 2.0f;
        node.bbox.y = det.box.y + det.box.height / 2.0f;
        node.bbox.width = static_cast<float>(det.box.width);
        node.bbox.height = static_cast<float>(det.box.height);
        
        nodes.push_back(node);
    }
    
    return nodes;
}

std::vector<Edge> Runner::relationsToEdges(const std::vector<RelationPrediction>& relations) {
    std::vector<Edge> edges;
    
    for (const auto& rel : relations) {
        Edge edge;
        edge.src_id = rel.subject_id;
        edge.dst_id = rel.object_id;
        edge.predicate_id = RelPredictor::getPredicateId(rel.predicate);
        edge.predicate_label = rel.predicate_label;
        edge.score = rel.confidence;
        
        edges.push_back(edge);
    }
    
    return edges;
}

} // namespace scene_graph
