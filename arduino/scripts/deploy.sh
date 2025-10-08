#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Deploying to Arduino Giga R1 WiFi${NC}"

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Check for PlatformIO
if ! command -v pio &> /dev/null; then
    echo -e "${RED}Error: PlatformIO CLI not found.${NC}"
    echo -e "${YELLOW}Please run ./scripts/build.sh first or install PlatformIO manually.${NC}"
    exit 1
fi

# Check if upload port is specified
UPLOAD_PORT="${1:-}"

if [ -n "$UPLOAD_PORT" ]; then
    echo -e "${YELLOW}Using specified upload port: ${UPLOAD_PORT}${NC}"
    pio run --target upload --upload-port "$UPLOAD_PORT"
else
    echo -e "${YELLOW}Auto-detecting upload port...${NC}"
    pio run --target upload
fi

# Check upload status
if [ $? -eq 0 ]; then
    echo -e ""
    echo -e "${GREEN}===============================================${NC}"
    echo -e "${GREEN}Upload successful! ✅${NC}"
    echo -e "${GREEN}===============================================${NC}"
    echo -e ""
    echo -e "${YELLOW}To monitor serial output: ./scripts/monitor.sh${NC}"
else
    echo -e "${RED}Upload failed!${NC}"
    echo -e ""
    echo -e "${YELLOW}Troubleshooting:${NC}"
    echo -e "${YELLOW}1. Ensure the Arduino Giga R1 WiFi is connected via USB${NC}"
    echo -e "${YELLOW}2. Check that you have proper USB permissions${NC}"
    echo -e "${YELLOW}3. Try specifying the port manually: ./scripts/deploy.sh /dev/ttyACM0${NC}"
    exit 1
fi
