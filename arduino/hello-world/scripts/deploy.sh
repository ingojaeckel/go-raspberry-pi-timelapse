#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Deploying to Arduino Board with Arduino CLI${NC}"

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Read configuration from arduino-cli.yaml
CONFIG_FILE="$PROJECT_ROOT/arduino-cli.yaml"

if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}Error: Configuration file not found: $CONFIG_FILE${NC}"
    exit 1
fi

# Extract FQBN from config
FQBN=$(grep "^fqbn:" "$CONFIG_FILE" | cut -d':' -f2- | xargs)

if [ -z "$FQBN" ]; then
    echo -e "${RED}Error: 'fqbn' not found in $CONFIG_FILE${NC}"
    exit 1
fi

echo -e "${YELLOW}Board FQBN: $FQBN${NC}"

# Check for Arduino CLI
if ! command -v arduino-cli &> /dev/null; then
    echo -e "${RED}Error: Arduino CLI not found.${NC}"
    echo -e "${YELLOW}Please run ./scripts/build.sh first or install Arduino CLI manually.${NC}"
    exit 1
fi

# Check if upload port is specified
UPLOAD_PORT="${1:-}"

if [ -n "$UPLOAD_PORT" ]; then
    echo -e "${YELLOW}Using specified upload port: ${UPLOAD_PORT}${NC}"
    arduino-cli upload -p "$UPLOAD_PORT" --fqbn "$FQBN" hello-world
else
    echo -e "${YELLOW}Auto-detecting upload port...${NC}"
    # List available boards
    echo -e "${YELLOW}Available boards:${NC}"
    arduino-cli board list
    
    # Upload with auto-detection
    arduino-cli upload --fqbn "$FQBN" hello-world
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
    echo -e "${YELLOW}1. Ensure the Arduino board is connected via USB${NC}"
    echo -e "${YELLOW}2. Check that you have proper USB permissions${NC}"
    echo -e "${YELLOW}3. List boards with: arduino-cli board list${NC}"
    echo -e "${YELLOW}4. Try specifying the port manually: ./scripts/deploy.sh /dev/ttyACM0${NC}"
    exit 1
fi
