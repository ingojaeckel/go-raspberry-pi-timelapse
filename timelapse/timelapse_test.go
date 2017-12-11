package timelapse

import (
	"github.com/facebookgo/ensure"
	"testing"
	"time"
)

func TestNoSleepTillBrooklyn(t *testing.T) {
	tl := createTimelapseForTesting(0)

	before := time.Now().Unix()
	tl.WaitForFirstCapture()
	ensure.DeepEqual(t, time.Now().Unix()-before, int64(0))
}

func TestGetInitialSleepTime(t *testing.T) {
	testInitialSleepTime(t, 9, 13, 47+5*60)
	testInitialSleepTime(t, 8, 0, 7*60)
	testInitialSleepTime(t, 16, 0, 29*60)
	testInitialSleepTime(t, 30, 59, 14*60+1)
}

func testInitialSleepTime(t *testing.T, minute int, second int, expectedSleepTime int) {
	tl := createTimelapseForTesting(900)
	theTime := time.Date(2017, time.January, 1, 1, minute, second, 0, time.UTC)

	ensure.DeepEqual(t, tl.SecondsToSleepUntilOffset(theTime), expectedSleepTime)
}

func createTimelapseForTesting(offsetWithinHourSeconds int64) Timelapse {
	res := Resolution{800, 600}
	c := NewCamera("timelapse-pictures", res.Width, res.Height)
	return Timelapse{*c, "timelapse-pictures", 1800, offsetWithinHourSeconds, res}
}
