package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func cacheHeaderFor(t *testing.T, path string) string {
	t.Helper()
	h := WithCacheHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec.Header().Get("Cache-Control")
}

func wantCacheHeader(t *testing.T, path, want string) {
	t.Helper()
	if got := cacheHeaderFor(t, path); got != want {
		t.Errorf("%s: Cache-Control %q, want %q", path, got, want)
	}
}

func TestHashedAssetsAreCachedForever(t *testing.T) {
	wantCacheHeader(t, "/static/frontend/build/assets/index-D2Xs6g67.js", "public, max-age=31536000, immutable")
	wantCacheHeader(t, "/static/frontend/build/assets/font-BgYH2mbd.woff2", "public, max-age=31536000, immutable")
}

func TestPagesAreRevalidated(t *testing.T) {
	// index.html references the hashed assets; it must never be served stale
	// after the binary on the Pi is updated.
	wantCacheHeader(t, "/static/frontend/build/", "no-cache")
	wantCacheHeader(t, "/static/frontend/build/index.html", "no-cache")
	wantCacheHeader(t, "/static/frontend/build/favicon.svg", "no-cache")
}
