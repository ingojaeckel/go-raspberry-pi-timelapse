#ifndef SCENE_GRAPH_DETECTOR_H
#define SCENE_GRAPH_DETECTOR_H

#include <string>
#include <vector>
#include <opencv2/opencv.hpp>
#include "graph.h"

namespace scene_graph {

// Object detection result
struct Detection {
    int class_id;
    std::string label;
    BBox bbox;
    float score;
    
    Detection() : class_id(-1), score(0.0f) {}
};

// Object detector interface using OpenCV DNN
class Detector {
public:
    Detector();
    ~Detector();
    
    // Initialize detector with ONNX model
    bool initialize(const std::string& model_path, 
                   const std::string& labels_path,
                   const std::string& backend = "cpu");
    
    // Run object detection on image
    std::vector<Detection> detect(const cv::Mat& image, float threshold = 0.25);
    
    // Get loaded class labels
    const std::vector<std::string>& getLabels() const { return labels_; }
    
    // Check if initialized
    bool isInitialized() const { return initialized_; }
    
    // Get the actual backend being used (for warning if fallback occurred)
    std::string getActualBackend() const { return actual_backend_; }
    
private:
    cv::dnn::Net net_;
    std::vector<std::string> labels_;
    bool initialized_;
    int input_width_;
    int input_height_;
    std::string actual_backend_;
    
    // Pre-processing and post-processing
    cv::Mat preprocessImage(const cv::Mat& image);
    std::vector<Detection> postprocess(const cv::Mat& blob, 
                                       const std::vector<cv::Mat>& outputs,
                                       const cv::Size& image_size,
                                       float threshold);
    
    // Non-maximum suppression
    void performNMS(std::vector<Detection>& detections, float nms_threshold = 0.45);
};

// Factory function to load detector
Detector* loadDetector(const std::string& model_path,
                      const std::string& labels_path,
                      const std::string& backend = "cpu");

} // namespace scene_graph

#endif // SCENE_GRAPH_DETECTOR_H
