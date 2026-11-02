package web

import (
	"io/fs"
	"testing"
)

func TestDistServesTheFrontendBuild(t *testing.T) {
	b, err := fs.ReadFile(Dist(), "index.html")
	if err != nil {
		t.Fatalf("index.html missing from the embedded build (run make frontend): %v", err)
	}
	if len(b) == 0 {
		t.Fatal("index.html is empty")
	}
}
