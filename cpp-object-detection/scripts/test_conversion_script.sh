#!/bin/bash
# Test script for YOLO model conversion
# This script tests the conversion script's interface and error handling without downloading large models

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Testing YOLO Model Conversion Script${NC}"
echo ""

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
CONVERT_SCRIPT="$SCRIPT_DIR/convert_yolo_to_onnx.py"

# Test 1: Script exists and is executable
echo -e "${YELLOW}Test 1: Script exists${NC}"
if [ -f "$CONVERT_SCRIPT" ]; then
    echo -e "${GREEN}✓ Script found: $CONVERT_SCRIPT${NC}"
else
    echo -e "${RED}✗ Script not found${NC}"
    exit 1
fi

# Test 2: Help output works
echo -e "\n${YELLOW}Test 2: Help output${NC}"
if python3 "$CONVERT_SCRIPT" --help > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Help output works${NC}"
else
    echo -e "${RED}✗ Help output failed${NC}"
    exit 1
fi

# Test 3: Missing required argument (model-size)
echo -e "\n${YELLOW}Test 3: Argument validation${NC}"
if python3 "$CONVERT_SCRIPT" --version 11 2>&1 | grep -q "model-size is required"; then
    echo -e "${GREEN}✓ Correctly requires --model-size argument${NC}"
else
    echo -e "${RED}✗ Argument validation not working${NC}"
    exit 1
fi

# Test 4: Invalid YOLO version
echo -e "\n${YELLOW}Test 4: Invalid version rejection${NC}"
if ! python3 "$CONVERT_SCRIPT" --version 99 --model-size n 2>&1 | grep -q "invalid choice"; then
    echo -e "${RED}✗ Should reject invalid version${NC}"
    exit 1
else
    echo -e "${GREEN}✓ Correctly rejects invalid version${NC}"
fi

# Test 5: Invalid model size
echo -e "\n${YELLOW}Test 5: Invalid model size rejection${NC}"
if ! python3 "$CONVERT_SCRIPT" --version 11 --model-size invalid 2>&1 | grep -q "invalid choice"; then
    echo -e "${RED}✗ Should reject invalid model size${NC}"
    exit 1
else
    echo -e "${GREEN}✓ Correctly rejects invalid model size${NC}"
fi

# Test 6: Missing dependencies error message
echo -e "\n${YELLOW}Test 6: Dependency check${NC}"
if python3 "$CONVERT_SCRIPT" --version 11 --model-size n 2>&1 | grep -q "pip install ultralytics"; then
    echo -e "${GREEN}✓ Provides helpful error for missing dependencies${NC}"
elif python3 -c "import ultralytics" 2>/dev/null; then
    echo -e "${GREEN}✓ Dependencies already installed (would work)${NC}"
else
    echo -e "${RED}✗ Dependency check not working properly${NC}"
    exit 1
fi

# Test 7: .pt file validation
echo -e "\n${YELLOW}Test 7: .pt file validation${NC}"
# This test only works if ultralytics is installed, otherwise we get import error first
if python3 -c "import ultralytics" 2>/dev/null; then
    if python3 "$CONVERT_SCRIPT" --pt-file nonexistent.pt 2>&1 | grep -q "not found"; then
        echo -e "${GREEN}✓ Correctly validates .pt file existence${NC}"
    else
        echo -e "${RED}✗ .pt file validation not working${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}⊙ Skipped (requires ultralytics to be installed)${NC}"
fi

# Test 8: Valid argument combinations
echo -e "\n${YELLOW}Test 8: Valid argument parsing${NC}"
VALID_COMBOS=(
    "--version 9 --model-size n"
    "--version 10 --model-size s"
    "--version 11 --model-size m"
)

for combo in "${VALID_COMBOS[@]}"; do
    if python3 "$CONVERT_SCRIPT" $combo --help > /dev/null 2>&1 || \
       python3 "$CONVERT_SCRIPT" $combo 2>&1 | grep -q "pip install\|Loading.*model"; then
        echo -e "${GREEN}✓ Valid combo accepted: $combo${NC}"
    else
        echo -e "${RED}✗ Valid combo rejected: $combo${NC}"
        exit 1
    fi
done

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✓ All tests passed!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${YELLOW}Note: Actual conversion requires 'pip install ultralytics onnx'${NC}"
echo -e "${YELLOW}To test full conversion:${NC}"
echo -e "${YELLOW}  python3 $CONVERT_SCRIPT --version 11 --model-size n${NC}"
echo ""
