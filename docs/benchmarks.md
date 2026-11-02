# Go Performance Benchmarks

This document describes the micro-benchmarks created for performance-sensitive parts of the Go implementation.

## Running Benchmarks

Run these from `src/backend/` (the Go module root). The frontend must have been built first (`make frontend` from the repo root), because the main package embeds it.

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

### Run Specific Benchmarks

Run only the scheduling benchmark:
```bash
go test -run='^$' -bench=BenchmarkNextAligned -benchmem ./timelapse/...
```

## Benchmark Coverage

Benchmarks use the `for b.Loop()` form (Go 1.24+).

### Timelapse Package (`timelapse/schedule_test.go`)

1. **BenchmarkNextAligned**
   - Time of the next capture aligned to the hour (runs at least once per minute while waiting)
   - **Typical Result:** ~60 ns/op, 0 B/op, 0 allocs/op (Apple M-series)

### Files Package (`files/files_test.go`)

1. **BenchmarkListFiles**
   - Lists and sorts a folder of 1000 photos (directory read, metadata, `slices.SortFunc`)
   - Used by `/photos`, `/file/last`, the archives and disk space reclaiming
   - **Typical Result:** ~3 ms/op, 600 KB/op, 6000 allocs/op (Apple M-series)

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

Compare using benchstat:
```bash
go install golang.org/x/perf/cmd/benchstat@latest
benchstat old.txt new.txt
```

### CPU Profiling

Generate CPU profile while benchmarking:
```bash
go test -run='^$' -bench=BenchmarkNextAligned -cpuprofile=cpu.prof ./timelapse/...
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
2. **File Operations** stay fast with many photos (a few ms for 1000 photos on a laptop; expect 10-20x slower on a Pi Zero)

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
