package rest

import "testing"

func TestGetBasename(t *testing.T) {
	if s := getBasename("timelapse-pictures/foo"); s != "foo" {
		t.Fatalf("Unexpected string: %s", s)
	}
	if s := getBasename("foo"); s != "foo" {
		t.Fatalf("Unexpected string: %s", s)
	}
}
