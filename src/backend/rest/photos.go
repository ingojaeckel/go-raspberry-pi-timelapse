package rest

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/internal/accept"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

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

// photos serves the gallery listing, which the frontend polls. The encoded
// (and gzipped) response is kept until the listing changes, and its ETag
// lets browsers revalidate with a 304 instead of downloading it again: for
// a month of photos at one per minute it is 4 MB of JSON.
func (s *Server) photos(w http.ResponseWriter, r *http.Request) {
	body, err := s.photosBody()
	if err != nil {
		slog.Warn("listing photos", "err", err)
		writeJSON(w, http.StatusOK, GetPhotosResponse{Photos: []Photo{}})
		return
	}
	h := w.Header()
	h.Set("Content-Type", "application/json")
	h.Set("Cache-Control", "no-cache") // always revalidate
	h.Set("Vary", "Accept-Encoding")
	data, etag := body.plain, body.etag
	if accept.Gzip(r) {
		data, etag = body.gzipped, body.etag+"-gz"
		h.Set("Content-Encoding", "gzip")
	}
	h.Set("ETag", `"`+etag+`"`)
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(data))
}

// cachedPhotos is the /photos response for one generation of the listing.
type cachedPhotos struct {
	gen            uint64
	etag           string
	plain, gzipped []byte
}

// processID distinguishes ETags of different runs, whose index generations
// start at the same number.
var processID = strconv.FormatInt(time.Now().UnixNano(), 36)

// gzipWriters keeps compressors, which hold about 1 MB of state each.
var gzipWriters = sync.Pool{New: func() any {
	zw, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed) // valid level
	return zw
}}

// photosBody returns the cached /photos response, encoding it again if the
// listing changed.
func (s *Server) photosBody() (*cachedPhotos, error) {
	s.photosMu.Lock()
	defer s.photosMu.Unlock()
	list, gen, err := s.Photos.Generation()
	if err != nil {
		return nil, err
	}
	if c := s.photosCache; c != nil && c.gen == gen {
		return c, nil
	}
	plain, err := json.Marshal(GetPhotosResponse{Photos: toPhotos(list)})
	if err != nil {
		return nil, err
	}
	s.photosCache = &cachedPhotos{gen: gen, etag: processID + "-" + strconv.FormatUint(gen, 36), plain: plain, gzipped: gzipBytes(plain)}
	return s.photosCache, nil
}

func gzipBytes(plain []byte) []byte {
	var gz bytes.Buffer
	zw := gzipWriters.Get().(*gzip.Writer)
	defer gzipWriters.Put(zw)
	zw.Reset(&gz)
	zw.Write(plain)
	zw.Close()
	return gz.Bytes()
}

func (s *Server) listFiles(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, ListFilesResponse{Files: s.listOrWarn()})
}

// listOrWarn lists the photos; on failure it logs and returns what it has.
func (s *Server) listOrWarn() []files.File {
	list, err := s.Photos.List()
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
		return strconv.FormatInt(n>>20, 10) + " MB"
	case n >= 1<<10:
		return strconv.FormatInt(n>>10, 10) + " KB"
	default:
		return strconv.FormatInt(n, 10) + " bytes"
	}
}

func (s *Server) file(w http.ResponseWriter, r *http.Request) {
	s.serveFromFolder(w, r, r.PathValue("fileName"))
}

func (s *Server) mostRecentFile(w http.ResponseWriter, r *http.Request) {
	list, _ := s.Photos.List()
	if len(list) == 0 {
		http.NotFound(w, r)
		return
	}
	s.serveFromFolder(w, r, list[len(list)-1].Name)
}

// serveFromFolder serves name from the photo folder; os.Root rejects names
// that would leave it.
func (s *Server) serveFromFolder(w http.ResponseWriter, r *http.Request, name string) {
	f, err := os.OpenInRoot(s.Photos.Dir(), name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	serveFile(w, r, f)
}

// maxDownloadBytes is the largest file served as a download.
const maxDownloadBytes = 100 << 20

// serveFile serves a regular file up to maxDownloadBytes as a download.
func serveFile(w http.ResponseWriter, r *http.Request, f *os.File) {
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxDownloadBytes {
		http.NotFound(w, r)
		return
	}
	name := filepath.Base(f.Name())
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	http.ServeContent(newDeadlineWriter(w), r, name, info.ModTime(), f)
}

func (s *Server) deleteFiles(w http.ResponseWriter, r *http.Request) {
	names := r.URL.Query()["f"]
	if len(names) == 0 {
		writeError(w, http.StatusBadRequest, errors.New("no files to delete provided"))
		return
	}
	s.remove(w, names)
}

func (s *Server) deleteAllPhotos(w http.ResponseWriter, _ *http.Request) {
	list, err := s.Photos.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.remove(w, files.Names(list))
}

// remove deletes the named photos and responds with how many were deleted.
func (s *Server) remove(w http.ResponseWriter, names []string) {
	deleted, err := files.RemoveFiles(s.Photos.Dir(), names, func(name string, err error) {
		slog.Warn("not deleting photo", "name", name, "err", err)
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	slog.Info("deleted photos", "count", deleted, "requested", len(names))
	writeJSON(w, http.StatusOK, deleted)
}

// archive streams the photos named by the f query parameters, or all photos
// if there are none. size returns the exact length of what write writes.
func (s *Server) archive(filename, contentType string, size func([]files.File) (int64, error), write func(io.Writer, string, []files.File) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.Photos.List()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		photos := selected(list, r.URL.Query()["f"])
		n, err := size(photos)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		h := w.Header()
		h.Set("Content-Type", contentType)
		h.Set("Content-Disposition", "attachment; filename="+filename)
		h.Set("Content-Length", strconv.FormatInt(n, 10))
		if r.Method == http.MethodHead {
			return
		}
		streamArchive(newDeadlineWriter(w), func(w io.Writer) error { return write(w, s.Photos.Dir(), photos) })
	}
}

// selected keeps the photos that are wanted, in listing order, or all of
// them if none are.
func selected(list []files.File, wanted []string) []files.File {
	if len(wanted) == 0 {
		return list
	}
	want := make(map[string]bool, len(wanted))
	for _, n := range wanted {
		want[n] = true
	}
	var out []files.File
	for _, f := range list {
		if want[f.Name] {
			out = append(out, f)
		}
	}
	return out
}

// streamArchive writes the archive straight to the client. Once streaming has
// started the status can no longer change, so a failure aborts the
// connection: the client sees a failed download instead of a truncated file.
func streamArchive(w io.Writer, write func(io.Writer) error) {
	if err := write(w); err != nil {
		slog.Warn("archive download aborted", "err", err)
		panic(http.ErrAbortHandler)
	}
}
