package timelapse

import (
	"strings"
	"testing"
	"time"

	"github.com/facebookgo/ensure"
)

func TestCreateCameraWithoutPath(t *testing.T) {
	_, err := NewCamera("", 200, 100, false, 100)
	// Should have failed since path must not be empty.
	ensure.NotNil(t, err)
}

func TestBuildingArguments(t *testing.T) {
	unrotatedCamera, err := NewCamera("foo", 200, 100, false, 100)
	ensure.Nil(t, err)
	ensure.False(t, unrotatedCamera.flipVertically)
	ensure.False(t, unrotatedCamera.flipHorizontally)

	p := unrotatedCamera.getAbsoluteFilepath()
	// Example: "foo/20210913-184442.jpg"
	ensure.DeepEqual(t, 0, strings.Index(p, "foo/"))
	ensure.DeepEqual(t, strings.Index(p, ".jpg"), 19)
}

func TestRaspistillArgs(t *testing.T) {
	unrotatedCamera, _ := NewCamera("foo", 200, 100, false, 100)
	args := unrotatedCamera.getRaspistillArgs("foo/someFile.jpg")
	ensure.DeepEqual(t, []string{"-w", "200", "-h", "100", "-q", "100", "-o", "foo/someFile.jpg"}, args)
}

func TestCreateRotatedCamera(t *testing.T) {
	rotatedCamera, err := NewCamera("foo", 200, 100, true, 100)
	ensure.Nil(t, err)
	ensure.True(t, rotatedCamera.flipVertically)
	ensure.True(t, rotatedCamera.flipHorizontally)
}

func TestCreateFileName(t *testing.T) {
	ensure.DeepEqual(t, getFileName(time.Date(1974, time.January, 1, 1, 0, 0, 0, time.UTC)), "19740101-010000.jpg")
	ensure.DeepEqual(t, getFileName(time.Date(1974, time.January, 19, 1, 2, 3, 0, time.UTC)), "19740119-010203.jpg")
	ensure.DeepEqual(t, getFileName(time.Date(1974, time.May, 19, 1, 2, 3, 0, time.UTC)), "19740519-010203.jpg")
	ensure.DeepEqual(t, getFileName(time.Date(1974, time.December, 19, 1, 2, 3, 4, time.UTC)), "19741219-010203.jpg")
}
