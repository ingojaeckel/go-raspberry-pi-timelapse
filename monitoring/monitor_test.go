package monitoring

import (
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// Clean up any existing debug log
	os.Remove(DebugLogFile)

	m, err := New()
	if err != nil {
		t.Fatalf("Failed to create monitor: %s", err)
	}
	defer m.Close()

	if m.startTime.IsZero() {
		t.Error("Start time should be set")
	}

	if m.debugLogFile == nil {
		t.Error("Debug log file should be initialized")
	}

	// Verify the debug log file was created
	if _, err := os.Stat(DebugLogFile); os.IsNotExist(err) {
		t.Error("Debug log file should be created")
	}

	// Clean up
	os.Remove(DebugLogFile)
}

func TestGetRuntimeSeconds(t *testing.T) {
	os.Remove(DebugLogFile)

	m, err := New()
	if err != nil {
		t.Fatalf("Failed to create monitor: %s", err)
	}
	defer m.Close()
	defer os.Remove(DebugLogFile)

	// Sleep for a bit to accumulate runtime
	time.Sleep(100 * time.Millisecond)

	runtime := m.getRuntimeSeconds()
	if runtime < 0 {
		t.Errorf("Runtime should be positive, got %d", runtime)
	}
}

func TestPeriodicCheckIntervalRespected(t *testing.T) {
	os.Remove(DebugLogFile)

	m, err := New()
	if err != nil {
		t.Fatalf("Failed to create monitor: %s", err)
	}
	defer m.Close()
	defer os.Remove(DebugLogFile)

	// First check should not log (interval not elapsed)
	initialLogTime := m.lastLogTime
	m.PeriodicCheck()

	// Last log time should not have changed since interval hasn't elapsed
	if m.lastLogTime != initialLogTime {
		t.Error("Last log time should not change if interval hasn't elapsed")
	}
}

func TestDailyRuntimeTracking(t *testing.T) {
	os.Remove(DebugLogFile)

	m, err := New()
	if err != nil {
		t.Fatalf("Failed to create monitor: %s", err)
	}
	defer m.Close()
	defer os.Remove(DebugLogFile)

	// Simulate some runtime
	m.startTime = time.Now().Add(-1 * time.Hour)

	// First periodic check
	m.PeriodicCheck()

	runtime := m.getRuntimeSeconds()
	if m.maxDailyRuntime < runtime {
		t.Errorf("Max daily runtime should be updated, got %d, expected at least %d", m.maxDailyRuntime, runtime)
	}
}

func TestCollectStats(t *testing.T) {
	os.Remove(DebugLogFile)

	m, err := New()
	if err != nil {
		t.Fatalf("Failed to create monitor: %s", err)
	}
	defer m.Close()
	defer os.Remove(DebugLogFile)

	stats := m.collectStats()

	if stats.Timestamp == "" {
		t.Error("Timestamp should be set")
	}

	if stats.RuntimeSeconds < 0 {
		t.Errorf("Runtime seconds should be non-negative, got %d", stats.RuntimeSeconds)
	}

	// Note: Other fields may be empty or contain error messages if commands fail
	// This is expected in a test environment
}
