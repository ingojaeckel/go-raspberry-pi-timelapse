package rest

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// photoBytes is the size of a typical 8 MP JPEG from the Pi camera.
const photoBytes = 3 << 20

// benchPhotos writes n photos of size bytes, one minute apart, and returns
// their total size.
func benchPhotos(b *testing.B, e *env, n, size int) int64 {
	b.Helper()
	content := make([]byte, size)
	for i := range content {
		content[i] = byte(i * 7)
	}
	start := time.Now().Add(-time.Duration(n) * time.Minute)
	for i := range n {
		p := filepath.Join(e.folder, fmt.Sprintf("2026%06d.jpg", i))
		if err := os.WriteFile(p, content, 0o600); err != nil {
			b.Fatal(err)
		}
		mt := start.Add(time.Duration(i) * time.Minute)
		os.Chtimes(p, mt, mt)
	}
	return int64(n * size)
}

// settleIndex lists the folder, then again after the index's racy window
// (3s), so that later requests see the steady state of an unchanged folder.
func settleIndex(e *env) {
	e.srv.photosBody()
	time.Sleep(3500 * time.Millisecond)
	e.srv.photosBody()
}

// download GETs url and discards the body, failing unless at least want
// bytes arrive.
func download(b *testing.B, url string, want int64) {
	b.Helper()
	res, err := http.Get(url)
	if err != nil {
		b.Fatal(err)
	}
	defer res.Body.Close()
	n, err := io.Copy(io.Discard, res.Body)
	if err != nil || res.StatusCode != http.StatusOK || n < want {
		b.Fatalf("status %d, %d bytes, err %v", res.StatusCode, n, err)
	}
}

// BenchmarkArchive downloads 20 photos as one archive over a loopback
// connection through the real mux and middleware. On loopback the network is
// not the limit, so MB/s shows the CPU cost per byte; run with -cpu=1 to
// approximate the single core of a Pi Zero.
func BenchmarkArchive(b *testing.B) {
	for _, format := range []string{"zip", "tar"} {
		b.Run(format, func(b *testing.B) {
			e := newEnv(b)
			total := benchPhotos(b, e, 20, photoBytes)
			b.SetBytes(total)
			for b.Loop() {
				download(b, e.url+"/archive/"+format, total)
			}
		})
	}
}

// BenchmarkFileDownload downloads a single photo.
func BenchmarkFileDownload(b *testing.B) {
	e := newEnv(b)
	benchPhotos(b, e, 1, photoBytes)
	b.SetBytes(photoBytes)
	for b.Loop() {
		download(b, e.url+"/file/last", photoBytes)
	}
}

// BenchmarkPhotos lists the photos as JSON for the gallery, which polls it:
// 1440 photos are a month at the default interval, 44640 a month at one
// photo per minute. "full" downloads the listing, "revalidate" is the
// browser's poll with If-None-Match while nothing changed, "new-photo" the
// first poll after a capture (the folder is read again and the JSON
// encoded and gzipped again).
func BenchmarkPhotos(b *testing.B) {
	for _, n := range []int{1440, 44640} {
		e := newEnv(b)
		benchPhotos(b, e, n, 0)
		settleIndex(e)
		b.Run(fmt.Sprintf("full/%d", n), func(b *testing.B) {
			benchRequest(b, e.handler, benchReq("/photos", "Accept-Encoding", "gzip"))
		})
		b.Run(fmt.Sprintf("revalidate/%d", n), func(b *testing.B) {
			etag := `"` + e.srv.photosCache.etag + `-gz"`
			benchRequest(b, e.handler, benchReq("/photos", "Accept-Encoding", "gzip", "If-None-Match", etag))
		})
		b.Run(fmt.Sprintf("new-photo/%d", n), func(b *testing.B) {
			req := benchReq("/photos", "Accept-Encoding", "gzip")
			i := 0
			for b.Loop() {
				os.WriteFile(filepath.Join(e.folder, fmt.Sprintf("new-%06d.jpg", i)), nil, 0o600)
				i++
				e.handler.ServeHTTP(&discardResponse{header: http.Header{}}, req)
			}
		})
	}
}

// BenchmarkMostRecentFile serves the newest of a month of photos (the
// frontend's live view).
func BenchmarkMostRecentFile(b *testing.B) {
	e := newEnv(b)
	benchPhotos(b, e, 44640, 0)
	settleIndex(e)
	benchRequest(b, e.handler, benchReq("/file/last"))
}

// BenchmarkSmallRequest measures the per-request overhead of mux and
// middleware.
func BenchmarkSmallRequest(b *testing.B) {
	benchRequest(b, newEnv(b).handler, benchReq("/version"))
}

func benchReq(path string, header ...string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Host = "localhost"
	for i := 0; i+1 < len(header); i += 2 {
		req.Header.Set(header[i], header[i+1])
	}
	return req
}

func benchRequest(b *testing.B, h http.Handler, req *http.Request) {
	for b.Loop() {
		w := &discardResponse{header: http.Header{}}
		h.ServeHTTP(w, req)
		if w.status >= 400 {
			b.Fatalf("status %d", w.status)
		}
	}
}

// discardResponse is a ResponseWriter that keeps nothing, so benchmarks
// measure the handler rather than a recorder.
type discardResponse struct {
	header http.Header
	status int
}

func (d *discardResponse) Header() http.Header         { return d.header }
func (d *discardResponse) Write(p []byte) (int, error) { return len(p), nil }
func (d *discardResponse) WriteHeader(s int)           { d.status = s }
