# Go Performance Benchmarks

This document describes the micro-benchmarks created for performance-sensitive parts of the Go implementation.

## Running Benchmarks

### Run All Benchmarks

To run all benchmarks in the project:

```bash
go test -bench=. -benchmem ./...
```

### Run Benchmarks for Specific Packages

Run timelapse package benchmarks:
```bash
go test -bench=. -benchmem ./timelapse/...
```

Run files package benchmarks:
```bash
go test -bench=. -benchmem ./files/...
```

Run rest package benchmarks:
```bash
go test -bench=. -benchmem ./rest/...
```

### Run Specific Benchmarks

Run only the getSecondsToFirstCapture benchmarks:
```bash
go test -bench=BenchmarkGetSecondsToFirstCapture -benchmem ./timelapse/...
```

## Benchmark Coverage

### Timelapse Package (`timelapse/timelapse_test.go`)

1. **BenchmarkGetSecondsToFirstCapture**
   - Tests the performance of calculating sleep time until next capture
   - Measures single time calculation
   - **Typical Result:** ~15 ns/op, 0 B/op, 0 allocs/op

2. **BenchmarkGetSecondsToFirstCaptureVariousTimes**
   - Tests the performance across multiple different time scenarios
   - Measures calculation with various times in an hour
   - **Typical Result:** ~62 ns/op, 0 B/op, 0 allocs/op

3. **BenchmarkGetSecondsToFirstCaptureShortInterval**
   - Tests performance with short capture intervals (60 seconds)
   - Simulates high-frequency capture scenarios
   - **Typical Result:** ~75 ns/op, 0 B/op, 0 allocs/op

### Files Package (`files/files_test.go`)

1. **BenchmarkListFiles**
   - Tests the performance of listing files in a directory
   - Includes directory reading and file metadata collection
   - **Typical Result:** ~15 μs/op, 1297 B/op, 16 allocs/op

2. **BenchmarkByAgeSort**
   - Tests the performance of sorting 100 files by age
   - Measures actual sorting performance using sort.Sort()
   - **Typical Result:** ~1.4 μs/op, 6168 B/op, 2 allocs/op

3. **BenchmarkByAgeSortFull**
   - Tests sorting performance with 1000 files to verify scalability
   - Simulates large timelapse photo collections
   - **Typical Result:** ~13 μs/op, 57368 B/op, 2 allocs/op

### Rest Package (`rest/template_test.go`)

1. **BenchmarkGetHumanReadableSize**
   - Tests the performance of formatting file sizes
   - Measures conversion across bytes, KB, and MB ranges
   - **Typical Result:** ~532 ns/op, 64 B/op, 8 allocs/op

## Interpreting Results

Benchmark output format:
```
BenchmarkName-4    1000000    1234 ns/op    56 B/op    3 allocs/op
```

- **1000000**: Number of iterations run
- **1234 ns/op**: Average time per operation (nanoseconds)
- **56 B/op**: Average bytes allocated per operation
- **3 allocs/op**: Average number of allocations per operation

## Advanced Usage

### Compare Performance Between Changes

Save baseline benchmark:
```bash
go test -bench=. -benchmem ./... > old.txt
```

After making changes, run benchmarks again:
```bash
go test -bench=. -benchmem ./... > new.txt
```

Compare using benchcmp (requires installation):
```bash
go get golang.org/x/tools/cmd/benchcmp
benchcmp old.txt new.txt
```

### CPU Profiling

Generate CPU profile while benchmarking:
```bash
go test -bench=BenchmarkGetSecondsToFirstCapture -cpuprofile=cpu.prof ./timelapse/...
go tool pprof cpu.prof
```

### Memory Profiling

Generate memory profile:
```bash
go test -bench=BenchmarkListFiles -memprofile=mem.prof ./files/...
go tool pprof mem.prof
```

## Performance Goals

These benchmarks help ensure that:

1. **Time Calculations** remain fast (< 100 ns/op) for timelapse scheduling
2. **File Operations** are efficient even with many photos (< 20 μs for listing)
3. **Sorting Operations** scale well (< 2 μs for 100 items, < 15 μs for 1000 items)
4. **Size Formatting** is efficient for UI rendering (< 1 μs/op)

## Continuous Performance Monitoring

Consider running benchmarks as part of CI/CD to catch performance regressions:

```bash
# In CI pipeline
go test -bench=. -benchmem ./... | tee benchmark_results.txt
```

## Notes

- Benchmarks run on the current system; results may vary on Raspberry Pi hardware
- For production performance on Raspberry Pi, test on actual hardware
- Memory allocations are important on resource-constrained devices
- Some benchmarks may show variance; run multiple times for statistical significance
