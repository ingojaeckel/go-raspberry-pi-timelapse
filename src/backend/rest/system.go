package rest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"time"

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

// vcgencmdPath finds vcgencmd: /usr/bin on current Raspberry Pi OS,
// /opt/vc/bin on older releases.
func vcgencmdPath() string {
	if p, err := exec.LookPath("vcgencmd"); err == nil {
		return p
	}
	return "/opt/vc/bin/vcgencmd"
}

// monitoring reports system status. Each field holds the command output, or
// the error if it failed.
func (s *Server) monitoring(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	writeJSON(w, http.StatusOK, MonitoringResponse{
		Time:           s.output(ctx, "/bin/date"),
		GpuTemperature: s.output(ctx, vcgencmdPath(), "measure_temp"),
		CpuTemperature: readOrError(s.ThermalFile),
		Uptime:         s.output(ctx, "/usr/bin/uptime"),
		FreeDiskSpace:  s.output(ctx, "/bin/df", "-h"),
	})
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
