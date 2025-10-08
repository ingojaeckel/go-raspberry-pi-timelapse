# Profiling Infrastructure Summary

This document summarizes the profiling infrastructure added to the C++ Object Detection application.

## Overview

The application now has comprehensive cross-platform profiling support for both **macOS (Darwin/AMD64)** and **Linux (64-bit Intel/AMD)**, enabling performance analysis and optimization.

## What Was Added

### Documentation

1. **[docs/PROFILING.md](docs/PROFILING.md)** (14.5 KB)
   - Complete profiling guide
   - Tool selection and installation
   - CPU and memory profiling workflows
   - Performance analysis techniques
   - Platform-specific best practices
   - Troubleshooting guide

2. **[docs/PROFILING_QUICKSTART.md](docs/PROFILING_QUICKSTART.md)** (6.1 KB)
   - Quick reference guide
   - Common profiling tasks
   - Command examples
   - Analysis tips
   - Optimization workflow

3. **[.vscode/README.md](../.vscode/README.md)** (Updated)
   - VSCode profiling configurations
   - Task usage guide
   - Profiling workflow examples

### Build Scripts

1. **`scripts/build_profile_mac.sh`** - macOS profiling build
   - RelWithDebInfo configuration
   - Frame pointers enabled
   - Debug symbols included
   - Optimizations enabled

2. **`scripts/build_profile_linux.sh`** - Linux profiling build
   - RelWithDebInfo configuration
   - Frame pointers enabled
   - Debug symbols included
   - Optimizations enabled

### CPU Profiling Scripts

1. **`scripts/profile_cpu_mac.sh`** - macOS CPU profiling
   - Uses Apple Instruments Time Profiler
   - Configurable duration and parameters
   - Generates trace files for GUI analysis
   - Extracts performance summaries
   - Command-line options for flexibility

2. **`scripts/profile_cpu_linux.sh`** - Linux CPU profiling
   - Uses Linux perf tool
   - Configurable sampling frequency
   - Generates flame graphs (if tools available)
   - Text reports and summaries
   - Handles permissions automatically

### Memory Profiling Scripts

1. **`scripts/profile_memory_mac.sh`** - macOS memory profiling
   - Instruments Allocations profile
   - Instruments Leaks detection
   - Configurable profile types
   - Trace files for detailed analysis

2. **`scripts/profile_memory_linux.sh`** - Linux memory profiling
   - Valgrind memcheck and massif
   - heaptrack heap profiling
   - Tool selection via command-line
   - Detailed leak reports

### Helper Scripts

1. **`scripts/download_flamegraph.sh`**
   - Downloads FlameGraph tools from GitHub
   - Sets up flame graph generation
   - Used by CPU profiling scripts

### VSCode Integration

1. **`.vscode/launch.json`** (Updated)
   - New profiling launch configurations:
     - `Profile: CPU (macOS)` - Run with profiling symbols on macOS
     - `Profile: CPU (Linux)` - Run with profiling symbols on Linux
     - `Profile: Memory - Valgrind (Linux)` - Run under Valgrind
     - `Profile: Performance Test` - High-load testing scenario

2. **`.vscode/tasks.json`** (Updated)
   - New profiling tasks:
     - `build-profile` - Build with profiling flags
     - `profile-cpu-mac` - Run macOS CPU profiling
     - `profile-cpu-linux` - Run Linux CPU profiling
     - `profile-memory-mac` - Run macOS memory profiling
     - `profile-memory-linux` - Run Linux memory profiling
     - `clean-profiling-results` - Clean profiling output

### Configuration Files

1. **`.gitignore`** (New)
   - Excludes build directories
   - Excludes profiling results
   - Excludes IDE files
   - Excludes temporary files

## Profiling Tools Supported

### macOS (Darwin/AMD64)

| Tool | Purpose | Usage |
|------|---------|-------|
| **Instruments** | CPU, memory, GPU profiling | GUI and CLI |
| **dtrace** | System-level tracing | CLI |
| **Valgrind** | Memory debugging | CLI (if installed) |
| **gprof** | Function call profiling | CLI |

### Linux (64-bit Intel/AMD)

| Tool | Purpose | Usage |
|------|---------|-------|
| **perf** | CPU profiling, performance counters | CLI |
| **Valgrind** | Memory debugging, leak detection | CLI |
| **heaptrack** | Heap profiling | CLI + GUI |
| **gprof** | Function call profiling | CLI |

## Quick Start Guide

### For macOS Users

```bash
# 1. Build for profiling
cd cpp-object-detection
./scripts/build_profile_mac.sh

# 2. Run CPU profiling
./scripts/profile_cpu_mac.sh --duration 60

# 3. View results in Instruments
open profiling_results/*_cpu_profile*.trace

# 4. Run memory profiling
./scripts/profile_memory_mac.sh --duration 60 --type allocations
open profiling_results/*_allocations.trace
```

### For Linux Users

```bash
# 1. Build for profiling
cd cpp-object-detection
./scripts/build_profile_linux.sh

# 2. Run CPU profiling
./scripts/profile_cpu_linux.sh --duration 60

# 3. View summary
cat profiling_results/*_summary.txt

# 4. Run memory profiling
./scripts/profile_memory_linux.sh --duration 60 --tool valgrind

# 5. Check for leaks
grep -i leak profiling_results/*_memcheck.log
```

### Using VSCode

1. **Open Run and Debug panel** (`Cmd/Ctrl+Shift+D`)
2. **Select profiling configuration**:
   - `Profile: CPU (macOS)` or `Profile: CPU (Linux)`
   - `Profile: Memory - Valgrind (Linux)`
3. **Press F5** to start profiling
4. **Results** saved to `profiling_results/`

Or use tasks:
1. **Press** `Cmd/Ctrl+Shift+P`
2. **Type** "Tasks: Run Task"
3. **Select** profiling task (e.g., `profile-cpu-linux`)

## Profiling Workflow

### 1. Build with Profiling Symbols

```bash
# macOS
./scripts/build_profile_mac.sh

# Linux
./scripts/build_profile_linux.sh
```

This creates a `build-profile/` directory with:
- **RelWithDebInfo** build type (optimized + debug symbols)
- **Frame pointers** enabled (for accurate call stacks)
- **Compile commands** exported (for IDE integration)

### 2. Run Profiling Session

**CPU Profiling:**
```bash
# macOS
./scripts/profile_cpu_mac.sh --duration 120

# Linux
./scripts/profile_cpu_linux.sh --duration 120
```

**Memory Profiling:**
```bash
# macOS - Allocations
./scripts/profile_memory_mac.sh --type allocations --duration 60

# Linux - Valgrind
./scripts/profile_memory_linux.sh --tool valgrind --duration 60
```

### 3. Analyze Results

**CPU Analysis:**
- Identify hot functions (>5% CPU time)
- Review call stacks
- Generate flame graphs (Linux)
- Check for unexpected bottlenecks

**Memory Analysis:**
- Look for memory leaks
- Check heap growth over time
- Identify large allocations
- Review allocation patterns

### 4. Optimize and Re-measure

1. Identify bottleneck from profile
2. Make targeted optimization
3. Re-run profiling
4. Compare before/after results
5. Document findings

## Output Files

All profiling results are saved to `profiling_results/` directory:

### macOS
- `*_cpu_profile*.trace` - Instruments CPU trace (open in Instruments.app)
- `*_allocations.trace` - Instruments memory trace
- `*_leaks.trace` - Instruments leak detection trace
- `*_summary.txt` - Text summary of results
- `*_app.log` - Application output during profiling

### Linux
- `*.perf.data` - perf profiling data
- `*_report.txt` - perf report in text format
- `*_flamegraph.svg` - Flame graph visualization (if FlameGraph tools installed)
- `*_memcheck.log` - Valgrind memcheck output
- `*_massif.out` - Valgrind massif heap profiling data
- `heaptrack.*.gz` - heaptrack data files
- `*_summary.txt` - Text summary of results

## Performance Metrics to Monitor

### CPU Profiling

| Metric | Target | Red Flag |
|--------|--------|----------|
| CPU usage | <80% of one core | >95% sustained |
| Frame processing time | <200ms @ 5fps | >400ms |
| FPS | >= configured max-fps | <50% of target |

**Common bottlenecks:**
- `cv::dnn::Net::forward()` - YOLO inference (expected hotspot)
- `cv::resize()` - Image scaling (optimize with `--detection-scale`)
- File I/O - Disk writes
- Lock contention - Thread synchronization

### Memory Profiling

| Metric | Target | Red Flag |
|--------|--------|----------|
| Heap growth | Stable | >10MB/min growth |
| Memory leaks | 0 bytes | Any leaks |
| Peak memory | <500MB | >1GB |

**Common issues:**
- Unbounded frame queue
- OpenCV Mat objects not released
- Image buffers not reused
- Memory fragmentation

## Integration with Existing Performance Features

The profiling infrastructure complements existing performance monitoring:

1. **PerformanceMonitor** - Runtime FPS and processing time tracking
2. **SystemMonitor** - Disk space and CPU temperature monitoring
3. **Profiling tools** - Deep analysis of bottlenecks

Workflow:
1. **Run application** with `--verbose` flag
2. **Monitor logs** for performance warnings
3. **If issues detected**, run profiling scripts
4. **Analyze** bottlenecks and optimize
5. **Verify** improvements with re-profiling

## Testing and Verification

### Verify Installation

**macOS:**
```bash
# Check Instruments
instruments -s

# Check build script
./scripts/build_profile_mac.sh
ls -lh build-profile/object_detection
```

**Linux:**
```bash
# Check perf
perf --version

# Check Valgrind
valgrind --version

# Check build script
./scripts/build_profile_linux.sh
ls -lh build-profile/object_detection
```

### Test Profiling Scripts

**CPU Profiling:**
```bash
# macOS
./scripts/profile_cpu_mac.sh --duration 30 --app-args "--help"

# Linux
./scripts/profile_cpu_linux.sh --duration 30 --app-args "--help"

# Check results
ls -lh profiling_results/
```

**Memory Profiling:**
```bash
# Linux (quick test with heaptrack)
./scripts/profile_memory_linux.sh --duration 10 --tool heaptrack --app-args "--help"

# Check results
ls -lh profiling_results/
```

### Validate VSCode Integration

1. Open project in VSCode
2. Open Run and Debug panel
3. Verify profiling configurations appear
4. Select `Profile: CPU (Linux)` or `Profile: CPU (macOS)`
5. Press F5 (should build and launch)
6. Verify `profiling_results/` is created

## Documentation Links

- **[PROFILING.md](docs/PROFILING.md)** - Complete profiling guide
- **[PROFILING_QUICKSTART.md](docs/PROFILING_QUICKSTART.md)** - Quick reference
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** - System architecture
- **[GPU_ACCELERATION.md](docs/GPU_ACCELERATION.md)** - GPU optimization
- **[README.md](../README.md)** - Main project documentation

## Next Steps

1. **Test profiling on target platforms** (macOS and Linux)
2. **Verify VSCode configurations** work as expected
3. **Run baseline profiling** to establish performance metrics
4. **Document optimization findings** in profiling results
5. **Create performance benchmarks** for different scenarios
6. **Set up continuous performance testing** (optional)

## Benefits

✅ **Cross-platform profiling** - Works on both macOS and Linux
✅ **Multiple profiling tools** - CPU, memory, heap, cache analysis
✅ **VSCode integration** - Launch profiling directly from IDE
✅ **Automated scripts** - One-command profiling sessions
✅ **Comprehensive documentation** - Detailed guides and quick references
✅ **Flame graph support** - Visual performance analysis
✅ **Flexible configuration** - Customizable profiling parameters

---

**For questions or issues**, refer to:
- Documentation: `docs/PROFILING.md`
- Quick reference: `docs/PROFILING_QUICKSTART.md`
- Script help: `./scripts/profile_*.sh --help`
