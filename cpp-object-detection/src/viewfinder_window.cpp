#include "viewfinder_window.hpp"

ViewfinderWindow::ViewfinderWindow(std::shared_ptr<Logger> logger, 
                                   const std::string& window_name)
    : logger_(logger), window_name_(window_name), initialized_(false) {
}

ViewfinderWindow::~ViewfinderWindow() {
    close();
}

bool ViewfinderWindow::initialize() {
    if (initialized_) {
        return true;
    }

    logger_->info("Initializing viewfinder window: " + window_name_);
    
    try {
        // Create the named window
        cv::namedWindow(window_name_, cv::WINDOW_NORMAL);
        
        // Set window to be resizable
        cv::resizeWindow(window_name_, 1280, 720);
        
        initialized_ = true;
        logger_->info("Viewfinder window initialized successfully");
        return true;
    } catch (const cv::Exception& e) {
        logger_->error("Failed to initialize viewfinder window: " + std::string(e.what()));
        return false;
    }
}

void ViewfinderWindow::showFrame(const cv::Mat& frame, const std::vector<Detection>& detections) {
    if (!initialized_ || frame.empty()) {
        return;
    }

    try {
        // Draw bounding boxes on the frame
        cv::Mat annotated_frame = drawBoundingBoxes(frame, detections);
        
        // Display the frame
        cv::imshow(window_name_, annotated_frame);
        
        // Process window events (1ms wait allows window to update)
        cv::waitKey(1);
    } catch (const cv::Exception& e) {
        logger_->error("Error displaying frame in viewfinder: " + std::string(e.what()));
    }
}

bool ViewfinderWindow::shouldClose() {
    if (!initialized_) {
        return true;
    }

    // Check if user pressed 'q' or 'ESC' key
    int key = cv::waitKey(1);
    return (key == 'q' || key == 27);  // 27 is ESC key
}

void ViewfinderWindow::close() {
    if (!initialized_) {
        return;
    }

    logger_->info("Closing viewfinder window");
    
    try {
        cv::destroyWindow(window_name_);
        initialized_ = false;
    } catch (const cv::Exception& e) {
        logger_->error("Error closing viewfinder window: " + std::string(e.what()));
    }
}

cv::Mat ViewfinderWindow::drawBoundingBoxes(const cv::Mat& frame, 
                                            const std::vector<Detection>& detections) {
    // Create a copy of the frame to draw on
    cv::Mat annotated_frame = frame.clone();
    
    // Draw bounding boxes for each detection
    for (const auto& detection : detections) {
        cv::Scalar color = getColorForClass(detection.class_name);
        
        // Draw rectangle around the object
        cv::rectangle(annotated_frame, detection.bbox, color, 2);
        
        // Draw label with class name and confidence
        std::string label = detection.class_name + " " + 
                           std::to_string(static_cast<int>(detection.confidence * 100)) + "%";
        int baseline;
        cv::Size text_size = cv::getTextSize(label, cv::FONT_HERSHEY_SIMPLEX, 0.5, 1, &baseline);
        
        // Draw label background
        cv::Point text_origin(detection.bbox.x, detection.bbox.y - 5);
        cv::rectangle(annotated_frame, 
                     cv::Point(text_origin.x, text_origin.y - text_size.height - 2),
                     cv::Point(text_origin.x + text_size.width, text_origin.y + 2),
                     color, cv::FILLED);
        
        // Draw label text
        cv::putText(annotated_frame, label, text_origin, 
                   cv::FONT_HERSHEY_SIMPLEX, 0.5, cv::Scalar(0, 0, 0), 1);
    }
    
    return annotated_frame;
}

cv::Scalar ViewfinderWindow::getColorForClass(const std::string& class_name) const {
    // Define colors for different object types (BGR format)
    // Green for persons, red for cats, blue for dogs, yellow for vehicles, etc.
    if (class_name == "person") {
        return cv::Scalar(0, 255, 0);  // Green
    } else if (class_name == "cat") {
        return cv::Scalar(0, 0, 255);  // Red
    } else if (class_name == "dog") {
        return cv::Scalar(255, 0, 0);  // Blue
    } else if (class_name == "car" || class_name == "truck" || class_name == "bus") {
        return cv::Scalar(0, 255, 255);  // Yellow
    } else if (class_name == "motorcycle" || class_name == "bicycle") {
        return cv::Scalar(255, 0, 255);  // Magenta
    } else {
        return cv::Scalar(255, 255, 255);  // White for unknown
    }
}
