package rest

import (
	"testing"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
)

func TestGetBasename(t *testing.T) {
	if s := getBasename(conf.StorageFolder + "/foo"); s != "foo" {
		t.Fatalf("Unexpected string: %s", s)
	}
	if s := getBasename("foo"); s != "foo" {
		t.Fatalf("Unexpected string: %s", s)
	}
}
