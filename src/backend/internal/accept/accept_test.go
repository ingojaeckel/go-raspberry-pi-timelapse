package accept

import (
	"net/http/httptest"
	"testing"
)

func TestGzip(t *testing.T) {
	for header, want := range map[string]bool{
		"": false, "gzip": true, "deflate, GZIP": true, "gzip;q=0": false, "gzip; q=0": false, "gzip;q=0.5": true, "x-gzip": false,
	} {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Accept-Encoding", header)
		if got := Gzip(r); got != want {
			t.Errorf("%q: %v, want %v", header, got, want)
		}
	}
}
