#!/bin/bash

# Manual End-to-End Testing Script for Object Detection Application
# This script provides a framework for manual testing of the object detection system

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}Manual E2E Testing Script for Object Detection Application${NC}"
echo -e "${BLUE}=====================================================${NC}"

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"

# Ensure application is built
if [ ! -f "$BUILD_DIR/object_detection" ]; then
    echo -e "${YELLOW}Application not built. Building first...${NC}"
    "$SCRIPT_DIR/build.sh"
fi

cd "$BUILD_DIR"

echo -e "${YELLOW}Available Test Scenarios:${NC}"
echo "1. Test with mock camera (camera not required)"
echo "2. Test with real webcam (requires USB webcam)"
echo "3. Performance test with different frame rates"
echo "4. Configuration validation test"
echo "5. Logging system test"
echo "6. Help and usage test"
echo "7. Exit"

while true; do
    echo
    read -p "Select test scenario (1-7): " choice
    
    case $choice in
        1)
            echo -e "${BLUE}Testing with mock camera (will fail gracefully)...${NC}"
            echo -e "${YELLOW}Expected: Application should start, fail to initialize camera, and exit cleanly${NC}"
            timeout 10s ./object_detection --camera-id 999 --verbose || true
            echo -e "${GREEN}Mock camera test completed${NC}"
            ;;
        2)
            echo -e "${BLUE}Testing with real webcam...${NC}"
            echo -e "${YELLOW}Expected: Application should detect camera and start processing${NC}"
            echo -e "${YELLOW}Press Ctrl+C to stop after observing some logs${NC}"
            read -p "Press Enter to start (or 'skip' to skip): " confirm
            if [ "$confirm" != "skip" ]; then
                timeout 30s ./object_detection --verbose --max-fps 2 || true
            fi
            echo -e "${GREEN}Real webcam test completed${NC}"
            ;;
        3)
            echo -e "${BLUE}Performance test with different frame rates...${NC}"
            echo -e "${YELLOW}Testing low frame rate (1 fps)...${NC}"
            timeout 15s ./object_detection --max-fps 1 --verbose || true
            echo
            echo -e "${YELLOW}Testing high frame rate (10 fps)...${NC}"
            timeout 15s ./object_detection --max-fps 10 --verbose || true
            echo -e "${GREEN}Performance test completed${NC}"
            ;;
        4)
            echo -e "${BLUE}Configuration validation test...${NC}"
            
            echo -e "${YELLOW}Testing invalid max-fps...${NC}"
            ./object_detection --max-fps 0 || true
            
            echo -e "${YELLOW}Testing invalid confidence...${NC}"
            ./object_detection --min-confidence 1.5 || true
            
            echo -e "${YELLOW}Testing invalid camera ID...${NC}"
            ./object_detection --camera-id -1 || true
            
            echo -e "${GREEN}Configuration validation test completed${NC}"
            ;;
        5)
            echo -e "${BLUE}Logging system test...${NC}"
            LOG_FILE="/tmp/test_object_detection.log"
            rm -f "$LOG_FILE"
            
            echo -e "${YELLOW}Running with custom log file...${NC}"
            timeout 10s ./object_detection --log-file "$LOG_FILE" --verbose || true
            
            if [ -f "$LOG_FILE" ]; then
                echo -e "${GREEN}Log file created successfully${NC}"
                echo -e "${YELLOW}Log file contents:${NC}"
                head -20 "$LOG_FILE"
                rm -f "$LOG_FILE"
            else
                echo -e "${RED}Log file was not created${NC}"
            fi
            ;;
        6)
            echo -e "${BLUE}Help and usage test...${NC}"
            echo -e "${YELLOW}Testing --help flag...${NC}"
            ./object_detection --help
            
            echo -e "${YELLOW}Testing invalid arguments...${NC}"
            ./object_detection --invalid-arg || true
            
            echo -e "${GREEN}Help and usage test completed${NC}"
            ;;
        7)
            echo -e "${GREEN}Exiting manual test suite${NC}"
            break
            ;;
        *)
            echo -e "${RED}Invalid choice. Please select 1-7.${NC}"
            ;;
    esac
done

echo
echo -e "${GREEN}Manual testing completed!${NC}"
echo -e "${BLUE}=====================================================${NC}"
echo -e "${YELLOW}Testing Notes:${NC}"
echo "- For real webcam testing, ensure a USB camera is connected"
echo "- Logitech C920 is recommended but any USB Video Class (UVC) camera should work"
echo "- The application requires a YOLO model file for object detection"
echo "- Download model: wget -O models/yolov5s.onnx https://github.com/ultralytics/yolov5/releases/download/v6.2/yolov5s.onnx"
echo "- Check log files for detailed operation information"
echo "- Performance may vary based on CPU capabilities and camera resolution"