package rest

import (
	"io"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"
)

// localSuffixes are domains that only resolve on the local network.
var localSuffixes = []string{".local", ".lan", ".home.arpa", ".internal"}

// Middleware protects the API and logs requests at debug level:
//
//   - Requests must use a local Host name (IP address, localhost, a single
//     label like "raspberrypi", a local domain like "timelapse.local") or one
//     of allowedHosts. This blocks DNS rebinding, where a public website
//     re-resolves its own name to the camera's IP and becomes same-origin.
//   - Cross-site POST/DELETE requests are rejected, so a page open on a phone
//     connected to the camera's WiFi cannot delete photos or shut it down.
//   - No CORS headers are sent, so other websites cannot read photos or
//     settings. The frontend is served by this server (same origin), and
//     `npm start` proxies the API.
func Middleware(next http.Handler, allowedHosts ...string) http.Handler {
	protected := http.NewCrossOriginProtection().Handler(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allowedHost(r.Host, allowedHosts) {
			slog.Warn("rejected request for unknown host", "host", r.Host, "uri", r.RequestURI, "remote", r.RemoteAddr)
			http.Error(w, "unknown host name; start the camera with -allowedHosts to allow it", http.StatusForbidden)
			return
		}
		if !slog.Default().Enabled(r.Context(), slog.LevelDebug) {
			protected.ServeHTTP(w, r) // no per-request cost unless debugging
			return
		}
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		protected.ServeHTTP(rec, r)
		slog.Debug("request", "method", r.Method, "uri", r.RequestURI, "status", rec.status, "took", time.Since(start))
	})
}

func allowedHost(hostport string, extra []string) bool {
	host := hostName(hostport)
	return isLocalName(host) || slices.ContainsFunc(extra, func(h string) bool { return strings.EqualFold(host, h) })
}

// hostName strips the port, IPv6 brackets and a trailing dot, and lowercases.
func hostName(hostport string) string {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	return strings.ToLower(strings.TrimSuffix(strings.Trim(host, "[]"), "."))
}

// isLocalName reports whether host can only refer to the local network.
func isLocalName(host string) bool {
	if net.ParseIP(host) != nil || host == "localhost" || !strings.Contains(host, ".") {
		return true
	}
	return slices.ContainsFunc(localSuffixes, func(s string) bool { return strings.HasSuffix(host, s) })
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// ReadFrom keeps the wrapped writer's io.ReaderFrom reachable, which lets
// net/http send files with sendfile(2) instead of copying them through user
// space.
func (r *statusRecorder) ReadFrom(src io.Reader) (int64, error) {
	return io.Copy(r.ResponseWriter, src)
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (r *statusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }
