#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Serial Monitor for Arduino Giga R1 WiFi${NC}"

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

# Check if monitor port is specified
MONITOR_PORT="${1:-}"

if [ -n "$MONITOR_PORT" ]; then
    echo -e "${YELLOW}Using specified monitor port: ${MONITOR_PORT}${NC}"
    pio device monitor --port "$MONITOR_PORT" --baud 115200
else
    echo -e "${YELLOW}Auto-detecting monitor port...${NC}"
    pio device monitor --baud 115200
fi
