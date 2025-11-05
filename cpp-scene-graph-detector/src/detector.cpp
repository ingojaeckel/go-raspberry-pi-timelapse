#include "scene_graph/Detector.h"
#include "scene_graph/logger.h"
#include <fstream>
#include <algorithm>

namespace scene_graph {

Detector::Detector() 
    : backend_(Backend::CPU), 
      model_loaded_(false),
      input_width_(640),
      input_height_(640) {
}

Detector::~Detector() {
}

bool Detector::loadDetector(const std::string& model_path, 
                            const std::string& labels_path,
                            Backend backend) {
    auto& logger = Logger::getInstance();
    
    try {
        // Load labels
        std::ifstream labels_file(labels_path);
        if (!labels_file.is_open()) {
            logger.error("Failed to open labels file: " + labels_path);
            return false;
        }
        
        std::string line;
        class_labels_.clear();
        while (std::getline(labels_file, line)) {
            if (!line.empty()) {
                class_labels_.push_back(line);
            }
        }
        labels_file.close();
        
        logger.info("Loaded " + std::to_string(class_labels_.size()) + " class labels");
        
        // Load model
        net_ = cv::dnn::readNetFromONNX(model_path);
        if (net_.empty()) {
            logger.error("Failed to load ONNX model: " + model_path);
            return false;
        }
        
        // Set backend and target
        backend_ = backend;
        if (backend == Backend::OPENCL) {
#ifdef WITH_OPENCL
            net_.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
            net_.setPreferableTarget(cv::dnn::DNN_TARGET_OPENCL);
            logger.info("Using OpenCL backend");
#else
            logger.warning("OpenCL requested but not available, falling back to CPU");
            net_.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
            net_.setPreferableTarget(cv::dnn::DNN_TARGET_CPU);
            backend_ = Backend::CPU;
#endif
        } else if (backend == Backend::AUTO) {
#ifdef WITH_OPENCL
            net_.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
            net_.setPreferableTarget(cv::dnn::DNN_TARGET_OPENCL);
            backend_ = Backend::OPENCL;
            logger.info("Auto-selected OpenCL backend");
#else
            net_.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
            net_.setPreferableTarget(cv::dnn::DNN_TARGET_CPU);
            backend_ = Backend::CPU;
            logger.info("Auto-selected CPU backend");
#endif
        } else {
            net_.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
            net_.setPreferableTarget(cv::dnn::DNN_TARGET_CPU);
            logger.info("Using CPU backend");
        }
        
        model_loaded_ = true;
        logger.info("Detector model loaded successfully");
        return true;
        
    } catch (const cv::Exception& e) {
        logger.error("OpenCV exception while loading detector: " + std::string(e.what()));
        return false;
    } catch (const std::exception& e) {
        logger.error("Exception while loading detector: " + std::string(e.what()));
        return false;
    }
}

std::vector<Detection> Detector::infer(const cv::Mat& frame, 
                                       float confidence_threshold,
                                       int max_objects) {
    std::vector<Detection> detections;
    
    if (!model_loaded_) {
        Logger::getInstance().error("Detector model not loaded");
        return detections;
    }
    
    if (frame.empty()) {
        Logger::getInstance().error("Empty frame provided for inference");
        return detections;
    }
    
    try {
        // Prepare input blob
        cv::Mat blob = cv::dnn::blobFromImage(frame, 1.0/255.0, 
                                              cv::Size(input_width_, input_height_), 
                                              cv::Scalar(), true, false);
        net_.setInput(blob);
        
        // Run inference
        std::vector<cv::Mat> outputs;
        net_.forward(outputs, net_.getUnconnectedOutLayersNames());
        
        // Post-process outputs
        detections = postProcess(outputs, frame.size(), confidence_threshold, max_objects);
        
        Logger::getInstance().debug("Detected " + std::to_string(detections.size()) + " objects");
        
    } catch (const cv::Exception& e) {
        Logger::getInstance().error("OpenCV exception during inference: " + std::string(e.what()));
    } catch (const std::exception& e) {
        Logger::getInstance().error("Exception during inference: " + std::string(e.what()));
    }
    
    return detections;
}

std::vector<Detection> Detector::postProcess(const std::vector<cv::Mat>& outputs,
                                             const cv::Size& frame_size,
                                             float confidence_threshold,
                                             int max_objects) {
    std::vector<Detection> detections;
    
    // YOLO output format: [batch, num_detections, 5 + num_classes]
    // [x, y, w, h, objectness, class_scores...]
    
    for (const auto& output : outputs) {
        const auto* data = (float*)output.data;
        int rows = output.size[1];  // number of detections
        int cols = output.size[2];  // 5 + num_classes
        
        for (int i = 0; i < rows; ++i) {
            const float* row = data + i * cols;
            
            float objectness = row[4];
            if (objectness < confidence_threshold) {
                continue;
            }
            
            // Find class with max score
            const float* class_scores = row + 5;
            int num_classes = cols - 5;
            int class_id = 0;
            float max_score = class_scores[0];
            
            for (int j = 1; j < num_classes; ++j) {
                if (class_scores[j] > max_score) {
                    max_score = class_scores[j];
                    class_id = j;
                }
            }
            
            float confidence = objectness * max_score;
            if (confidence < confidence_threshold) {
                continue;
            }
            
            // Get bounding box (center format)
            float cx = row[0];
            float cy = row[1];
            float w = row[2];
            float h = row[3];
            
            // Scale to frame size
            float scale_x = static_cast<float>(frame_size.width) / input_width_;
            float scale_y = static_cast<float>(frame_size.height) / input_height_;
            
            int x = static_cast<int>((cx - w / 2) * scale_x);
            int y = static_cast<int>((cy - h / 2) * scale_y);
            int width = static_cast<int>(w * scale_x);
            int height = static_cast<int>(h * scale_y);
            
            // Clamp to frame boundaries
            x = std::max(0, std::min(x, frame_size.width - 1));
            y = std::max(0, std::min(y, frame_size.height - 1));
            width = std::min(width, frame_size.width - x);
            height = std::min(height, frame_size.height - y);
            
            cv::Rect box(x, y, width, height);
            
            std::string label = (class_id < static_cast<int>(class_labels_.size())) 
                              ? class_labels_[class_id] 
                              : "unknown";
            
            detections.emplace_back(class_id, label, confidence, box);
        }
    }
    
    // Apply NMS
    applyNMS(detections);
    
    // Limit to max objects
    if (static_cast<int>(detections.size()) > max_objects) {
        // Sort by confidence and keep top N
        std::sort(detections.begin(), detections.end(), 
                 [](const Detection& a, const Detection& b) {
                     return a.confidence > b.confidence;
                 });
        detections.resize(max_objects);
    }
    
    return detections;
}

void Detector::applyNMS(std::vector<Detection>& detections, float nms_threshold) {
    if (detections.empty()) return;
    
    // Sort by confidence
    std::sort(detections.begin(), detections.end(),
             [](const Detection& a, const Detection& b) {
                 return a.confidence > b.confidence;
             });
    
    std::vector<bool> suppressed(detections.size(), false);
    
    for (size_t i = 0; i < detections.size(); ++i) {
        if (suppressed[i]) continue;
        
        const auto& box1 = detections[i].box;
        
        for (size_t j = i + 1; j < detections.size(); ++j) {
            if (suppressed[j]) continue;
            
            const auto& box2 = detections[j].box;
            
            // Calculate IoU
            int x1 = std::max(box1.x, box2.x);
            int y1 = std::max(box1.y, box2.y);
            int x2 = std::min(box1.x + box1.width, box2.x + box2.width);
            int y2 = std::min(box1.y + box1.height, box2.y + box2.height);
            
            int intersection_area = std::max(0, x2 - x1) * std::max(0, y2 - y1);
            int union_area = box1.area() + box2.area() - intersection_area;
            
            float iou = (union_area > 0) ? static_cast<float>(intersection_area) / union_area : 0.0f;
            
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
    
    detections = std::move(filtered);
}

std::string Detector::getBackendInfo() const {
    std::string info = "Backend: ";
    switch (backend_) {
        case Backend::CPU: info += "CPU"; break;
        case Backend::OPENCL: info += "OpenCL"; break;
        case Backend::AUTO: info += "Auto"; break;
    }
    return info;
}

} // namespace scene_graph
