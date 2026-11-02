package web

import (
	"net/http"
	"strings"
)

// WithCacheHeaders sets Cache-Control for the embedded frontend.
//
// Vite puts a content hash in every file name under assets/, so those can be
// cached forever. Everything else (index.html, favicon) must be revalidated,
// otherwise phones keep showing the old UI after the binary is updated.
func WithCacheHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		next.ServeHTTP(w, r)
	})
}
