package timelapse

import (
	"testing"
	"time"

	"github.com/facebookgo/ensure"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
)

type abbrevTime struct {
	year, day, hour, min, sec int
	month                     time.Month
	location                  *time.Location
}

func (t abbrevTime) toDate() time.Time {
	return time.Date(t.year, t.month, t.day, t.hour, t.min, t.sec, 0, t.location)
}

func TestSecondsToSleepUntilOffset(t *testing.T) {
	tl := Timelapse{
		Settings: conf.Settings{
			SecondsBetweenCaptures: 60 * 30,
			OffsetWithinHour:       15 * 60,
		},
	}
	s := tl.secondsToSleepUntilOffset(time.Now())
	ensure.True(t, 0 <= s)
	ensure.True(t, s <= tl.Settings.SecondsBetweenCaptures)
}

func TestSecondsToSleepUntilOffset2(t *testing.T) {
	tl := Timelapse{
		Settings: conf.Settings{
			SecondsBetweenCaptures: 60 * 30,
			OffsetWithinHour:       15 * 60,
		},
	}

	loc := time.Now().Location()
	ensure.DeepEqual(t, 29*60-1, tl.secondsToSleepUntilOffset(abbrevTime{year: 2017, month: 12, day: 1, hour: 8, min: 46, sec: 1, location: loc}.toDate()))
	ensure.DeepEqual(t, 13*60-1, tl.secondsToSleepUntilOffset(abbrevTime{year: 2017, month: 12, day: 1, hour: 8, min: 32, sec: 1, location: loc}.toDate()))
	ensure.DeepEqual(t, 29*60-1, tl.secondsToSleepUntilOffset(abbrevTime{year: 2017, month: 12, day: 1, hour: 8, min: 16, sec: 1, location: loc}.toDate()))
	ensure.DeepEqual(t, 7*60-1, tl.secondsToSleepUntilOffset(abbrevTime{year: 2017, month: 12, day: 1, hour: 8, min: 8, sec: 1, location: loc}.toDate()))
}

func TestFewPicsPerDay(t *testing.T) {
	// TODO prepare test for scenarios where pictures are taken more infrequently e.g. once or few times per day.
	tl := Timelapse{
		Settings: conf.Settings{
			SecondsBetweenCaptures: 60 * 30,
			OffsetWithinHour:       0,
		},
	}

	loc := time.Now().Location()
	ensure.DeepEqual(t, tl.secondsToSleepUntilOffset(abbrevTime{year: 2017, month: 12, day: 1, hour: 8, min: 45, location: loc}.toDate()), 15*60)
}

func TestNoSleepTillBrooklyn(t *testing.T) {
	tl := createTimelapseForTesting(t, 0)
	tl.Settings.SecondsBetweenCaptures = 1

	before := time.Now().Unix()
	tl.waitForCapture()
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

	ensure.DeepEqual(t, tl.secondsToSleepUntilOffset(theTime), expectedSleepTime)
}

func createTimelapseForTesting(t *testing.T, offsetWithinHourSeconds int) Timelapse {
	return Timelapse{
		Folder: conf.StorageFolder,
		Settings: conf.Settings{
			SecondsBetweenCaptures: 1800,
			OffsetWithinHour:       offsetWithinHourSeconds,
			DebugEnabled:           true,
			PhotoResolutionWidth:   800,
			PhotoResolutionHeight:  600,
		},
	}
}
