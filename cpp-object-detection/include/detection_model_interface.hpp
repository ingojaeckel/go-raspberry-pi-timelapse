#pragma once

#include <opencv2/opencv.hpp>
#include <vector>
#include <string>
#include <memory>

/**
 * Detection result structure
 */
struct Detection {
    std::string class_name;
    double confidence;
    cv::Rect bbox;
    int class_id;
};

/**
 * Model performance metrics
 */
struct ModelMetrics {
    std::string model_name;
    std::string model_type;
    double accuracy_score;      // Relative accuracy (0.0 - 1.0)
    int avg_inference_time_ms;  // Average inference time in milliseconds
    int model_size_mb;          // Model file size in MB
    std::string description;
};

/**
 * Abstract base interface for object detection models
 * 
 * This interface allows plugging in different detection models
 * without changing the calling code. Engineers can implement
 * this interface for different model types (YOLO, SSD, etc.)
 */
class IDetectionModel {
public:
    virtual ~IDetectionModel() = default;
    
    /**
     * Initialize the model with configuration
     * @param model_path Path to the model file
     * @param config_path Path to configuration file (optional)
     * @param classes_path Path to class names file
     * @param confidence_threshold Minimum confidence for detections
     * @return true if initialization successful
     */
    virtual bool initialize(const std::string& model_path,
                           const std::string& config_path,
                           const std::string& classes_path,
                           double confidence_threshold) = 0;
    
    /**
     * Detect objects in a frame
     * @param frame Input image frame
     * @return Vector of detections
     */
    virtual std::vector<Detection> detect(const cv::Mat& frame) = 0;
    
    /**
     * Get model performance metrics
     * @return ModelMetrics structure with performance information
     */
    virtual ModelMetrics getMetrics() const = 0;
    
    /**
     * Get list of supported target classes
     * @return Vector of class names this model can detect
     */
    virtual std::vector<std::string> getSupportedClasses() const = 0;
    
    /**
     * Check if model is initialized and ready
     * @return true if model is ready for inference
     */
    virtual bool isInitialized() const = 0;
    
    /**
     * Get model name for identification
     * @return Human-readable model name
     */
    virtual std::string getModelName() const = 0;
    
    /**
     * Warm up the model with a dummy inference
     * This helps get accurate timing measurements
     */
    virtual void warmUp() = 0;
    
protected:
    // Common target classes for security monitoring
    static std::vector<std::string> getTargetClasses() {
        return {
            "person",      // People
            "car",         // Vehicles  
            "truck",
            "bus",
            "motorcycle", 
            "bicycle",
            "cat",         // Small animals
            "dog"
        };
    }
};

/**
 * Factory for creating detection models
 */
class DetectionModelFactory {
public:
    enum class ModelType {
        YOLO_V5_SMALL,    // Fast, lower accuracy
        YOLO_V5_LARGE,    // Slower, higher accuracy
        YOLO_V8_NANO,     // Fastest, good for embedded
        YOLO_V8_MEDIUM    // Balanced speed/accuracy
    };
    
    /**
     * Create a detection model of the specified type
     * @param type Model type to create
     * @param logger Logger instance for the model
     * @return Unique pointer to the model instance
     */
    static std::unique_ptr<IDetectionModel> createModel(
        ModelType type, 
        std::shared_ptr<class Logger> logger);
    
    /**
     * Get available model types with their characteristics
     * @return Vector of available models and their metrics
     */
    static std::vector<ModelMetrics> getAvailableModels();
    
    /**
     * Parse model type from string
     * @param model_name String representation of model type
     * @return ModelType enum value
     */
    static ModelType parseModelType(const std::string& model_name);
    
    /**
     * Get model type as string
     * @param type ModelType enum value
     * @return String representation
     */
    static std::string modelTypeToString(ModelType type);
};