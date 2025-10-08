#!/bin/bash
# CPU profiling script for macOS using Instruments
# Profiles the object detection application and generates performance reports

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="$PROJECT_DIR/build-profile"
RESULTS_DIR="$PROJECT_DIR/profiling_results"
APP_BINARY="$BUILD_DIR/object_detection"

# Default parameters
DURATION=60
OUTPUT_NAME="cpu_profile_$(date +%Y%m%d_%H%M%S)"
APP_ARGS="--max-fps 10 --verbose"

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
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --duration SECONDS    Profiling duration in seconds (default: 60)"
            echo "  --output NAME         Output file base name (default: cpu_profile_TIMESTAMP)"
            echo "  --app-args ARGS       Arguments to pass to application (default: '--max-fps 10 --verbose')"
            echo "  --help                Show this help message"
            echo ""
            echo "Example:"
            echo "  $0 --duration 120 --output burst_mode_test --app-args '--max-fps 15 --enable-burst-mode'"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

echo "CPU Profiling for macOS (Instruments Time Profiler)"
echo "===================================================="
echo ""

# Check if we're on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "Error: This script is for macOS only. Use profile_cpu_linux.sh for Linux."
    exit 1
fi

# Check if Instruments is available
if ! command -v instruments &> /dev/null; then
    echo "Error: Instruments not found. Please install Xcode Command Line Tools:"
    echo "  xcode-select --install"
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
TRACE_FILE="$RESULTS_DIR/${OUTPUT_NAME}.trace"
REPORT_FILE="$RESULTS_DIR/${OUTPUT_NAME}_report.txt"
SUMMARY_FILE="$RESULTS_DIR/${OUTPUT_NAME}_summary.txt"

echo "Configuration:"
echo "  Application: $APP_BINARY"
echo "  Arguments: $APP_ARGS"
echo "  Duration: ${DURATION}s"
echo "  Results directory: $RESULTS_DIR"
echo "  Trace file: $TRACE_FILE"
echo ""

# Method 1: Using Instruments CLI
echo "Starting application and profiling with Instruments..."
echo "(This will run for ${DURATION} seconds)"
echo ""

# Start the application in background
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
echo "Profiling..."

# Run Instruments Time Profiler
instruments -t "Time Profiler" \
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

# Extract report from trace file (if possible)
if [ -f "$TRACE_FILE" ]; then
    echo "Trace file created: $TRACE_FILE"
    echo "To view in Instruments GUI: open \"$TRACE_FILE\""
    echo ""
    
    # Try to extract text summary
    echo "Extracting summary..." 
    xctrace export --input "$TRACE_FILE" --output "$RESULTS_DIR/${OUTPUT_NAME}_export" 2>/dev/null || true
    
    if [ -d "$RESULTS_DIR/${OUTPUT_NAME}_export" ]; then
        echo "Exported data to: $RESULTS_DIR/${OUTPUT_NAME}_export"
    fi
fi

# Create a summary from application logs
echo "Creating summary report..."
cat > "$SUMMARY_FILE" << EOF
CPU Profiling Summary
=====================

Date: $(date)
Application: $APP_BINARY
Arguments: $APP_ARGS
Duration: ${DURATION}s
PID: $APP_PID

Files Generated:
- Trace file: $TRACE_FILE
- App log: $RESULTS_DIR/${OUTPUT_NAME}_app.log
- Instruments log: $RESULTS_DIR/${OUTPUT_NAME}_instruments.log

Application Performance Metrics (from logs):
============================================
EOF

# Extract performance metrics from app logs if available
if [ -f "$RESULTS_DIR/${OUTPUT_NAME}_app.log" ]; then
    echo "" >> "$SUMMARY_FILE"
    grep -i "fps\|performance\|processing time" "$RESULTS_DIR/${OUTPUT_NAME}_app.log" | tail -20 >> "$SUMMARY_FILE" 2>/dev/null || echo "No performance metrics found in logs" >> "$SUMMARY_FILE"
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
echo "  3. Analyze hot functions and optimize bottlenecks"
echo "  4. See docs/PROFILING.md for analysis guide"
echo ""
