package conf

import (
	"os"
	"testing"

	"github.com/facebookgo/ensure"
)

func TestAreSettingsMissing(t *testing.T) {
	ensure.True(t, areSettingsMissing("non-existing-file"))
	ensure.False(t, areSettingsMissing("settings_test.go"))
}

func TestSettingsToString(t *testing.T) {
	expected := `{"SecondsBetweenCaptures":0,"OffsetWithinHour":0,"PhotoResolutionWidth":0,"PhotoResolutionHeight":0,"PreviewResolutionWidth":0,"PreviewResolutionHeight":0,"RotateBy":0,"ResolutionSetting":0,"Quality":0,"DebugEnabled":false}`
	ensure.DeepEqual(t, expected, Settings{}.String())
}

func TestWriteAndLoadConfiguration(t *testing.T) {
	// Use a temp file for the settings DB to avoid polluting the working directory
	origSettingsFile := settingsFile
	defer func() {
		// We can't reassign the const, so we clean up the temp file
		os.Remove("test-settings.db")
	}()

	// Remove any leftover test DB
	os.Remove("test-settings.db")

	s := Settings{
		SecondsBetweenCaptures:  120,
		OffsetWithinHour:        300,
		PhotoResolutionWidth:    1920,
		PhotoResolutionHeight:   1080,
		PreviewResolutionWidth:  640,
		PreviewResolutionHeight: 480,
		RotateBy:                180,
		ResolutionSetting:       1,
		Quality:                 85,
		DebugEnabled:            true,
	}

	// settingsFile is a const so we test via the public API which uses it
	// Write then load to verify round-trip
	_ = origSettingsFile // acknowledge we can't change the const

	written, err := WriteConfiguration(s)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, s.Quality, written.Quality)
	ensure.DeepEqual(t, s.SecondsBetweenCaptures, written.SecondsBetweenCaptures)

	loaded, err := LoadConfiguration()
	ensure.Nil(t, err)
	ensure.DeepEqual(t, s.Quality, loaded.Quality)
	ensure.DeepEqual(t, s.SecondsBetweenCaptures, loaded.SecondsBetweenCaptures)
	ensure.DeepEqual(t, s.OffsetWithinHour, loaded.OffsetWithinHour)
	ensure.DeepEqual(t, s.PhotoResolutionWidth, loaded.PhotoResolutionWidth)
	ensure.DeepEqual(t, s.RotateBy, loaded.RotateBy)
	ensure.DeepEqual(t, s.DebugEnabled, loaded.DebugEnabled)
}

func TestLoadConfigurationEnforcesMinInterval(t *testing.T) {
	// Remove any leftover DB to trigger fresh creation
	os.Remove(settingsFile)
	defer os.Remove(settingsFile)

	// Write with a very short interval
	s := Settings{
		SecondsBetweenCaptures: 10, // Below minimum of 60
		Quality:                100,
	}
	_, err := WriteConfiguration(s)
	ensure.Nil(t, err)

	// Load should enforce the minimum
	loaded, err := LoadConfiguration()
	ensure.Nil(t, err)
	ensure.True(t, loaded.SecondsBetweenCaptures >= minSecondsBetweenCaptures)
}

func TestLoadConfigurationCreatesInitialSettings(t *testing.T) {
	os.Remove(settingsFile)
	defer os.Remove(settingsFile)

	loaded, err := LoadConfiguration()
	ensure.Nil(t, err)
	ensure.DeepEqual(t, initialConfiguration.Quality, loaded.Quality)
	ensure.DeepEqual(t, initialConfiguration.PhotoResolutionWidth, loaded.PhotoResolutionWidth)
}
