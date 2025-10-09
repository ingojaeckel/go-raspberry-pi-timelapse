#!/bin/bash
# Memory profiling script for macOS using Instruments
# Profiles memory allocations, leaks, and heap usage

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="$PROJECT_DIR/build-profile"
RESULTS_DIR="$PROJECT_DIR/profiling_results"
APP_BINARY="$BUILD_DIR/object_detection"

# Default parameters
DURATION=60
OUTPUT_NAME="memory_profile_$(date +%Y%m%d_%H%M%S)"
APP_ARGS="--max-fps 5 --verbose"
PROFILE_TYPE="allocations"  # allocations or leaks

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --duration)
            DURATION="$2"
            shift 2
            ;;
        --output)
            OUTPUT_NAME="$2"
            shift 2
            ;;
        --app-args)
            APP_ARGS="$2"
            shift 2
            ;;
        --type)
            PROFILE_TYPE="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --duration SECONDS    Profiling duration in seconds (default: 60)"
            echo "  --output NAME         Output file base name (default: memory_profile_TIMESTAMP)"
            echo "  --app-args ARGS       Arguments to pass to application (default: '--max-fps 5 --verbose')"
            echo "  --type TYPE           Profile type: 'allocations' or 'leaks' (default: allocations)"
            echo "  --help                Show this help message"
            echo ""
            echo "Example:"
            echo "  $0 --duration 120 --type leaks --output leak_check"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

echo "Memory Profiling for macOS (Instruments)"
echo "========================================="
echo ""

# Check if we're on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "Error: This script is for macOS only. Use profile_memory_linux.sh for Linux."
    exit 1
fi

# Check if Instruments is available
if ! command -v instruments &> /dev/null; then
    echo "Error: Instruments not found. Please install Xcode Command Line Tools:"
    echo "  xcode-select --install"
    exit 1
fi

# Validate profile type
if [[ "$PROFILE_TYPE" != "allocations" && "$PROFILE_TYPE" != "leaks" ]]; then
    echo "Error: Invalid profile type '$PROFILE_TYPE'. Must be 'allocations' or 'leaks'"
    exit 1
fi

# Check if application is built
if [ ! -f "$APP_BINARY" ]; then
    echo "Error: Application not found at $APP_BINARY"
    echo "Please build first with: ./scripts/build_profile_mac.sh"
    exit 1
fi

# Create results directory
mkdir -p "$RESULTS_DIR"

# Output files
if [ "$PROFILE_TYPE" == "allocations" ]; then
    TEMPLATE="Allocations"
    TRACE_FILE="$RESULTS_DIR/${OUTPUT_NAME}_allocations.trace"
else
    TEMPLATE="Leaks"
    TRACE_FILE="$RESULTS_DIR/${OUTPUT_NAME}_leaks.trace"
fi

SUMMARY_FILE="$RESULTS_DIR/${OUTPUT_NAME}_summary.txt"

echo "Configuration:"
echo "  Application: $APP_BINARY"
echo "  Arguments: $APP_ARGS"
echo "  Profile type: $PROFILE_TYPE"
echo "  Duration: ${DURATION}s"
echo "  Results directory: $RESULTS_DIR"
echo "  Trace file: $TRACE_FILE"
echo ""

# Start the application in background
echo "Starting application..."
$APP_BINARY $APP_ARGS > "$RESULTS_DIR/${OUTPUT_NAME}_app.log" 2>&1 &
APP_PID=$!

# Give app time to initialize
sleep 2

# Check if app is still running
if ! ps -p $APP_PID > /dev/null; then
    echo "Error: Application failed to start. Check $RESULTS_DIR/${OUTPUT_NAME}_app.log"
    exit 1
fi

echo "Application started (PID: $APP_PID)"
echo "Profiling memory with $TEMPLATE template..."
echo "(This will run for ${DURATION} seconds)"
echo ""

# Run Instruments with specified template
instruments -t "$TEMPLATE" \
    -D "$TRACE_FILE" \
    -p $APP_PID \
    -l $((DURATION * 1000)) \
    2>&1 | tee "$RESULTS_DIR/${OUTPUT_NAME}_instruments.log" || true

# Wait a moment then kill the app if still running
sleep 2
if ps -p $APP_PID > /dev/null; then
    echo "Stopping application..."
    kill $APP_PID 2>/dev/null || true
    sleep 1
    kill -9 $APP_PID 2>/dev/null || true
fi

echo ""
echo "✅ Profiling complete!"
echo ""

# Create summary
echo "Creating summary report..."
cat > "$SUMMARY_FILE" << EOF
Memory Profiling Summary
========================

Date: $(date)
Application: $APP_BINARY
Arguments: $APP_ARGS
Profile type: $PROFILE_TYPE
Duration: ${DURATION}s
PID: $APP_PID

Files Generated:
- Trace file: $TRACE_FILE
- App log: $RESULTS_DIR/${OUTPUT_NAME}_app.log
- Instruments log: $RESULTS_DIR/${OUTPUT_NAME}_instruments.log

EOF

# Add specific information based on profile type
if [ "$PROFILE_TYPE" == "allocations" ]; then
    cat >> "$SUMMARY_FILE" << EOF
Allocations Profile:
- Shows all heap memory allocations
- Tracks allocation growth over time
- Identifies largest allocations
- Useful for finding memory bloat

Analysis tips:
1. Look for constantly growing allocations
2. Check for unexpectedly large buffers
3. Review allocation call stacks
4. Compare peak memory to expected usage

EOF
else
    cat >> "$SUMMARY_FILE" << EOF
Leaks Profile:
- Detects memory leaks
- Shows leak call stacks
- Categorizes leak types
- Useful for finding resource leaks

Analysis tips:
1. Focus on "Leaked" memory category
2. Review call stacks for leaked allocations
3. Check for circular references
4. Verify all resources are properly freed

EOF
fi

# Extract performance metrics from app logs if available
if [ -f "$RESULTS_DIR/${OUTPUT_NAME}_app.log" ]; then
    echo "" >> "$SUMMARY_FILE"
    echo "Application Performance Metrics:" >> "$SUMMARY_FILE"
    echo "================================" >> "$SUMMARY_FILE"
    grep -i "fps\|performance\|memory\|heap" "$RESULTS_DIR/${OUTPUT_NAME}_app.log" | tail -20 >> "$SUMMARY_FILE" 2>/dev/null || echo "No relevant metrics found in logs" >> "$SUMMARY_FILE"
fi

echo ""
echo "Results saved to:"
echo "  - Trace file (GUI): $TRACE_FILE"
echo "  - Summary: $SUMMARY_FILE"
echo "  - App output: $RESULTS_DIR/${OUTPUT_NAME}_app.log"
echo ""
echo "Next steps:"
echo "  1. Open trace in Instruments: open \"$TRACE_FILE\""
echo "  2. Review summary: cat \"$SUMMARY_FILE\""
if [ "$PROFILE_TYPE" == "allocations" ]; then
    echo "  3. In Instruments, check 'Allocations Summary' for growing allocations"
    echo "  4. Use 'Mark Generation' to track memory growth between snapshots"
else
    echo "  3. In Instruments, review 'Leaks' section for leaked memory"
    echo "  4. Check call stacks to identify leak sources"
fi
echo "  5. See docs/PROFILING.md for detailed analysis guide"
echo ""
