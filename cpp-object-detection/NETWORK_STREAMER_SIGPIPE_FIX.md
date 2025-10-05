# Network Streamer SIGPIPE Fix

## Problem

The object detection program was crashing/exiting when clients connected to the network video stream. This occurred because:

1. When a client connects and then disconnects (either gracefully or abruptly)
2. The server attempts to send data via `send()` system call
3. If the client has closed the connection, `send()` generates a **SIGPIPE** signal
4. By default, SIGPIPE terminates the program

## Root Cause

All `send()` calls in `network_streamer.cpp` were using `0` as the flags parameter:

```cpp
send(client_socket, data, size, 0)  // Can trigger SIGPIPE
```

When writing to a socket whose peer has closed the connection, the kernel sends a SIGPIPE signal to the process. Without proper signal handling, this terminates the entire application.

## Solution

Added `MSG_NOSIGNAL` flag to all `send()` calls:

```cpp
send(client_socket, data, size, MSG_NOSIGNAL)  // Returns error instead of SIGPIPE
```

The `MSG_NOSIGNAL` flag prevents the SIGPIPE signal from being raised. Instead, `send()` returns an error (EPIPE), which the code already handles gracefully by logging and breaking out of the send loop.

## Changes Made

### Modified Files

1. **cpp-object-detection/src/network_streamer.cpp**
   - Updated 4 `send()` calls to use `MSG_NOSIGNAL` flag:
     - Line 206: HTTP headers send
     - Line 243: Frame header send
     - Line 249: JPEG data send
     - Line 256: Boundary send

2. **cpp-object-detection/tests/test_network_streamer.cpp**
   - Added socket includes for testing
   - Added `ClientConnectionDoesNotCrash` test case
   - Test simulates client connecting and disconnecting abruptly
   - Verifies server continues running after disconnect

## Testing

### Unit Tests
- All existing tests pass (103/104)
- New integration test validates fix:
  - Creates network streamer
  - Connects a client
  - Immediately disconnects (simulating abrupt disconnect)
  - Verifies server remains running
  - Test passes consistently

### Pre-existing Test Failure
- `NetworkStreamerTest.MultipleStartStopCycles` - unrelated to this fix
- Failure is due to re-initialization issue, not signal handling

## Verification

Build and test:
```bash
cd cpp-object-detection
./scripts/build.sh
./scripts/test.sh
```

All tests pass except the one pre-existing failure.

## Impact

This fix ensures that:
- ✅ Network streaming server remains stable when clients connect/disconnect
- ✅ Object detection program continues running during network activity
- ✅ Graceful handling of abrupt client disconnections
- ✅ No crashes from SIGPIPE signals
- ✅ Existing error handling remains effective

## Technical Details

### MSG_NOSIGNAL Flag

From `man 2 send`:
```
MSG_NOSIGNAL
    Requests not to send SIGPIPE on errors on stream oriented
    sockets when the other end breaks the connection. The EPIPE
    error is still returned.
```

This is a standard POSIX flag available on Linux and provides a per-call mechanism to avoid SIGPIPE without requiring process-wide signal handling changes.

### Alternative Approaches Considered

1. **Signal Handler** - Could ignore SIGPIPE globally via `signal(SIGPIPE, SIG_IGN)`
   - Rejected: Affects entire process, may mask other issues
   
2. **SO_NOSIGPIPE Socket Option** - Set socket option to prevent SIGPIPE
   - Rejected: Not portable (BSD-specific), MSG_NOSIGNAL is more standard

3. **MSG_NOSIGNAL Flag** - ✅ Selected
   - Most portable and targeted solution
   - Only affects specific send operations
   - Standard on Linux systems

## Deployment Notes

- No configuration changes required
- Fix is backward compatible
- Works on all Linux distributions
- No performance impact
- Minimal code changes (4 lines modified, 1 test added)

## References

- Issue: "cpp-object-detection: connecting to network video stream can exit/crash detection program"
- Related Documentation:
  - `IMPLEMENTATION_SUMMARY.md` - Network Streaming Feature
  - `NETWORK_STREAM_STATS_INTEGRATION.md` - Stats overlay feature
