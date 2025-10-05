# Google Sheets Integration for Object Detection

## Overview

The C++ Object Detection Application now includes optional Google Sheets integration for logging detection events. When enabled, the application will automatically append rows to a Google Sheet whenever objects are detected, providing a cloud-based log of all detection events.

## Features

- **Off by Default**: Google Sheets integration is disabled by default and requires explicit configuration
- **Entry and Movement Logging**: Logs both object entry events (new objects entering the frame) and movement events (tracked objects moving)
- **Rich Event Data**: Each row includes:
  - Timestamp (ISO 8601 format)
  - Object type (e.g., "person", "cat", "car")
  - Event type ("entry" or "movement")
  - X, Y coordinates
  - Movement distance (for movement events)
  - Additional description with confidence scores and movement details
- **Cross-Platform**: Works on 64-bit Intel macOS and Linux
- **URL or ID Support**: Accepts either a full Google Sheets URL or just the spreadsheet ID
- **Configurable Sheet**: Specify which sheet/tab to use (defaults to "Sheet1")

## Prerequisites

1. **Google Sheets Setup**:
   - Create a Google Sheet where you want to log detection events
   - Note the Spreadsheet ID from the URL (see below)

2. **Google API Key**:
   - Create a Google Cloud project
   - Enable the Google Sheets API
   - Create an API key with Sheets API access
   - Make sure the sheet is shared with "Anyone with the link" (read/write permissions) OR use OAuth2 (requires code modifications)

### Getting Your Spreadsheet ID

From a Google Sheets URL like:
```
https://docs.google.com/spreadsheets/d/1ABC123DEF456GHI789/edit#gid=0
```

The Spreadsheet ID is: `1ABC123DEF456GHI789`

You can provide either:
- The full URL (the application will extract the ID automatically)
- Just the ID

### Getting a Google API Key

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google Sheets API for your project
4. Go to "Credentials" and create an API key
5. Restrict the API key to only allow Google Sheets API access (recommended for security)

**Important**: With an API key, the spreadsheet must be publicly accessible. For private sheets, you would need to implement OAuth2 authentication.

## Configuration

### Command-Line Options

```bash
--enable-google-sheets         Enable Google Sheets integration (default: disabled)
--google-sheets-id ID          Google Sheets spreadsheet ID or full URL (required)
--google-sheets-api-key KEY    Google API key for Sheets API access (required)
--google-sheets-name NAME      Sheet name/tab within spreadsheet (default: Sheet1)
```

## Usage Examples

### Basic Usage

```bash
# Enable Google Sheets logging
./object_detection \
  --enable-google-sheets \
  --google-sheets-id "1ABC123DEF456GHI789" \
  --google-sheets-api-key "AIzaSyABC123DEF456GHI789"
```

### With Full URL

```bash
# Using the full Google Sheets URL
./object_detection \
  --enable-google-sheets \
  --google-sheets-id "https://docs.google.com/spreadsheets/d/1ABC123DEF456GHI789/edit#gid=0" \
  --google-sheets-api-key "AIzaSyABC123DEF456GHI789"
```

### Custom Sheet Name

```bash
# Log to a specific sheet/tab called "Detections"
./object_detection \
  --enable-google-sheets \
  --google-sheets-id "1ABC123DEF456GHI789" \
  --google-sheets-api-key "AIzaSyABC123DEF456GHI789" \
  --google-sheets-name "Detections"
```

### Combined with Other Features

```bash
# Google Sheets + Network Streaming + Verbose logging
./object_detection \
  --enable-google-sheets \
  --google-sheets-id "1ABC123DEF456GHI789" \
  --google-sheets-api-key "AIzaSyABC123DEF456GHI789" \
  --enable-streaming \
  --verbose
```

## Sheet Format

The application will append rows to your Google Sheet with the following columns:

| Timestamp | Object Type | Event Type | X | Y | Distance | Description |
|-----------|-------------|------------|---|---|----------|-------------|
| 2024-10-05T14:30:15.123 | person | entry | 320.5 | 240.8 | | Confidence: 87% |
| 2024-10-05T14:30:16.456 | person | movement | 325.2 | 245.3 | 6.8 | From (320,240) to (325,245) [avg step: 6.8 px, overall path: 6.8 px] |
| 2024-10-05T14:30:20.789 | cat | entry | 150.0 | 180.5 | | Confidence: 92% |

### Column Descriptions

- **Timestamp**: ISO 8601 timestamp with millisecond precision
- **Object Type**: Type of detected object (person, cat, dog, car, etc.)
- **Event Type**: Either "entry" (new object) or "movement" (tracked object moved)
- **X**: X coordinate of object center
- **Y**: Y coordinate of object center
- **Distance**: Movement distance in pixels (only for movement events)
- **Description**: Additional context:
  - Entry events: Confidence score
  - Movement events: From/to coordinates and movement statistics

## Setting Up Your Sheet

It's recommended to create a header row in your Google Sheet before starting the application:

```
Timestamp | Object Type | Event Type | X | Y | Distance | Description
```

The application will append rows below this header.

## Security Considerations

⚠️ **Important Security Notes**:

1. **API Key Security**: 
   - Never commit your API key to version control
   - Consider using environment variables to store the key
   - Restrict the API key to only Google Sheets API access

2. **Sheet Privacy**:
   - When using an API key, the sheet must be publicly accessible
   - Anyone with the sheet link can view the detection data
   - For private sheets, implement OAuth2 authentication (code modifications required)

3. **Data Sensitivity**:
   - Detection logs may reveal patterns about your location/property
   - Review Google Sheets privacy settings carefully
   - Consider using a dedicated Google account for this purpose

## Using Environment Variables (Recommended)

Instead of passing the API key on the command line (which may be visible in process lists), consider using environment variables:

```bash
# Set environment variables
export GOOGLE_SHEETS_ID="1ABC123DEF456GHI789"
export GOOGLE_SHEETS_API_KEY="AIzaSyABC123DEF456GHI789"

# Then modify the application to read from environment variables
# (requires code changes in config_manager.cpp)
```

## Troubleshooting

### "Failed to initialize Google Sheets client"

Check that you've provided both:
- `--google-sheets-id` or full URL
- `--google-sheets-api-key`

### "Google Sheets API request failed with HTTP 403"

- Verify that the spreadsheet is shared publicly or with the appropriate permissions
- Check that your API key has Google Sheets API access enabled
- Ensure the API key hasn't been restricted to specific domains/IPs that don't include your system

### "Google Sheets API request failed with HTTP 400"

- Verify the spreadsheet ID is correct
- Check that the sheet name exists in the spreadsheet
- Ensure the sheet isn't protected or locked

### No rows appearing in the sheet

- Check verbose logs with `--verbose` to see if detections are happening
- Verify the camera is working and detecting objects
- Ensure the confidence threshold isn't too high (`--min-confidence`)
- Check that network connectivity is working

## Performance Considerations

- Each detection event requires an HTTP API call to Google Sheets
- In high-traffic scenarios, this could impact performance
- Consider the rate of detections when tuning `--max-fps` and `--min-confidence`
- Google Sheets API has rate limits (100 requests per 100 seconds per user)

## Implementation Details

The Google Sheets integration uses:
- **libcurl**: For HTTP requests (portable across macOS and Linux)
- **Google Sheets API v4**: RESTful API for appending rows
- **JSON**: For request/response formatting
- **Thread-safe**: Uses mutex locks for concurrent access

The integration is implemented in:
- `include/google_sheets_client.hpp`: Client interface
- `src/google_sheets_client.cpp`: Client implementation
- `src/object_detector.cpp`: Integration into detection event logging

## Testing

The implementation includes comprehensive tests:

```bash
# Run Google Sheets client tests
cd build
./tests/object_detection_tests --gtest_filter="GoogleSheetsClientTest.*"

# Run config manager tests (includes Google Sheets config)
./tests/object_detection_tests --gtest_filter="ConfigManagerTest.*Google*"
```

## Future Enhancements

Potential improvements for the future:
- OAuth2 authentication for private sheets
- Batch API requests to improve performance
- Configurable column mapping
- Support for Google Drive folder organization
- Automatic sheet creation
- Rate limiting and retry logic
