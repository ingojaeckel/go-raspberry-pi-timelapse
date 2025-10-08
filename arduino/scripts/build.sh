#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building Arduino Giga R1 WiFi Project${NC}"

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Check for PlatformIO
if ! command -v pio &> /dev/null; then
    echo -e "${YELLOW}PlatformIO CLI not found. Installing...${NC}"
    
    # Install PlatformIO CLI
    if command -v python3 &> /dev/null; then
        python3 -m pip install --user platformio
        
        # Add to PATH if not already there
        export PATH=$PATH:~/.local/bin
        
        if ! command -v pio &> /dev/null; then
            echo -e "${RED}Error: PlatformIO installation failed.${NC}"
            echo -e "${YELLOW}Please install manually: https://platformio.org/install/cli${NC}"
            exit 1
        fi
    else
        echo -e "${RED}Error: Python 3 not found. Please install Python 3.${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}PlatformIO found: $(pio --version)${NC}"

# Install/update platform and dependencies
echo -e "${YELLOW}Installing/updating STM32 platform and dependencies...${NC}"
pio pkg install --platform ststm32

# Build the project
echo -e "${YELLOW}Building project...${NC}"
pio run

# Check build status
if [ $? -eq 0 ]; then
    echo -e ""
    echo -e "${GREEN}===============================================${NC}"
    echo -e "${GREEN}Build successful! ✅${NC}"
    echo -e "${GREEN}===============================================${NC}"
    echo -e ""
    echo -e "${YELLOW}Binary location:${NC}"
    find .pio/build -name "*.bin" -o -name "*.elf" | head -5
    echo -e ""
    echo -e "${YELLOW}To upload to board: ./scripts/deploy.sh${NC}"
else
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi
