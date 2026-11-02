package rest

import (
	"testing"

	"github.com/facebookgo/ensure"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
)

func TestGetBasename(t *testing.T) {
	ensure.DeepEqual(t, "foo", getBasename(conf.StorageFolder+"/foo"))
	ensure.DeepEqual(t, "foo", getBasename("foo"))
}
