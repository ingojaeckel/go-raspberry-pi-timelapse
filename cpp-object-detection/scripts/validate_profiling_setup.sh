#!/bin/bash
# Validation script to verify profiling infrastructure setup
# Run this to check if profiling tools and scripts are properly configured

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Profiling Infrastructure Validation"
echo "===================================="
echo ""

ERRORS=0
WARNINGS=0

# Detect platform
if [[ "$OSTYPE" == "darwin"* ]]; then
    PLATFORM="macOS"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    PLATFORM="Linux"
else
    echo "❌ Unsupported platform: $OSTYPE"
    exit 1
fi

echo "Platform: $PLATFORM"
echo ""

# Check documentation
echo "Checking documentation..."
DOCS=(
    "docs/PROFILING.md"
    "docs/PROFILING_QUICKSTART.md"
    "docs/PROFILING_INFRASTRUCTURE_SUMMARY.md"
    ".vscode/README.md"
)

for doc in "${DOCS[@]}"; do
    if [ -f "$PROJECT_DIR/$doc" ]; then
        echo "  ✓ $doc"
    else
        echo "  ❌ Missing: $doc"
        ((ERRORS++))
    fi
done
echo ""

# Check scripts
echo "Checking profiling scripts..."
SCRIPTS=(
    "scripts/build_profile_mac.sh"
    "scripts/build_profile_linux.sh"
    "scripts/profile_cpu_mac.sh"
    "scripts/profile_cpu_linux.sh"
    "scripts/profile_memory_mac.sh"
    "scripts/profile_memory_linux.sh"
    "scripts/download_flamegraph.sh"
)

for script in "${SCRIPTS[@]}"; do
    if [ -f "$PROJECT_DIR/$script" ]; then
        if [ -x "$PROJECT_DIR/$script" ]; then
            echo "  ✓ $script (executable)"
        else
            echo "  ⚠ $script (not executable - run: chmod +x $script)"
            ((WARNINGS++))
        fi
    else
        echo "  ❌ Missing: $script"
        ((ERRORS++))
    fi
done
echo ""

# Check VSCode configuration
echo "Checking VSCode configuration..."
VSCODE_FILES=(
    ".vscode/launch.json"
    ".vscode/tasks.json"
)

for file in "${VSCODE_FILES[@]}"; do
    if [ -f "$PROJECT_DIR/$file" ]; then
        # Validate JSON
        if python3 -m json.tool "$PROJECT_DIR/$file" > /dev/null 2>&1; then
            echo "  ✓ $file (valid JSON)"
        else
            echo "  ❌ $file (invalid JSON)"
            ((ERRORS++))
        fi
    else
        echo "  ❌ Missing: $file"
        ((ERRORS++))
    fi
done
echo ""

# Check platform-specific tools
echo "Checking profiling tools..."

if [ "$PLATFORM" == "macOS" ]; then
    # macOS tools
    if command -v instruments &> /dev/null; then
        echo "  ✓ Instruments (Apple profiling tool)"
    else
        echo "  ⚠ Instruments not found (install Xcode Command Line Tools)"
        ((WARNINGS++))
    fi
    
    if command -v cmake &> /dev/null; then
        echo "  ✓ cmake"
    else
        echo "  ❌ cmake not found (install with: brew install cmake)"
        ((ERRORS++))
    fi
    
    if brew list opencv &> /dev/null || pkg-config --exists opencv4 2> /dev/null; then
        echo "  ✓ OpenCV"
    else
        echo "  ⚠ OpenCV not found (install with: brew install opencv)"
        ((WARNINGS++))
    fi
    
    if command -v valgrind &> /dev/null; then
        echo "  ✓ Valgrind (optional)"
    else
        echo "  ℹ Valgrind not installed (optional: brew install valgrind)"
    fi
    
elif [ "$PLATFORM" == "Linux" ]; then
    # Linux tools
    if command -v perf &> /dev/null; then
        echo "  ✓ perf (Linux profiling tool)"
    else
        echo "  ⚠ perf not found (install with: sudo apt-get install linux-tools-generic)"
        ((WARNINGS++))
    fi
    
    if command -v valgrind &> /dev/null; then
        echo "  ✓ Valgrind"
    else
        echo "  ⚠ Valgrind not found (install with: sudo apt-get install valgrind)"
        ((WARNINGS++))
    fi
    
    if command -v heaptrack &> /dev/null; then
        echo "  ✓ heaptrack (heap profiler)"
    else
        echo "  ℹ heaptrack not installed (optional: sudo apt-get install heaptrack)"
    fi
    
    if command -v cmake &> /dev/null; then
        echo "  ✓ cmake"
    else
        echo "  ❌ cmake not found (install with: sudo apt-get install cmake)"
        ((ERRORS++))
    fi
    
    if pkg-config --exists opencv4 2> /dev/null || pkg-config --exists opencv 2> /dev/null; then
        echo "  ✓ OpenCV"
    else
        echo "  ⚠ OpenCV not found (install with: sudo apt-get install libopencv-dev)"
        ((WARNINGS++))
    fi
fi
echo ""

# Check build compiler
echo "Checking build tools..."
if command -v g++ &> /dev/null; then
    echo "  ✓ g++ ($(g++ --version | head -1))"
elif command -v clang++ &> /dev/null; then
    echo "  ✓ clang++ ($(clang++ --version | head -1))"
else
    echo "  ❌ No C++ compiler found"
    ((ERRORS++))
fi

if command -v make &> /dev/null; then
    echo "  ✓ make ($(make --version | head -1))"
else
    echo "  ❌ make not found"
    ((ERRORS++))
fi
echo ""

# Check .gitignore
echo "Checking .gitignore..."
if [ -f "$PROJECT_DIR/.gitignore" ]; then
    if grep -q "profiling_results" "$PROJECT_DIR/.gitignore"; then
        echo "  ✓ .gitignore includes profiling_results"
    else
        echo "  ⚠ .gitignore missing profiling_results entry"
        ((WARNINGS++))
    fi
else
    echo "  ⚠ .gitignore not found"
    ((WARNINGS++))
fi
echo ""

# Summary
echo "======================================"
echo "Validation Summary"
echo "======================================"
echo ""

if [ $ERRORS -eq 0 ]; then
    if [ $WARNINGS -eq 0 ]; then
        echo "✅ All checks passed!"
    else
        echo "✅ Core infrastructure is set up correctly"
        echo "⚠  $WARNINGS warning(s) found (optional dependencies)"
    fi
    echo ""
    echo "Profiling infrastructure is ready to use."
    echo ""
    echo "Next steps:"
    if [ "$PLATFORM" == "macOS" ]; then
        echo "  1. Build for profiling: ./scripts/build_profile_mac.sh"
        echo "  2. Run CPU profiling: ./scripts/profile_cpu_mac.sh"
        echo "  3. View results: open profiling_results/*.trace"
    else
        echo "  1. Build for profiling: ./scripts/build_profile_linux.sh"
        echo "  2. Run CPU profiling: ./scripts/profile_cpu_linux.sh"
        echo "  3. View results: cat profiling_results/*_summary.txt"
    fi
    echo ""
    echo "See docs/PROFILING.md for complete guide."
    exit 0
else
    echo "❌ $ERRORS error(s) found"
    if [ $WARNINGS -gt 0 ]; then
        echo "⚠  $WARNINGS warning(s) found"
    fi
    echo ""
    echo "Please fix the errors above before using profiling tools."
    echo ""
    echo "For help, see:"
    echo "  - docs/PROFILING.md"
    echo "  - docs/PROFILING_QUICKSTART.md"
    exit 1
fi
