#include <gtest/gtest.h>
#include <memory>
#include "../include/detection_model_interface.hpp"
#include "../include/yolo_v5_model.hpp"
#include "../include/logger.hpp"

class DetectionModelInterfaceTest : public ::testing::Test {
protected:
    void SetUp() override {
        logger_ = std::make_shared<Logger>("test.log", false);
    }

    std::shared_ptr<Logger> logger_;
};

class MockDetectionModel : public IDetectionModel {
public:
    explicit MockDetectionModel(std::shared_ptr<Logger> logger) : logger_(logger), initialized_(false) {}
    
    bool initialize(const std::string& model_path,
                   const std::string& config_path,
                   const std::string& classes_path,
                   double confidence_threshold) override {
        model_path_ = model_path;
        confidence_threshold_ = confidence_threshold;
        initialized_ = true;
        return true;
    }
    
    std::vector<Detection> detect(const cv::Mat& frame) override {
        if (!initialized_ || frame.empty()) {
            return {};
        }
        
        // Mock detection: return a single person detection
        Detection det;
        det.class_name = "person";
        det.confidence = 0.8;
        det.bbox = cv::Rect(100, 100, 200, 300);
        det.class_id = 0;
        
        return {det};
    }
    
    ModelMetrics getMetrics() const override {
        return {"MockModel", "Test", 0.9, 50, 10, "Mock model for testing"};
    }
    
    std::vector<std::string> getSupportedClasses() const override {
        return getTargetClasses();
    }
    
    bool isInitialized() const override {
        return initialized_;
    }
    
    std::string getModelName() const override {
        return "Mock Detection Model";
    }
    
    void warmUp() override {
        // Mock warm up - do nothing
    }

private:
    std::shared_ptr<Logger> logger_;
    bool initialized_;
    std::string model_path_;
    double confidence_threshold_;
};

TEST_F(DetectionModelInterfaceTest, MockModelInitialization) {
    auto model = std::make_unique<MockDetectionModel>(logger_);
    
    EXPECT_FALSE(model->isInitialized());
    EXPECT_EQ(model->getModelName(), "Mock Detection Model");
    
    bool result = model->initialize("test_model.onnx", "", "test_classes.names", 0.5);
    
    EXPECT_TRUE(result);
    EXPECT_TRUE(model->isInitialized());
}

TEST_F(DetectionModelInterfaceTest, MockModelDetection) {
    auto model = std::make_unique<MockDetectionModel>(logger_);
    model->initialize("test_model.onnx", "", "test_classes.names", 0.5);
    
    // Create a test frame
    cv::Mat test_frame(480, 640, CV_8UC3, cv::Scalar(128, 128, 128));
    
    auto detections = model->detect(test_frame);
    
    EXPECT_EQ(detections.size(), 1);
    EXPECT_EQ(detections[0].class_name, "person");
    EXPECT_DOUBLE_EQ(detections[0].confidence, 0.8);
    EXPECT_EQ(detections[0].bbox.x, 100);
    EXPECT_EQ(detections[0].bbox.y, 100);
}

TEST_F(DetectionModelInterfaceTest, MockModelMetrics) {
    auto model = std::make_unique<MockDetectionModel>(logger_);
    auto metrics = model->getMetrics();
    
    EXPECT_EQ(metrics.model_name, "MockModel");
    EXPECT_EQ(metrics.model_type, "Test");
    EXPECT_DOUBLE_EQ(metrics.accuracy_score, 0.9);
    EXPECT_EQ(metrics.avg_inference_time_ms, 50);
    EXPECT_EQ(metrics.model_size_mb, 10);
    EXPECT_EQ(metrics.description, "Mock model for testing");
}

TEST_F(DetectionModelInterfaceTest, MockModelSupportedClasses) {
    auto model = std::make_unique<MockDetectionModel>(logger_);
    auto classes = model->getSupportedClasses();
    
    EXPECT_FALSE(classes.empty());
    EXPECT_NE(std::find(classes.begin(), classes.end(), "person"), classes.end());
    EXPECT_NE(std::find(classes.begin(), classes.end(), "car"), classes.end());
    EXPECT_NE(std::find(classes.begin(), classes.end(), "cat"), classes.end());
    EXPECT_NE(std::find(classes.begin(), classes.end(), "dog"), classes.end());
}

TEST_F(DetectionModelInterfaceTest, DetectionModelFactoryCreateModel) {
    auto model = DetectionModelFactory::createModel(
        DetectionModelFactory::ModelType::YOLO_V5_SMALL, logger_);
    
    EXPECT_NE(model, nullptr);
    EXPECT_EQ(model->getModelName(), "YOLOv5 Small");
    EXPECT_FALSE(model->isInitialized());
}

TEST_F(DetectionModelInterfaceTest, DetectionModelFactoryGetAvailableModels) {
    auto available_models = DetectionModelFactory::getAvailableModels();
    
    EXPECT_GE(available_models.size(), 2);  // At least YOLOv5s and YOLOv5l
    
    // Check that YOLOv5s is included
    bool found_yolov5s = false;
    for (const auto& model : available_models) {
        if (model.model_name == "YOLOv5s") {
            found_yolov5s = true;
            EXPECT_GT(model.accuracy_score, 0.0);
            EXPECT_GT(model.avg_inference_time_ms, 0);
            EXPECT_FALSE(model.description.empty());
            break;
        }
    }
    EXPECT_TRUE(found_yolov5s);
}

TEST_F(DetectionModelInterfaceTest, DetectionModelFactoryParseModelType) {
    EXPECT_EQ(DetectionModelFactory::parseModelType("yolov5s"), 
              DetectionModelFactory::ModelType::YOLO_V5_SMALL);
    EXPECT_EQ(DetectionModelFactory::parseModelType("yolov5l"), 
              DetectionModelFactory::ModelType::YOLO_V5_LARGE);
    EXPECT_EQ(DetectionModelFactory::parseModelType("small"), 
              DetectionModelFactory::ModelType::YOLO_V5_SMALL);
    EXPECT_EQ(DetectionModelFactory::parseModelType("large"), 
              DetectionModelFactory::ModelType::YOLO_V5_LARGE);
              
    EXPECT_THROW(DetectionModelFactory::parseModelType("invalid_model"), 
                 std::invalid_argument);
}

TEST_F(DetectionModelInterfaceTest, DetectionModelFactoryModelTypeToString) {
    EXPECT_EQ(DetectionModelFactory::modelTypeToString(
        DetectionModelFactory::ModelType::YOLO_V5_SMALL), "yolov5s");
    EXPECT_EQ(DetectionModelFactory::modelTypeToString(
        DetectionModelFactory::ModelType::YOLO_V5_LARGE), "yolov5l");
    EXPECT_EQ(DetectionModelFactory::modelTypeToString(
        DetectionModelFactory::ModelType::YOLO_V8_NANO), "yolov8n");
    EXPECT_EQ(DetectionModelFactory::modelTypeToString(
        DetectionModelFactory::ModelType::YOLO_V8_MEDIUM), "yolov8m");
}

TEST_F(DetectionModelInterfaceTest, YoloV5SmallModelCreation) {
    auto model = std::make_unique<YoloV5SmallModel>(logger_);
    
    EXPECT_EQ(model->getModelName(), "YOLOv5 Small");
    EXPECT_FALSE(model->isInitialized());
    
    auto metrics = model->getMetrics();
    EXPECT_EQ(metrics.model_name, "YOLOv5s");
    EXPECT_EQ(metrics.model_type, "YOLO");
    EXPECT_GT(metrics.accuracy_score, 0.0);
    EXPECT_LT(metrics.accuracy_score, 1.0);
}

TEST_F(DetectionModelInterfaceTest, YoloV5LargeModelCreation) {
    auto model = std::make_unique<YoloV5LargeModel>(logger_);
    
    EXPECT_EQ(model->getModelName(), "YOLOv5 Large");
    EXPECT_FALSE(model->isInitialized());
    
    auto metrics = model->getMetrics();
    EXPECT_EQ(metrics.model_name, "YOLOv5l");
    EXPECT_EQ(metrics.model_type, "YOLO");
    EXPECT_GT(metrics.accuracy_score, 0.0);
    EXPECT_LT(metrics.accuracy_score, 1.0);
    
    // YOLOv5l should have higher accuracy than YOLOv5s
    auto small_model = std::make_unique<YoloV5SmallModel>(logger_);
    auto small_metrics = small_model->getMetrics();
    EXPECT_GT(metrics.accuracy_score, small_metrics.accuracy_score);
}

TEST_F(DetectionModelInterfaceTest, ModelPerformanceComparison) {
    auto available_models = DetectionModelFactory::getAvailableModels();
    
    // Find YOLOv5s and YOLOv5l models
    ModelMetrics yolov5s_metrics, yolov5l_metrics;
    bool found_both = false;
    
    for (const auto& model : available_models) {
        if (model.model_name == "YOLOv5s") {
            yolov5s_metrics = model;
        } else if (model.model_name == "YOLOv5l") {
            yolov5l_metrics = model;
            found_both = true;
        }
    }
    
    if (found_both) {
        // YOLOv5l should have higher accuracy but slower inference
        EXPECT_GT(yolov5l_metrics.accuracy_score, yolov5s_metrics.accuracy_score);
        EXPECT_GT(yolov5l_metrics.avg_inference_time_ms, yolov5s_metrics.avg_inference_time_ms);
        EXPECT_GT(yolov5l_metrics.model_size_mb, yolov5s_metrics.model_size_mb);
    }
}