package conf

import (
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

func TestSanitizeQuality(t *testing.T) {
	// Test quality=0 gets sanitized to minimum
	s := Settings{Quality: 0}
	sanitized := s.Sanitize()
	ensure.DeepEqual(t, MinQuality, sanitized.Quality)

	// Test negative quality gets sanitized to minimum
	s = Settings{Quality: -10}
	sanitized = s.Sanitize()
	ensure.DeepEqual(t, MinQuality, sanitized.Quality)

	// Test quality > 100 gets sanitized to maximum
	s = Settings{Quality: 150}
	sanitized = s.Sanitize()
	ensure.DeepEqual(t, MaxQuality, sanitized.Quality)

	// Test valid quality is unchanged
	s = Settings{Quality: 75}
	sanitized = s.Sanitize()
	ensure.DeepEqual(t, 75, sanitized.Quality)
}

func TestSanitizeSecondsBetweenCaptures(t *testing.T) {
	// Test 0 seconds gets sanitized to minimum
	s := Settings{SecondsBetweenCaptures: 0}
	sanitized := s.Sanitize()
	ensure.DeepEqual(t, MinSecondsBetweenCaptures, sanitized.SecondsBetweenCaptures)

	// Test negative seconds gets sanitized to minimum
	s = Settings{SecondsBetweenCaptures: -100}
	sanitized = s.Sanitize()
	ensure.DeepEqual(t, MinSecondsBetweenCaptures, sanitized.SecondsBetweenCaptures)

	// Test value less than minimum gets sanitized
	s = Settings{SecondsBetweenCaptures: 5}
	sanitized = s.Sanitize()
	ensure.DeepEqual(t, MinSecondsBetweenCaptures, sanitized.SecondsBetweenCaptures)

	// Test valid value is unchanged
	s = Settings{SecondsBetweenCaptures: 1800}
	sanitized = s.Sanitize()
	ensure.DeepEqual(t, 1800, sanitized.SecondsBetweenCaptures)
}

func TestSanitizeOffsetWithinHour(t *testing.T) {
	// Test -1 (disabled) is allowed and unchanged
	s := Settings{OffsetWithinHour: -1}
	sanitized := s.Sanitize()
	ensure.DeepEqual(t, -1, sanitized.OffsetWithinHour)

	// Test negative value other than -1 gets sanitized to minimum
	s = Settings{OffsetWithinHour: -100}
	sanitized = s.Sanitize()
	ensure.DeepEqual(t, MinOffsetWithinHour, sanitized.OffsetWithinHour)

	// Test value > 3599 gets sanitized to maximum
	s = Settings{OffsetWithinHour: 5000}
	sanitized = s.Sanitize()
	ensure.DeepEqual(t, MaxOffsetWithinHour, sanitized.OffsetWithinHour)

	// Test valid value is unchanged
	s = Settings{OffsetWithinHour: 900}
	sanitized = s.Sanitize()
	ensure.DeepEqual(t, 900, sanitized.OffsetWithinHour)

	// Test boundary values
	s = Settings{OffsetWithinHour: 0}
	sanitized = s.Sanitize()
	ensure.DeepEqual(t, 0, sanitized.OffsetWithinHour)

	s = Settings{OffsetWithinHour: 3599}
	sanitized = s.Sanitize()
	ensure.DeepEqual(t, 3599, sanitized.OffsetWithinHour)
}
