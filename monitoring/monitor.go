package monitoring

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/admin"
)

const (
	// DebugLogFile is the file where system monitoring debug info is stored
	DebugLogFile = "system-debug.log"
	// MonitoringInterval is how often to log system stats (5 minutes)
	MonitoringInterval = 5 * time.Minute
)

// SystemStats contains system monitoring information
type SystemStats struct {
	Timestamp      string `json:"timestamp"`
	Uptime         string `json:"uptime"`
	CpuTemperature string `json:"cpu_temperature"`
	GpuTemperature string `json:"gpu_temperature"`
	FreeDiskSpace  string `json:"free_disk_space"`
	SystemClock    string `json:"system_clock"`
	RuntimeSeconds int64  `json:"runtime_seconds"`
}

// Monitor handles periodic system monitoring and logging
type Monitor struct {
	startTime       time.Time
	lastLogTime     time.Time
	debugLogFile    *os.File
	dailyStartTime  time.Time
	maxDailyRuntime int64
}

// New creates a new system monitor
func New() (*Monitor, error) {
	f, err := os.OpenFile(DebugLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	m := &Monitor{
		startTime:      time.Now(),
		lastLogTime:    time.Now(),
		debugLogFile:   f,
		dailyStartTime: time.Now(),
	}

	// Log startup event
	m.logEvent("STARTUP", "System started")
	return m, nil
}

// Close closes the debug log file
func (m *Monitor) Close() error {
	m.logEvent("SHUTDOWN", fmt.Sprintf("System shutdown after %d seconds runtime", m.getRuntimeSeconds()))
	if m.debugLogFile != nil {
		return m.debugLogFile.Close()
	}
	return nil
}

// PeriodicCheck performs periodic system monitoring and logs the information
func (m *Monitor) PeriodicCheck() {
	now := time.Now()

	// Check if we've moved to a new day to reset daily runtime tracking
	if now.Day() != m.dailyStartTime.Day() {
		m.logEvent("DAILY_RUNTIME", fmt.Sprintf("Max runtime for previous day: %d seconds", m.maxDailyRuntime))
		m.dailyStartTime = now
		m.maxDailyRuntime = 0
	}

	// Update max daily runtime
	currentRuntime := m.getRuntimeSeconds()
	if currentRuntime > m.maxDailyRuntime {
		m.maxDailyRuntime = currentRuntime
	}

	// Only log if the monitoring interval has elapsed
	if now.Sub(m.lastLogTime) < MonitoringInterval {
		return
	}

	stats := m.collectStats()
	m.logStats(stats)
	m.lastLogTime = now
}

// collectStats gathers current system statistics
func (m *Monitor) collectStats() SystemStats {
	return SystemStats{
		Timestamp:      time.Now().Format(time.RFC3339),
		Uptime:         admin.RunCommandOrPanic("/usr/bin/uptime"),
		CpuTemperature: admin.RunCommandOrPanic("/bin/cat", "/sys/class/thermal/thermal_zone0/temp"),
		GpuTemperature: admin.RunCommandOrPanic("/opt/vc/bin/vcgencmd", "measure_temp"),
		FreeDiskSpace:  admin.RunCommandOrPanic("/bin/df", "-h"),
		SystemClock:    admin.RunCommandOrPanic("/bin/date", "+%Y-%m-%d %H:%M:%S %Z"),
		RuntimeSeconds: m.getRuntimeSeconds(),
	}
}

// getRuntimeSeconds returns how long the system has been running in seconds
func (m *Monitor) getRuntimeSeconds() int64 {
	return int64(time.Since(m.startTime).Seconds())
}

// logStats writes system statistics to the debug log
func (m *Monitor) logStats(stats SystemStats) {
	jsonData, err := json.Marshal(stats)
	if err != nil {
		log.Printf("Error marshaling system stats: %s\n", err)
		return
	}

	logLine := fmt.Sprintf("[STATS] %s\n", string(jsonData))
	if _, err := m.debugLogFile.WriteString(logLine); err != nil {
		log.Printf("Error writing to debug log: %s\n", err)
	}

	// Also log to main log for visibility
	log.Printf("System stats recorded - Runtime: %ds, CPU temp: %s, Disk: %s",
		stats.RuntimeSeconds,
		stats.CpuTemperature,
		stats.FreeDiskSpace)
}

// logEvent writes an event to the debug log
func (m *Monitor) logEvent(eventType, message string) {
	timestamp := time.Now().Format(time.RFC3339)
	logLine := fmt.Sprintf("[%s] %s - %s\n", eventType, timestamp, message)

	if m.debugLogFile != nil {
		if _, err := m.debugLogFile.WriteString(logLine); err != nil {
			log.Printf("Error writing event to debug log: %s\n", err)
		}
	}

	log.Printf("[%s] %s", eventType, message)
}
