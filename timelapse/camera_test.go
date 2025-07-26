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
	ensure.DeepEqual(t, []string{"--width", "200", "--height", "100", "--quality", "100", "--output", "foo/someFile.jpg"}, args)
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

func TestIsNightTime(t *testing.T) {
	tenPM := time.Date(2000, time.January, 1, 22, 0, 0, 0, time.UTC)
	elevenPM := time.Date(2000, time.January, 1, 23, 0, 0, 0, time.UTC)
	fourAM := time.Date(2000, time.January, 1, 4, 0, 0, 0, time.UTC)
	sixAM := time.Date(2000, time.January, 1, 6, 0, 0, 0, time.UTC)

	ensure.True(t, isNightPhoto(elevenPM))
	ensure.True(t, isNightPhoto(fourAM))

	ensure.True(t, isNightPhoto(tenPM))
	ensure.False(t, isNightPhoto(sixAM))
}
