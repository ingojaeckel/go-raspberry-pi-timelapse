package timelapse

import (
	"github.com/facebookgo/ensure"
	"testing"
	"time"
)

func TestSecondsToSleepUntilOffset(t *testing.T) {
	tl := Timelapse{SecondsBetweenCapture: 60 * 30, OffsetWithinHourSeconds: 15 * 60}
	s := tl.SecondsToSleepUntilOffset(time.Now())
	ensure.True(t, 0 <= s)
	ensure.True(t, s <= int(tl.SecondsBetweenCapture))
}

func TestSecondsToSleepUntilOffset2(t *testing.T) {
	tl := Timelapse{SecondsBetweenCapture: 60 * 30, OffsetWithinHourSeconds: 15 * 60}

	l := time.Now().Location()
	hour := 8
	min := 46
	sec := 1
	ensure.DeepEqual(t, 29*60-1, tl.SecondsToSleepUntilOffset(time.Date(2017, 12, 1, hour, min, sec, 0, l)))

	min2 := 32
	sec2 := 1
	ensure.DeepEqual(t, 13*60-1, tl.SecondsToSleepUntilOffset(time.Date(2017, 12, 1, hour, min2, sec2, 0, l)))

	min3 := 16
	sec3 := 1
	ensure.DeepEqual(t, 29*60-1, tl.SecondsToSleepUntilOffset(time.Date(2017, 12, 1, hour, min3, sec3, 0, l)))

	min4 := 8
	sec4 := 1
	ensure.DeepEqual(t, 7*60-1, tl.SecondsToSleepUntilOffset(time.Date(2017, 12, 1, hour, min4, sec4, 0, l)))
}

func TestNoSleepTillBrooklyn(t *testing.T) {
	tl := createTimelapseForTesting(t, 0)
	tl.SecondsBetweenCapture = 1

	before := time.Now().Unix()
	tl.WaitForCapture()
	ensure.DeepEqual(t, time.Now().Unix()-before, int64(0))
}

func TestGetInitialSleepTime(t *testing.T) {
	testInitialSleepTime(t, 9, 13, 47+5*60)
	testInitialSleepTime(t, 8, 0, 7*60)
	testInitialSleepTime(t, 16, 0, 29*60)
	testInitialSleepTime(t, 30, 59, 14*60+1)
}

func testInitialSleepTime(t *testing.T, minute int, second int, expectedSleepTime int) {
	tl := createTimelapseForTesting(t, 900)
	theTime := time.Date(2017, time.January, 1, 1, minute, second, 0, time.UTC)

	ensure.DeepEqual(t, tl.SecondsToSleepUntilOffset(theTime), expectedSleepTime)
}

func createTimelapseForTesting(t *testing.T, offsetWithinHourSeconds int) Timelapse {
	res := Resolution{800, 600}
	c, err := NewCamera("timelapse-pictures", res.Width, res.Height, false)
	ensure.Nil(t, err)
	return Timelapse{c, "timelapse-pictures", 1800, offsetWithinHourSeconds, res, true}
}
