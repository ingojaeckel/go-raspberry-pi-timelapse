package timelapse

import (
	"github.com/facebookgo/ensure"
	"strings"
	"testing"
)

func TestBuildingArguments(t *testing.T) {
	unrotatedCamera, err := NewCamera("foo", 200, 100, false)
	ensure.Nil(t, err)
	ensure.False(t, unrotatedCamera.flipVertically)
	ensure.False(t, unrotatedCamera.flipHorizontally)

	p := unrotatedCamera.getAbsoluteFilepath()
	ensure.DeepEqual(t, 0, strings.Index(p, "foo/"))
	ensure.DeepEqual(t, 23, strings.Index(p, ".jpg"))
	ensure.DeepEqual(t, 27, len(p))
}

func TestRaspistillArgs(t *testing.T) {
	unrotatedCamera, _ := NewCamera("foo", 200, 100, false)
	args := unrotatedCamera.getRaspistillArgs("foo/someFile.jpg")
	ensure.DeepEqual(t, []string{"-w", "200", "-h", "100", "-o", "foo/someFile.jpg"}, args)
}

func TestCreateRotatedCamera(t *testing.T) {
	rotatedCamera, _ := NewCamera("foo", 200, 100, true)
	ensure.True(t, rotatedCamera.flipVertically)
	ensure.True(t, rotatedCamera.flipHorizontally)
}
