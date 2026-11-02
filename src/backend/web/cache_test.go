package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/facebookgo/ensure"
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

func TestHashedAssetsAreCachedForever(t *testing.T) {
	ensure.DeepEqual(t, cacheHeaderFor(t, "/static/frontend/build/assets/index-D2Xs6g67.js"), "public, max-age=31536000, immutable")
	ensure.DeepEqual(t, cacheHeaderFor(t, "/static/frontend/build/assets/font-BgYH2mbd.woff2"), "public, max-age=31536000, immutable")
}

func TestPagesAreRevalidated(t *testing.T) {
	// index.html references the hashed assets; it must never be served stale
	// after the binary on the Pi is updated.
	ensure.DeepEqual(t, cacheHeaderFor(t, "/static/frontend/build/"), "no-cache")
	ensure.DeepEqual(t, cacheHeaderFor(t, "/static/frontend/build/index.html"), "no-cache")
	ensure.DeepEqual(t, cacheHeaderFor(t, "/static/frontend/build/favicon.svg"), "no-cache")
}
