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
