#pragma once

#include "detection_model_interface.hpp"
#include "logger.hpp"
#include <opencv2/dnn.hpp>
#include <chrono>

/**
 * YOLOv5 Small model implementation - Fast inference, good accuracy
 * Optimized for real-time applications with moderate hardware requirements
 */
class YoloV5SmallModel : public IDetectionModel {
public:
    explicit YoloV5SmallModel(std::shared_ptr<Logger> logger);
    ~YoloV5SmallModel() override = default;
    
    bool initialize(const std::string& model_path,
                   const std::string& config_path,
                   const std::string& classes_path,
                   double confidence_threshold,
                   double detection_scale_factor = 1.0) override;
    
    std::vector<Detection> detect(const cv::Mat& frame) override;
    
    ModelMetrics getMetrics() const override;
    
    std::vector<std::string> getSupportedClasses() const override;
    
    bool isInitialized() const override;
    
    std::string getModelName() const override;
    
    void warmUp() override;

private:
    std::shared_ptr<Logger> logger_;
    cv::dnn::Net net_;
    std::vector<std::string> class_names_;
    double confidence_threshold_;
    double detection_scale_factor_;
    bool initialized_;
    mutable std::chrono::steady_clock::time_point last_inference_start_;
    mutable int avg_inference_time_ms_;
    
    // YOLOv5s specific parameters
    static constexpr int INPUT_WIDTH = 640;
    static constexpr int INPUT_HEIGHT = 640;
    static constexpr float SCALE_FACTOR = 1.0 / 255.0;
    static const cv::Scalar MEAN;
    
    bool loadClassNames(const std::string& classes_path);
    bool loadModel(const std::string& model_path);
    std::vector<Detection> postProcess(const cv::Mat& frame, 
                                     const std::vector<cv::Mat>& outputs);
    void updateInferenceTime(int inference_time_ms) const;
};

/**
 * YOLOv5 Large model implementation - Higher accuracy, slower inference
 * Better for applications where accuracy is more important than speed
 */
class YoloV5LargeModel : public IDetectionModel {
public:
    explicit YoloV5LargeModel(std::shared_ptr<Logger> logger);
    ~YoloV5LargeModel() override = default;
    
    bool initialize(const std::string& model_path,
                   const std::string& config_path,
                   const std::string& classes_path,
                   double confidence_threshold,
                   double detection_scale_factor = 1.0) override;
    
    std::vector<Detection> detect(const cv::Mat& frame) override;
    
    ModelMetrics getMetrics() const override;
    
    std::vector<std::string> getSupportedClasses() const override;
    
    bool isInitialized() const override;
    
    std::string getModelName() const override;
    
    void warmUp() override;

private:
    std::shared_ptr<Logger> logger_;
    cv::dnn::Net net_;
    std::vector<std::string> class_names_;
    double confidence_threshold_;
    double detection_scale_factor_;
    bool initialized_;
    mutable std::chrono::steady_clock::time_point last_inference_start_;
    mutable int avg_inference_time_ms_;
    
    // YOLOv5l specific parameters (larger input size for better accuracy)
    static constexpr int INPUT_WIDTH = 832;   // Larger input for better accuracy
    static constexpr int INPUT_HEIGHT = 832;
    static constexpr float SCALE_FACTOR = 1.0 / 255.0;
    static const cv::Scalar MEAN;
    
    bool loadClassNames(const std::string& classes_path);
    bool loadModel(const std::string& model_path);
    std::vector<Detection> postProcess(const cv::Mat& frame, 
                                     const std::vector<cv::Mat>& outputs);
    void updateInferenceTime(int inference_time_ms) const;
};