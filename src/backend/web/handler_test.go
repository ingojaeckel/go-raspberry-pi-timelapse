package web

import (
	"bytes"
	"compress/gzip"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func gzipped(t *testing.T, s string) []byte {
	t.Helper()
	var b bytes.Buffer
	zw := gzip.NewWriter(&b)
	zw.Write([]byte(s))
	zw.Close()
	return b.Bytes()
}

// testBuild looks like the frontend build: text files gzipped, fonts as they are.
func testBuild(t *testing.T) fstest.MapFS {
	return fstest.MapFS{
		"index.html.gz":           {Data: gzipped(t, "<html>app</html>")},
		"assets/index-1a2b.js":    {Data: []byte("not served: only the .gz is embedded")},
		"assets/index-1a2b.js.gz": {Data: gzipped(t, "console.log('app')")},
		"assets/font-3c4d.woff2":  {Data: []byte("woff2")},
	}
}

const prefix = "/static/frontend/build/"

// get requests prefix+path.
func get(t *testing.T, fsys fs.FS, path, acceptEncoding string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, prefix+strings.TrimPrefix(path, "/"), nil)
	if acceptEncoding != "" {
		r.Header.Set("Accept-Encoding", acceptEncoding)
	}
	rec := httptest.NewRecorder()
	http.StripPrefix(prefix, Handler(fsys)).ServeHTTP(rec, r) // as the app mounts it
	return rec
}

func TestBrowsersGetTheGzippedFile(t *testing.T) {
	build := testBuild(t)
	rec := get(t, build, "/assets/index-1a2b.js", "gzip, deflate, br")
	h := rec.Header()
	if rec.Code != 200 || h.Get("Content-Encoding") != "gzip" || h.Get("Vary") != "Accept-Encoding" {
		t.Fatalf("status %d, headers %v", rec.Code, h)
	}
	if ct := h.Get("Content-Type"); !strings.HasPrefix(ct, "text/javascript") {
		t.Errorf("Content-Type %q, want the original file's", ct)
	}
	if !bytes.Equal(rec.Body.Bytes(), build["assets/index-1a2b.js.gz"].Data) {
		t.Error("the embedded gzip bytes should be sent as they are")
	}
}

func TestClientsWithoutGzipGetTheFileDecompressed(t *testing.T) {
	rec := get(t, testBuild(t), "/assets/index-1a2b.js", "")
	if rec.Code != 200 || rec.Header().Get("Content-Encoding") != "" || rec.Body.String() != "console.log('app')" {
		t.Fatalf("status %d, encoding %q, body %q", rec.Code, rec.Header().Get("Content-Encoding"), rec.Body)
	}
}

func TestTheFolderServesIndexHTML(t *testing.T) {
	rec := get(t, testBuild(t), "", "")
	if rec.Code != 200 || rec.Body.String() != "<html>app</html>" || !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("status %d, type %q, body %q", rec.Code, rec.Header().Get("Content-Type"), rec.Body)
	}
}

func TestIndexHTMLRedirectsToTheFolder(t *testing.T) {
	rec := get(t, testBuild(t), "/index.html", "gzip")
	if rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != "./" {
		t.Fatalf("status %d, location %q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestCompressedFilesAreServedAsTheyAre(t *testing.T) {
	rec := get(t, testBuild(t), "/assets/font-3c4d.woff2", "gzip")
	if rec.Code != 200 || rec.Header().Get("Content-Encoding") != "" || rec.Body.String() != "woff2" {
		t.Fatalf("status %d, encoding %q, body %q", rec.Code, rec.Header().Get("Content-Encoding"), rec.Body)
	}
}

func TestMissingFilesAre404(t *testing.T) {
	if rec := get(t, testBuild(t), "/assets/missing.js", "gzip"); rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestTheBuildIsEmbeddedGzipped(t *testing.T) {
	if _, err := fs.Stat(Dist(), "index.html.gz"); err != nil {
		t.Fatalf("index.html.gz missing from the embedded build (run make frontend): %v", err)
	}
	texts, _ := fs.Glob(Dist(), "assets/*.js")
	if len(texts) != 0 {
		t.Fatalf("%v embedded uncompressed", texts)
	}
	rec := get(t, Dist(), "/", "")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "<div id=\"root\">") {
		t.Fatalf("status %d, body %.200q", rec.Code, rec.Body)
	}
}
