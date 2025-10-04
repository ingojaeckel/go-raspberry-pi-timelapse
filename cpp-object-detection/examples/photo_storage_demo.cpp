// Example usage of photo storage feature
// This demonstrates how the photo storage with bounding boxes works

#include "parallel_frame_processor.hpp"
#include "object_detector.hpp"
#include "logger.hpp"
#include "performance_monitor.hpp"
#include <opencv2/opencv.hpp>
#include <memory>
#include <iostream>

int main() {
    // Example 1: Create a frame processor with default output directory
    {
        std::cout << "=== Example 1: Default output directory ===" << std::endl;
        
        auto logger = std::make_shared<Logger>("example.log", true);
        auto perf_monitor = std::make_shared<PerformanceMonitor>(logger, 1.0);
        auto detector = std::make_shared<ObjectDetector>(
            "models/yolov5s.onnx", 
            "models/yolov5s.yaml", 
            "models/coco.names", 
            0.5, 
            logger);
        
        // Uses default output directory "detections"
        auto processor = std::make_shared<ParallelFrameProcessor>(
            detector, logger, perf_monitor, 1, 10);
        
        processor->initialize();
        
        std::cout << "Photos will be saved to: detections/" << std::endl;
        std::cout << "Example filename: 2025-10-04 143022 person detected.jpg" << std::endl;
        
        processor->shutdown();
    }
    
    // Example 2: Create a frame processor with custom output directory
    {
        std::cout << "\n=== Example 2: Custom output directory ===" << std::endl;
        
        auto logger = std::make_shared<Logger>("example.log", true);
        auto perf_monitor = std::make_shared<PerformanceMonitor>(logger, 1.0);
        auto detector = std::make_shared<ObjectDetector>(
            "models/yolov5s.onnx", 
            "models/yolov5s.yaml", 
            "models/coco.names", 
            0.5, 
            logger);
        
        // Custom output directory
        auto processor = std::make_shared<ParallelFrameProcessor>(
            detector, logger, perf_monitor, 1, 10, "/tmp/my-detections");
        
        processor->initialize();
        
        std::cout << "Photos will be saved to: /tmp/my-detections/" << std::endl;
        
        processor->shutdown();
    }
    
    // Example 3: Simulated detection scenario
    {
        std::cout << "\n=== Example 3: Simulated detection scenario ===" << std::endl;
        std::cout << "When a frame with detections is processed:" << std::endl;
        std::cout << "1. Objects are detected (person, cat, dog, etc.)" << std::endl;
        std::cout << "2. Center coordinates are logged:" << std::endl;
        std::cout << "   detected person at coordinates: (640, 360) with confidence 92%" << std::endl;
        std::cout << "   detected cat at coordinates: (320, 240) with confidence 87%" << std::endl;
        std::cout << "3. Photo is saved (if 10 seconds have passed since last photo):" << std::endl;
        std::cout << "   Saved detection photo: detections/2025-10-04 143022 person cat detected.jpg" << std::endl;
        std::cout << "4. Bounding boxes are drawn with colors:" << std::endl;
        std::cout << "   - Green box around person" << std::endl;
        std::cout << "   - Red box around cat" << std::endl;
    }
    
    // Example 4: Color mapping reference
    {
        std::cout << "\n=== Example 4: Color mapping reference ===" << std::endl;
        std::cout << "Object Type        | Color   | BGR Value" << std::endl;
        std::cout << "-------------------|---------|---------------" << std::endl;
        std::cout << "person             | Green   | (0, 255, 0)" << std::endl;
        std::cout << "cat                | Red     | (0, 0, 255)" << std::endl;
        std::cout << "dog                | Blue    | (255, 0, 0)" << std::endl;
        std::cout << "car/truck/bus      | Yellow  | (0, 255, 255)" << std::endl;
        std::cout << "motorcycle/bicycle | Magenta | (255, 0, 255)" << std::endl;
        std::cout << "other              | White   | (255, 255, 255)" << std::endl;
    }
    
    std::cout << "\n=== Feature Summary ===" << std::endl;
    std::cout << "✅ Photos saved with timestamped filenames" << std::endl;
    std::cout << "✅ Bounding boxes drawn around detected objects" << std::endl;
    std::cout << "✅ Different colors for different object types" << std::endl;
    std::cout << "✅ Center coordinates logged to console" << std::endl;
    std::cout << "✅ Rate limited to 1 photo every 10 seconds" << std::endl;
    std::cout << "✅ Thread-safe implementation" << std::endl;
    
    return 0;
}
