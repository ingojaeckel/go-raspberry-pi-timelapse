#include "network_streamer.hpp"
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
    
    if (send(client_socket, headers.c_str(), headers.length(), 0) < 0) {
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
        if (send(client_socket, header_str.c_str(), header_str.length(), 0) < 0) {
            logger_->debug("Client disconnected (header send failed)");
            break;
        }

        // Send JPEG data
        if (send(client_socket, jpeg_data.data(), jpeg_data.size(), 0) < 0) {
            logger_->debug("Client disconnected (data send failed)");
            break;
        }

        // Send boundary
        const char* boundary = "\r\n";
        if (send(client_socket, boundary, strlen(boundary), 0) < 0) {
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
            std::string addr(addr_str);
            if (addr != "127.0.0.1") {
                ip_address = addr;
                break;
            }
        }
    }

    freeifaddrs(ifaddr);
    return ip_address;
}
