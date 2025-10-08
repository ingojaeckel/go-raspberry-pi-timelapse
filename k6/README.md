# k6 Load Testing

This directory contains k6 load test scripts for the Go Raspberry Pi Timelapse API.

## Prerequisites

1. **Install k6**: Follow the [k6 installation guide](https://k6.io/docs/get-started/installation/)
   - macOS: `brew install k6`
   - Linux: `sudo gpg -k && sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69 && echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list && sudo apt-get update && sudo apt-get install k6`
   - Windows: `choco install k6` or download from [k6 releases](https://github.com/grafana/k6/releases)

2. **Build and run the Go application** with profiling enabled (see below)

## Running Load Tests

### Quick Start

```bash
# From the repository root, run the load test script
./scripts/run-load-test.sh
```

This script will:
1. Start the Go application with profiling enabled
2. Run the k6 load test
3. Display results
4. Optionally collect and save CPU/memory profiles

### Manual Execution

1. **Start the Go application with profiling enabled**:
   ```bash
   # From repository root
   go run -ldflags="-X 'main.gitCommit=local' -X 'main.builtAt=$(date)'" . -pprof
   ```

2. **Run the load test**:
   ```bash
   # Basic load test
   k6 run k6/load-test.js
   
   # Test against a different URL
   K6_BASE_URL=http://192.168.50.1:8080 k6 run k6/load-test.js
   
   # Run with custom duration
   k6 run --duration 5m k6/load-test.js
   ```

## Success Criteria

The load test enforces the following success criteria:

1. **100% Success Rate**: Less than 1% failed requests (HTTP 2xx responses)
2. **P95 Latency < 100ms**: 95th percentile response time under 100ms
3. **Additional Monitoring**:
   - P50 (median) < 50ms
   - P99 < 200ms

If any threshold is violated, k6 will exit with a non-zero status code.

## Profiling the Go Application

### Enabling Profiling

The Go application supports CPU and memory profiling via the standard `net/http/pprof` package when started with the `-pprof` flag.

### Collecting Profiles During Load Test

1. **CPU Profile** (30 seconds):
   ```bash
   curl http://localhost:8080/debug/pprof/profile?seconds=30 -o cpu.prof
   ```

2. **Memory Heap Profile**:
   ```bash
   curl http://localhost:8080/debug/pprof/heap -o mem.prof
   ```

3. **Goroutine Profile**:
   ```bash
   curl http://localhost:8080/debug/pprof/goroutine -o goroutine.prof
   ```

### Analyzing Profiles

Use the Go `pprof` tool to analyze collected profiles:

```bash
# Interactive analysis
go tool pprof cpu.prof

# Generate a visualization (requires graphviz)
go tool pprof -http=:8081 cpu.prof

# Top functions by CPU usage
go tool pprof -top cpu.prof

# Memory allocation analysis
go tool pprof -alloc_space mem.prof
```

### Common pprof Commands

Once in the interactive pprof shell:

- `top`: Show top functions by resource usage
- `list <function>`: Show source code of a function
- `web`: Open graph visualization in browser (requires graphviz)
- `pdf`: Generate PDF graph (requires graphviz)
- `help`: Show all available commands

## Load Test Configuration

The `load-test.js` script tests the following read-only endpoints:

- `/version` - Version information
- `/monitoring` - System monitoring data (CPU, GPU temp, disk space, etc.)
- `/photos` - List of timelapse photos
- `/configuration` - Current configuration
- `/logs` - Application logs

### Test Stages

1. **Ramp-up (30s)**: 0 → 10 virtual users
2. **Steady (1m)**: 10 virtual users
3. **Scale-up (30s)**: 10 → 50 virtual users
4. **Peak load (2m)**: 50 virtual users
5. **Ramp-down (30s)**: 50 → 0 virtual users

**Total duration**: ~4.5 minutes

## Interpreting Results

After the test completes, k6 will display:

```
checks.........................: 100.00% ✓ 2450      ✗ 0     
http_req_duration..............: avg=25ms    p(95)=45ms  p(99)=85ms
http_req_failed................: 0.00%   ✓ 0         ✗ 2450
```

Key metrics to monitor:

- `http_req_duration`: Response time percentiles
- `http_req_failed`: Failed request rate
- `checks`: Validation check pass rate
- `iterations`: Total number of test iterations
- `vus`: Virtual users (concurrent load)

## Troubleshooting

### High Error Rate

If you see a high error rate:
1. Check that the Go application is running
2. Verify the base URL is correct
3. Check application logs for errors
4. Reduce virtual users to isolate the issue

### High Latency

If P95 latency exceeds 100ms:
1. Collect CPU and memory profiles during the test
2. Analyze profiles to find bottlenecks
3. Check system resources (CPU, memory, disk I/O)
4. Consider optimizing hot code paths

### Connection Refused

If k6 can't connect:
1. Ensure the Go app is running: `curl http://localhost:8080/version`
2. Check the listen address matches the test URL
3. Verify no firewall is blocking the connection

## Advanced Usage

### Custom Test Scenarios

Edit `load-test.js` to:
- Add more endpoints
- Adjust load patterns
- Modify success criteria
- Add custom metrics

### Continuous Load Testing

For long-running tests:

```bash
# 1 hour continuous test with 25 users
k6 run --vus 25 --duration 1h k6/load-test.js
```

### Results Export

Export results for analysis:

```bash
# JSON output
k6 run --out json=results.json k6/load-test.js

# CSV output (requires xk6-output-prometheus-remote or similar)
k6 run --out csv=results.csv k6/load-test.js
```

## References

- [k6 Documentation](https://k6.io/docs/)
- [Go pprof Documentation](https://pkg.go.dev/net/http/pprof)
- [Profiling Go Programs](https://go.dev/blog/pprof)
