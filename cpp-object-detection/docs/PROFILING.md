# Performance Profiling Guide

This guide provides comprehensive instructions for profiling the C++ Object Detection application on both macOS (Darwin/AMD64) and Linux (64-bit Intel/AMD) platforms.

## Table of Contents

1. [Overview](#overview)
2. [Profiling Tools](#profiling-tools)
3. [Quick Start](#quick-start)
4. [Building for Profiling](#building-for-profiling)
5. [CPU Profiling](#cpu-profiling)
6. [Memory Profiling](#memory-profiling)
7. [Performance Analysis](#performance-analysis)
8. [VSCode Integration](#vscode-integration)
9. [Best Practices](#best-practices)

## Overview

Performance profiling helps identify bottlenecks in:
- Frame processing pipeline
- Object detection inference
- Image preprocessing and scaling
- Multi-threaded operations
- Memory allocation patterns
- I/O operations (disk, network streaming)

## Profiling Tools

### macOS (Darwin/AMD64)

| Tool | Purpose | Installation |
|------|---------|--------------|
| **Instruments** | Apple's profiling tool (CPU, Memory, GPU) | Built-in with Xcode |
| **perf** (dtrace) | Low-level system profiling | Built-in |
| **Valgrind** | Memory profiling, leak detection | `brew install valgrind` |
| **gprof** | Function call profiling | Built-in with gcc/clang |
| **heaptrack** | Heap memory profiler | `brew install heaptrack` |

### Linux (64-bit Intel/AMD)

| Tool | Purpose | Installation |
|------|---------|--------------|
| **perf** | CPU profiling and performance counters | `sudo apt-get install linux-tools-generic` |
| **Valgrind** | Memory profiling, leak detection | `sudo apt-get install valgrind` |
| **gprof** | Function call profiling | Built-in with gcc |
| **heaptrack** | Heap memory profiler | `sudo apt-get install heaptrack` |
| **perf-tools** | Flame graphs and visualization | `sudo apt-get install linux-tools-common` |

## Quick Start

### 1. Build with Profiling Enabled

**macOS:**
```bash
./scripts/build_profile_mac.sh
```

**Linux:**
```bash
./scripts/build_profile_linux.sh
```

### 2. Run Profiling Script

**CPU Profiling:**
```bash
# macOS
./scripts/profile_cpu_mac.sh

# Linux
./scripts/profile_cpu_linux.sh
```

**Memory Profiling:**
```bash
# macOS
./scripts/profile_memory_mac.sh

# Linux
./scripts/profile_memory_linux.sh
```

### 3. Analyze Results

Profiling scripts generate reports in `profiling_results/` directory.

## Building for Profiling

### Debug Build with Profiling Symbols

```bash
cd cpp-object-detection
mkdir -p build-profile
cd build-profile
cmake .. -DCMAKE_BUILD_TYPE=RelWithDebInfo \
         -DCMAKE_CXX_FLAGS="-fno-omit-frame-pointer -g" \
         -DCMAKE_EXPORT_COMPILE_COMMANDS=ON
make -j$(nproc)
```

**RelWithDebInfo** provides:
- Optimizations enabled (like Release)
- Debug symbols included (for profiling)
- Frame pointers for accurate call stacks

### gprof Build (Function Profiling)

```bash
cd build-profile
cmake .. -DCMAKE_BUILD_TYPE=Debug \
         -DCMAKE_CXX_FLAGS="-pg -g -fno-omit-frame-pointer" \
         -DCMAKE_EXE_LINKER_FLAGS="-pg"
make -j$(nproc)
```

## CPU Profiling

### Using Instruments (macOS)

**Time Profiler for CPU Usage:**
```bash
# Start the application
./build-profile/object_detection --max-fps 10 --verbose &
APP_PID=$!

# Profile with Instruments
instruments -t "Time Profiler" -p $APP_PID -D profiling_results/cpu_profile.trace

# Stop after 60 seconds
sleep 60
kill $APP_PID

# Open in Instruments GUI
open profiling_results/cpu_profile.trace
```

**Or use the automated script:**
```bash
./scripts/profile_cpu_mac.sh --duration 60 --output cpu_profile
```

### Using perf (Linux)

**Record CPU samples:**
```bash
# Run application with perf
sudo perf record -F 99 -g --call-graph dwarf \
     ./build-profile/object_detection --max-fps 10 --verbose

# Press Ctrl+C after desired duration

# Generate report
sudo perf report > profiling_results/perf_report.txt

# Generate flame graph
sudo perf script | ./scripts/flamegraph.pl > profiling_results/flamegraph.svg
```

**Or use the automated script:**
```bash
./scripts/profile_cpu_linux.sh --duration 60 --output cpu_profile
```

### Using gprof

**After building with -pg flag:**
```bash
# Run application (generates gmon.out)
./build-profile/object_detection --max-fps 5 --verbose
# Press Ctrl+C after collecting data

# Generate profile report
gprof ./build-profile/object_detection gmon.out > profiling_results/gprof_analysis.txt

# View top functions
gprof ./build-profile/object_detection gmon.out | head -50
```

## Memory Profiling

### Using Valgrind (Both Platforms)

**Memory leak detection:**
```bash
valgrind --leak-check=full \
         --show-leak-kinds=all \
         --track-origins=yes \
         --verbose \
         --log-file=profiling_results/valgrind_memcheck.txt \
         ./build-profile/object_detection --max-fps 1 --verbose
```

**Heap profiling:**
```bash
valgrind --tool=massif \
         --massif-out-file=profiling_results/massif.out \
         ./build-profile/object_detection --max-fps 1 --verbose

# Visualize heap usage
ms_print profiling_results/massif.out > profiling_results/massif_report.txt
```

**Cache profiling:**
```bash
valgrind --tool=cachegrind \
         --cachegrind-out-file=profiling_results/cachegrind.out \
         ./build-profile/object_detection --max-fps 5

# Analyze cache misses
cg_annotate profiling_results/cachegrind.out > profiling_results/cachegrind_report.txt
```

### Using heaptrack (Both Platforms)

```bash
# Record heap allocations
heaptrack ./build-profile/object_detection --max-fps 5 --verbose

# This creates heaptrack.object_detection.PID.gz

# Analyze in GUI (if available)
heaptrack_gui heaptrack.object_detection.*.gz

# Or generate text report
heaptrack_print heaptrack.object_detection.*.gz > profiling_results/heaptrack_report.txt
```

### Using Instruments (macOS)

**Allocations profiling:**
```bash
./build-profile/object_detection --max-fps 10 --verbose &
APP_PID=$!

instruments -t "Allocations" -p $APP_PID -D profiling_results/memory_profile.trace

sleep 60
kill $APP_PID

open profiling_results/memory_profile.trace
```

## Performance Analysis

### Key Metrics to Monitor

1. **Frame Processing Rate**
   - Target: Maintain configured max FPS
   - Monitor: `PerformanceMonitor::getCurrentFPS()`
   - Check logs for "Performance warning" messages

2. **Average Processing Time**
   - Target: < (1000/max_fps) ms per frame
   - Monitor: `PerformanceMonitor::getAverageProcessingTime()`
   - Bottleneck if consistently high

3. **Object Detection Latency**
   - Critical path in frame processing
   - Profile `ObjectDetector::detectObjects()`
   - Check YOLOv5 inference time

4. **Memory Growth**
   - Monitor heap allocations over time
   - Check for memory leaks in long-running sessions
   - Review frame queue size and image buffers

### Interpreting Results

#### CPU Profiling

**Hot functions to investigate:**
- `cv::dnn::Net::forward()` - YOLO inference (expected)
- `cv::resize()` - Image scaling operations
- `ObjectDetector::detectObjects()` - Detection pipeline
- Frame queue operations in `ParallelFrameProcessor`

**Optimization opportunities:**
- High CPU in preprocessing → Adjust `detection_scale_factor`
- Inference dominates → Enable GPU acceleration or use lighter model
- Thread synchronization overhead → Review parallel processing settings

#### Memory Profiling

**Common issues:**
- Growing heap → Frame queue not being cleared, memory leak
- High peak memory → Large image buffers, model weights
- Frequent allocations → Inefficient buffer reuse

**Solutions:**
- Reuse OpenCV Mat objects
- Limit frame queue size (`--max-frame-queue`)
- Reduce image resolution or detection scale
- Review object lifetime in parallel processing

### Flame Graphs

Generate visual call stack analysis:

**macOS (using dtrace):**
```bash
sudo dtrace -x ustackframes=100 -n 'profile-997 /pid == $target/ { @[ustack()] = count(); }' \
     -p $(pgrep object_detection) -o profiling_results/stacks.txt

# Convert to flame graph (requires flamegraph.pl)
./scripts/flamegraph.pl profiling_results/stacks.txt > profiling_results/flamegraph.svg
```

**Linux (using perf):**
```bash
sudo perf record -F 99 -a -g -- sleep 30
sudo perf script | ./scripts/flamegraph.pl > profiling_results/flamegraph.svg
```

Open `flamegraph.svg` in a browser to visualize CPU time distribution.

## VSCode Integration

### Profiling Launch Configurations

The `.vscode/launch.json` includes profiling configurations:

1. **CPU Profile (macOS)** - Launch with Instruments Time Profiler
2. **CPU Profile (Linux)** - Launch with perf record
3. **Memory Profile (macOS)** - Launch with Instruments Allocations
4. **Memory Profile (Linux)** - Launch with Valgrind Massif

### Usage in VSCode

1. Open the Run and Debug panel (Cmd/Ctrl+Shift+D)
2. Select a profiling configuration from dropdown
3. Press F5 to start profiling
4. Application runs with profiler attached
5. Results saved to `profiling_results/` directory

### Custom Profiling Tasks

VSCode tasks in `.vscode/tasks.json`:
- `build-profile` - Build with profiling symbols
- `profile-cpu-report` - Generate CPU profile report
- `profile-memory-report` - Generate memory profile report
- `clean-profiling-results` - Clean profiling output directory

## Best Practices

### 1. Profiling Workflow

```bash
# 1. Build with profiling enabled
./scripts/build_profile_linux.sh  # or build_profile_mac.sh

# 2. Run realistic workload
# Use actual camera or representative test video
./build-profile/object_detection --max-fps 10 --camera-id 0 --verbose

# 3. Profile during peak load
# Trigger burst mode or high object count scenarios

# 4. Collect data for sufficient duration
# At least 60 seconds for statistical significance

# 5. Analyze results iteratively
# Focus on top 10 hot functions first
```

### 2. Platform-Specific Considerations

**macOS:**
- Use Instruments for GUI-based analysis
- Xcode Command Line Tools required
- May need to disable System Integrity Protection (SIP) for some tools
- Frame pointers always available on x86_64

**Linux:**
- Use `sudo` for system-wide perf profiling
- Enable kernel symbols: `sudo sysctl -w kernel.kptr_restrict=0`
- For best results: `sudo sysctl -w kernel.perf_event_paranoid=-1`
- Install debug symbols for system libraries

### 3. Representative Workloads

**Test scenarios for profiling:**

1. **Idle Camera** - No objects detected
   ```bash
   ./build-profile/object_detection --max-fps 5
   ```

2. **High Activity** - Multiple objects moving
   ```bash
   ./build-profile/object_detection --max-fps 15 --enable-burst-mode
   ```

3. **Sustained Load** - Long-running operation
   ```bash
   ./build-profile/object_detection --max-fps 10 --verbose
   # Run for 5+ minutes
   ```

4. **GPU Accelerated** - Compare GPU vs CPU
   ```bash
   ./build-profile/object_detection --max-fps 10 --enable-gpu
   ```

### 4. Optimization Cycle

1. **Profile** → Identify bottleneck
2. **Hypothesize** → Form optimization idea
3. **Implement** → Make targeted change
4. **Measure** → Profile again to verify improvement
5. **Iterate** → Repeat for next bottleneck

### 5. Profiling Checklist

- [ ] Build with RelWithDebInfo for accurate profiling
- [ ] Use realistic test data (actual camera feed preferred)
- [ ] Profile for sufficient duration (60+ seconds)
- [ ] Focus on hot paths (functions using >5% CPU)
- [ ] Check both CPU time and wall-clock time
- [ ] Monitor memory growth over time
- [ ] Profile with different configurations (FPS, resolution, GPU)
- [ ] Compare before/after optimization results
- [ ] Document findings in profiling_results/NOTES.md

## Troubleshooting

### Perf Issues (Linux)

**"perf_event_open failed: Permission denied"**
```bash
# Temporarily allow perf for all users
sudo sysctl -w kernel.perf_event_paranoid=-1

# Or run as root
sudo perf record ...
```

### Valgrind Issues

**"vgdb: error: no such process"**
- Ensure Valgrind is up to date: `sudo apt-get update && sudo apt-get upgrade valgrind`

**Too slow to be practical**
- Valgrind has 10-50x slowdown, use shorter test duration
- For CPU profiling, prefer perf/Instruments instead

### Instruments Issues (macOS)

**"Instruments cannot profile this process"**
- Check that application is built with debug symbols
- Ensure System Integrity Protection allows profiling

**Trace file too large**
- Limit profiling duration
- Reduce sampling rate: Use "System Trace" instead of "Time Profiler"

## Additional Resources

### Documentation
- [Instruments User Guide](https://help.apple.com/instruments/mac/current/)
- [Linux perf Tutorial](https://perf.wiki.kernel.org/index.php/Tutorial)
- [Valgrind Documentation](https://valgrind.org/docs/manual/manual.html)
- [Brendan Gregg's Performance Tools](http://www.brendangregg.com/linuxperf.html)

### Flame Graph Tools
- [FlameGraph Repository](https://github.com/brendangregg/FlameGraph)
- [Speedscope (Interactive Viewer)](https://www.speedscope.app/)

### Related Documentation
- [ARCHITECTURE.md](ARCHITECTURE.md) - System design and components
- [GPU_ACCELERATION.md](GPU_ACCELERATION.md) - GPU optimization guide
- [PERFORMANCE.md](PERFORMANCE.md) - Performance tuning guide (if exists)

## Example Analysis Session

Here's a complete example of finding and fixing a performance issue:

```bash
# 1. Build for profiling
cd cpp-object-detection
./scripts/build_profile_linux.sh

# 2. Run with CPU profiler
./scripts/profile_cpu_linux.sh --duration 120 --output baseline

# 3. Analyze results
less profiling_results/baseline_report.txt

# Example findings:
# - cv::resize() taking 35% of CPU time
# - YOLOv5 inference taking 45% of CPU time
# - Rest: 20% (frame queue, I/O, etc.)

# 4. Hypothesis: Reduce image resizing overhead
# Strategy: Increase detection_scale_factor to reduce resize operations

# 5. Test optimization
./build-profile/object_detection --detection-scale 0.75 --max-fps 10 &
./scripts/profile_cpu_linux.sh --duration 120 --output optimized

# 6. Compare results
diff profiling_results/baseline_report.txt profiling_results/optimized_report.txt

# 7. Verify FPS improvement
# Check application logs for actual FPS metrics

# 8. Document findings
echo "Optimization: Increased detection scale from 0.5 to 0.75" >> profiling_results/NOTES.md
echo "Result: cv::resize() CPU usage reduced from 35% to 20%" >> profiling_results/NOTES.md
echo "Impact: FPS increased from 8.2 to 11.5 avg" >> profiling_results/NOTES.md
```

---

For questions or issues with profiling, please refer to the issue tracker or consult the team.
