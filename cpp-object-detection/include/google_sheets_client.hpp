#pragma once

#include <string>
#include <vector>
#include <memory>
#include <mutex>
#include "logger.hpp"

/**
 * Google Sheets API client for logging detection events.
 * Uses Google Sheets API v4 to append rows to a spreadsheet.
 * Portable across 64-bit Intel macOS and Linux.
 */
class GoogleSheetsClient {
public:
    struct Config {
        std::string spreadsheet_id;     // Google Sheets spreadsheet ID
        std::string api_key;            // Google API key for authentication
        std::string sheet_name;         // Name of the sheet/tab within spreadsheet (default: "Sheet1")
        bool enabled = false;           // Enable/disable Google Sheets integration
    };

    GoogleSheetsClient(const Config& config, std::shared_ptr<Logger> logger);
    ~GoogleSheetsClient();

    /**
     * Initialize the client (validate configuration, test connectivity)
     */
    bool initialize();

    /**
     * Log a detection event to Google Sheets
     * @param timestamp ISO 8601 timestamp string
     * @param object_type Type of object detected (e.g., "person", "cat")
     * @param event_type Type of event ("entry", "movement", "exit")
     * @param x X coordinate of object
     * @param y Y coordinate of object
     * @param distance Movement distance (for movement events)
     * @param description Additional description/details
     */
    bool logDetection(const std::string& timestamp,
                     const std::string& object_type,
                     const std::string& event_type,
                     float x,
                     float y,
                     float distance = 0.0f,
                     const std::string& description = "");

    /**
     * Check if client is enabled and properly initialized
     */
    bool isEnabled() const { return config_.enabled && initialized_; }

private:
    Config config_;
    std::shared_ptr<Logger> logger_;
    bool initialized_;
    std::mutex mutex_;  // Thread-safe access to API calls

    /**
     * Append a row to the Google Sheet
     */
    bool appendRow(const std::vector<std::string>& values);

    /**
     * Make HTTP request to Google Sheets API
     */
    bool makeApiRequest(const std::string& endpoint, const std::string& json_data, std::string& response);

    /**
     * Extract spreadsheet ID from URL if full URL is provided
     */
    static std::string extractSpreadsheetId(const std::string& url_or_id);
};
