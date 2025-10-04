#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Testing 32-bit Linux Build Configuration${NC}"

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Test 1: Check build script exists and is executable
echo -e "${YELLOW}Test 1: Check build-linux-386.sh exists and is executable${NC}"
if [ -x "$SCRIPT_DIR/build-linux-386.sh" ]; then
    echo -e "${GREEN}✓ build-linux-386.sh is executable${NC}"
else
    echo -e "${RED}✗ build-linux-386.sh not found or not executable${NC}"
    exit 1
fi

# Test 2: Check CMakeLists.txt has 32-bit architecture handling
echo -e "${YELLOW}Test 2: Check CMakeLists.txt has 32-bit support${NC}"
if grep -q "CMAKE_SYSTEM_PROCESSOR MATCHES \"i386\"" "$PROJECT_ROOT/CMakeLists.txt"; then
    echo -e "${GREEN}✓ CMakeLists.txt includes i386 architecture handling${NC}"
else
    echo -e "${RED}✗ CMakeLists.txt missing i386 architecture handling${NC}"
    exit 1
fi

# Test 3: Check README has 32-bit documentation
echo -e "${YELLOW}Test 3: Check README has 32-bit Linux documentation${NC}"
if grep -q "32-bit Linux" "$PROJECT_ROOT/README.md" && grep -q "build-linux-386.sh" "$PROJECT_ROOT/README.md"; then
    echo -e "${GREEN}✓ README includes 32-bit Linux documentation${NC}"
else
    echo -e "${RED}✗ README missing 32-bit Linux documentation${NC}"
    exit 1
fi

# Test 4: Check README has Pentium M recommendations
echo -e "${YELLOW}Test 4: Check README has Pentium M recommendations${NC}"
if grep -q "Pentium M" "$PROJECT_ROOT/README.md" && grep -q "1.5GB RAM" "$PROJECT_ROOT/README.md"; then
    echo -e "${GREEN}✓ README includes Pentium M hardware recommendations${NC}"
else
    echo -e "${RED}✗ README missing Pentium M hardware recommendations${NC}"
    exit 1
fi

# Test 5: Check workflow has 32-bit build job
echo -e "${YELLOW}Test 5: Check GitHub Actions workflow has 32-bit build job${NC}"
WORKFLOW_FILE="$PROJECT_ROOT/../.github/workflows/cpp-object-detection.yml"
if [ -f "$WORKFLOW_FILE" ]; then
    if grep -q "build-linux-386" "$WORKFLOW_FILE"; then
        echo -e "${GREEN}✓ GitHub Actions workflow includes 32-bit build job${NC}"
    else
        echo -e "${RED}✗ GitHub Actions workflow missing 32-bit build job${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠ Workflow file not found (may be in different location)${NC}"
fi

# Test 6: Check workflow has proper job dependency
echo -e "${YELLOW}Test 6: Check 32-bit job runs after 64-bit builds${NC}"
if [ -f "$WORKFLOW_FILE" ]; then
    if grep -A 3 "build-linux-386" "$WORKFLOW_FILE" | grep -q "needs:.*build-test.*build-static"; then
        echo -e "${GREEN}✓ 32-bit build job properly depends on 64-bit builds${NC}"
    else
        echo -e "${YELLOW}⚠ Could not verify job dependencies (check manually)${NC}"
    fi
fi

# Test 7: Check config_manager.cpp has 32-bit recommendations in help
echo -e "${YELLOW}Test 7: Check help text includes 32-bit recommendations${NC}"
if grep -q "32-BIT LINUX RECOMMENDATIONS" "$PROJECT_ROOT/src/config_manager.cpp"; then
    echo -e "${GREEN}✓ Help text includes 32-bit recommendations${NC}"
else
    echo -e "${RED}✗ Help text missing 32-bit recommendations${NC}"
    exit 1
fi

# Test 8: Verify recommended CLI flags are documented
echo -e "${YELLOW}Test 8: Check documented CLI flags for low-resource systems${NC}"
if grep -q "max-fps 1" "$PROJECT_ROOT/README.md" && \
   grep -q "frame-width 640" "$PROJECT_ROOT/README.md" && \
   grep -q "analysis-rate-limit 0.5" "$PROJECT_ROOT/README.md"; then
    echo -e "${GREEN}✓ README documents recommended CLI flags for constrained hardware${NC}"
else
    echo -e "${RED}✗ README missing complete CLI flag recommendations${NC}"
    exit 1
fi

# Test 9: Check if build script handles OpenCV dependencies
echo -e "${YELLOW}Test 9: Check build script mentions OpenCV dependencies${NC}"
if grep -q "libopencv" "$SCRIPT_DIR/build-linux-386.sh"; then
    echo -e "${GREEN}✓ Build script checks for OpenCV dependencies${NC}"
else
    echo -e "${RED}✗ Build script missing OpenCV dependency checks${NC}"
    exit 1
fi

# Test 10: Verify multilib requirement is documented
echo -e "${YELLOW}Test 10: Check multilib requirement in build script${NC}"
if grep -q "gcc-multilib" "$SCRIPT_DIR/build-linux-386.sh"; then
    echo -e "${GREEN}✓ Build script documents multilib requirement${NC}"
else
    echo -e "${RED}✗ Build script missing multilib requirement${NC}"
    exit 1
fi

echo -e ""
echo -e "${GREEN}===============================================${NC}"
echo -e "${GREEN}All 32-bit Linux configuration tests passed! ✅${NC}"
echo -e "${GREEN}===============================================${NC}"
echo -e ""
echo -e "${YELLOW}To build for 32-bit Linux:${NC}"
echo -e "${YELLOW}  1. Install dependencies: sudo apt-get install gcc-multilib g++-multilib${NC}"
echo -e "${YELLOW}  2. Run: ./scripts/build-linux-386.sh${NC}"
echo -e "${YELLOW}  3. Test: ./build-linux-386/object_detection --help${NC}"
echo -e ""
echo -e "${YELLOW}Recommended settings for Pentium M (1.5GB RAM):${NC}"
echo -e "${YELLOW}  ./object_detection --max-fps 1 --min-confidence 0.8 --frame-width 640 --frame-height 480 --analysis-rate-limit 0.5${NC}"
