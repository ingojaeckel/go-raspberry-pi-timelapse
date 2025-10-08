# VSCode Debug and Profiling Configuration for cpp-object-detection

This directory contains VSCode debug and profiling configurations for the cpp-object-detection project, optimized for debugging and performance analysis on macOS and Linux.

## Prerequisites

### macOS
1. Install the **CodeLLDB** extension in VSCode:
   - Open VSCode
   - Press `Cmd+Shift+X` to open Extensions
   - Search for "CodeLLDB"
   - Install the extension by Vadim Chugunov

2. Ensure you have the C++ development tools:
   ```bash
   xcode-select --install
   ```

### Linux
1. Install the **C/C++** extension in VSCode:
   - Search for "C/C++" by Microsoft in Extensions
   - Install the extension

2. Install GDB:
   ```bash
   sudo apt-get install gdb
   ```

## Debug Configurations

### macOS Debug Configurations

#### 1. **(lldb) Launch - macOS Debug**
- **Purpose**: Debug the application with `--help` argument
- **Stops at**: Entry point (main function)
- **Best for**: Debugging why help output isn't working
- **Usage**:
  1. Open `src/main.cpp`
  2. Set breakpoints where you want to inspect (e.g., line 24 at `main()`, line 35 at `ConfigManager`, line 38 at help check)
  3. Press `F5` or select this configuration from the Debug panel
  4. Step through code with `F10` (step over) or `F11` (step into)

#### 2. **(lldb) Launch - macOS Debug (No Args)**
- **Purpose**: Debug the application without any arguments
- **Stops at**: Entry point
- **Best for**: Debugging default behavior and initialization issues
- **Usage**: Same as above but tests behavior without command-line arguments

#### 3. **(lldb) Attach to Process**
- **Purpose**: Attach to an already-running instance
- **Best for**: Debugging a hanging or stuck process
- **Usage**:
  1. Start the application manually in Terminal
  2. Select this configuration
  3. Choose the `object_detection` process from the list

### Linux Debug Configurations

Similar configurations using GDB instead of LLDB for Linux development.

## Profiling Configurations

### CPU Profiling Configurations

#### 1. **Profile: CPU (macOS)**
- **Purpose**: Run application with profiling symbols for macOS Instruments
- **Build**: Uses `build-profile` task (RelWithDebInfo)
- **Arguments**: `--max-fps 10 --verbose`
- **Best for**: Identifying CPU hotspots and performance bottlenecks
- **Usage**:
  1. Select this configuration from Debug panel
  2. Press F5 to launch
  3. Use external profiling tools (Instruments, dtrace) to attach
  4. Analyze with: `./scripts/profile_cpu_mac.sh`

#### 2. **Profile: CPU (Linux)**
- **Purpose**: Run application with profiling symbols for perf/gprof
- **Build**: Uses `build-profile` task (RelWithDebInfo)
- **Arguments**: `--max-fps 10 --verbose`
- **Best for**: Identifying CPU hotspots on Linux
- **Usage**:
  1. Select this configuration from Debug panel
  2. Press F5 to launch
  3. Use external profiling tools (perf, gprof) to attach
  4. Analyze with: `./scripts/profile_cpu_linux.sh`

### Memory Profiling Configurations

#### 3. **Profile: Memory - Valgrind (Linux)**
- **Purpose**: Run application under Valgrind for memory leak detection
- **Tool**: Valgrind memcheck
- **Arguments**: `--max-fps 3 --verbose` (reduced FPS due to Valgrind overhead)
- **Output**: `profiling_results/valgrind_vscode.log`
- **Best for**: Finding memory leaks and invalid memory access
- **Note**: Expect 10-50x slowdown
- **Usage**:
  1. Select this configuration from Debug panel
  2. Press F5 to launch (will be slow)
  3. Let run for a few minutes
  4. Stop and review `profiling_results/valgrind_vscode.log`

### Performance Testing Configuration

#### 4. **Profile: Performance Test**
- **Purpose**: Test high-load scenario with burst mode
- **Arguments**: `--max-fps 15 --enable-burst-mode --verbose`
- **Best for**: Stress testing and identifying bottlenecks under load
- **Usage**:
  1. Select this configuration
  2. Press F5 to launch
  3. Trigger object detection to activate burst mode
  4. Monitor FPS and processing times in output

## Profiling Tasks

VSCode tasks available in the Tasks menu (Terminal → Run Task):

### Build Tasks

- **build-profile** - Build with RelWithDebInfo (optimized with debug symbols)
- **create-profile-build-dir** - Create build-profile directory
- **build-profile-make** - Compile profiling build with make

### Profiling Tasks

- **profile-cpu-mac** - Run CPU profiling on macOS (60s)
- **profile-cpu-linux** - Run CPU profiling on Linux (60s)
- **profile-memory-mac** - Run memory profiling on macOS (60s)
- **profile-memory-linux** - Run memory profiling on Linux (60s)
- **clean-profiling-results** - Delete profiling_results directory

### Running Tasks

1. Press `Cmd+Shift+P` (macOS) or `Ctrl+Shift+P` (Linux)
2. Type "Tasks: Run Task"
3. Select the desired profiling task
4. Results will be in `profiling_results/` directory

## Profiling Workflow

### Quick Start: CPU Profiling

**macOS:**
```bash
# 1. Build for profiling (or use VSCode task)
./scripts/build_profile_mac.sh

# 2. Run CPU profiling (or use VSCode task)
./scripts/profile_cpu_mac.sh --duration 60

# 3. Open results in Instruments
open profiling_results/*_cpu_profile*.trace
```

**Linux:**
```bash
# 1. Build for profiling (or use VSCode task)
./scripts/build_profile_linux.sh

# 2. Run CPU profiling (or use VSCode task)
./scripts/profile_cpu_linux.sh --duration 60

# 3. Review results
cat profiling_results/*_summary.txt
less profiling_results/*_report.txt
```

### Quick Start: Memory Profiling

**macOS:**
```bash
# Run memory profiling
./scripts/profile_memory_mac.sh --duration 60 --type allocations

# Open results in Instruments
open profiling_results/*_memory_profile*.trace
```

**Linux:**
```bash
# Run Valgrind memory profiling
./scripts/profile_memory_linux.sh --duration 60 --tool valgrind

# Review leak summary
cat profiling_results/*_summary.txt

# Or use heaptrack for faster profiling
./scripts/profile_memory_linux.sh --duration 60 --tool heaptrack
```

## Analyzing Profiling Results

### CPU Profiling

**Key metrics to look for:**
1. **Hot functions** - Functions consuming >5% CPU time
2. **Call stacks** - Where hot functions are called from
3. **Self time vs. total time** - Direct computation vs. called functions

**Common bottlenecks:**
- `cv::dnn::Net::forward()` - YOLO inference (expected to be high)
- `cv::resize()` - Image scaling (optimize with `--detection-scale`)
- Frame queue operations - Consider parallel processing settings
- I/O operations - Disk writes, network streaming

### Memory Profiling

**Key metrics to look for:**
1. **Memory leaks** - Allocations never freed
2. **Peak memory usage** - Maximum heap size
3. **Allocation rate** - Frequency of allocations
4. **Large allocations** - Big buffers or images

**Common issues:**
- Growing heap - Memory leak or unbounded frame queue
- High peak - Large image buffers or model weights
- Frequent allocations - Inefficient object reuse

### Interpreting Results

See `docs/PROFILING.md` for detailed analysis guide including:
- How to read flame graphs
- Optimization strategies
- Before/after comparison techniques
- Platform-specific profiling tips

## Debugging the macOS Issue

Based on the reported issue where nothing happens when running the executable on macOS, here's a debugging strategy:

### Step 1: Check if main() is reached
1. Open `src/main.cpp`
2. Set a breakpoint on line 24 (first line of `main()`)
3. Run **(lldb) Launch - macOS Debug**
4. If breakpoint is NOT hit:
   - The issue is before main() (static initialization, library loading)
   - Check the Debug Console for errors
5. If breakpoint IS hit:
   - Continue to Step 2

### Step 2: Check argument parsing
1. Set breakpoints at:
   - Line 35: `ConfigManager config_manager;`
   - Line 36: `auto parse_result = config_manager.parseArgs(argc, argv);`
   - Line 38: `if (parse_result == ConfigManager::ParseResult::HELP_REQUESTED...`
2. Step through with `F10`
3. Inspect variables:
   - Check `argc` and `argv` values
   - Check `parse_result` value
   - See if execution reaches line 38

### Step 3: Check output flushing
1. Set breakpoint at line 41: `std::cout.flush();`
2. Check if this line is reached
3. If reached but no output appears:
   - Check Debug Console output
   - The issue might be with terminal/console redirection

### Step 4: Check for blocking operations
1. In `src/config_manager.cpp`, set breakpoints in:
   - `parseArgs()` function
   - `printUsage()` function
2. Step through to see if any operation hangs
3. Check for:
   - File I/O operations that might block
   - Library calls that might hang on macOS

### Step 5: Check for exceptions
1. Enable "All Exceptions" in VSCode:
   - In Debug view, go to Breakpoints panel
   - Check "All Exceptions"
2. Run debug session
3. See if any exception is thrown before output

## Useful Debug Commands

While debugging in the Debug Console, you can use LLDB commands:

```lldb
# Print variable value
p argc
p argv[0]
p parse_result

# Print with formatting
p/x some_int_variable  # hexadecimal
p/s some_string        # as string

# Check call stack
bt

# Continue execution
c

# Step commands
n  # next (step over)
s  # step (step into)
finish  # step out

# Breakpoint commands
br list  # list breakpoints
br del 1 # delete breakpoint 1
```

## Tips for macOS Debugging

1. **Check library dependencies**: Run `otool -L build/object_detection` to see if all libraries are found
2. **Check for signing issues**: macOS Gatekeeper might block unsigned executables
3. **Run with verbose output**: Add `-v` or `--verbose` flags
4. **Check Console.app**: Open macOS Console.app to see system logs for crashes

## Building for Debug

The debug configurations automatically build with debug symbols. To manually build:

```bash
cd cpp-object-detection
mkdir -p build
cd build
cmake .. -DCMAKE_BUILD_TYPE=Debug -DCMAKE_EXPORT_COMPILE_COMMANDS=ON
make -j$(sysctl -n hw.ncpu)
```

## Troubleshooting

### "Program not found" error
- Ensure you've built the project first
- Check that `build/object_detection` exists

### CodeLLDB extension not working
- Update to latest version
- Try uninstalling and reinstalling
- Check VSCode Output panel for errors

### Breakpoints not being hit
- Verify you're building with Debug configuration (not Release)
- Check that symbols are included: `dsymutil build/object_detection`
- Rebuild the project

### Cannot attach to process
- Ensure the process is running
- Check you have permissions (may need `sudo` on some systems)

## Additional Resources

- [LLDB Tutorial](https://lldb.llvm.org/use/tutorial.html)
- [VSCode C++ Debugging](https://code.visualstudio.com/docs/cpp/cpp-debug)
- [CodeLLDB Documentation](https://github.com/vadimcn/vscode-lldb/blob/master/MANUAL.md)
