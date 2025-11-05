#ifndef SCENE_GRAPH_RUNNER_HPP
#define SCENE_GRAPH_RUNNER_HPP

#include <string>
#include <memory>
#include <opencv2/opencv.hpp>
#include "Graph.h"
#include "Detector.h"
#include "RelPredictor.h"

namespace scene_graph {

// Configuration for the runner
struct RunnerConfig {
    std::string detector_model_path;
    std::string relations_model_path;
    std::string labels_path;
    Backend backend;
    float obj_threshold;
    float rel_threshold;
    int max_objects;
    bool use_geometric_relations;  // Use geometric relations if no model

    RunnerConfig() 
        : detector_model_path(""),
          relations_model_path(""),
          labels_path(""),
          backend(Backend::CPU),
          obj_threshold(0.25f),
          rel_threshold(0.5f),
          max_objects(128),
          use_geometric_relations(true) {}
};

// Runner orchestrates the scene graph pipeline
class Runner {
public:
    Runner();
    ~Runner();

    // Initialize with configuration
    bool initialize(const RunnerConfig& config);

    // Process a single image
    SceneGraph processImage(const cv::Mat& image);

    // Process video (returns graph for each frame processed)
    std::vector<SceneGraph> processVideo(const std::string& video_path, 
                                         int frames_per_second = 1);

    // Check if runner is initialized
    bool isInitialized() const { return initialized_; }

    // Get statistics
    struct Stats {
        int total_frames_processed;
        int total_objects_detected;
        int total_relations_found;
        double avg_processing_time_ms;

        Stats() : total_frames_processed(0), total_objects_detected(0), 
                  total_relations_found(0), avg_processing_time_ms(0.0) {}
    };

    Stats getStats() const { return stats_; }

private:
    std::unique_ptr<Detector> detector_;
    std::unique_ptr<RelPredictor> rel_predictor_;
    RunnerConfig config_;
    bool initialized_;
    Stats stats_;

    int next_node_id_;

    // Convert detections to nodes
    std::vector<Node> detectionsToNodes(const std::vector<Detection>& detections);

    // Convert relation predictions to edges
    std::vector<Edge> relationsToEdges(const std::vector<RelationPrediction>& relations);
};

} // namespace scene_graph

#endif // SCENE_GRAPH_RUNNER_HPP
