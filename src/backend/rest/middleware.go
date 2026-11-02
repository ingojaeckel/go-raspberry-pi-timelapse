package rest

import (
	"log/slog"
	"net"
	"net/http"
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
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		protected.ServeHTTP(rec, r)
		slog.Debug("request", "method", r.Method, "uri", r.RequestURI, "status", rec.status, "took", time.Since(start))
	})
}

func allowedHost(hostport string, extra []string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	host = strings.ToLower(strings.TrimSuffix(strings.Trim(host, "[]"), "."))
	if net.ParseIP(host) != nil || host == "localhost" || !strings.Contains(host, ".") {
		return true
	}
	for _, suffix := range localSuffixes {
		if strings.HasSuffix(host, suffix) {
			return true
		}
	}
	for _, h := range extra {
		if strings.EqualFold(host, h) {
			return true
		}
	}
	return false
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
