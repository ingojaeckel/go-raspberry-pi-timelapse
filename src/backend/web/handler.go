package web

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/internal/accept"
)

// Handler serves the frontend build in fsys. The build embeds its text files
// (HTML, JS, CSS, SVG) only gzipped, as name.gz (see gzipBuild in
// src/frontend/vite.config.ts): a third of the size in the binary and on the
// network. They are sent as they are to clients that accept gzip, i.e. every
// browser, and decompressed for the others. Other files (fonts, already
// compressed) are served as they are.
func Handler(fsys fs.FS) http.Handler {
	files := http.FileServerFS(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Behind http.StripPrefix the folder itself is "", not "/".
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if path.Base(name) == "index.html" {
			w.Header().Set("Location", "./") // relative, like http.FileServer
			w.WriteHeader(http.StatusMovedPermanently)
			return
		}
		if name == "" || strings.HasSuffix(r.URL.Path, "/") {
			name = path.Join(name, "index.html")
		}
		gz, err := fs.ReadFile(fsys, name+".gz")
		if err != nil {
			files.ServeHTTP(w, r) // stored as it is, or missing
			return
		}
		serveGzipped(w, r, name, gz)
	})
}

// serveGzipped serves gz, the gzipped content of the file name.
func serveGzipped(w http.ResponseWriter, r *http.Request, name string, gz []byte) {
	h := w.Header()
	h.Set("Content-Type", mime.TypeByExtension(path.Ext(name)))
	h.Add("Vary", "Accept-Encoding")
	body := gz
	if accept.Gzip(r) {
		h.Set("Content-Encoding", "gzip")
	} else {
		var err error
		if body, err = gunzip(gz); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(body))
}

func gunzip(gz []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		return nil, err
	}
	return io.ReadAll(zr)
}
