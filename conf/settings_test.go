package conf

import (
	"github.com/facebookgo/ensure"
	"testing"
)

func TestAreSettingsMissing(t *testing.T) {
	ensure.True(t, areSettingsMissing("non-existing-file"))
	ensure.False(t, areSettingsMissing("settings_test.go"))
}

func TestGet(t *testing.T) {
	set()
}
