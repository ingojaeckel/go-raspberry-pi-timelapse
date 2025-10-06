#include <gtest/gtest.h>
#include "google_sheets_client.hpp"
#include "logger.hpp"
#include <memory>

class GoogleSheetsClientTest : public ::testing::Test {
protected:
    void SetUp() override {
        logger = std::make_shared<Logger>("/tmp/google_sheets_test.log", false);
    }

    void TearDown() override {
        // Cleanup
        std::remove("/tmp/google_sheets_test.log");
    }

    std::shared_ptr<Logger> logger;
};

TEST_F(GoogleSheetsClientTest, DisabledByDefault) {
    GoogleSheetsClient::Config config;
    config.enabled = false;
    
    GoogleSheetsClient client(config, logger);
    EXPECT_TRUE(client.initialize());  // Should succeed even though disabled
    EXPECT_FALSE(client.isEnabled());
}

TEST_F(GoogleSheetsClientTest, RequiresSpreadsheetId) {
    GoogleSheetsClient::Config config;
    config.enabled = true;
    config.spreadsheet_id = "";
    config.api_key = "test_key";
    
    GoogleSheetsClient client(config, logger);
    EXPECT_FALSE(client.initialize());  // Should fail without spreadsheet ID
}

TEST_F(GoogleSheetsClientTest, RequiresApiKey) {
    GoogleSheetsClient::Config config;
    config.enabled = true;
    config.spreadsheet_id = "test_id";
    config.api_key = "";
    
    GoogleSheetsClient client(config, logger);
    EXPECT_FALSE(client.initialize());  // Should fail without API key
}

TEST_F(GoogleSheetsClientTest, AcceptsFullUrl) {
    GoogleSheetsClient::Config config;
    config.enabled = true;
    config.spreadsheet_id = "https://docs.google.com/spreadsheets/d/1ABC123DEF456/edit#gid=0";
    config.api_key = "test_key";
    config.sheet_name = "Sheet1";
    
    GoogleSheetsClient client(config, logger);
    EXPECT_TRUE(client.initialize());  // Should extract ID from URL
    EXPECT_TRUE(client.isEnabled());
}

TEST_F(GoogleSheetsClientTest, AcceptsPlainId) {
    GoogleSheetsClient::Config config;
    config.enabled = true;
    config.spreadsheet_id = "1ABC123DEF456";
    config.api_key = "test_key";
    config.sheet_name = "Sheet1";
    
    GoogleSheetsClient client(config, logger);
    EXPECT_TRUE(client.initialize());
    EXPECT_TRUE(client.isEnabled());
}

TEST_F(GoogleSheetsClientTest, UsesDefaultSheetName) {
    GoogleSheetsClient::Config config;
    config.enabled = true;
    config.spreadsheet_id = "test_id";
    config.api_key = "test_key";
    config.sheet_name = "";  // Empty sheet name
    
    GoogleSheetsClient client(config, logger);
    EXPECT_TRUE(client.initialize());  // Should use default "Sheet1"
}

// Note: We cannot test actual API calls without valid credentials,
// but we've tested the configuration validation logic
