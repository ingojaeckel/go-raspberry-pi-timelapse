#include "network_streamer.hpp"
#include "drawing_utils.hpp"
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <unistd.h>
#include <ifaddrs.h>
#include <cerrno>
#include <cstring>
#include <sstream>

NetworkStreamer::NetworkStreamer(std::shared_ptr<Logger> logger, int port)
    : logger_(logger), port_(port), running_(false), initialized_(false), server_socket_(-1) {
}

NetworkStreamer::~NetworkStreamer() {
    stop();
}

bool NetworkStreamer::initialize() {
    if (initialized_) {
        return true;
    }

    logger_->info("Initializing network streamer on port " + std::to_string(port_));
    
    // Create socket
    server_socket_ = socket(AF_INET, SOCK_STREAM, 0);
    if (server_socket_ < 0) {
        logger_->error("Failed to create socket: " + std::string(strerror(errno)));
        return false;
    }

    // Allow socket reuse
    int opt = 1;
    if (setsockopt(server_socket_, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt)) < 0) {
        logger_->error("Failed to set socket options: " + std::string(strerror(errno)));
        close(server_socket_);
        server_socket_ = -1;
        return false;
    }

    // Bind socket to port
    struct sockaddr_in server_addr;
    memset(&server_addr, 0, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_addr.s_addr = INADDR_ANY;
    server_addr.sin_port = htons(port_);

    if (bind(server_socket_, (struct sockaddr*)&server_addr, sizeof(server_addr)) < 0) {
        logger_->error("Failed to bind socket to port " + std::to_string(port_) + ": " + std::string(strerror(errno)));
        close(server_socket_);
        server_socket_ = -1;
        return false;
    }

    // Start listening
    if (listen(server_socket_, 5) < 0) {
        logger_->error("Failed to listen on socket: " + std::string(strerror(errno)));
        close(server_socket_);
        server_socket_ = -1;
        return false;
    }

    initialized_ = true;
    logger_->info("Network streamer initialized successfully");
    return true;
}

bool NetworkStreamer::start() {
    if (!initialized_) {
        logger_->error("Cannot start streamer - not initialized");
        return false;
    }

    if (running_) {
        logger_->warning("Streamer already running");
        return true;
    }

    running_ = true;
    server_thread_ = std::thread(&NetworkStreamer::serverLoop, this);
    
    std::string url = getStreamingUrl();
    logger_->info("Network streaming server started");
    logger_->info("Stream available at: " + url);
    logger_->info("Open in browser or VLC to view live feed with object detection");
    
    return true;
}

void NetworkStreamer::stop() {
    if (!running_) {
        return;
    }

    logger_->info("Stopping network streamer...");
    running_ = false;

    // Close server socket to unblock accept()
    if (server_socket_ >= 0) {
        shutdown(server_socket_, SHUT_RDWR);
        close(server_socket_);
        server_socket_ = -1;
    }

    // Wait for server thread to finish
    if (server_thread_.joinable()) {
        server_thread_.join();
    }

    initialized_ = false;
    logger_->info("Network streamer stopped");
}

void NetworkStreamer::updateFrame(const cv::Mat& frame, const std::vector<Detection>& detections) {
    if (frame.empty()) {
        return;
    }

    // Draw bounding boxes on the frame
    cv::Mat annotated_frame = drawBoundingBoxes(frame, detections);

    // Update current frame (thread-safe)
    std::lock_guard<std::mutex> lock(frame_mutex_);
    current_frame_ = annotated_frame.clone();
}

void NetworkStreamer::updateFrameWithStats(const cv::Mat& frame, 
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
                                          bool brightness_filter_active) {
    if (frame.empty()) {
        return;
    }

    // Draw bounding boxes on the frame
    cv::Mat annotated_frame = drawBoundingBoxes(frame, detections);
    
    // Draw debug info overlay
    drawDebugInfo(annotated_frame, current_fps, avg_processing_time_ms,
                 total_objects_detected, total_images_saved, start_time,
                 top_objects, camera_width, camera_height, camera_id,
                 camera_name, detection_width, detection_height, brightness_filter_active);

    // Update current frame (thread-safe)
    std::lock_guard<std::mutex> lock(frame_mutex_);
    current_frame_ = annotated_frame.clone();
}

std::string NetworkStreamer::getStreamingUrl() const {
    std::string ip = getLocalIpAddress();
    return "http://" + ip + ":" + std::to_string(port_) + "/stream";
}

void NetworkStreamer::serverLoop() {
    logger_->info("Server loop started");

    while (running_) {
        // Accept incoming connection
        struct sockaddr_in client_addr;
        socklen_t client_len = sizeof(client_addr);
        int client_socket = accept(server_socket_, (struct sockaddr*)&client_addr, &client_len);

        if (client_socket < 0) {
            if (running_) {
                logger_->error("Failed to accept connection: " + std::string(strerror(errno)));
            }
            continue;
        }

        // Get client IP
        char client_ip[INET_ADDRSTRLEN];
        inet_ntop(AF_INET, &client_addr.sin_addr, client_ip, INET_ADDRSTRLEN);
        logger_->info("New client connected from " + std::string(client_ip));

        // Handle client in this thread (simple single-client implementation)
        // For multiple clients, spawn a new thread here
        handleClient(client_socket);
        
        close(client_socket);
        logger_->info("Client disconnected");
    }

    logger_->info("Server loop ended");
}

void NetworkStreamer::handleClient(int client_socket) {
    // Send HTTP headers for MJPEG stream
    std::string headers = 
        "HTTP/1.1 200 OK\r\n"
        "Content-Type: multipart/x-mixed-replace; boundary=frame\r\n"
        "Cache-Control: no-cache\r\n"
        "Connection: keep-alive\r\n"
        "\r\n";
    
    if (send(client_socket, headers.c_str(), headers.length(), MSG_NOSIGNAL) < 0) {
        logger_->error("Failed to send headers to client");
        return;
    }

    // Stream frames
    while (running_) {
        cv::Mat frame_to_send;
        
        // Get current frame (thread-safe)
        {
            std::lock_guard<std::mutex> lock(frame_mutex_);
            if (current_frame_.empty()) {
                // No frame available yet, wait a bit
                std::this_thread::sleep_for(std::chrono::milliseconds(100));
                continue;
            }
            frame_to_send = current_frame_.clone();
        }

        // Encode frame as JPEG
        std::vector<uchar> jpeg_data = encodeFrameAsJpeg(frame_to_send);
        if (jpeg_data.empty()) {
            logger_->warning("Failed to encode frame as JPEG");
            continue;
        }

        // Send frame
        std::ostringstream frame_header;
        frame_header << "--frame\r\n"
                    << "Content-Type: image/jpeg\r\n"
                    << "Content-Length: " << jpeg_data.size() << "\r\n"
                    << "\r\n";
        
        std::string header_str = frame_header.str();
        
        // Send header
        if (send(client_socket, header_str.c_str(), header_str.length(), MSG_NOSIGNAL) < 0) {
            logger_->debug("Client disconnected (header send failed)");
            break;
        }

        // Send JPEG data
        if (send(client_socket, jpeg_data.data(), jpeg_data.size(), MSG_NOSIGNAL) < 0) {
            logger_->debug("Client disconnected (data send failed)");
            break;
        }

        // Send boundary
        const char* boundary = "\r\n";
        if (send(client_socket, boundary, strlen(boundary), MSG_NOSIGNAL) < 0) {
            logger_->debug("Client disconnected (boundary send failed)");
            break;
        }

        // Control frame rate (aim for ~10 fps to reduce bandwidth)
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
    }
}

std::vector<uchar> NetworkStreamer::encodeFrameAsJpeg(const cv::Mat& frame) {
    std::vector<uchar> jpeg_data;
    std::vector<int> params = {cv::IMWRITE_JPEG_QUALITY, 80}; // 80% quality
    
    try {
        cv::imencode(".jpg", frame, jpeg_data, params);
    } catch (const cv::Exception& e) {
        logger_->error("Error encoding frame as JPEG: " + std::string(e.what()));
        return std::vector<uchar>();
    }
    
    return jpeg_data;
}

cv::Mat NetworkStreamer::drawBoundingBoxes(const cv::Mat& frame, 
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
        }
        
        // Draw label with auto-positioning to avoid cutoff at screen edges
        DrawingUtils::drawBoundingBoxLabel(annotated_frame, label, detection.bbox, color);
    }
    
    return annotated_frame;
}

void NetworkStreamer::drawDebugInfo(cv::Mat& frame,
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
                                   bool brightness_filter_active) {
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
    lines.push_back("FPS: " + std::to_string(static_cast<int>(current_fps)));
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

cv::Scalar NetworkStreamer::getColorForClass(const std::string& class_name) const {
    // Define colors for different object types (BGR format)
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

std::string NetworkStreamer::getLocalIpAddress() const {
    struct ifaddrs *ifaddr, *ifa;
    std::string ip_address = "127.0.0.1";  // Default to localhost

    if (getifaddrs(&ifaddr) == -1) {
        logger_->error("Failed to get network interfaces");
        return ip_address;
    }

    // Look for first non-loopback IPv4 address
    for (ifa = ifaddr; ifa != nullptr; ifa = ifa->ifa_next) {
        if (ifa->ifa_addr == nullptr) {
            continue;
        }

        // Check for IPv4
        if (ifa->ifa_addr->sa_family == AF_INET) {
            char addr_str[INET_ADDRSTRLEN];
            struct sockaddr_in* addr = (struct sockaddr_in*)ifa->ifa_addr;
            inet_ntop(AF_INET, &addr->sin_addr, addr_str, INET_ADDRSTRLEN);
            
            // Skip loopback
            std::string addr_string(addr_str);
            if (addr_string != "127.0.0.1") {
                ip_address = addr_string;
                break;
            }
        }
    }

    freeifaddrs(ifaddr);
    return ip_address;
}
