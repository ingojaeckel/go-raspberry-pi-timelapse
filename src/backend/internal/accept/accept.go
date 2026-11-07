// Package accept reads what a client accepts from its request headers.
package accept

import (
	"net/http"
	"strings"
)

// Gzip reports whether the client accepts gzip-encoded responses
// (Accept-Encoding lists gzip, and not with q=0).
func Gzip(r *http.Request) bool {
	for _, v := range r.Header.Values("Accept-Encoding") {
		for part := range strings.SplitSeq(v, ",") {
			coding, params, _ := strings.Cut(strings.TrimSpace(part), ";")
			if strings.EqualFold(strings.TrimSpace(coding), "gzip") && strings.ReplaceAll(params, " ", "") != "q=0" {
				return true
			}
		}
	}
	return false
}
