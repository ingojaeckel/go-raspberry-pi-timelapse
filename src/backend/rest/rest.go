// Package rest implements the HTTP API used by the frontend.
//
// GET requests never change state. Photos are deleted with DELETE, the
// system is restarted or shut down with POST.
package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"sync"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
)

// Settings is implemented by *conf.Store.
type Settings interface {
	Get() conf.Settings
	Update(func(*conf.Settings)) (conf.Settings, error)
}

// Server holds the dependencies of the HTTP handlers.
type Server struct {
	Settings Settings
	Camera   timelapse.Capturer
	// Photos lists the photo folder; TempFolder holds the preview captures.
	Photos     *files.Index
	TempFolder string
	// LogFile is served by /logs.
	LogFile string
	// ThermalFile holds the CPU temperature in millidegrees Celsius.
	ThermalFile string
	Version     string
	// Run executes system commands (monitoring, shutdown, restart).
	Run func(ctx context.Context, name string, args ...string) ([]byte, error)

	photosMu    sync.Mutex
	photosCache *cachedPhotos
}

// Register adds the API routes to mux.
func (s *Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /version", s.version)
	mux.HandleFunc("GET /configuration", s.getConfiguration)
	mux.HandleFunc("POST /configuration", s.updateConfiguration)
	mux.HandleFunc("GET /capture", s.preview)
	mux.HandleFunc("GET /photos", s.photos)
	mux.HandleFunc("DELETE /photos", s.deleteAllPhotos)
	mux.HandleFunc("GET /file", s.listFiles)
	mux.HandleFunc("DELETE /file", s.deleteFiles)
	mux.HandleFunc("GET /file/last", s.mostRecentFile)
	mux.HandleFunc("GET /file/{fileName}", s.file)
	mux.HandleFunc("GET /archive/zip", s.archive("archive.zip", "application/zip", files.ZipSize, files.WriteZip))
	mux.HandleFunc("GET /archive/tar", s.archive("archive.tar", "application/tar", files.TarSize, files.WriteTar))
	mux.HandleFunc("GET /monitoring", s.monitoring)
	mux.HandleFunc("GET /logs", s.logs)
	mux.HandleFunc("POST /admin/shutdown", s.command("/usr/bin/sudo", "/sbin/shutdown", "-h", "now"))
	mux.HandleFunc("POST /admin/restart", s.command("/usr/bin/sudo", "/sbin/shutdown", "-r", "now"))
}

func (s *Server) version(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "Hello from %s on %s [version:%s]", runtime.GOARCH, runtime.GOOS, s.Version)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set(conf.HeaderContentType, conf.HeaderContentTypeJSON)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Warn("writing response", "err", err)
	}
}

// writeError sends err as a JSON string, which the frontend shows to the user.
func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, err.Error())
}
