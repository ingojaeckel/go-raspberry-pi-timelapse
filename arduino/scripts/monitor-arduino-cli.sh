#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Serial Monitor for Arduino Giga R1 WiFi with Arduino CLI${NC}"

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Check for Arduino CLI
if ! command -v arduino-cli &> /dev/null; then
    echo -e "${RED}Error: Arduino CLI not found.${NC}"
    echo -e "${YELLOW}Please run ./scripts/build-arduino-cli.sh first or install Arduino CLI manually.${NC}"
    exit 1
fi

# Check if monitor port is specified
MONITOR_PORT="${1:-}"

if [ -n "$MONITOR_PORT" ]; then
    echo -e "${YELLOW}Using specified monitor port: ${MONITOR_PORT}${NC}"
    arduino-cli monitor -p "$MONITOR_PORT" --config baudrate=115200
else
    echo -e "${YELLOW}Auto-detecting monitor port...${NC}"
    # List available boards
    echo -e "${YELLOW}Available boards:${NC}"
    arduino-cli board list
    
    # Get first available port
    PORT=$(arduino-cli board list | grep -v "Unknown" | awk 'NR==2 {print $1}')
    
    if [ -n "$PORT" ]; then
        echo -e "${YELLOW}Using port: $PORT${NC}"
        arduino-cli monitor -p "$PORT" --config baudrate=115200
    else
        echo -e "${RED}No board detected!${NC}"
        echo -e "${YELLOW}Please specify port manually: ./scripts/monitor-arduino-cli.sh /dev/ttyACM0${NC}"
        exit 1
    fi
fi
