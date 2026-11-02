package rest

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
)

// preview takes a photo with the preview size and serves it without keeping it.
func (s *Server) preview(w http.ResponseWriter, r *http.Request) {
	path, err := s.Camera.Capture(r.Context(), s.TempFolder, timelapse.PreviewOptions(s.Settings.Get()))
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
	writeJSON(w, http.StatusOK, GetPhotosResponse{Photos: toPhotos(s.listOrWarn())})
}

func (s *Server) listFiles(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, ListFilesResponse{Files: s.listOrWarn()})
}

// listOrWarn lists the photos; on failure it logs and returns what it has.
func (s *Server) listOrWarn() []files.File {
	list, err := files.ListFiles(s.Folder)
	if err != nil {
		slog.Warn("listing photos", "err", err)
	}
	return list
}

func toPhotos(list []files.File) []Photo {
	photos := make([]Photo, len(list))
	for i, f := range list {
		photos[i] = Photo{Name: f.Name, ModTime: f.ModTime, Size: humanReadableSize(f.Bytes)}
	}
	return photos
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

// serveFile serves a regular file up to conf.MaxFileSizeBytes as a download.
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
	writeJSON(w, http.StatusOK, s.remove(files.Names(list)))
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
		names := selected(files.Names(list), r.URL.Query()["f"])
		streamArchive(w, filename, contentType, func(w io.Writer) error { return write(w, s.Folder, names) })
	}
}

// selected keeps the names that are wanted, or all of them if none are.
func selected(names, wanted []string) []string {
	if len(wanted) == 0 {
		return names
	}
	return slices.DeleteFunc(names, func(n string) bool { return !slices.Contains(wanted, n) })
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
