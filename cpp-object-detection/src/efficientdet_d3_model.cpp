#include "efficientdet_d3_model.hpp"
#include <fstream>
#include <algorithm>

// Check if filesystem is available
#if __has_include(<filesystem>)
    #include <filesystem>
    namespace fs = std::filesystem;
#else
    // Fallback for older systems
    #include <sys/stat.h>
    namespace fs {
        inline bool exists(const std::string& path) {
            struct stat buffer;
            return (stat(path.c_str(), &buffer) == 0);
        }
    }
#endif

// Static member definitions
const cv::Scalar EfficientDetD3Model::MEAN = cv::Scalar(0, 0, 0);

EfficientDetD3Model::EfficientDetD3Model(std::shared_ptr<Logger> logger)
    : logger_(logger), confidence_threshold_(0.5), detection_scale_factor_(1.0), 
      initialized_(false), enable_gpu_(false), avg_inference_time_ms_(95) {
}

bool EfficientDetD3Model::initialize(const std::string& model_path,
                                     const std::string& /* config_path */,
                                     const std::string& classes_path,
                                     double confidence_threshold,
                                     double detection_scale_factor) {
    if (initialized_) {
        return true;
    }

    confidence_threshold_ = confidence_threshold;
    detection_scale_factor_ = detection_scale_factor;
    
    logger_->info("Initializing EfficientDet-D3 model...");
    logger_->debug("Model path: " + model_path);
    logger_->debug("Classes path: " + classes_path);
    logger_->debug("Confidence threshold: " + std::to_string(confidence_threshold_));
    logger_->debug("Detection scale factor: " + std::to_string(detection_scale_factor_));

    // Load class names
    if (!loadClassNames(classes_path)) {
        logger_->error("Failed to load class names");
        return false;
    }

    // Load the model
    if (!loadModel(model_path)) {
        logger_->error("Failed to load EfficientDet-D3 model");
        return false;
    }

    initialized_ = true;
    logger_->info("EfficientDet-D3 model initialized successfully");
    
    return true;
}

std::vector<Detection> EfficientDetD3Model::detect(const cv::Mat& frame) {
    if (!initialized_ || frame.empty()) {
        return {};
    }

    auto start_time = std::chrono::steady_clock::now();
    std::vector<Detection> detections;

    try {
        // Downscale frame if scale factor is less than 1.0
        cv::Mat detection_frame = frame;
        if (detection_scale_factor_ < 1.0) {
            int new_width = static_cast<int>(frame.cols * detection_scale_factor_);
            int new_height = static_cast<int>(frame.rows * detection_scale_factor_);
            cv::resize(frame, detection_frame, cv::Size(new_width, new_height), 0, 0, cv::INTER_LINEAR);
        }
        
        // Create blob from image (potentially downscaled)
        cv::Mat blob;
        cv::dnn::blobFromImage(detection_frame, blob, SCALE_FACTOR, 
                              cv::Size(INPUT_WIDTH, INPUT_HEIGHT), 
                              MEAN, true, false, CV_32F);

        // Set input to the network
        net_.setInput(blob);

        // Forward pass
        std::vector<cv::Mat> outputs;
        net_.forward(outputs, net_.getUnconnectedOutLayersNames());

        // Post-process the outputs (pass original frame for proper bbox scaling)
        detections = postProcess(frame, outputs);

    } catch (const cv::Exception& e) {
        logger_->error("OpenCV error during EfficientDet-D3 detection: " + std::string(e.what()));
    } catch (const std::exception& e) {
        logger_->error("Error during EfficientDet-D3 detection: " + std::string(e.what()));
    }

    // Update inference timing
    auto end_time = std::chrono::steady_clock::now();
    auto inference_time = std::chrono::duration_cast<std::chrono::milliseconds>(end_time - start_time);
    updateInferenceTime(static_cast<int>(inference_time.count()));

    return detections;
}

ModelMetrics EfficientDetD3Model::getMetrics() const {
    return {
        "EfficientDet-D3",
        "EfficientDet",
        0.89,                    // 89% mAP on COCO
        avg_inference_time_ms_,  // Current measured average
        45,                      // ~45MB model size
        "Compound-scaled efficient detection model with BiFPN. "
        "Excellent balance of accuracy and speed for outdoor scenes."
    };
}

std::vector<std::string> EfficientDetD3Model::getSupportedClasses() const {
    return class_names_;
}

bool EfficientDetD3Model::isInitialized() const {
    return initialized_;
}

std::string EfficientDetD3Model::getModelName() const {
    return "EfficientDet-D3";
}

void EfficientDetD3Model::setEnableGpu(bool enable_gpu) {
    enable_gpu_ = enable_gpu;
}

void EfficientDetD3Model::warmUp() {
    if (!initialized_) {
        return;
    }
    
    logger_->debug("Warming up EfficientDet-D3 model...");
    
    // Create a dummy frame for warm-up
    cv::Mat dummy_frame(INPUT_HEIGHT, INPUT_WIDTH, CV_8UC3, cv::Scalar(128, 128, 128));
    
    // Run a few inference passes to warm up
    for (int i = 0; i < 3; ++i) {
        detect(dummy_frame);
    }
    
    logger_->debug("EfficientDet-D3 model warm-up complete");
}

bool EfficientDetD3Model::loadClassNames(const std::string& classes_path) {
    std::ifstream class_file(classes_path);
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

bool EfficientDetD3Model::loadModel(const std::string& model_path) {
    try {
        // Check if model file exists
        if (!fs::exists(model_path)) {
            logger_->error("EfficientDet-D3 model file not found: " + model_path);
            logger_->error("Please download efficientdet-d3.onnx and place it at the specified path");
            logger_->error("Run: ./scripts/download_models.sh to download all models");
            return false;
        }

        // Load the network
        net_ = cv::dnn::readNetFromONNX(model_path);
        
        if (net_.empty()) {
            logger_->error("Failed to load EfficientDet-D3 neural network from: " + model_path);
            return false;
        }

        // Select backend based on platform and available hardware
#ifdef __APPLE__
        if (enable_gpu_) {
            // Try to use GPU acceleration on macOS via OpenCL
            try {
                net_.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
                net_.setPreferableTarget(cv::dnn::DNN_TARGET_OPENCL);
                logger_->info("EfficientDet-D3 using OpenCL backend for GPU acceleration (macOS)");
            } catch (const std::exception& e) {
                logger_->info("EfficientDet-D3 using CPU backend for inference (OpenCL failed): " + std::string(e.what()));
                net_.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
                net_.setPreferableTarget(cv::dnn::DNN_TARGET_CPU);
            }
        } else {
            logger_->info("EfficientDet-D3 using CPU backend for inference (macOS)");
            net_.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
            net_.setPreferableTarget(cv::dnn::DNN_TARGET_CPU);
        }
#else
        // Try to use GPU if available and enabled on other platforms
        if (enable_gpu_) {
            try {
                net_.setPreferableBackend(cv::dnn::DNN_BACKEND_CUDA);
                net_.setPreferableTarget(cv::dnn::DNN_TARGET_CUDA);
                logger_->info("EfficientDet-D3 using CUDA backend for GPU acceleration");
            } catch (const std::exception& e) {
                logger_->info("EfficientDet-D3 using CPU backend for inference (CUDA failed): " + std::string(e.what()));
                net_.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
                net_.setPreferableTarget(cv::dnn::DNN_TARGET_CPU);
            }
        } else {
            logger_->info("EfficientDet-D3 using CPU backend for inference");
            net_.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
            net_.setPreferableTarget(cv::dnn::DNN_TARGET_CPU);
        }
#endif

        logger_->debug("EfficientDet-D3 neural network loaded successfully");
        return true;

    } catch (const cv::Exception& e) {
        logger_->error("OpenCV error loading EfficientDet-D3 model: " + std::string(e.what()));
        return false;
    } catch (const std::exception& e) {
        logger_->error("Error loading EfficientDet-D3 model: " + std::string(e.what()));
        return false;
    }
}

std::vector<Detection> EfficientDetD3Model::postProcess(
    const cv::Mat& frame, const std::vector<cv::Mat>& outputs) {
    
    std::vector<Detection> detections;
    
    if (outputs.empty()) {
        return detections;
    }

    // EfficientDet output format is similar to YOLO: [batch, num_detections, 85] 
    // where 85 = 4 (bbox) + 1 (confidence) + 80 (classes)
    const cv::Mat& output = outputs[0];
    
    if (output.dims != 3) {
        logger_->debug("Unexpected EfficientDet-D3 output dimensions: " + std::to_string(output.dims));
        return detections;
    }

    const int num_detections = output.size[1];
    const int num_classes = output.size[2] - 5; // First 5 are x, y, w, h, confidence
    
    float* data = (float*)output.data;
    
    // Temporary storage for NMS
    std::vector<cv::Rect> boxes;
    std::vector<float> confidences;
    std::vector<int> class_ids;
    
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
        
        cv::Rect bbox(cv::Point(static_cast<int>(x1), static_cast<int>(y1)),
                      cv::Point(static_cast<int>(x2), static_cast<int>(y2)));
        
        // Collect boxes for NMS
        boxes.push_back(bbox);
        confidences.push_back(final_confidence);
        class_ids.push_back(max_class_id);
    }
    
    // Apply Non-Maximum Suppression to eliminate overlapping boxes
    std::vector<int> indices;
    if (!boxes.empty()) {
        // NMS threshold of 0.45 is standard
        // This means boxes with IoU > 0.45 (45% overlap) will be suppressed
        cv::dnn::NMSBoxes(boxes, confidences, confidence_threshold_, 0.45f, indices);
    }
    
    // Build final detections from NMS results
    for (int idx : indices) {
        Detection det;
        det.bbox = boxes[idx];
        det.confidence = confidences[idx];
        det.class_id = class_ids[idx];
        det.class_name = class_names_[class_ids[idx]];
        detections.push_back(det);
    }
    
    return detections;
}

void EfficientDetD3Model::updateInferenceTime(int inference_time_ms) const {
    // Simple moving average
    avg_inference_time_ms_ = (avg_inference_time_ms_ * 9 + inference_time_ms) / 10;
}
