#ifndef SCENE_GRAPH_RUNNER_H
#define SCENE_GRAPH_RUNNER_H

#include <string>
#include <memory>
#include <opencv2/opencv.hpp>
#include "graph.h"
#include "detector.h"
#include "rel_predictor.h"

namespace scene_graph {

// Configuration for the runner
struct RunnerConfig {
    std::string detector_model_path;
    std::string relation_model_path;
    std::string labels_path;
    std::string backend;
    float object_threshold;
    float relation_threshold;
    int max_objects;
    bool show_preview;
    
    RunnerConfig() : backend("cpu"), object_threshold(0.25f), 
                    relation_threshold(0.5f), max_objects(128),
                    show_preview(false) {}
};

// Main pipeline runner
class Runner {
public:
    Runner();
    ~Runner();
    
    // Initialize with configuration
    bool initialize(const RunnerConfig& config);
    
    // Process single image
    SceneGraph processImage(const cv::Mat& image);
    
    // Process video file
    bool processVideo(const std::string& video_path,
                     const std::string& output_dir,
                     int frames_per_second = 1);
    
    // Process webcam stream
    bool processWebcam(int camera_id,
                      const std::string& output_dir,
                      int frames_per_second = 1);
    
    // Get current configuration
    const RunnerConfig& getConfig() const { return config_; }
    
    // Visualization
    cv::Mat visualize(const cv::Mat& image, const SceneGraph& graph);
    
private:
    RunnerConfig config_;
    std::unique_ptr<Detector> detector_;
    std::unique_ptr<RelPredictor> rel_predictor_;
    bool initialized_;
    
    // Scene change detection
    bool hasSceneChanged(const SceneGraph& prev, const SceneGraph& current);
    
    // Draw scene graph overlay
    void drawSceneGraphOverlay(cv::Mat& image, const SceneGraph& graph);
    void drawBoundingBoxes(cv::Mat& image, const SceneGraph& graph);
    void drawSceneDescription(cv::Mat& image, const SceneGraph& graph);
};

} // namespace scene_graph

#endif // SCENE_GRAPH_RUNNER_H
