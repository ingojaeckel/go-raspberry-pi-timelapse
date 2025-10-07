#include "notification_manager.hpp"
#include <curl/curl.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <unistd.h>
#include <sstream>
#include <iomanip>
#include <fstream>
#include <ctime>

NotificationManager::NotificationManager(std::shared_ptr<Logger> logger, const NotificationConfig& config)
    : logger_(logger), config_(config), running_(false), initialized_(false), sse_server_socket_(-1), sse_clients_() {
}

NotificationManager::~NotificationManager() {
    stop();
}

bool NotificationManager::initialize() {
    if (initialized_) {
        return true;
    }
    
    logger_->info("Initializing notification manager...");
    
    if (config_.enable_webhook) {
        logger_->info("Webhook notifications enabled: " + config_.webhook_url);
        // Initialize CURL globally (once)
        curl_global_init(CURL_GLOBAL_DEFAULT);
    }
    
    if (config_.enable_sse) {
        logger_->info("SSE notifications enabled on port " + std::to_string(config_.sse_port));
        startSSEServer();
    }
    
    if (config_.enable_file_notification) {
        logger_->info("File notifications enabled: " + config_.notification_file_path);
    }
    
    if (config_.enable_stdio_notification) {
        logger_->info("Stdio notifications enabled");
    }
    
    initialized_ = true;
    return true;
}

void NotificationManager::notifyNewObject(const NotificationData& data) {
    if (!initialized_ || !isEnabled()) {
        return;
    }
    
    logger_->debug("Sending notifications for new object: " + data.object_type);
    
    // Send notifications through all enabled channels
    if (config_.enable_webhook) {
        sendWebhookNotification(data);
    }
    
    if (config_.enable_sse) {
        sendSSENotification(data);
    }
    
    if (config_.enable_file_notification) {
        sendFileNotification(data);
    }
    
    if (config_.enable_stdio_notification) {
        sendStdioNotification(data);
    }
}

void NotificationManager::stop() {
    if (!running_) {
        return;
    }
    
    running_ = false;
    
    // Close SSE server
    if (sse_server_socket_ >= 0) {
        close(sse_server_socket_);
        sse_server_socket_ = -1;
    }
    
    // Close all SSE client connections
    {
        std::lock_guard<std::mutex> lock(sse_clients_mutex_);
        for (int client : sse_clients_) {
            close(client);
        }
        sse_clients_.clear();
    }
    
    // Wait for SSE server thread
    if (sse_server_thread_.joinable()) {
        sse_server_thread_.join();
    }
    
    // Cleanup CURL
    if (config_.enable_webhook) {
        curl_global_cleanup();
    }
    
    initialized_ = false;
}

bool NotificationManager::isEnabled() const {
    return config_.enable_webhook || config_.enable_sse || 
           config_.enable_file_notification || config_.enable_stdio_notification;
}

// Callback for CURL to capture response
static size_t WriteCallback(void* contents, size_t size, size_t nmemb, void* userp) {
    ((std::string*)userp)->append((char*)contents, size * nmemb);
    return size * nmemb;
}

void NotificationManager::sendWebhookNotification(const NotificationData& data) {
    CURL* curl = curl_easy_init();
    if (!curl) {
        logger_->error("Failed to initialize CURL for webhook notification");
        return;
    }
    
    std::string json_payload = createNotificationJSON(data);
    std::string response;
    
    struct curl_slist* headers = nullptr;
    headers = curl_slist_append(headers, "Content-Type: application/json");
    
    curl_easy_setopt(curl, CURLOPT_URL, config_.webhook_url.c_str());
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json_payload.c_str());
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, WriteCallback);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &response);
    curl_easy_setopt(curl, CURLOPT_TIMEOUT, 5L); // 5 second timeout
    
    CURLcode res = curl_easy_perform(curl);
    
    if (res != CURLE_OK) {
        logger_->error("Webhook notification failed: " + std::string(curl_easy_strerror(res)));
    } else {
        long response_code;
        curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &response_code);
        logger_->debug("Webhook notification sent, response code: " + std::to_string(response_code));
    }
    
    curl_slist_free_all(headers);
    curl_easy_cleanup(curl);
}

void NotificationManager::startSSEServer() {
    running_ = true;
    
    sse_server_thread_ = std::thread([this]() {
        // Create socket
        sse_server_socket_ = socket(AF_INET, SOCK_STREAM, 0);
        if (sse_server_socket_ < 0) {
            logger_->error("Failed to create SSE server socket");
            return;
        }
        
        // Set socket options
        int opt = 1;
        if (setsockopt(sse_server_socket_, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt)) < 0) {
            logger_->error("Failed to set SSE socket options");
            close(sse_server_socket_);
            return;
        }
        
        // Bind socket
        struct sockaddr_in address;
        address.sin_family = AF_INET;
        address.sin_addr.s_addr = INADDR_ANY;
        address.sin_port = htons(config_.sse_port);
        
        if (bind(sse_server_socket_, (struct sockaddr*)&address, sizeof(address)) < 0) {
            logger_->error("Failed to bind SSE server socket to port " + std::to_string(config_.sse_port));
            close(sse_server_socket_);
            return;
        }
        
        // Listen
        if (listen(sse_server_socket_, 10) < 0) {
            logger_->error("Failed to listen on SSE server socket");
            close(sse_server_socket_);
            return;
        }
        
        logger_->info("SSE server listening on port " + std::to_string(config_.sse_port));
        
        // Accept connections
        while (running_) {
            fd_set readfds;
            FD_ZERO(&readfds);
            FD_SET(sse_server_socket_, &readfds);
            
            struct timeval tv;
            tv.tv_sec = 1;
            tv.tv_usec = 0;
            
            int activity = select(sse_server_socket_ + 1, &readfds, nullptr, nullptr, &tv);
            
            if (activity < 0 && errno != EINTR) {
                logger_->error("SSE server select error");
                break;
            }
            
            if (activity > 0 && FD_ISSET(sse_server_socket_, &readfds)) {
                struct sockaddr_in client_address;
                socklen_t addrlen = sizeof(client_address);
                int client_socket = accept(sse_server_socket_, (struct sockaddr*)&client_address, &addrlen);
                
                if (client_socket >= 0) {
                    logger_->info("New SSE client connected: " + std::string(inet_ntoa(client_address.sin_addr)));
                    
                    // Send SSE headers
                    std::string sse_headers = 
                        "HTTP/1.1 200 OK\r\n"
                        "Content-Type: text/event-stream\r\n"
                        "Cache-Control: no-cache\r\n"
                        "Connection: keep-alive\r\n"
                        "Access-Control-Allow-Origin: *\r\n"
                        "\r\n";
                    send(client_socket, sse_headers.c_str(), sse_headers.length(), 0);
                    
                    // Add to clients list
                    {
                        std::lock_guard<std::mutex> lock(sse_clients_mutex_);
                        sse_clients_.push_back(client_socket);
                    }
                }
            }
        }
    });
}

void NotificationManager::sendSSENotification(const NotificationData& data) {
    std::string json_payload = createNotificationJSON(data);
    std::string sse_message = "data: " + json_payload + "\n\n";
    broadcastSSEMessage(sse_message);
}

void NotificationManager::broadcastSSEMessage(const std::string& message) {
    std::lock_guard<std::mutex> lock(sse_clients_mutex_);
    
    std::vector<int> disconnected_clients;
    
    for (int client : sse_clients_) {
        ssize_t sent = send(client, message.c_str(), message.length(), MSG_NOSIGNAL);
        if (sent < 0) {
            disconnected_clients.push_back(client);
        }
    }
    
    // Remove disconnected clients
    for (int client : disconnected_clients) {
        close(client);
        sse_clients_.erase(std::remove(sse_clients_.begin(), sse_clients_.end(), client), sse_clients_.end());
        logger_->debug("SSE client disconnected");
    }
}

void NotificationManager::sendFileNotification(const NotificationData& data) {
    std::string json_payload = createNotificationJSON(data);
    
    try {
        std::ofstream file(config_.notification_file_path, std::ios::app);
        if (file.is_open()) {
            file << json_payload << std::endl;
            file.close();
            logger_->debug("File notification written to: " + config_.notification_file_path);
        } else {
            logger_->error("Failed to open notification file: " + config_.notification_file_path);
        }
    } catch (const std::exception& e) {
        logger_->error("File notification error: " + std::string(e.what()));
    }
}

void NotificationManager::sendStdioNotification(const NotificationData& data) {
    std::string json_payload = createNotificationJSON(data);
    
    // Output to stdout with clear delimiters
    std::cout << "=== NEW OBJECT NOTIFICATION ===" << std::endl;
    std::cout << json_payload << std::endl;
    std::cout << "===============================" << std::endl;
    std::cout.flush();
}

std::string NotificationManager::createNotificationJSON(const NotificationData& data) {
    std::stringstream ss;
    
    // Format timestamp
    auto time_t = std::chrono::system_clock::to_time_t(data.timestamp);
    std::tm* tm = std::localtime(&time_t);
    char timestamp_str[100];
    std::strftime(timestamp_str, sizeof(timestamp_str), "%Y-%m-%d %H:%M:%S", tm);
    
    ss << "{";
    ss << "\"event\":\"new_object_detected\",";
    ss << "\"timestamp\":\"" << timestamp_str << "\",";
    ss << "\"object\":{";
    ss << "\"type\":\"" << data.object_type << "\",";
    ss << "\"x\":" << data.x << ",";
    ss << "\"y\":" << data.y << ",";
    ss << "\"confidence\":" << std::fixed << std::setprecision(2) << data.confidence;
    ss << "},";
    
    // Include all detections
    ss << "\"all_detections\":[";
    for (size_t i = 0; i < data.all_detections.size(); i++) {
        const auto& det = data.all_detections[i];
        ss << "{";
        ss << "\"class\":\"" << det.class_name << "\",";
        ss << "\"confidence\":" << std::fixed << std::setprecision(2) << det.confidence << ",";
        ss << "\"bbox\":{";
        ss << "\"x\":" << det.bbox.x << ",";
        ss << "\"y\":" << det.bbox.y << ",";
        ss << "\"width\":" << det.bbox.width << ",";
        ss << "\"height\":" << det.bbox.height;
        ss << "}";
        ss << "}";
        if (i < data.all_detections.size() - 1) ss << ",";
    }
    ss << "],";
    
    // Status information
    ss << "\"status\":{";
    ss << "\"fps\":" << std::fixed << std::setprecision(2) << data.current_fps << ",";
    ss << "\"avg_processing_time_ms\":" << std::fixed << std::setprecision(2) << data.avg_processing_time_ms << ",";
    ss << "\"total_objects_detected\":" << data.total_objects_detected << ",";
    ss << "\"total_images_saved\":" << data.total_images_saved << ",";
    ss << "\"brightness_filter_active\":" << (data.brightness_filter_active ? "true" : "false") << ",";
    ss << "\"gpu_enabled\":" << (data.gpu_enabled ? "true" : "false") << ",";
    ss << "\"burst_mode_enabled\":" << (data.burst_mode_enabled ? "true" : "false");
    ss << "},";
    
    // Top detected objects
    ss << "\"top_objects\":[";
    for (size_t i = 0; i < data.top_objects.size(); i++) {
        ss << "{";
        ss << "\"type\":\"" << data.top_objects[i].first << "\",";
        ss << "\"count\":" << data.top_objects[i].second;
        ss << "}";
        if (i < data.top_objects.size() - 1) ss << ",";
    }
    ss << "],";
    
    // Encode image as base64 if available
    if (!data.frame_with_boxes.empty()) {
        std::string base64_image = encodeImageToBase64(data.frame_with_boxes);
        ss << "\"image\":\"" << base64_image << "\"";
    } else {
        ss << "\"image\":null";
    }
    
    ss << "}";
    
    return ss.str();
}

std::string NotificationManager::encodeImageToBase64(const cv::Mat& image) {
    // Encode image to JPEG
    std::vector<uchar> buffer;
    std::vector<int> params = {cv::IMWRITE_JPEG_QUALITY, 80};
    cv::imencode(".jpg", image, buffer, params);
    
    // Base64 encode
    static const char* base64_chars = 
        "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
        "abcdefghijklmnopqrstuvwxyz"
        "0123456789+/";
    
    std::string encoded;
    int i = 0;
    int j = 0;
    unsigned char char_array_3[3];
    unsigned char char_array_4[4];
    
    for (size_t idx = 0; idx < buffer.size(); idx++) {
        char_array_3[i++] = buffer[idx];
        if (i == 3) {
            char_array_4[0] = (char_array_3[0] & 0xfc) >> 2;
            char_array_4[1] = ((char_array_3[0] & 0x03) << 4) + ((char_array_3[1] & 0xf0) >> 4);
            char_array_4[2] = ((char_array_3[1] & 0x0f) << 2) + ((char_array_3[2] & 0xc0) >> 6);
            char_array_4[3] = char_array_3[2] & 0x3f;
            
            for (i = 0; i < 4; i++)
                encoded += base64_chars[char_array_4[i]];
            i = 0;
        }
    }
    
    if (i) {
        for (j = i; j < 3; j++)
            char_array_3[j] = '\0';
        
        char_array_4[0] = (char_array_3[0] & 0xfc) >> 2;
        char_array_4[1] = ((char_array_3[0] & 0x03) << 4) + ((char_array_3[1] & 0xf0) >> 4);
        char_array_4[2] = ((char_array_3[1] & 0x0f) << 2) + ((char_array_3[2] & 0xc0) >> 6);
        
        for (j = 0; j < i + 1; j++)
            encoded += base64_chars[char_array_4[j]];
        
        while (i++ < 3)
            encoded += '=';
    }
    
    return encoded;
}
