#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building Arduino Project with Arduino CLI${NC}"

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Read configuration from arduino-cli.yaml
CONFIG_FILE="$PROJECT_ROOT/arduino-cli.yaml"

if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}Error: Configuration file not found: $CONFIG_FILE${NC}"
    echo -e "${YELLOW}Expected format:${NC}"
    echo -e "${YELLOW}  fqbn: arduino:mbed_giga:giga${NC}"
    exit 1
fi

# Extract FQBN from config
FQBN=$(grep "^fqbn:" "$CONFIG_FILE" | cut -d':' -f2- | xargs)

if [ -z "$FQBN" ]; then
    echo -e "${RED}Error: 'fqbn' not found in $CONFIG_FILE${NC}"
    exit 1
fi

echo -e "${YELLOW}Board FQBN: $FQBN${NC}"

# Extract platform from FQBN (e.g., arduino:mbed_giga from arduino:mbed_giga:giga)
PLATFORM=$(echo "$FQBN" | cut -d':' -f1-2)

# Check for Arduino CLI
if ! command -v arduino-cli &> /dev/null; then
    echo -e "${YELLOW}Arduino CLI not found. Installing...${NC}"
    
    # Detect OS and architecture
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$ARCH" in
        x86_64) ARCH="64bit" ;;
        aarch64|arm64) ARCH="ARM64" ;;
        armv7l) ARCH="ARMv7" ;;
        *) echo -e "${RED}Unsupported architecture: $ARCH${NC}"; exit 1 ;;
    esac
    
    case "$OS" in
        linux) OS="Linux" ;;
        darwin) OS="macOS" ;;
        *) echo -e "${RED}Unsupported OS: $OS${NC}"; exit 1 ;;
    esac
    
    # Download and install Arduino CLI
    ARDUINO_CLI_URL="https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_${OS}_${ARCH}.tar.gz"
    echo -e "${YELLOW}Downloading from: $ARDUINO_CLI_URL${NC}"
    
    mkdir -p ~/.local/bin
    curl -fsSL "$ARDUINO_CLI_URL" | tar -xz -C ~/.local/bin
    
    # Add to PATH if not already there
    export PATH=$PATH:~/.local/bin
    
    if ! command -v arduino-cli &> /dev/null; then
        echo -e "${RED}Error: Arduino CLI installation failed.${NC}"
        echo -e "${YELLOW}Please install manually: https://arduino.github.io/arduino-cli/latest/installation/${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}Arduino CLI found: $(arduino-cli version)${NC}"

# Initialize Arduino CLI config if needed
if [ ! -f ~/.arduino15/arduino-cli.yaml ]; then
    echo -e "${YELLOW}Initializing Arduino CLI configuration...${NC}"
    arduino-cli config init
fi

# Update package index
echo -e "${YELLOW}Updating package index...${NC}"
arduino-cli core update-index

# Install board platform
echo -e "${YELLOW}Installing platform: $PLATFORM${NC}"
arduino-cli core install "$PLATFORM" || echo -e "${YELLOW}Platform may already be installed${NC}"

# Compile the sketch
echo -e "${YELLOW}Compiling sketch...${NC}"
arduino-cli compile --fqbn "$FQBN" src/hello-world.ino

# Check build status
if [ $? -eq 0 ]; then
    echo -e ""
    echo -e "${GREEN}===============================================${NC}"
    echo -e "${GREEN}Build successful! ✅${NC}"
    echo -e "${GREEN}===============================================${NC}"
    echo -e ""
    echo -e "${YELLOW}Binary location:${NC}"
    find src/build -name "*.bin" -o -name "*.elf" 2>/dev/null | head -5
    echo -e ""
    echo -e "${YELLOW}To upload to board: ./scripts/deploy.sh${NC}"
else
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi
