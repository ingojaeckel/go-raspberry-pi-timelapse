package rest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/health"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/logfile"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
)

const maxLogBytes = 256 << 10

// RunCommand runs a system command and returns its combined output. It is
// the production implementation of Server.Run.
var RunCommand = timelapse.CommandOutput

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
	st.SampleBytesPerPixel, st.SampleQuality, st.SamplePhotos = samplePhotos(s.Photos.Dir(), list)
	if st.FreeBytes, err = s.FreeBytes(s.Photos.Dir()); err != nil {
		st.Error = err.Error()
	}
	return st
}

// sampleSize is how many of the newest photos samplePhotos reads.
const sampleSize = 20

// samplePhotos measures the size on disk per pixel of the newest photos that
// share the newest readable photo's JPEG quality. list is oldest first.
// Only the headers are read.
func samplePhotos(dir string, list []files.File) (bytesPerPixel float64, quality, n int) {
	var bytes, pixels int64
	for i := len(list) - 1; i >= 0 && i >= len(list)-sampleSize; i-- {
		f, err := os.Open(filepath.Join(dir, list[i].Name))
		if err != nil {
			continue
		}
		info, err := files.ReadJPEGInfo(f)
		f.Close()
		if err != nil || quality != 0 && info.Quality != quality {
			continue
		}
		quality = info.Quality
		bytes += list[i].Bytes
		pixels += int64(info.Width) * int64(info.Height)
		n++
	}
	if pixels == 0 {
		return 0, 0, 0
	}
	return float64(bytes) / float64(pixels), quality, n
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
	if !readJSON(w, r, 1<<10, &req) {
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
	writeJSON(w, http.StatusOK, SetTimeResponse{UnixMillis: time.Now().UnixMilli(), RTC: s.updateRTC(r.Context())})
}

// updateRTC copies the system clock to the clock module, if there is one,
// and says how that went ("" without a clock module).
func (s *Server) updateRTC(ctx context.Context) string {
	if _, err := os.Stat(s.RTCDevice); err != nil { // always for ""
		return ""
	}
	out, err := s.Run(ctx, "/usr/bin/sudo", "-n", "/usr/sbin/hwclock", "--rtc="+s.RTCDevice, "--systohc", "--utc")
	if err != nil {
		slog.Warn("updating the clock module", "err", err, "output", string(out))
		return fmt.Sprintf("not updated: %v %s", err, out)
	}
	return "updated"
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
