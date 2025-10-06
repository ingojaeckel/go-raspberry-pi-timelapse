#include "viewfinder_window.hpp"
#include "drawing_utils.hpp"

ViewfinderWindow::ViewfinderWindow(std::shared_ptr<Logger> logger, 
                                   const std::string& window_name)
    : logger_(logger), window_name_(window_name), initialized_(false), show_debug_info_(true) {
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
    
    // Toggle debug info with space key
    if (key == ' ' || key == 32) {
        show_debug_info_ = !show_debug_info_;
        logger_->info(show_debug_info_ ? "Debug info enabled" : "Debug info disabled");
    }
    
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
        std::string label = detection.class_name + " (" + 
                           std::to_string(static_cast<int>(detection.confidence * 100)) + "%)";
        
        // Add stationary indicator if object is stationary
        if (detection.is_stationary) {
            label += ", stationary";
            
            // Add duration if available
            if (detection.stationary_duration_seconds > 0) {
                int duration = detection.stationary_duration_seconds;
                if (duration < 60) {
                    label += " for " + std::to_string(duration) + " sec";
                } else {
                    int minutes = duration / 60;
                    label += " for " + std::to_string(minutes) + " min";
                }
            }
        }
        
        // Draw label with auto-positioning to avoid cutoff at screen edges
        DrawingUtils::drawBoundingBoxLabel(annotated_frame, label, detection.bbox, color);
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

void ViewfinderWindow::showFrameWithStats(const cv::Mat& frame, 
                                         const std::vector<Detection>& detections,
                                         double current_fps,
                                         double avg_processing_time_ms,
                                         int total_objects_detected,
                                         int total_images_saved,
                                         const std::chrono::steady_clock::time_point& start_time,
                                         const std::vector<std::pair<std::string, int>>& top_objects,
                                         int camera_width,
                                         int camera_height,
                                         int camera_id,
                                         const std::string& camera_name,
                                         int detection_width,
                                         int detection_height,
                                         bool brightness_filter_active,
                                         bool gpu_enabled,
                                         bool burst_mode_enabled,
                                         double disk_usage_percent,
                                         double cpu_temp_celsius) {
    if (!initialized_ || frame.empty()) {
        return;
    }

    try {
        // Draw bounding boxes on the frame
        cv::Mat annotated_frame = drawBoundingBoxes(frame, detections);
        
        // Draw debug info if enabled
        if (show_debug_info_) {
            drawDebugInfo(annotated_frame, current_fps, avg_processing_time_ms,
                         total_objects_detected, total_images_saved, start_time,
                         top_objects, camera_width, camera_height, camera_id,
                         camera_name, detection_width, detection_height, brightness_filter_active,
                         gpu_enabled, burst_mode_enabled, disk_usage_percent, cpu_temp_celsius);
        }
        
        // Display the frame
        cv::imshow(window_name_, annotated_frame);
        
        // Process window events (1ms wait allows window to update)
        cv::waitKey(1);
    } catch (const cv::Exception& e) {
        logger_->error("Error displaying frame in viewfinder: " + std::string(e.what()));
    }
}

void ViewfinderWindow::drawDebugInfo(cv::Mat& frame,
                                    double current_fps,
                                    double avg_processing_time_ms,
                                    int total_objects_detected,
                                    int total_images_saved,
                                    const std::chrono::steady_clock::time_point& start_time,
                                    const std::vector<std::pair<std::string, int>>& top_objects,
                                    int camera_width,
                                    int camera_height,
                                    int camera_id,
                                    const std::string& camera_name,
                                    int detection_width,
                                    int detection_height,
                                    bool brightness_filter_active,
                                    bool gpu_enabled,
                                    bool burst_mode_enabled,
                                    double disk_usage_percent,
                                    double cpu_temp_celsius) {
    // Use small font to minimize screen coverage
    const double font_scale = 0.4;
    const int font_thickness = 1;
    const int font_face = cv::FONT_HERSHEY_SIMPLEX;
    const cv::Scalar text_color(255, 255, 255);  // White text
    const cv::Scalar bg_color(0, 0, 0);  // Black background with transparency
    const int line_spacing = 15;
    const int padding = 5;
    
    int y_offset = padding + 12;
    int x_offset = padding;
    
    // Calculate uptime
    auto now = std::chrono::steady_clock::now();
    auto uptime_duration = std::chrono::duration_cast<std::chrono::seconds>(now - start_time);
    int uptime_hours = uptime_duration.count() / 3600;
    int uptime_minutes = (uptime_duration.count() % 3600) / 60;
    int uptime_seconds = uptime_duration.count() % 60;
    
    std::vector<std::string> lines;
    
    // Performance metrics
    char fps_str[32];
    snprintf(fps_str, sizeof(fps_str), "FPS: %.1f", current_fps);
    lines.push_back(std::string(fps_str));
    lines.push_back("Avg proc: " + std::to_string(static_cast<int>(avg_processing_time_ms)) + " ms");
    
    // Counters
    lines.push_back("Objects: " + std::to_string(total_objects_detected));
    lines.push_back("Images: " + std::to_string(total_images_saved));
    
    // Uptime
    char uptime_str[32];
    snprintf(uptime_str, sizeof(uptime_str), "Uptime: %02d:%02d:%02d", uptime_hours, uptime_minutes, uptime_seconds);
    lines.push_back(std::string(uptime_str));
    
    // Camera info
    std::string cam_name = camera_name.empty() ? "Camera " + std::to_string(camera_id) : camera_name;
    lines.push_back(cam_name + ": " + std::to_string(camera_width) + "x" + std::to_string(camera_height));
    
    // Detection resolution
    lines.push_back("Detection: " + std::to_string(detection_width) + "x" + std::to_string(detection_height));
    
    // GPU and burst mode status
    lines.push_back("GPU: " + std::string(gpu_enabled ? "ON" : "OFF"));
    lines.push_back("Burst: " + std::string(burst_mode_enabled ? "ON" : "OFF"));
    
    // System monitor metrics
    if (disk_usage_percent >= 0.0) {
        char disk_str[32];
        snprintf(disk_str, sizeof(disk_str), "Disk: %.1f%%", disk_usage_percent);
        lines.push_back(std::string(disk_str));
    }
    if (cpu_temp_celsius >= 0.0) {
        char temp_str[32];
        snprintf(temp_str, sizeof(temp_str), "CPU: %.1f°C", cpu_temp_celsius);
        lines.push_back(std::string(temp_str));
    }
    
    // Top detected objects
    if (!top_objects.empty()) {
        lines.push_back("--- Top Objects ---");
        int count = 0;
        for (const auto& [obj_type, obj_count] : top_objects) {
            if (count >= 10) break;
            lines.push_back(obj_type + ": " + std::to_string(obj_count));
            count++;
        }
    }
    
    // Draw semi-transparent background for all text
    int max_width = 0;
    for (const auto& line : lines) {
        int baseline;
        cv::Size text_size = cv::getTextSize(line, font_face, font_scale, font_thickness, &baseline);
        max_width = std::max(max_width, text_size.width);
    }
    
    // Draw background rectangle with transparency
    cv::Mat overlay = frame.clone();
    cv::rectangle(overlay, 
                 cv::Point(0, 0),
                 cv::Point(max_width + 2 * padding, y_offset + lines.size() * line_spacing + padding),
                 bg_color, cv::FILLED);
    cv::addWeighted(overlay, 0.6, frame, 0.4, 0, frame);
    
    // Draw all text lines
    for (const auto& line : lines) {
        cv::putText(frame, line, cv::Point(x_offset, y_offset),
                   font_face, font_scale, text_color, font_thickness);
        y_offset += line_spacing;
    }
    
    // Draw brightness filter indicator in top right corner if active
    if (brightness_filter_active) {
        const std::string filter_text = "High brightness filter ON";
        const double indicator_font_scale = 0.5;
        const int indicator_font_thickness = 1;
        const cv::Scalar indicator_bg_color(0, 100, 200);  // Orange background
        const cv::Scalar indicator_text_color(255, 255, 255);  // White text
        
        // Get text size
        int baseline;
        cv::Size text_size = cv::getTextSize(filter_text, font_face, indicator_font_scale, 
                                             indicator_font_thickness, &baseline);
        
        // Position in top right corner with some padding
        int indicator_x = frame.cols - text_size.width - 10;
        int indicator_y = 10;
        
        // Draw background rectangle with transparency
        cv::Mat indicator_overlay = frame.clone();
        cv::rectangle(indicator_overlay, 
                     cv::Point(indicator_x - 5, indicator_y - 2),
                     cv::Point(indicator_x + text_size.width + 5, indicator_y + text_size.height + 5),
                     indicator_bg_color, cv::FILLED);
        cv::addWeighted(indicator_overlay, 0.7, frame, 0.3, 0, frame);
        
        // Draw text
        cv::putText(frame, filter_text, cv::Point(indicator_x, indicator_y + text_size.height),
                   font_face, indicator_font_scale, indicator_text_color, indicator_font_thickness);
    }
}

