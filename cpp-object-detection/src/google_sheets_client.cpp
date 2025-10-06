#include "google_sheets_client.hpp"
#include <curl/curl.h>
#include <sstream>
#include <iomanip>
#include <regex>

// Callback for libcurl to write response data
static size_t WriteCallback(void* contents, size_t size, size_t nmemb, void* userp) {
    ((std::string*)userp)->append((char*)contents, size * nmemb);
    return size * nmemb;
}

GoogleSheetsClient::GoogleSheetsClient(const Config& config, std::shared_ptr<Logger> logger)
    : config_(config), logger_(logger), initialized_(false) {
    // Extract spreadsheet ID from URL if a full URL was provided
    config_.spreadsheet_id = extractSpreadsheetId(config_.spreadsheet_id);
}

GoogleSheetsClient::~GoogleSheetsClient() {
}

bool GoogleSheetsClient::initialize() {
    if (!config_.enabled) {
        logger_->debug("Google Sheets integration is disabled");
        return true;  // Not an error, just disabled
    }

    if (config_.spreadsheet_id.empty()) {
        logger_->error("Google Sheets: spreadsheet_id is required when enabled");
        return false;
    }

    if (config_.api_key.empty()) {
        logger_->error("Google Sheets: api_key is required when enabled");
        return false;
    }

    if (config_.sheet_name.empty()) {
        config_.sheet_name = "Sheet1";  // Default sheet name
    }

    // Initialize curl globally (should be done once)
    curl_global_init(CURL_GLOBAL_DEFAULT);

    logger_->info("Google Sheets integration initialized successfully");
    logger_->info("  Spreadsheet ID: " + config_.spreadsheet_id);
    logger_->info("  Sheet name: " + config_.sheet_name);

    initialized_ = true;
    return true;
}

bool GoogleSheetsClient::logDetection(const std::string& timestamp,
                                     const std::string& object_type,
                                     const std::string& event_type,
                                     float x,
                                     float y,
                                     float distance,
                                     const std::string& description) {
    if (!isEnabled()) {
        return false;
    }

    std::lock_guard<std::mutex> lock(mutex_);

    // Prepare row data
    std::vector<std::string> values;
    values.push_back(timestamp);
    values.push_back(object_type);
    values.push_back(event_type);
    values.push_back(std::to_string(x));
    values.push_back(std::to_string(y));
    
    if (event_type == "movement") {
        values.push_back(std::to_string(distance));
    } else {
        values.push_back("");
    }
    
    values.push_back(description);

    logger_->debug("Logging to Google Sheets: " + object_type + " " + event_type + 
                  " at (" + std::to_string(x) + ", " + std::to_string(y) + ")");

    return appendRow(values);
}

bool GoogleSheetsClient::appendRow(const std::vector<std::string>& values) {
    // Build JSON for the request
    std::ostringstream json;
    json << "{\"values\": [[";
    for (size_t i = 0; i < values.size(); ++i) {
        if (i > 0) json << ", ";
        // Escape quotes in values
        std::string escaped_value = values[i];
        size_t pos = 0;
        while ((pos = escaped_value.find('"', pos)) != std::string::npos) {
            escaped_value.replace(pos, 1, "\\\"");
            pos += 2;
        }
        json << "\"" << escaped_value << "\"";
    }
    json << "]]}";

    // Build API endpoint
    std::string endpoint = "https://sheets.googleapis.com/v4/spreadsheets/" + 
                          config_.spreadsheet_id + 
                          "/values/" + config_.sheet_name + 
                          ":append?valueInputOption=RAW&key=" + config_.api_key;

    std::string response;
    bool success = makeApiRequest(endpoint, json.str(), response);

    if (!success) {
        logger_->error("Failed to append row to Google Sheets");
        logger_->debug("Response: " + response);
    }

    return success;
}

bool GoogleSheetsClient::makeApiRequest(const std::string& endpoint, 
                                       const std::string& json_data, 
                                       std::string& response) {
    CURL* curl = curl_easy_init();
    if (!curl) {
        logger_->error("Failed to initialize CURL");
        return false;
    }

    struct curl_slist* headers = nullptr;
    headers = curl_slist_append(headers, "Content-Type: application/json");

    curl_easy_setopt(curl, CURLOPT_URL, endpoint.c_str());
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json_data.c_str());
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, WriteCallback);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &response);
    curl_easy_setopt(curl, CURLOPT_TIMEOUT, 10L);  // 10 second timeout
    curl_easy_setopt(curl, CURLOPT_FOLLOWLOCATION, 1L);

    CURLcode res = curl_easy_perform(curl);
    long http_code = 0;
    curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &http_code);

    curl_slist_free_all(headers);
    curl_easy_cleanup(curl);

    if (res != CURLE_OK) {
        logger_->error("CURL request failed: " + std::string(curl_easy_strerror(res)));
        return false;
    }

    if (http_code < 200 || http_code >= 300) {
        logger_->error("Google Sheets API request failed with HTTP " + std::to_string(http_code));
        return false;
    }

    return true;
}

std::string GoogleSheetsClient::extractSpreadsheetId(const std::string& url_or_id) {
    // If it looks like a URL, extract the ID from it
    // Format: https://docs.google.com/spreadsheets/d/{SPREADSHEET_ID}/edit...
    std::regex url_pattern(R"(/spreadsheets/d/([a-zA-Z0-9-_]+))");
    std::smatch match;
    
    if (std::regex_search(url_or_id, match, url_pattern)) {
        return match[1].str();
    }
    
    // Otherwise, assume it's already an ID
    return url_or_id;
}
