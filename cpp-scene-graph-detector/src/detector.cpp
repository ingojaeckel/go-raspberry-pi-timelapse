#include "scene_graph/detector.h"
#include <fstream>
#include <algorithm>
#include <iostream>

namespace scene_graph {

Detector::Detector() : initialized_(false), input_width_(640), input_height_(640) {}

Detector::~Detector() {}

bool Detector::initialize(const std::string& model_path, 
                         const std::string& labels_path,
                         const std::string& backend) {
    try {
        // Load ONNX model
        net_ = cv::dnn::readNetFromONNX(model_path);
        if (net_.empty()) {
            std::cerr << "Failed to load model from: " << model_path << std::endl;
            return false;
        }
        
        // Set backend and target
        if (backend == "opencl") {
            net_.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
            net_.setPreferableTarget(cv::dnn::DNN_TARGET_OPENCL);
        } else if (backend == "auto") {
            // Try OpenCL first, fall back to CPU
            try {
                net_.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
                net_.setPreferableTarget(cv::dnn::DNN_TARGET_OPENCL);
            } catch (...) {
                net_.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
                net_.setPreferableTarget(cv::dnn::DNN_TARGET_CPU);
            }
        } else {
            // CPU backend (default)
            net_.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
            net_.setPreferableTarget(cv::dnn::DNN_TARGET_CPU);
        }
        
        // Load labels
        std::ifstream ifs(labels_path);
        if (!ifs.is_open()) {
            std::cerr << "Failed to load labels from: " << labels_path << std::endl;
            return false;
        }
        
        std::string label;
        while (std::getline(ifs, label)) {
            if (!label.empty()) {
                labels_.push_back(label);
            }
        }
        ifs.close();
        
        initialized_ = true;
        return true;
    } catch (const cv::Exception& e) {
        std::cerr << "OpenCV exception in Detector::initialize: " << e.what() << std::endl;
        return false;
    } catch (const std::exception& e) {
        std::cerr << "Exception in Detector::initialize: " << e.what() << std::endl;
        return false;
    }
}

cv::Mat Detector::preprocessImage(const cv::Mat& image) {
    // Create blob from image (letterbox resize to input size)
    cv::Mat blob = cv::dnn::blobFromImage(image, 1.0/255.0, 
                                         cv::Size(input_width_, input_height_),
                                         cv::Scalar(0, 0, 0), true, false);
    return blob;
}

std::vector<Detection> Detector::detect(const cv::Mat& image, float threshold) {
    std::vector<Detection> detections;
    
    if (!initialized_) {
        std::cerr << "Detector not initialized" << std::endl;
        return detections;
    }
    
    try {
        // Preprocess
        cv::Mat blob = preprocessImage(image);
        net_.setInput(blob);
        
        // Forward pass
        std::vector<cv::Mat> outputs;
        net_.forward(outputs, net_.getUnconnectedOutLayersNames());
        
        // Postprocess
        detections = postprocess(blob, outputs, image.size(), threshold);
        
        // Apply NMS
        performNMS(detections);
        
    } catch (const cv::Exception& e) {
        std::cerr << "OpenCV exception in Detector::detect: " << e.what() << std::endl;
    } catch (const std::exception& e) {
        std::cerr << "Exception in Detector::detect: " << e.what() << std::endl;
    }
    
    return detections;
}

std::vector<Detection> Detector::postprocess(const cv::Mat& blob,
                                            const std::vector<cv::Mat>& outputs,
                                            const cv::Size& image_size,
                                            float threshold) {
    std::vector<Detection> detections;
    
    if (outputs.empty()) {
        return detections;
    }
    
    // YOLOv5 output format: [batch, 25200, 85] or similar
    // Format: [x, y, w, h, obj_conf, class1_conf, class2_conf, ...]
    const cv::Mat& output = outputs[0];
    
    // Get dimensions
    int num_detections = output.size[1];
    int num_classes = output.size[2] - 5;
    
    for (int i = 0; i < num_detections; ++i) {
        const float* data = output.ptr<float>(0, i);
        
        float obj_conf = data[4];
        if (obj_conf < threshold) {
            continue;
        }
        
        // Find max class score
        int class_id = -1;
        float max_class_score = 0.0f;
        for (int c = 0; c < num_classes; ++c) {
            float score = data[5 + c];
            if (score > max_class_score) {
                max_class_score = score;
                class_id = c;
            }
        }
        
        float final_score = obj_conf * max_class_score;
        if (final_score < threshold) {
            continue;
        }
        
        // Get bounding box (YOLO format: center_x, center_y, width, height)
        float cx = data[0] / input_width_;
        float cy = data[1] / input_height_;
        float w = data[2] / input_width_;
        float h = data[3] / input_height_;
        
        Detection det;
        det.class_id = class_id;
        det.label = (class_id >= 0 && class_id < static_cast<int>(labels_.size())) 
                    ? labels_[class_id] : "unknown";
        det.bbox = BBox(cx, cy, w, h);
        det.score = final_score;
        
        detections.push_back(det);
    }
    
    return detections;
}

void Detector::performNMS(std::vector<Detection>& detections, float nms_threshold) {
    if (detections.size() <= 1) {
        return;
    }
    
    // Sort by score (descending)
    std::sort(detections.begin(), detections.end(), 
              [](const Detection& a, const Detection& b) {
                  return a.score > b.score;
              });
    
    std::vector<bool> suppressed(detections.size(), false);
    
    for (size_t i = 0; i < detections.size(); ++i) {
        if (suppressed[i]) continue;
        
        const BBox& box1 = detections[i].bbox;
        
        for (size_t j = i + 1; j < detections.size(); ++j) {
            if (suppressed[j]) continue;
            
            const BBox& box2 = detections[j].bbox;
            
            // Compute IoU
            float x1 = std::max(box1.x - box1.width/2, box2.x - box2.width/2);
            float y1 = std::max(box1.y - box1.height/2, box2.y - box2.height/2);
            float x2 = std::min(box1.x + box1.width/2, box2.x + box2.width/2);
            float y2 = std::min(box1.y + box1.height/2, box2.y + box2.height/2);
            
            float inter_area = std::max(0.0f, x2 - x1) * std::max(0.0f, y2 - y1);
            float box1_area = box1.width * box1.height;
            float box2_area = box2.width * box2.height;
            float union_area = box1_area + box2_area - inter_area;
            
            float iou = (union_area > 0) ? (inter_area / union_area) : 0.0f;
            
            if (iou > nms_threshold) {
                suppressed[j] = true;
            }
        }
    }
    
    // Remove suppressed detections
    std::vector<Detection> filtered;
    for (size_t i = 0; i < detections.size(); ++i) {
        if (!suppressed[i]) {
            filtered.push_back(detections[i]);
        }
    }
    
    detections = filtered;
}

Detector* loadDetector(const std::string& model_path,
                      const std::string& labels_path,
                      const std::string& backend) {
    Detector* detector = new Detector();
    if (!detector->initialize(model_path, labels_path, backend)) {
        delete detector;
        return nullptr;
    }
    return detector;
}

} // namespace scene_graph
