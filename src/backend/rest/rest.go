// Package rest implements the HTTP API used by the frontend.
//
// GET requests never change state. Photos are deleted with DELETE, the
// system is restarted or shut down with POST.
package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/logfile"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
)

const (
	maxLogBytes      = 256 << 10
	maxSettingsBytes = 64 << 10
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
	// Folder holds the photos; TempFolder the preview captures.
	Folder, TempFolder string
	// LogFile is served by /logs.
	LogFile string
	// ThermalFile holds the CPU temperature in millidegrees Celsius.
	ThermalFile string
	Version     string
	// Run executes system commands (monitoring, shutdown, restart).
	Run func(ctx context.Context, name string, args ...string) ([]byte, error)
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
	mux.HandleFunc("GET /archive/zip", s.archive("archive.zip", "application/zip", files.WriteZip))
	mux.HandleFunc("GET /archive/tar", s.archive("archive.tar", "application/tar", files.WriteTar))
	mux.HandleFunc("GET /monitoring", s.monitoring)
	mux.HandleFunc("GET /logs", s.logs)
	mux.HandleFunc("POST /admin/shutdown", s.command("/usr/bin/sudo", "/sbin/shutdown", "-h", "now"))
	mux.HandleFunc("POST /admin/restart", s.command("/usr/bin/sudo", "/sbin/shutdown", "-r", "now"))
}

func (s *Server) version(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "Hello from %s on %s [version:%s]", runtime.GOARCH, runtime.GOOS, s.Version)
}

func (s *Server) getConfiguration(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.Settings.Get())
}

// updateConfiguration applies the schedule and photo fields of the request.
// Preview size is not editable through the API.
func (s *Server) updateConfiguration(w http.ResponseWriter, r *http.Request) {
	var req conf.Settings
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxSettingsBytes)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	updated, err := s.Settings.Update(func(c *conf.Settings) {
		c.SecondsBetweenCaptures = req.SecondsBetweenCaptures
		c.OffsetWithinHour = req.OffsetWithinHour
		c.Quality = req.Quality
		c.RotateBy = req.RotateBy
		c.ResolutionSetting = req.ResolutionSetting
		c.DebugEnabled = req.DebugEnabled
	})
	switch {
	case errors.Is(err, conf.ErrInvalid):
		writeError(w, http.StatusBadRequest, err)
	case err != nil:
		writeError(w, http.StatusInternalServerError, err)
	default:
		writeJSON(w, http.StatusOK, updated)
	}
}

// preview takes a photo with the preview size and serves it without keeping it.
func (s *Server) preview(w http.ResponseWriter, r *http.Request) {
	set := s.Settings.Get()
	path, err := s.Camera.Capture(r.Context(), s.TempFolder, timelapse.Options{
		Width:   set.PreviewResolutionWidth,
		Height:  set.PreviewResolutionHeight,
		Quality: set.Quality,
		Rotate:  set.RotateBy == 180,
	})
	if err != nil {
		http.Error(w, "Failed to take capture: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(path)
	f, err := os.Open(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	serveFile(w, r, f)
}

func (s *Server) photos(w http.ResponseWriter, _ *http.Request) {
	list, err := files.ListFiles(s.Folder)
	if err != nil {
		slog.Warn("listing photos", "err", err)
	}
	photos := make([]Photo, len(list))
	for i, f := range list {
		photos[i] = Photo{Name: f.Name, ModTime: f.ModTime, Size: humanReadableSize(f.Bytes)}
	}
	writeJSON(w, http.StatusOK, GetPhotosResponse{Photos: photos})
}

func (s *Server) listFiles(w http.ResponseWriter, _ *http.Request) {
	list, err := files.ListFiles(s.Folder)
	if err != nil {
		slog.Warn("listing photos", "err", err)
	}
	writeJSON(w, http.StatusOK, ListFilesResponse{Files: list})
}

func (s *Server) file(w http.ResponseWriter, r *http.Request) {
	s.serveFromFolder(w, r, r.PathValue("fileName"))
}

func (s *Server) mostRecentFile(w http.ResponseWriter, r *http.Request) {
	list, _ := files.ListFiles(s.Folder)
	if len(list) == 0 {
		http.NotFound(w, r)
		return
	}
	s.serveFromFolder(w, r, list[len(list)-1].Name)
}

// serveFromFolder serves name from the photo folder; os.Root rejects names
// that would leave it.
func (s *Server) serveFromFolder(w http.ResponseWriter, r *http.Request, name string) {
	f, err := os.OpenInRoot(s.Folder, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	serveFile(w, r, f)
}

func serveFile(w http.ResponseWriter, r *http.Request, f *os.File) {
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > conf.MaxFileSizeBytes {
		http.NotFound(w, r)
		return
	}
	name := filepath.Base(f.Name())
	w.Header().Set(conf.HeaderContentDisposition, mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	http.ServeContent(w, r, name, info.ModTime(), f)
}

func (s *Server) deleteFiles(w http.ResponseWriter, r *http.Request) {
	names := r.URL.Query()["f"]
	if len(names) == 0 {
		writeError(w, http.StatusBadRequest, errors.New("no files to delete provided"))
		return
	}
	writeJSON(w, http.StatusOK, s.remove(names))
}

func (s *Server) deleteAllPhotos(w http.ResponseWriter, _ *http.Request) {
	list, err := files.ListFiles(s.Folder)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	names := make([]string, len(list))
	for i, f := range list {
		names[i] = f.Name
	}
	writeJSON(w, http.StatusOK, s.remove(names))
}

// remove deletes the named photos and returns how many were deleted.
func (s *Server) remove(names []string) int {
	deleted := 0
	for _, name := range names {
		if err := files.RemoveFile(s.Folder, name); err != nil {
			slog.Warn("not deleting photo", "name", name, "err", err)
			continue
		}
		deleted++
	}
	slog.Info("deleted photos", "count", deleted, "requested", len(names))
	return deleted
}

// archive streams the photos named by the f query parameters, or all photos
// if there are none.
func (s *Server) archive(filename, contentType string, write func(io.Writer, string, []string) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := files.ListFiles(s.Folder)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		wanted := r.URL.Query()["f"]
		var names []string
		for _, f := range list {
			if len(wanted) == 0 || slices.Contains(wanted, f.Name) {
				names = append(names, f.Name)
			}
		}
		streamArchive(w, filename, contentType, func(w io.Writer) error { return write(w, s.Folder, names) })
	}
}

// streamArchive writes the archive straight to the client. Once streaming has
// started the status can no longer change, so a failure aborts the
// connection: the client sees a failed download instead of a truncated file.
func streamArchive(w http.ResponseWriter, filename, contentType string, write func(io.Writer) error) {
	w.Header().Set(conf.HeaderContentType, contentType)
	w.Header().Set(conf.HeaderContentDisposition, "attachment; filename="+filename)
	if err := write(w); err != nil {
		slog.Warn("archive download aborted", "err", err)
		panic(http.ErrAbortHandler)
	}
}

func (s *Server) monitoring(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	run := func(name string, args ...string) string {
		out, err := s.Run(ctx, name, args...)
		if err != nil {
			return err.Error()
		}
		return string(out)
	}
	cpu, err := os.ReadFile(s.ThermalFile)
	if err != nil {
		cpu = []byte(err.Error())
	}
	writeJSON(w, http.StatusOK, MonitoringResponse{
		Time:           run("/bin/date"),
		GpuTemperature: run(vcgencmdPath(), "measure_temp"),
		CpuTemperature: string(cpu),
		Uptime:         run("/usr/bin/uptime"),
		FreeDiskSpace:  run("/bin/df", "-h"),
	})
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

func humanReadableSize(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%d MB", n>>20)
	case n >= 1<<10:
		return fmt.Sprintf("%d KB", n>>10)
	default:
		return fmt.Sprintf("%d bytes", n)
	}
}
