#pragma once

#include "detection_model_interface.hpp"
#include "logger.hpp"
#include <opencv2/dnn.hpp>
#include <chrono>

/**
 * EfficientDet-D3 model implementation - High accuracy with efficient detection
 * Uses compound scaling and BiFPN for excellent balance of accuracy and speed
 * Optimized for outdoor scenes with better multi-scale detection
 */
class EfficientDetD3Model : public IDetectionModel {
public:
    explicit EfficientDetD3Model(std::shared_ptr<Logger> logger);
    ~EfficientDetD3Model() override = default;
    
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
    
    /**
     * Set GPU acceleration preference
     * @param enable_gpu Whether to enable GPU/CUDA acceleration
     */
    void setEnableGpu(bool enable_gpu);

private:
    std::shared_ptr<Logger> logger_;
    cv::dnn::Net net_;
    std::vector<std::string> class_names_;
    double confidence_threshold_;
    double detection_scale_factor_;
    bool initialized_;
    bool enable_gpu_;
    mutable std::chrono::steady_clock::time_point last_inference_start_;
    mutable int avg_inference_time_ms_;
    
    // EfficientDet-D3 specific parameters
    static constexpr int INPUT_WIDTH = 896;   // EfficientDet-D3 input size
    static constexpr int INPUT_HEIGHT = 896;
    static constexpr float SCALE_FACTOR = 1.0 / 255.0;
    static const cv::Scalar MEAN;
    
    bool loadClassNames(const std::string& classes_path);
    bool loadModel(const std::string& model_path);
    std::vector<Detection> postProcess(const cv::Mat& frame, 
                                     const std::vector<cv::Mat>& outputs);
    void updateInferenceTime(int inference_time_ms) const;
};
