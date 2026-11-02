package rest

import (
	"log/slog"
	"net/http"
	"time"
)

// Middleware wraps the routes with CORS headers, preflight handling,
// protection against cross-site requests that change state (a page open on a
// phone connected to the camera's WiFi must not be able to delete photos or
// shut the camera down), and debug-level request logging.
func Middleware(next http.Handler) http.Handler {
	protected := http.NewCrossOriginProtection().Handler(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "*")
		h.Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		protected.ServeHTTP(rec, r)
		slog.Debug("request", "method", r.Method, "uri", r.RequestURI, "status", rec.status, "took", time.Since(start))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (r *statusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }
