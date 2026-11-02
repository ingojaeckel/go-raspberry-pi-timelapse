package rest

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/facebookgo/ensure"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
)

func TestSafeFilePath(t *testing.T) {
	// Normal file name should work
	p, err := safeFilePath("/tmp/photos", "image.jpg")
	ensure.Nil(t, err)
	ensure.DeepEqual(t, filepath.Join("/tmp/photos", "image.jpg"), p)

	// Path traversal should fail
	_, err = safeFilePath("/tmp/photos", "../etc/passwd")
	ensure.NotNil(t, err)

	_, err = safeFilePath("/tmp/photos", "../../etc/shadow")
	ensure.NotNil(t, err)
}

func TestMakeGetVersionFn(t *testing.T) {
	handler := MakeGetVersionFn("test-v1")
	req := httptest.NewRequest("GET", "/version", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	ensure.DeepEqual(t, http.StatusOK, w.Code)
	ensure.True(t, strings.Contains(w.Body.String(), "test-v1"))
}

func TestGetConfigurationHandler(t *testing.T) {
	// Ensure a settings file exists
	os.Remove("timelapse-settings.db")
	defer os.Remove("timelapse-settings.db")

	req := httptest.NewRequest("GET", "/configuration", nil)
	w := httptest.NewRecorder()

	GetConfiguration(w, req)

	ensure.DeepEqual(t, http.StatusOK, w.Code)
	ensure.True(t, strings.Contains(w.Body.String(), "SecondsBetweenCaptures"))
}

func TestMakeUpdateConfigurationFn(t *testing.T) {
	os.Remove("timelapse-settings.db")
	defer os.Remove("timelapse-settings.db")

	// Initialize settings first
	_, err := conf.LoadConfiguration()
	ensure.Nil(t, err)

	ch := make(chan conf.Settings, 1)
	handler := MakeUpdateConfigurationFn(ch)

	body := `{"SecondsBetweenCaptures":120,"OffsetWithinHour":300,"Quality":85,"ResolutionSetting":1}`
	req := httptest.NewRequest("POST", "/configuration", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler(w, req)

	ensure.DeepEqual(t, http.StatusOK, w.Code)

	// Verify the channel received the update
	select {
	case updated := <-ch:
		ensure.DeepEqual(t, 120, updated.SecondsBetweenCaptures)
		ensure.DeepEqual(t, 85, updated.Quality)
	default:
		t.Fatal("Expected config update on channel")
	}
}

func TestMakeUpdateConfigurationFnInvalidJSON(t *testing.T) {
	ch := make(chan conf.Settings, 1)
	handler := MakeUpdateConfigurationFn(ch)

	req := httptest.NewRequest("POST", "/configuration", strings.NewReader("not json"))
	w := httptest.NewRecorder()

	handler(w, req)

	ensure.DeepEqual(t, http.StatusBadRequest, w.Code)
}

func TestMakeUpdateConfigurationFnInvalidSettings(t *testing.T) {
	ch := make(chan conf.Settings, 1)
	handler := MakeUpdateConfigurationFn(ch)

	// Quality of 0 should fail validation
	body := `{"SecondsBetweenCaptures":120,"Quality":0}`
	req := httptest.NewRequest("POST", "/configuration", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler(w, req)

	ensure.DeepEqual(t, http.StatusBadRequest, w.Code)
}

func TestDeleteFilesNoneProvided(t *testing.T) {
	req := httptest.NewRequest("GET", "/file/delete", nil)
	w := httptest.NewRecorder()

	DeleteFiles(w, req)

	ensure.DeepEqual(t, http.StatusBadRequest, w.Code)
}

func TestGetPhotosHandler(t *testing.T) {
	// Use a temp dir as storage
	origFolder := conf.StorageFolder
	tmpDir, _ := os.MkdirTemp("", "test-photos")
	conf.StorageFolder = tmpDir
	defer func() {
		conf.StorageFolder = origFolder
		os.RemoveAll(tmpDir)
	}()

	req := httptest.NewRequest("GET", "/photos", nil)
	w := httptest.NewRecorder()

	GetPhotos(w, req)

	ensure.DeepEqual(t, http.StatusOK, w.Code)
	ensure.True(t, strings.Contains(w.Body.String(), "Photos"))
}

func TestUpdatePartialConfigurationResolutionMapping(t *testing.T) {
	os.Remove("timelapse-settings.db")
	defer os.Remove("timelapse-settings.db")

	// Initialize settings
	_, err := conf.LoadConfiguration()
	ensure.Nil(t, err)

	// Test resolution setting 2 → 1640x1232
	s, err := updatePartialConfiguration(UpdateConfigurationRequest{
		Settings: conf.Settings{
			SecondsBetweenCaptures: 1800,
			OffsetWithinHour:       900,
			Quality:                100,
			ResolutionSetting:      2,
		},
	})
	ensure.Nil(t, err)
	ensure.DeepEqual(t, 1640, s.PhotoResolutionWidth)
	ensure.DeepEqual(t, 1232, s.PhotoResolutionHeight)

	// Test resolution setting 1 → 2186x1642
	s, err = updatePartialConfiguration(UpdateConfigurationRequest{
		Settings: conf.Settings{
			SecondsBetweenCaptures: 1800,
			OffsetWithinHour:       900,
			Quality:                100,
			ResolutionSetting:      1,
		},
	})
	ensure.Nil(t, err)
	ensure.DeepEqual(t, 2186, s.PhotoResolutionWidth)
	ensure.DeepEqual(t, 1642, s.PhotoResolutionHeight)

	// Test resolution setting 0 → 3280x2464, switching from a previous setting
	s, err = updatePartialConfiguration(UpdateConfigurationRequest{
		Settings: conf.Settings{
			SecondsBetweenCaptures: 1800,
			OffsetWithinHour:       900,
			Quality:                100,
			ResolutionSetting:      0,
		},
	})
	ensure.Nil(t, err)
	ensure.DeepEqual(t, 3280, s.PhotoResolutionWidth)
	ensure.DeepEqual(t, 2464, s.PhotoResolutionHeight)

	// Test unknown resolution setting → falls back to 3280x2464
	s, err = updatePartialConfiguration(UpdateConfigurationRequest{
		Settings: conf.Settings{
			SecondsBetweenCaptures: 1800,
			OffsetWithinHour:       900,
			Quality:                100,
			ResolutionSetting:      1,
		},
	})
	ensure.Nil(t, err)
	s, err = updatePartialConfiguration(UpdateConfigurationRequest{
		Settings: conf.Settings{
			SecondsBetweenCaptures: 1800,
			OffsetWithinHour:       900,
			Quality:                100,
			ResolutionSetting:      99, // unknown → falls through to default
		},
	})
	ensure.Nil(t, err)
	ensure.DeepEqual(t, 3280, s.PhotoResolutionWidth)
	ensure.DeepEqual(t, 2464, s.PhotoResolutionHeight)
}

func TestCurrentSettingsReflectsSavedConfiguration(t *testing.T) {
	os.Remove("timelapse-settings.db")
	defer os.Remove("timelapse-settings.db")

	startup, err := conf.LoadConfiguration()
	ensure.Nil(t, err)
	ensure.DeepEqual(t, 0, startup.RotateBy)

	// Saving new settings after startup must be visible to the preview capture.
	saved := *startup
	saved.RotateBy = 180
	saved.PreviewResolutionWidth = 320
	_, err = conf.WriteConfiguration(saved)
	ensure.Nil(t, err)

	current := currentSettings(startup)
	ensure.DeepEqual(t, 180, current.RotateBy)
	ensure.DeepEqual(t, 320, current.PreviewResolutionWidth)
}
