#!/bin/bash
# Memory profiling script for Linux using Valgrind and heaptrack
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
TOOL="valgrind"  # valgrind or heaptrack

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
        --tool)
            TOOL="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --duration SECONDS    Profiling duration in seconds (default: 60)"
            echo "  --output NAME         Output file base name (default: memory_profile_TIMESTAMP)"
            echo "  --app-args ARGS       Arguments to pass to application (default: '--max-fps 5 --verbose')"
            echo "  --tool TOOL           Profiling tool: 'valgrind' or 'heaptrack' (default: valgrind)"
            echo "  --help                Show this help message"
            echo ""
            echo "Example:"
            echo "  $0 --duration 120 --tool heaptrack --output heap_analysis"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

echo "Memory Profiling for Linux"
echo "=========================="
echo ""

# Check if we're on Linux
if [[ "$OSTYPE" != "linux-gnu"* ]]; then
    echo "Error: This script is for Linux only. Use profile_memory_mac.sh for macOS."
    exit 1
fi

# Validate tool
if [[ "$TOOL" != "valgrind" && "$TOOL" != "heaptrack" ]]; then
    echo "Error: Invalid tool '$TOOL'. Must be 'valgrind' or 'heaptrack'"
    exit 1
fi

# Check if tool is available
if ! command -v $TOOL &> /dev/null; then
    echo "Error: $TOOL not found. Install with:"
    echo "  Ubuntu/Debian: sudo apt-get install $TOOL"
    echo "  CentOS/RHEL: sudo yum install $TOOL"
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
SUMMARY_FILE="$RESULTS_DIR/${OUTPUT_NAME}_summary.txt"

echo "Configuration:"
echo "  Application: $APP_BINARY"
echo "  Arguments: $APP_ARGS"
echo "  Tool: $TOOL"
echo "  Duration: ${DURATION}s"
echo "  Results directory: $RESULTS_DIR"
echo ""

if [ "$TOOL" == "valgrind" ]; then
    # Valgrind profiling
    MEMCHECK_LOG="$RESULTS_DIR/${OUTPUT_NAME}_memcheck.log"
    MASSIF_OUT="$RESULTS_DIR/${OUTPUT_NAME}_massif.out"
    MASSIF_REPORT="$RESULTS_DIR/${OUTPUT_NAME}_massif_report.txt"
    
    echo "Running Valgrind memory profiling..."
    echo "(This will be SLOW - expect 10-50x slowdown)"
    echo ""
    
    # Create a timeout wrapper script
    TIMEOUT_SCRIPT="$RESULTS_DIR/${OUTPUT_NAME}_timeout.sh"
    cat > "$TIMEOUT_SCRIPT" << EOF
#!/bin/bash
timeout ${DURATION}s valgrind \\
    --leak-check=full \\
    --show-leak-kinds=all \\
    --track-origins=yes \\
    --verbose \\
    --log-file=$MEMCHECK_LOG \\
    --tool=memcheck \\
    $APP_BINARY $APP_ARGS
EOF
    chmod +x "$TIMEOUT_SCRIPT"
    
    # Run memcheck for leak detection
    echo "1/2: Running memcheck (leak detection)..."
    "$TIMEOUT_SCRIPT" > "$RESULTS_DIR/${OUTPUT_NAME}_app_memcheck.log" 2>&1 || true
    
    # Run massif for heap profiling
    echo "2/2: Running massif (heap profiling)..."
    timeout ${DURATION}s valgrind \
        --tool=massif \
        --massif-out-file="$MASSIF_OUT" \
        $APP_BINARY $APP_ARGS \
        > "$RESULTS_DIR/${OUTPUT_NAME}_app_massif.log" 2>&1 || true
    
    # Generate massif report
    if [ -f "$MASSIF_OUT" ]; then
        ms_print "$MASSIF_OUT" > "$MASSIF_REPORT" 2>/dev/null || true
    fi
    
    echo ""
    echo "✅ Valgrind profiling complete!"
    echo ""
    
    # Create summary
    cat > "$SUMMARY_FILE" << EOF
Memory Profiling Summary (Valgrind)
====================================

Date: $(date)
Application: $APP_BINARY
Arguments: $APP_ARGS
Duration: ${DURATION}s

Files Generated:
- Memcheck log: $MEMCHECK_LOG
- Massif output: $MASSIF_OUT
- Massif report: $MASSIF_REPORT
- App log (memcheck): $RESULTS_DIR/${OUTPUT_NAME}_app_memcheck.log
- App log (massif): $RESULTS_DIR/${OUTPUT_NAME}_app_massif.log

Memory Leak Summary:
====================
EOF
    
    # Extract leak summary from memcheck log
    if [ -f "$MEMCHECK_LOG" ]; then
        echo "" >> "$SUMMARY_FILE"
        grep -A 10 "LEAK SUMMARY" "$MEMCHECK_LOG" >> "$SUMMARY_FILE" 2>/dev/null || echo "No leak summary found" >> "$SUMMARY_FILE"
        echo "" >> "$SUMMARY_FILE"
        echo "Error Summary:" >> "$SUMMARY_FILE"
        echo "==============" >> "$SUMMARY_FILE"
        grep -A 5 "ERROR SUMMARY" "$MEMCHECK_LOG" >> "$SUMMARY_FILE" 2>/dev/null || echo "No error summary found" >> "$SUMMARY_FILE"
    fi
    
    # Extract heap peak from massif
    if [ -f "$MASSIF_REPORT" ]; then
        echo "" >> "$SUMMARY_FILE"
        echo "" >> "$SUMMARY_FILE"
        echo "Heap Usage Peak:" >> "$SUMMARY_FILE"
        echo "================" >> "$SUMMARY_FILE"
        grep -A 2 "peak" "$MASSIF_REPORT" | head -5 >> "$SUMMARY_FILE" 2>/dev/null || echo "No heap peak found" >> "$SUMMARY_FILE"
    fi
    
    echo ""
    echo "Results saved to:"
    echo "  - Memcheck log: $MEMCHECK_LOG"
    echo "  - Massif report: $MASSIF_REPORT"
    echo "  - Summary: $SUMMARY_FILE"
    echo ""
    echo "Next steps:"
    echo "  1. Review summary: cat \"$SUMMARY_FILE\""
    echo "  2. Check for leaks: grep -i leak \"$MEMCHECK_LOG\""
    echo "  3. Heap analysis: less \"$MASSIF_REPORT\""
    
else
    # heaptrack profiling
    HEAPTRACK_OUT="$RESULTS_DIR/${OUTPUT_NAME}_heaptrack.gz"
    HEAPTRACK_REPORT="$RESULTS_DIR/${OUTPUT_NAME}_heaptrack_report.txt"
    
    echo "Running heaptrack memory profiling..."
    echo ""
    
    # Create timeout wrapper
    TIMEOUT_SCRIPT="$RESULTS_DIR/${OUTPUT_NAME}_timeout.sh"
    cat > "$TIMEOUT_SCRIPT" << EOF
#!/bin/bash
cd $RESULTS_DIR
timeout ${DURATION}s heaptrack -o ${OUTPUT_NAME}_heaptrack $APP_BINARY $APP_ARGS
EOF
    chmod +x "$TIMEOUT_SCRIPT"
    
    # Run heaptrack
    "$TIMEOUT_SCRIPT" > "$RESULTS_DIR/${OUTPUT_NAME}_app.log" 2>&1 || true
    
    # Find the generated file (heaptrack adds PID to filename)
    ACTUAL_FILE=$(ls -t $RESULTS_DIR/${OUTPUT_NAME}_heaptrack.*.gz 2>/dev/null | head -1)
    
    if [ -f "$ACTUAL_FILE" ]; then
        # Generate text report
        heaptrack_print "$ACTUAL_FILE" > "$HEAPTRACK_REPORT" 2>/dev/null || true
        
        echo ""
        echo "✅ heaptrack profiling complete!"
        echo ""
        
        # Create summary
        cat > "$SUMMARY_FILE" << EOF
Memory Profiling Summary (heaptrack)
=====================================

Date: $(date)
Application: $APP_BINARY
Arguments: $APP_ARGS
Duration: ${DURATION}s

Files Generated:
- heaptrack data: $ACTUAL_FILE
- heaptrack report: $HEAPTRACK_REPORT
- App log: $RESULTS_DIR/${OUTPUT_NAME}_app.log

Heap Analysis:
==============
EOF
        
        # Extract key info from report
        if [ -f "$HEAPTRACK_REPORT" ]; then
            echo "" >> "$SUMMARY_FILE"
            grep -A 10 "MOST ALLOC" "$HEAPTRACK_REPORT" | head -15 >> "$SUMMARY_FILE" 2>/dev/null || true
            echo "" >> "$SUMMARY_FILE"
            echo "" >> "$SUMMARY_FILE"
            echo "Peak Heap Memory:" >> "$SUMMARY_FILE"
            grep -i "peak" "$HEAPTRACK_REPORT" | head -5 >> "$SUMMARY_FILE" 2>/dev/null || true
        fi
        
        echo ""
        echo "Results saved to:"
        echo "  - heaptrack data: $ACTUAL_FILE"
        echo "  - heaptrack report: $HEAPTRACK_REPORT"
        echo "  - Summary: $SUMMARY_FILE"
        echo ""
        echo "Next steps:"
        echo "  1. Review summary: cat \"$SUMMARY_FILE\""
        echo "  2. Detailed report: less \"$HEAPTRACK_REPORT\""
        if command -v heaptrack_gui &> /dev/null; then
            echo "  3. Visual analysis: heaptrack_gui \"$ACTUAL_FILE\""
        fi
    else
        echo "Error: heaptrack output file not found"
        exit 1
    fi
fi

echo "  4. See docs/PROFILING.md for analysis guide"
echo ""
