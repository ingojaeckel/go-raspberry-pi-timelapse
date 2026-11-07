package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/health"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/logfile"
)

const maxLogBytes = 256 << 10

// RunCommand runs a system command and returns its combined output. It is
// the production implementation of Server.Run.
func RunCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = 2 * time.Second
	return cmd.CombinedOutput()
}

// monitoring reports system status. Each field holds the command output, or
// the error if it failed.
func (s *Server) monitoring(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	writeJSON(w, http.StatusOK, MonitoringResponse{
		Time:           s.output(ctx, "/bin/date"),
		GpuTemperature: s.output(ctx, health.VcgencmdPath(), "measure_temp"),
		CpuTemperature: readOrError(s.ThermalFile),
		Uptime:         s.output(ctx, "/usr/bin/uptime"),
		FreeDiskSpace:  s.output(ctx, "/bin/df", "-h"),
		UnixMillis:     time.Now().UnixMilli(),
		Power:          s.power(ctx),
		CPU:            s.cpu(),
		Storage:        s.storage(),
	})
}

func (s *Server) power(ctx context.Context) PowerStatus {
	t, err := health.ReadThrottled(ctx, s.Run)
	if err != nil {
		return PowerStatus{Now: []string{}, SinceBoot: []string{}, Error: err.Error()}
	}
	return PowerStatus{Now: t.Now(), SinceBoot: t.SinceBoot()}
}

func (s *Server) cpu() CPUStatus {
	c, err := health.ReadCPUFreq(s.CPUFreqDir)
	if err != nil {
		return CPUStatus{Error: err.Error()}
	}
	return CPUStatus{CPUFreq: c}
}

func (s *Server) storage() StorageStatus {
	st := StorageStatus{ReserveBytes: s.ReserveBytes, StopWhenDiskFull: s.Settings.Get().StopWhenDiskFull}
	if s.DeletedOldest != nil {
		st.DeletedOldest = s.DeletedOldest()
	}
	list, err := s.Photos.List()
	if err != nil {
		st.Error = err.Error()
		return st
	}
	st.Photos = len(list)
	for _, f := range list {
		st.PhotoBytes += f.Bytes
	}
	if st.FreeBytes, err = s.FreeBytes(s.Photos.Dir()); err != nil {
		st.Error = err.Error()
	}
	return st
}

// Clocks outside this range are rejected: a browser can be wrong too.
var (
	minClock = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	maxClock = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
)

// setTime sets the camera's clock, which a Pi without a clock module (and
// no network time) loses at every power cut, and then the clock module if
// there is one.
func (s *Server) setTime(w http.ResponseWriter, r *http.Request) {
	var req SetTimeRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	t := time.UnixMilli(req.UnixMillis)
	if t.Before(minClock) || t.After(maxClock) {
		writeError(w, http.StatusBadRequest, fmt.Errorf("UnixMillis must be a time between %d and %d, got %v", minClock.Year(), maxClock.Year(), t.UTC()))
		return
	}
	was := time.Now()
	if err := s.SetClock(t); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("setting the clock: %w", err))
		return
	}
	slog.Warn("clock set from the web UI", "was", was.Round(0), "now", t, "change", t.Sub(was.Round(0)).Round(time.Second))
	res := SetTimeResponse{UnixMillis: time.Now().UnixMilli()}
	if _, err := os.Stat(s.RTCDevice); err == nil { // never for ""
		res.RTC = "updated"
		if out, err := s.Run(r.Context(), "/usr/bin/sudo", "-n", "/usr/sbin/hwclock", "--rtc="+s.RTCDevice, "--systohc", "--utc"); err != nil {
			res.RTC = fmt.Sprintf("not updated: %v %s", err, out)
			slog.Warn("updating the clock module", "err", err, "output", string(out))
		}
	}
	writeJSON(w, http.StatusOK, res)
}

// output returns the output of a command, or its error message.
func (s *Server) output(ctx context.Context, name string, args ...string) string {
	out, err := s.Run(ctx, name, args...)
	if err != nil {
		return err.Error()
	}
	return string(out)
}

// readOrError returns the content of a file, or the error message.
func readOrError(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (s *Server) logs(w http.ResponseWriter, _ *http.Request) {
	if s.LogFile == "" {
		writeError(w, http.StatusInternalServerError, errors.New("no log file configured"))
		return
	}
	text, err := logfile.Tail(s.LogFile, maxLogBytes)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, LogResponse{Logs: text})
}

// command returns a handler that runs a fixed system command.
func (s *Server) command(name string, args ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("running system command", "command", name, "args", args)
		if out, err := s.Run(r.Context(), name, args...); err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf("%w: %s", err, out))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
