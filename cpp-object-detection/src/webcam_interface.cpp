#include "webcam_interface.hpp"
#include <sstream>

WebcamInterface::WebcamInterface(int camera_id, int width, int height, 
                                std::shared_ptr<Logger> logger)
    : camera_id_(camera_id), width_(width), height_(height), 
      logger_(logger), initialized_(false) {
    capture_ = std::make_unique<cv::VideoCapture>();
}

WebcamInterface::~WebcamInterface() {
    release();
}

bool WebcamInterface::initialize() {
    if (initialized_) {
        return true;
    }

    logger_->info("Initializing webcam interface...");
    logger_->debug("Camera ID: " + std::to_string(camera_id_));
    logger_->debug("Target resolution: " + std::to_string(width_) + "x" + std::to_string(height_));

    // Try to open the camera
    if (!capture_->open(camera_id_)) {
        logger_->error("Failed to open camera with ID: " + std::to_string(camera_id_));
        return false;
    }

    // Set camera properties
    setCameraProperties();

    // Test camera capabilities
    if (!testCameraCapabilities()) {
        logger_->error("Camera capability test failed");
        capture_->release();
        return false;
    }

    initialized_ = true;
    logger_->info("Webcam interface initialized successfully");
    logger_->info(getCameraInfo());
    
    return true;
}

bool WebcamInterface::captureFrame(cv::Mat& frame) {
    if (!initialized_ || !capture_->isOpened()) {
        logger_->error("Camera not initialized or not opened");
        return false;
    }

    if (!capture_->read(frame)) {
        logger_->warning("Failed to read frame from camera");
        return false;
    }

    if (frame.empty()) {
        logger_->warning("Captured frame is empty");
        return false;
    }

    return true;
}

bool WebcamInterface::isReady() const {
    return initialized_ && capture_ && capture_->isOpened();
}

std::string WebcamInterface::getCameraInfo() const {
    if (!isReady()) {
        return "Camera not initialized";
    }

    std::stringstream ss;
    ss << "Camera " << camera_id_ << ": ";
    
    // Get actual resolution
    int actual_width = static_cast<int>(capture_->get(cv::CAP_PROP_FRAME_WIDTH));
    int actual_height = static_cast<int>(capture_->get(cv::CAP_PROP_FRAME_HEIGHT));
    ss << actual_width << "x" << actual_height;
    
    // Get FPS
    double fps = capture_->get(cv::CAP_PROP_FPS);
    if (fps > 0) {
        ss << " @ " << fps << " fps";
    }
    
    // Get backend info
    int backend = static_cast<int>(capture_->get(cv::CAP_PROP_BACKEND));
    ss << " (backend: " << backend << ")";
    
    return ss.str();
}

void WebcamInterface::release() {
    if (capture_ && capture_->isOpened()) {
        capture_->release();
        logger_->info("Camera released");
    }
    initialized_ = false;
}

bool WebcamInterface::testCameraCapabilities() {
    cv::Mat test_frame;
    
    // Try to capture a test frame
    if (!capture_->read(test_frame)) {
        logger_->error("Failed to capture test frame");
        return false;
    }
    
    if (test_frame.empty()) {
        logger_->error("Test frame is empty");
        return false;
    }
    
    logger_->debug("Test frame captured successfully: " + 
                   std::to_string(test_frame.cols) + "x" + std::to_string(test_frame.rows));
    
    // Check if we got a reasonable frame size
    if (test_frame.cols < 320 || test_frame.rows < 240) {
        logger_->warning("Camera resolution seems very low: " + 
                        std::to_string(test_frame.cols) + "x" + std::to_string(test_frame.rows));
    }
    
    return true;
}

void WebcamInterface::setCameraProperties() {
    // Set resolution
    capture_->set(cv::CAP_PROP_FRAME_WIDTH, width_);
    capture_->set(cv::CAP_PROP_FRAME_HEIGHT, height_);
    
    // Set FPS (try for 30 fps, camera will use what it supports)
    capture_->set(cv::CAP_PROP_FPS, 30);
    
    // Set buffer size to minimize latency
    capture_->set(cv::CAP_PROP_BUFFERSIZE, 1);
    
    // Set format to MJPEG if available (better for USB cameras)
    capture_->set(cv::CAP_PROP_FOURCC, cv::VideoWriter::fourcc('M', 'J', 'P', 'G'));
    
    // Log actual settings
    int actual_width = static_cast<int>(capture_->get(cv::CAP_PROP_FRAME_WIDTH));
    int actual_height = static_cast<int>(capture_->get(cv::CAP_PROP_FRAME_HEIGHT));
    double actual_fps = capture_->get(cv::CAP_PROP_FPS);
    
    logger_->debug("Camera properties set - Actual resolution: " + 
                   std::to_string(actual_width) + "x" + std::to_string(actual_height) + 
                   ", FPS: " + std::to_string(actual_fps));
    
    if (actual_width != width_ || actual_height != height_) {
        logger_->warning("Camera resolution differs from requested: got " + 
                        std::to_string(actual_width) + "x" + std::to_string(actual_height) + 
                        ", requested " + std::to_string(width_) + "x" + std::to_string(height_));
    }
}