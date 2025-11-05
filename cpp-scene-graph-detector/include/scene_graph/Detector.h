#ifndef SCENE_GRAPH_DETECTOR_HPP
#define SCENE_GRAPH_DETECTOR_HPP

#include <string>
#include <vector>
#include <memory>
#include <opencv2/opencv.hpp>
#include "Graph.h"

namespace scene_graph {

// Detection result
struct Detection {
    int class_id;
    std::string label;
    float confidence;
    cv::Rect box;

    Detection() : class_id(0), label(""), confidence(0.0f), box() {}
    Detection(int cid, const std::string& lbl, float conf, const cv::Rect& b)
        : class_id(cid), label(lbl), confidence(conf), box(b) {}
};

// Backend enumeration
enum class Backend {
    CPU,
    OPENCL,
    AUTO
};

// Detector class for object detection
class Detector {
public:
    Detector();
    ~Detector();

    // Load detector model
    bool loadDetector(const std::string& model_path, 
                     const std::string& labels_path,
                     Backend backend = Backend::CPU);

    // Run inference on a frame
    std::vector<Detection> infer(const cv::Mat& frame, 
                                 float confidence_threshold = 0.25f,
                                 int max_objects = 128);

    // Check if detector is loaded
    bool isLoaded() const { return model_loaded_; }

    // Get backend info
    std::string getBackendInfo() const;

private:
    cv::dnn::Net net_;
    std::vector<std::string> class_labels_;
    Backend backend_;
    bool model_loaded_;

    // Model parameters
    int input_width_;
    int input_height_;
    
    // Post-processing
    std::vector<Detection> postProcess(const std::vector<cv::Mat>& outputs,
                                       const cv::Size& frame_size,
                                       float confidence_threshold,
                                       int max_objects);
    
    void applyNMS(std::vector<Detection>& detections, float nms_threshold = 0.45f);
};

} // namespace scene_graph

#endif // SCENE_GRAPH_DETECTOR_HPP
