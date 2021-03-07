package rest

import (
	"testing"

	"github.com/facebookgo/ensure"
)

func TestGetHumanReadableSize(t *testing.T) {
	ensure.DeepEqual(t, "0 bytes", getHumanReadableSize(0))
	ensure.DeepEqual(t, "1023 bytes", getHumanReadableSize(1023))
	ensure.DeepEqual(t, "1 KB", getHumanReadableSize(1024))
	ensure.DeepEqual(t, "10 KB", getHumanReadableSize(10*1024))
	ensure.DeepEqual(t, "1 MB", getHumanReadableSize(1024*1024))
}
