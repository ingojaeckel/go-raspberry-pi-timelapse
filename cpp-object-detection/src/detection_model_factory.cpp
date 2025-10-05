#include "detection_model_interface.hpp"
#include "yolo_v5_model.hpp"
#include <stdexcept>
#include <algorithm>

std::unique_ptr<IDetectionModel> DetectionModelFactory::createModel(
    ModelType type, 
    std::shared_ptr<Logger> logger,
    bool enable_cuda) {
    
    switch (type) {
        case ModelType::YOLO_V5_SMALL:
            return std::make_unique<YoloV5SmallModel>(logger, enable_cuda);
            
        case ModelType::YOLO_V5_LARGE:
            return std::make_unique<YoloV5LargeModel>(logger, enable_cuda);
            
        case ModelType::YOLO_V8_NANO:
            // Future implementation - for now fall back to V5 small
            if (logger) {
                logger->warning("YOLOv8 Nano not yet implemented, using YOLOv5 Small");
            }
            return std::make_unique<YoloV5SmallModel>(logger, enable_cuda);
            
        case ModelType::YOLO_V8_MEDIUM:
            // Future implementation - for now fall back to V5 large
            if (logger) {
                logger->warning("YOLOv8 Medium not yet implemented, using YOLOv5 Large");
            }
            return std::make_unique<YoloV5LargeModel>(logger, enable_cuda);
            
        default:
            throw std::invalid_argument("Unknown model type");
    }
}

std::vector<ModelMetrics> DetectionModelFactory::getAvailableModels() {
    return {
        {
            "YOLOv5s", 
            "YOLO", 
            0.75,           // 75% relative accuracy
            65,             // ~65ms average inference on modern CPU
            14,             // ~14MB model size
            "Fast and efficient model optimized for real-time detection. "
            "Good balance of speed and accuracy for most applications."
        },
        {
            "YOLOv5l", 
            "YOLO", 
            0.85,           // 85% relative accuracy
            120,            // ~120ms average inference on modern CPU
            47,             // ~47MB model size
            "Higher accuracy model with larger network. Better for applications "
            "where precision is more important than speed. ~2x slower than YOLOv5s."
        },
        {
            "YOLOv8n", 
            "YOLO", 
            0.70,           // 70% relative accuracy
            35,             // ~35ms average inference (when implemented)
            6,              // ~6MB model size
            "Ultra-fast nano model for embedded systems and edge devices. "
            "Optimized for maximum speed with acceptable accuracy. (Future implementation)"
        },
        {
            "YOLOv8m", 
            "YOLO", 
            0.88,           // 88% relative accuracy
            150,            // ~150ms average inference (when implemented)
            52,             // ~52MB model size
            "High-accuracy medium model with state-of-the-art performance. "
            "Best accuracy available but requires more computational resources. (Future implementation)"
        }
    };
}

DetectionModelFactory::ModelType DetectionModelFactory::parseModelType(const std::string& model_name) {
    std::string lower_name = model_name;
    std::transform(lower_name.begin(), lower_name.end(), lower_name.begin(), ::tolower);
    
    if (lower_name == "yolov5s" || lower_name == "yolo_v5_small" || lower_name == "small") {
        return ModelType::YOLO_V5_SMALL;
    } else if (lower_name == "yolov5l" || lower_name == "yolo_v5_large" || lower_name == "large") {
        return ModelType::YOLO_V5_LARGE;
    } else if (lower_name == "yolov8n" || lower_name == "yolo_v8_nano" || lower_name == "nano") {
        return ModelType::YOLO_V8_NANO;
    } else if (lower_name == "yolov8m" || lower_name == "yolo_v8_medium" || lower_name == "medium") {
        return ModelType::YOLO_V8_MEDIUM;
    } else {
        throw std::invalid_argument("Unknown model name: " + model_name);
    }
}

std::string DetectionModelFactory::modelTypeToString(ModelType type) {
    switch (type) {
        case ModelType::YOLO_V5_SMALL:
            return "yolov5s";
        case ModelType::YOLO_V5_LARGE:
            return "yolov5l";
        case ModelType::YOLO_V8_NANO:
            return "yolov8n";
        case ModelType::YOLO_V8_MEDIUM:
            return "yolov8m";
        default:
            return "unknown";
    }
}