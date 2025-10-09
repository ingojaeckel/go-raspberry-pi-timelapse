#!/bin/bash
# Download FlameGraph tools for flame graph visualization
# These tools work on both macOS and Linux

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FLAMEGRAPH_DIR="$SCRIPT_DIR"

echo "FlameGraph Tools Setup"
echo "======================"
echo ""

# Check if flamegraph.pl already exists
if [ -f "$FLAMEGRAPH_DIR/flamegraph.pl" ]; then
    echo "✓ flamegraph.pl already exists at $FLAMEGRAPH_DIR/flamegraph.pl"
    echo ""
    echo "FlameGraph tools are ready to use!"
    echo ""
    echo "Usage:"
    echo "  - CPU profiling scripts will automatically use these tools"
    echo "  - Manual usage: perf script | $FLAMEGRAPH_DIR/flamegraph.pl > output.svg"
    exit 0
fi

echo "Downloading FlameGraph tools from GitHub..."
echo ""

# Download the main flame graph script
curl -L -o "$FLAMEGRAPH_DIR/flamegraph.pl" \
    https://raw.githubusercontent.com/brendangregg/FlameGraph/master/flamegraph.pl

# Download the stack collapse scripts (useful for different profilers)
curl -L -o "$FLAMEGRAPH_DIR/stackcollapse-perf.pl" \
    https://raw.githubusercontent.com/brendangregg/FlameGraph/master/stackcollapse-perf.pl

curl -L -o "$FLAMEGRAPH_DIR/stackcollapse-sample.pl" \
    https://raw.githubusercontent.com/brendangregg/FlameGraph/master/stackcollapse-sample.pl

# Make scripts executable
chmod +x "$FLAMEGRAPH_DIR/flamegraph.pl"
chmod +x "$FLAMEGRAPH_DIR/stackcollapse-perf.pl"
chmod +x "$FLAMEGRAPH_DIR/stackcollapse-sample.pl"

echo ""
echo "✅ FlameGraph tools downloaded successfully!"
echo ""
echo "Downloaded files:"
echo "  - $FLAMEGRAPH_DIR/flamegraph.pl"
echo "  - $FLAMEGRAPH_DIR/stackcollapse-perf.pl"
echo "  - $FLAMEGRAPH_DIR/stackcollapse-sample.pl"
echo ""
echo "Usage examples:"
echo ""
echo "Linux (with perf):"
echo "  sudo perf record -F 99 -a -g -- sleep 30"
echo "  sudo perf script | $FLAMEGRAPH_DIR/flamegraph.pl > flamegraph.svg"
echo ""
echo "macOS (with dtrace):"
echo "  sudo dtrace -x ustackframes=100 -n 'profile-997 /execname == \"object_detection\"/ { @[ustack()] = count(); }' -o out.stacks"
echo "  $FLAMEGRAPH_DIR/flamegraph.pl out.stacks > flamegraph.svg"
echo ""
echo "Note: CPU profiling scripts will automatically use these tools if available."
echo ""
