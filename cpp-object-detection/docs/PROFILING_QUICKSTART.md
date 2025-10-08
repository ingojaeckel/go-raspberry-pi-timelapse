# Performance Profiling Quick Reference

This is a quick reference for common profiling tasks. For detailed information, see [PROFILING.md](PROFILING.md).

## Quick Start

### macOS

```bash
# 1. Build for profiling
./scripts/build_profile_mac.sh

# 2. Profile CPU (60 seconds)
./scripts/profile_cpu_mac.sh

# 3. View results
open profiling_results/*_cpu_profile*.trace
```

### Linux

```bash
# 1. Build for profiling
./scripts/build_profile_linux.sh

# 2. Profile CPU (60 seconds)
./scripts/profile_cpu_linux.sh

# 3. View results
cat profiling_results/*_summary.txt
```

## Common Tasks

### Find CPU Hotspots

**macOS:**
```bash
./scripts/profile_cpu_mac.sh --duration 60
open profiling_results/*.trace
# In Instruments: Look at "Heaviest Stack Trace" view
```

**Linux:**
```bash
./scripts/profile_cpu_linux.sh --duration 60
cat profiling_results/*_summary.txt  # Top functions
```

### Find Memory Leaks

**macOS:**
```bash
./scripts/profile_memory_mac.sh --type leaks --duration 120
open profiling_results/*_leaks.trace
```

**Linux:**
```bash
./scripts/profile_memory_linux.sh --tool valgrind --duration 120
grep -i leak profiling_results/*_memcheck.log
```

### Heap Memory Analysis

**macOS:**
```bash
./scripts/profile_memory_mac.sh --type allocations --duration 60
open profiling_results/*_allocations.trace
# Check for growing allocations
```

**Linux:**
```bash
./scripts/profile_memory_linux.sh --tool heaptrack --duration 60
less profiling_results/*_heaptrack_report.txt
# Or use heaptrack_gui for visual analysis
```

### Generate Flame Graph (Linux)

```bash
# Download FlameGraph tools (one time)
./scripts/download_flamegraph.sh

# Profile and generate flame graph
./scripts/profile_cpu_linux.sh --duration 60
# Flame graph automatically created at:
# profiling_results/*_flamegraph.svg

# View in browser
firefox profiling_results/*_flamegraph.svg
```

## VSCode Integration

### Launch Profiling Build

1. Press `Cmd/Ctrl+Shift+P`
2. Type "Tasks: Run Task"
3. Select `build-profile`

### Run with Profiler Attached

1. Press `F5` or click Run and Debug
2. Select configuration:
   - `Profile: CPU (macOS)` or `Profile: CPU (Linux)`
   - `Profile: Memory - Valgrind (Linux)`
   - `Profile: Performance Test`
3. Application runs with profiling enabled
4. Results in `profiling_results/`

## Profiling Different Scenarios

### Low FPS (Standard)
```bash
./scripts/profile_cpu_mac.sh --app-args "--max-fps 5 --verbose"
```

### High FPS (Burst Mode)
```bash
./scripts/profile_cpu_mac.sh --app-args "--max-fps 15 --enable-burst-mode --verbose"
```

### GPU Accelerated
```bash
./scripts/profile_cpu_linux.sh --app-args "--max-fps 10 --enable-gpu --verbose"
```

### Long-Running Test
```bash
./scripts/profile_memory_linux.sh --duration 300 --tool heaptrack
```

## Analyzing Results

### CPU Profiling

**Look for:**
- Functions using >5% CPU time
- Unexpected hot paths
- Lock contention in threading

**Common bottlenecks:**
- `cv::dnn::Net::forward()` - YOLO inference (expected)
- `cv::resize()` - Image preprocessing
- File I/O operations
- Lock waits in parallel processing

### Memory Profiling

**Look for:**
- Growing heap size over time
- Large single allocations
- Frequent small allocations
- Leaked memory

**Common issues:**
- Unbounded frame queue
- Image buffers not reused
- OpenCV Mat objects not released

## Tool Comparison

| Tool | Platform | What it finds | Speed | Best for |
|------|----------|---------------|-------|----------|
| **Instruments** | macOS | CPU, memory, GPU, I/O | Fast | All-purpose profiling |
| **perf** | Linux | CPU, cache, branches | Fast | CPU hotspots |
| **Valgrind** | Both | Memory leaks, errors | Slow (10-50x) | Finding leaks |
| **heaptrack** | Linux | Heap allocations | Medium (2-3x) | Heap analysis |
| **gprof** | Both | Function calls | Medium | Call graph |

## Performance Targets

| Metric | Target | Red Flag |
|--------|--------|----------|
| **CPU Usage** | <80% on one core | >95% sustained |
| **Memory Growth** | Stable | Growing >10MB/min |
| **FPS** | >= configured max-fps | <50% of max-fps |
| **Frame Processing** | <200ms @ 5fps | >400ms |

## Optimization Workflow

1. **Profile** - Identify bottleneck
2. **Hypothesize** - Form theory about cause
3. **Implement** - Make targeted change
4. **Measure** - Profile again
5. **Compare** - Verify improvement
6. **Document** - Record findings

## Example: Optimizing Frame Processing

```bash
# 1. Baseline profile
./scripts/profile_cpu_linux.sh --output baseline

# 2. Review results
cat profiling_results/baseline_summary.txt
# Example finding: cv::resize() using 35% CPU

# 3. Hypothesis: Reduce resize operations
# Change: Increase detection_scale_factor from 0.5 to 0.75

# 4. Profile optimized version
./build-profile/object_detection --detection-scale 0.75 --max-fps 10 &
sleep 60
killall object_detection

./scripts/profile_cpu_linux.sh --output optimized

# 5. Compare
diff profiling_results/baseline_summary.txt \
     profiling_results/optimized_summary.txt

# 6. Document in profiling_results/NOTES.md
echo "Optimization: detection_scale 0.5 -> 0.75" >> profiling_results/NOTES.md
echo "Result: cv::resize CPU: 35% -> 18%" >> profiling_results/NOTES.md
echo "Impact: Average FPS: 8.2 -> 12.1" >> profiling_results/NOTES.md
```

## Troubleshooting

### "Permission denied" (Linux perf)
```bash
sudo sysctl -w kernel.perf_event_paranoid=-1
```

### "Instruments not found" (macOS)
```bash
xcode-select --install
```

### "Valgrind too slow"
Use heaptrack instead:
```bash
./scripts/profile_memory_linux.sh --tool heaptrack
```

### "No flame graph generated"
Download FlameGraph tools:
```bash
./scripts/download_flamegraph.sh
```

## Resources

- **Full Guide**: [PROFILING.md](PROFILING.md)
- **Architecture**: [ARCHITECTURE.md](ARCHITECTURE.md)
- **GPU Guide**: [GPU_ACCELERATION.md](GPU_ACCELERATION.md)
- **FlameGraph**: https://github.com/brendangregg/FlameGraph
- **Linux perf**: http://www.brendangregg.com/perf.html
- **Instruments**: https://help.apple.com/instruments/

---

💡 **Tip**: Always profile with realistic workloads and representative test data for accurate results.
