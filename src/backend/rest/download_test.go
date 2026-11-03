package rest

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
)

func TestArchivesHaveAnExactContentLength(t *testing.T) {
	e := newEnv(t)
	e.photo(t, "a.jpg", strings.Repeat("a", 1000), 2*time.Minute)
	e.photo(t, "b.jpg", "b", time.Minute)
	for _, format := range []string{"zip", "tar"} {
		res, body := do(t, "GET", e.url+"/archive/"+format, "")
		wantStatus(t, res, body, 200)
		if res.ContentLength != int64(len(body)) || len(res.TransferEncoding) != 0 {
			t.Errorf("%s: Content-Length %d, chunked %v, body %d bytes", format, res.ContentLength, res.TransferEncoding, len(body))
		}
	}
}

func TestArchiveHeadSendsOnlyTheHeaders(t *testing.T) {
	e := newEnv(t)
	e.photo(t, "a.jpg", strings.Repeat("a", 1000), time.Minute)
	res, _ := do(t, "GET", e.url+"/archive/tar", "")
	head, body := do(t, "HEAD", e.url+"/archive/tar", "")
	if head.StatusCode != 200 || head.ContentLength != res.ContentLength || body != "" {
		t.Fatalf("HEAD: %d, Content-Length %d (GET %d), body %q", head.StatusCode, head.ContentLength, res.ContentLength, body)
	}
}

func TestPhotosRevalidatesWithETag(t *testing.T) {
	e := newEnv(t)
	e.photo(t, "a.jpg", "a", time.Minute)
	res, body := do(t, "GET", e.url+"/photos", "")
	wantStatus(t, res, body, 200)
	etag := res.Header.Get("ETag")
	if etag == "" || res.Header.Get("Cache-Control") != "no-cache" || !strings.Contains(body, `"a.jpg"`) {
		t.Fatalf("headers %v, body %s", res.Header, body)
	}

	res, body = do(t, "GET", e.url+"/photos", "", "If-None-Match", etag)
	wantStatus(t, res, body, http.StatusNotModified)

	e.photo(t, "b.jpg", "b", 0)
	os.Chtimes(e.folder, time.Now(), time.Now()) // recent: never trusted
	res, body = do(t, "GET", e.url+"/photos", "", "If-None-Match", etag)
	wantStatus(t, res, body, 200)
	if res.Header.Get("ETag") == etag || !strings.Contains(body, `"b.jpg"`) {
		t.Fatalf("stale listing after a new photo: %s", body)
	}
}

func TestPhotosIsGzippedForClientsThatAcceptIt(t *testing.T) {
	e := newEnv(t)
	e.photo(t, "a.jpg", "a", time.Minute)
	_, plain := do(t, "GET", e.url+"/photos", "", "Accept-Encoding", "identity")
	res, body := do(t, "GET", e.url+"/photos", "", "Accept-Encoding", "br, gzip;q=0.8")
	if res.Header.Get("Content-Encoding") != "gzip" || res.Header.Get("Vary") != "Accept-Encoding" {
		t.Fatalf("headers %v", res.Header)
	}
	zr, err := gzip.NewReader(strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if unzipped, _ := io.ReadAll(zr); string(unzipped) != plain {
		t.Fatalf("gzipped %q, plain %q", unzipped, plain)
	}
}

func TestAcceptsGzip(t *testing.T) {
	for header, want := range map[string]bool{
		"": false, "gzip": true, "deflate, GZIP": true, "gzip;q=0": false, "gzip; q=0": false, "gzip;q=0.5": true, "x-gzip": false,
	} {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Accept-Encoding", header)
		if got := acceptsGzip(r); got != want {
			t.Errorf("%q: %v, want %v", header, got, want)
		}
	}
}

// A client that stops reading, like a phone that left the WiFi, must not
// keep its download (and an open photo) alive.
func TestStalledDownloadsAreDropped(t *testing.T) {
	defer func(d time.Duration) { stallTimeout = d }(stallTimeout)
	stallTimeout = 300 * time.Millisecond
	e := newEnv(t)
	content := bytes.Repeat([]byte("x"), 4<<20)
	for i := range 20 {
		os.WriteFile(fmt.Sprintf("%s/%02d.jpg", e.folder, i), content, 0o600)
	}
	for _, format := range []string{"zip", "tar"} {
		done := make(chan struct{})
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer close(done)
			defer func() { recover() }() // the expected http.ErrAbortHandler
			e.handler.ServeHTTP(w, r)
		}))
		conn, err := net.Dial("tcp", ts.Listener.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(conn, "GET /archive/%s HTTP/1.1\r\nHost: localhost\r\n\r\n", format)
		conn.Read(make([]byte, 1024)) // then stop reading
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Errorf("%s: the download of a client that stopped reading is still running", format)
		}
		conn.Close()
		ts.Close()
	}
}

// recordingWriter is a ResponseWriter that reports how ReadFrom was called.
type recordingWriter struct {
	discardResponse
	files, others int
}

func (r *recordingWriter) ReadFrom(src io.Reader) (int64, error) {
	if lr, ok := src.(*io.LimitedReader); ok {
		if _, ok := lr.R.(*os.File); ok && lr.N <= sendChunk {
			r.files++
			return io.Copy(io.Discard, src)
		}
	}
	r.others++
	return io.Copy(io.Discard, src)
}

func TestDeadlineWriterKeepsFilesForSendfile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "photo")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	f.Write(make([]byte, 3*sendChunk+10))
	f.Seek(0, io.SeekStart)
	rec := &recordingWriter{discardResponse: discardResponse{header: http.Header{}}}
	d := newDeadlineWriter(rec)
	if n, err := d.ReadFrom(&io.LimitedReader{R: f, N: 3*sendChunk + 10}); err != nil || n != 3*sendChunk+10 {
		t.Fatalf("copied %d, %v", n, err)
	}
	if rec.files != 4 || rec.others != 0 {
		t.Fatalf("%d chunks as files, %d otherwise", rec.files, rec.others)
	}
	// Other readers are copied until EOF.
	if n, err := d.ReadFrom(strings.NewReader("abc")); err != nil || n != 3 {
		t.Fatalf("copied %d, %v", n, err)
	}
	if d.Unwrap() != http.ResponseWriter(rec) {
		t.Fatal("Unwrap must return the wrapped writer")
	}
}

// With debug logging the middleware records the status, and must still let
// net/http use sendfile.
func TestDebugLoggingKeepsDownloadsWorking(t *testing.T) {
	var logs syncBuffer
	defer func(l *slog.Logger) { slog.SetDefault(l) }(slog.Default())
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	e := newEnv(t)
	e.photo(t, "a.jpg", strings.Repeat("a", 100000), time.Minute)
	res, body := do(t, "GET", e.url+"/file/a.jpg", "")
	wantStatus(t, res, body, 200)
	if len(body) != 100000 {
		t.Fatalf("%d bytes", len(body))
	}
	// The request is logged once the handler returns, which may be after
	// the client has the response.
	for deadline := time.Now().Add(5 * time.Second); !strings.Contains(logs.String(), "uri=/file/a.jpg status=200"); {
		if time.Now().After(deadline) {
			t.Fatalf("logs %s", logs.String())
		}
		time.Sleep(10 * time.Millisecond)
	}
	rec := &recordingWriter{discardResponse: discardResponse{header: http.Header{}}}
	if n, _ := (&statusRecorder{ResponseWriter: rec}).ReadFrom(io.LimitReader(strings.NewReader("abc"), 3)); n != 3 || rec.others != 1 {
		t.Fatalf("statusRecorder.ReadFrom copied %d, forwarded %d", n, rec.others)
	}
}

func TestArchiveSizeErrorsAre500(t *testing.T) {
	e := newEnv(t)
	failing := func([]files.File) (int64, error) { return 0, errors.New("bad name") }
	rec := httptest.NewRecorder()
	e.srv.archive("archive.zip", "application/zip", failing, nil)(rec, httptest.NewRequest("GET", "/archive/zip", nil))
	if rec.Code != 500 || !strings.Contains(rec.Body.String(), "bad name") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
