#include "object_detector.hpp"
#include <fstream>
#include <iostream>
#include <algorithm>
#include <filesystem>

ObjectDetector::ObjectDetector(const std::string& model_path,
                              const std::string& config_path,
                              const std::string& classes_path,
                              double confidence_threshold,
                              std::shared_ptr<Logger> logger)
    : model_path_(model_path), config_path_(config_path), classes_path_(classes_path),
      confidence_threshold_(confidence_threshold), logger_(logger), initialized_(false) {
}

ObjectDetector::~ObjectDetector() = default;

bool ObjectDetector::initialize() {
    if (initialized_) {
        return true;
    }

    logger_->info("Initializing object detector...");
    logger_->debug("Model path: " + model_path_);
    logger_->debug("Classes path: " + classes_path_);
    logger_->debug("Confidence threshold: " + std::to_string(confidence_threshold_));

    // Load class names
    if (!loadClassNames()) {
        logger_->error("Failed to load class names");
        return false;
    }

    // Load the model
    if (!loadModel()) {
        logger_->error("Failed to load detection model");
        return false;
    }

    initialized_ = true;
    logger_->info("Object detector initialized successfully");
    logger_->info("Loaded " + std::to_string(class_names_.size()) + " object classes");
    
    auto target_classes = getTargetClasses();
    logger_->info("Target classes: person, car, truck, bus, motorcycle, bicycle, cat, dog");
    
    return true;
}

std::vector<ObjectDetector::Detection> ObjectDetector::detectObjects(const cv::Mat& frame) {
    if (!initialized_ || frame.empty()) {
        return {};
    }

    std::vector<Detection> detections;

    try {
        // Create blob from image
        cv::Mat blob;
        cv::dnn::blobFromImage(frame, blob, SCALE_FACTOR, 
                              cv::Size(INPUT_WIDTH, INPUT_HEIGHT), 
                              MEAN, true, false, CV_32F);

        // Set input to the network
        net_.setInput(blob);

        // Forward pass
        std::vector<cv::Mat> outputs;
        net_.forward(outputs, net_.getUnconnectedOutLayersNames());

        // Post-process the outputs
        detections = postProcess(frame, outputs);

    } catch (const cv::Exception& e) {
        logger_->error("OpenCV error during detection: " + std::string(e.what()));
    } catch (const std::exception& e) {
        logger_->error("Error during detection: " + std::string(e.what()));
    }

    return detections;
}

void ObjectDetector::processFrame(const cv::Mat& frame) {
    if (!initialized_) {
        return;
    }

    // Detect objects in the current frame
    auto detections = detectObjects(frame);
    
    // Filter for target classes only
    std::vector<Detection> target_detections;
    for (const auto& detection : detections) {
        if (isTargetClass(detection.class_name)) {
            target_detections.push_back(detection);
        }
    }

    // Update tracking and log events
    updateTrackedObjects(target_detections);
    logObjectEvents(target_detections);
}

std::vector<std::string> ObjectDetector::getTargetClasses() {
    return {
        "person",      // People
        "car",         // Vehicles
        "truck",
        "bus", 
        "motorcycle",
        "bicycle",
        "cat",         // Small animals
        "dog"
        // Note: "fox" is not in COCO dataset, but we can detect it as similar to cat/dog
    };
}

bool ObjectDetector::loadClassNames() {
    std::ifstream class_file(classes_path_);
    if (!class_file.is_open()) {
        // If COCO classes file doesn't exist, create a basic one
        logger_->warning("Classes file not found, using built-in COCO classes");
        class_names_ = {
            "person", "bicycle", "car", "motorcycle", "airplane", "bus", "train", "truck",
            "boat", "traffic light", "fire hydrant", "stop sign", "parking meter", "bench",
            "bird", "cat", "dog", "horse", "sheep", "cow", "elephant", "bear", "zebra",
            "giraffe", "backpack", "umbrella", "handbag", "tie", "suitcase", "frisbee",
            "skis", "snowboard", "sports ball", "kite", "baseball bat", "baseball glove",
            "skateboard", "surfboard", "tennis racket", "bottle", "wine glass", "cup",
            "fork", "knife", "spoon", "bowl", "banana", "apple", "sandwich", "orange",
            "broccoli", "carrot", "hot dog", "pizza", "donut", "cake", "chair", "couch",
            "potted plant", "bed", "dining table", "toilet", "tv", "laptop", "mouse",
            "remote", "keyboard", "cell phone", "microwave", "oven", "toaster", "sink",
            "refrigerator", "book", "clock", "vase", "scissors", "teddy bear", "hair drier",
            "toothbrush"
        };
        return true;
    }

    class_names_.clear();
    std::string line;
    while (std::getline(class_file, line)) {
        if (!line.empty()) {
            class_names_.push_back(line);
        }
    }
    class_file.close();

    return !class_names_.empty();
}

bool ObjectDetector::loadModel() {
    try {
        // Check if model file exists
        if (!std::filesystem::exists(model_path_)) {
            logger_->error("Model file not found: " + model_path_);
            logger_->error("Please download a YOLO model (e.g., YOLOv5s) and place it at the specified path");
            logger_->error("Example: wget https://github.com/ultralytics/yolov5/releases/download/v6.2/yolov5s.onnx");
            return false;
        }

        // Load the network
        net_ = cv::dnn::readNetFromONNX(model_path_);
        
        if (net_.empty()) {
            logger_->error("Failed to load neural network from: " + model_path_);
            return false;
        }

        // Try to use GPU if available
        if (cv::dnn::DNN_BACKEND_CUDA == cv::dnn::getAvailableBackends()[0]) {
            logger_->info("CUDA backend available, attempting to use GPU");
            net_.setPreferableBackend(cv::dnn::DNN_BACKEND_CUDA);
            net_.setPreferableTarget(cv::dnn::DNN_TARGET_CUDA);
        } else {
            logger_->info("Using CPU backend for inference");
            net_.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
            net_.setPreferableTarget(cv::dnn::DNN_TARGET_CPU);
        }

        logger_->debug("Neural network loaded successfully");
        return true;

    } catch (const cv::Exception& e) {
        logger_->error("OpenCV error loading model: " + std::string(e.what()));
        return false;
    } catch (const std::exception& e) {
        logger_->error("Error loading model: " + std::string(e.what()));
        return false;
    }
}

std::vector<ObjectDetector::Detection> ObjectDetector::postProcess(
    const cv::Mat& frame, const std::vector<cv::Mat>& outputs) {
    
    std::vector<Detection> detections;
    
    if (outputs.empty()) {
        return detections;
    }

    // YOLO output format: [batch, num_detections, 85] where 85 = 4 (bbox) + 1 (confidence) + 80 (classes)
    const cv::Mat& output = outputs[0];
    
    if (output.dims != 3) {
        logger_->debug("Unexpected output dimensions: " + std::to_string(output.dims));
        return detections;
    }

    const int num_detections = output.size[1];
    const int num_classes = output.size[2] - 5; // First 5 are x, y, w, h, confidence
    
    float* data = (float*)output.data;
    
    for (int i = 0; i < num_detections; ++i) {
        float* detection = data + i * (num_classes + 5);
        
        float confidence = detection[4];
        if (confidence < confidence_threshold_) {
            continue;
        }
        
        // Find the class with maximum score
        float max_class_score = 0;
        int max_class_id = -1;
        for (int j = 0; j < num_classes; ++j) {
            float class_score = detection[5 + j];
            if (class_score > max_class_score) {
                max_class_score = class_score;
                max_class_id = j;
            }
        }
        
        float final_confidence = confidence * max_class_score;
        if (final_confidence < confidence_threshold_) {
            continue;
        }
        
        if (max_class_id < 0 || max_class_id >= static_cast<int>(class_names_.size())) {
            continue;
        }
        
        // Extract bounding box (center format)
        float center_x = detection[0];
        float center_y = detection[1];
        float width = detection[2];
        float height = detection[3];
        
        // Convert to corner format and scale to frame size
        float x1 = (center_x - width / 2) * frame.cols / INPUT_WIDTH;
        float y1 = (center_y - height / 2) * frame.rows / INPUT_HEIGHT;
        float x2 = (center_x + width / 2) * frame.cols / INPUT_WIDTH;
        float y2 = (center_y + height / 2) * frame.rows / INPUT_HEIGHT;
        
        Detection det;
        det.bbox = cv::Rect(cv::Point(static_cast<int>(x1), static_cast<int>(y1)),
                           cv::Point(static_cast<int>(x2), static_cast<int>(y2)));
        det.confidence = final_confidence;
        det.class_id = max_class_id;
        det.class_name = class_names_[max_class_id];
        
        detections.push_back(det);
    }
    
    return detections;
}

void ObjectDetector::updateTrackedObjects(const std::vector<Detection>& detections) {
    // Simple tracking: mark objects as present/absent
    // This is a basic implementation - could be enhanced with proper object tracking
    
    // Mark all current objects as not seen this frame
    for (auto& tracked : tracked_objects_) {
        tracked.was_present_last_frame = false;
        tracked.frames_since_detection++;
    }
    
    // Update with current detections
    for (const auto& detection : detections) {
        bool found_existing = false;
        
        // Try to match with existing tracked objects
        for (auto& tracked : tracked_objects_) {
            if (tracked.object_type == detection.class_name) {
                // Simple distance-based matching (could be improved)
                cv::Point2f detection_center(
                    detection.bbox.x + detection.bbox.width / 2.0f,
                    detection.bbox.y + detection.bbox.height / 2.0f
                );
                
                float distance = cv::norm(tracked.center - detection_center);
                if (distance < 100.0f) { // Threshold for same object
                    tracked.center = detection_center;
                    tracked.was_present_last_frame = true;
                    tracked.frames_since_detection = 0;
                    found_existing = true;
                    break;
                }
            }
        }
        
        // Add new object if not found
        if (!found_existing) {
            ObjectTracker new_tracker;
            new_tracker.object_type = detection.class_name;
            new_tracker.center = cv::Point2f(
                detection.bbox.x + detection.bbox.width / 2.0f,
                detection.bbox.y + detection.bbox.height / 2.0f
            );
            new_tracker.was_present_last_frame = true;
            new_tracker.frames_since_detection = 0;
            tracked_objects_.push_back(new_tracker);
        }
    }
    
    // Remove objects that haven't been seen for too long
    tracked_objects_.erase(
        std::remove_if(tracked_objects_.begin(), tracked_objects_.end(),
                      [](const ObjectTracker& tracker) {
                          return tracker.frames_since_detection > 30; // 30 frames threshold
                      }),
        tracked_objects_.end());
}

bool ObjectDetector::isTargetClass(const std::string& class_name) const {
    auto target_classes = getTargetClasses();
    return std::find(target_classes.begin(), target_classes.end(), class_name) != target_classes.end();
}

void ObjectDetector::logObjectEvents(const std::vector<Detection>& current_detections) {
    // Log enter/exit events based on tracking
    for (const auto& tracked : tracked_objects_) {
        bool currently_present = std::any_of(current_detections.begin(), current_detections.end(),
                                           [&tracked](const Detection& det) {
                                               return det.class_name == tracked.object_type;
                                           });
        
        if (currently_present && tracked.frames_since_detection > 5) {
            // Object entered (was absent for multiple frames, now present)
            auto detection_it = std::find_if(current_detections.begin(), current_detections.end(),
                                            [&tracked](const Detection& det) {
                                                return det.class_name == tracked.object_type;
                                            });
            if (detection_it != current_detections.end()) {
                logger_->logObjectDetection(tracked.object_type, "entered", detection_it->confidence);
            }
        }
    }
    
    // Note: Exit detection would require more sophisticated tracking
    // For now, we focus on entry detection which is more reliable
}