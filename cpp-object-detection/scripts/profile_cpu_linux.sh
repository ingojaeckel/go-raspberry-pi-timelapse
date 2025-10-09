#!/bin/bash
# CPU profiling script for Linux using perf
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
FREQUENCY=99  # Sampling frequency in Hz

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
        --frequency)
            FREQUENCY="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --duration SECONDS    Profiling duration in seconds (default: 60)"
            echo "  --output NAME         Output file base name (default: cpu_profile_TIMESTAMP)"
            echo "  --app-args ARGS       Arguments to pass to application (default: '--max-fps 10 --verbose')"
            echo "  --frequency HZ        Sampling frequency in Hz (default: 99)"
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

echo "CPU Profiling for Linux (perf)"
echo "==============================="
echo ""

# Check if we're on Linux
if [[ "$OSTYPE" != "linux-gnu"* ]]; then
    echo "Error: This script is for Linux only. Use profile_cpu_mac.sh for macOS."
    exit 1
fi

# Check if perf is available
if ! command -v perf &> /dev/null; then
    echo "Error: perf not found. Install with:"
    echo "  Ubuntu/Debian: sudo apt-get install linux-tools-generic linux-tools-\$(uname -r)"
    echo "  CentOS/RHEL: sudo yum install perf"
    exit 1
fi

# Check if application is built
if [ ! -f "$APP_BINARY" ]; then
    echo "Error: Application not found at $APP_BINARY"
    echo "Please build first with: ./scripts/build_profile_linux.sh"
    exit 1
fi

# Create results directory
mkdir -p "$RESULTS_DIR"

# Output files
PERF_DATA="$RESULTS_DIR/${OUTPUT_NAME}.perf.data"
REPORT_FILE="$RESULTS_DIR/${OUTPUT_NAME}_report.txt"
FLAMEGRAPH_FILE="$RESULTS_DIR/${OUTPUT_NAME}_flamegraph.svg"
SUMMARY_FILE="$RESULTS_DIR/${OUTPUT_NAME}_summary.txt"

echo "Configuration:"
echo "  Application: $APP_BINARY"
echo "  Arguments: $APP_ARGS"
echo "  Duration: ${DURATION}s"
echo "  Sampling frequency: ${FREQUENCY} Hz"
echo "  Results directory: $RESULTS_DIR"
echo ""

# Check perf permissions
echo "Checking perf permissions..."
PARANOID=$(cat /proc/sys/kernel/perf_event_paranoid 2>/dev/null || echo "999")
if [ "$PARANOID" -gt 1 ]; then
    echo "Warning: perf_event_paranoid is set to $PARANOID"
    echo "For best results, run:"
    echo "  sudo sysctl -w kernel.perf_event_paranoid=-1"
    echo ""
    echo "Continuing with current settings (may need sudo)..."
    USE_SUDO="sudo"
else
    USE_SUDO=""
fi

echo ""
echo "Starting application and profiling with perf..."
echo "(This will run for ${DURATION} seconds)"
echo ""

# Create a script to run the application with a timeout
RUNNER_SCRIPT="$RESULTS_DIR/${OUTPUT_NAME}_runner.sh"
cat > "$RUNNER_SCRIPT" << EOF
#!/bin/bash
$APP_BINARY $APP_ARGS &
APP_PID=\$!
sleep $DURATION
kill \$APP_PID 2>/dev/null || true
sleep 1
kill -9 \$APP_PID 2>/dev/null || true
EOF
chmod +x "$RUNNER_SCRIPT"

# Run perf record
$USE_SUDO perf record \
    -F $FREQUENCY \
    -g \
    --call-graph dwarf \
    -o "$PERF_DATA" \
    -- "$RUNNER_SCRIPT" \
    > "$RESULTS_DIR/${OUTPUT_NAME}_app.log" 2>&1

echo ""
echo "✅ Profiling complete!"
echo ""

# Generate report
echo "Generating perf report..."
$USE_SUDO perf report -i "$PERF_DATA" --stdio > "$REPORT_FILE" 2>/dev/null || true

# Generate flame graph if flamegraph.pl is available
if command -v flamegraph.pl &> /dev/null || [ -f "$SCRIPT_DIR/flamegraph.pl" ]; then
    echo "Generating flame graph..."
    FLAMEGRAPH_PL=$(command -v flamegraph.pl 2>/dev/null || echo "$SCRIPT_DIR/flamegraph.pl")
    $USE_SUDO perf script -i "$PERF_DATA" | "$FLAMEGRAPH_PL" > "$FLAMEGRAPH_FILE" 2>/dev/null || true
    
    if [ -f "$FLAMEGRAPH_FILE" ]; then
        echo "Flame graph created: $FLAMEGRAPH_FILE"
    fi
else
    echo "Note: flamegraph.pl not found. Skipping flame graph generation."
    echo "Download from: https://github.com/brendangregg/FlameGraph"
fi

# Create summary
echo ""
echo "Creating summary report..."
cat > "$SUMMARY_FILE" << EOF
CPU Profiling Summary
=====================

Date: $(date)
Application: $APP_BINARY
Arguments: $APP_ARGS
Duration: ${DURATION}s
Sampling frequency: ${FREQUENCY} Hz

Files Generated:
- perf data: $PERF_DATA
- Report: $REPORT_FILE
- App log: $RESULTS_DIR/${OUTPUT_NAME}_app.log
$([ -f "$FLAMEGRAPH_FILE" ] && echo "- Flame graph: $FLAMEGRAPH_FILE")

Top 20 Functions by CPU Usage:
===============================
EOF

# Extract top functions from report
if [ -f "$REPORT_FILE" ]; then
    echo "" >> "$SUMMARY_FILE"
    grep -A 20 "# Overhead" "$REPORT_FILE" | head -25 >> "$SUMMARY_FILE" 2>/dev/null || echo "Could not extract top functions" >> "$SUMMARY_FILE"
fi

# Extract performance metrics from app logs
if [ -f "$RESULTS_DIR/${OUTPUT_NAME}_app.log" ]; then
    echo "" >> "$SUMMARY_FILE"
    echo "" >> "$SUMMARY_FILE"
    echo "Application Performance Metrics:" >> "$SUMMARY_FILE"
    echo "================================" >> "$SUMMARY_FILE"
    grep -i "fps\|performance\|processing time" "$RESULTS_DIR/${OUTPUT_NAME}_app.log" | tail -20 >> "$SUMMARY_FILE" 2>/dev/null || echo "No performance metrics found in logs" >> "$SUMMARY_FILE"
fi

echo ""
echo "Results saved to:"
echo "  - perf data: $PERF_DATA"
echo "  - Report: $REPORT_FILE"
echo "  - Summary: $SUMMARY_FILE"
[ -f "$FLAMEGRAPH_FILE" ] && echo "  - Flame graph: $FLAMEGRAPH_FILE"
echo ""
echo "Next steps:"
echo "  1. Review summary: cat \"$SUMMARY_FILE\""
echo "  2. Detailed report: less \"$REPORT_FILE\""
[ -f "$FLAMEGRAPH_FILE" ] && echo "  3. View flame graph: xdg-open \"$FLAMEGRAPH_FILE\""
echo "  4. See docs/PROFILING.md for analysis guide"
echo ""
