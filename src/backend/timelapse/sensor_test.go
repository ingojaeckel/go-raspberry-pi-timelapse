package timelapse

import (
	"context"
	"errors"
	"slices"
	"testing"
	"testing/synctest"
	"time"
)

const listIMX290 = `Available cameras
-----------------
0 : imx290 [1920x1080 12-bit RGGB] (/base/soc/i2c0mux/i2c@1/imx290@1a)
    Modes: 'SRGGB10_CSI2P' : 1280x720 [30.00 fps - (65535, 65535)/65535x65535 crop]
                             1920x1080 [30.00 fps - (65535, 65535)/65535x65535 crop]
           'SRGGB12_CSI2P' : 1280x720 [30.00 fps - (65535, 65535)/65535x65535 crop]
                             1920x1080 [30.00 fps - (65535, 65535)/65535x65535 crop]
`

const listTwo = `Available cameras
-----------------
0 : imx219 [3280x2464 10-bit RGGB] (/base/axi/pcie@120000/rp1/i2c@88000/imx219@10)
    Modes: 'SRGGB10_CSI2P' : 640x480 [103.33 fps - (1000, 752)/1280x960 crop]
                             1640x1232 [41.85 fps - (0, 0)/3280x2464 crop]
                             1920x1080 [47.57 fps - (680, 692)/1920x1080 crop]
                             3280x2464 [21.19 fps - (0, 0)/3280x2464 crop]
1 : imx708_wide [4608x2592 10-bit RGGB] (/base/axi/pcie@120000/rp1/i2c@80000/imx708@1a)
    Modes: 'SRGGB10_CSI2P' : 1536x864 [120.13 fps - (0, 0)/4608x2592 crop]
                             2304x1296 [56.03 fps - (0, 0)/4608x2592 crop]
                             4608x2592 [14.35 fps - (0, 0)/4608x2592 crop]
`

func TestParseCameras(t *testing.T) {
	got := ParseCameras(listTwo)
	if len(got) != 2 {
		t.Fatalf("got %+v", got)
	}
	if s := got[0]; s.Model != "imx219" || s.Width != 3280 || s.Height != 2464 ||
		!slices.Equal(s.Resolutions(), []Size{{3280, 2464}, {1920, 1080}, {1640, 1232}, {640, 480}}) {
		t.Errorf("first: %+v %v", s, s.Resolutions())
	}
	if s := got[1]; s.Model != "imx708_wide" || !slices.Equal(s.Resolutions(), []Size{{4608, 2592}, {2304, 1296}, {1536, 864}}) {
		t.Errorf("second: %+v %v", s, s.Resolutions())
	}
}

func TestParseCamerasMergesModesOfEveryFormat(t *testing.T) {
	got := ParseCameras(listIMX290)
	if len(got) != 1 || !slices.Equal(got[0].Resolutions(), []Size{{1920, 1080}, {1280, 720}}) {
		t.Fatalf("got %+v", got)
	}
}

func TestParseCamerasWithoutACamera(t *testing.T) {
	if got := ParseCameras("No cameras available!\n"); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestSensorWithoutModesOffersItsFullSize(t *testing.T) {
	s := Sensor{Model: "x", Width: 800, Height: 600}
	if !slices.Equal(s.Resolutions(), []Size{{800, 600}}) {
		t.Fatal(s.Resolutions())
	}
}

func TestDefaultResolutionIsTheLargestFullViewUpTo4MP(t *testing.T) {
	cams := ParseCameras(listTwo)
	for _, tc := range []struct {
		s    Sensor
		want Size
	}{
		{cams[0], Size{1640, 1232}}, // 1920x1080 is larger but a 16:9 crop of the 4:3 sensor
		{cams[1], Size{2304, 1296}},
		{ParseCameras(listIMX290)[0], Size{1920, 1080}},
		{Sensor{Width: 9000, Height: 9000}, Size{9000, 9000}}, // nothing that small: the smallest
	} {
		if got := tc.s.DefaultResolution(); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.s.Model, got, tc.want)
		}
	}
}

func TestSupports(t *testing.T) {
	s := ParseCameras(listIMX290)[0]
	if !s.Supports(1280, 720) || s.Supports(2186, 1642) {
		t.Fatal("Supports")
	}
}

func TestSensorsCachesADetectedCamera(t *testing.T) {
	calls := 0
	d := NewSensors(func(context.Context, string, ...string) ([]byte, error) {
		calls++
		return []byte(listIMX290), nil
	})
	for range 3 {
		got, err := d.Get(t.Context())
		if err != nil || got.Model != "imx290" {
			t.Fatalf("got %+v, %v", got, err)
		}
	}
	if calls != 1 {
		t.Fatalf("listed %d times", calls)
	}
}

func TestSensorsRetriesAfterAMinute(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		out, calls := "", 0
		d := NewSensors(func(context.Context, string, ...string) ([]byte, error) {
			calls++
			return []byte(out), nil
		})
		if _, err := d.Get(t.Context()); !errors.Is(err, ErrNoCamera) {
			t.Fatalf("err %v", err)
		}
		out = listIMX290
		if _, err := d.Get(t.Context()); !errors.Is(err, ErrNoCamera) || calls != 1 {
			t.Fatalf("retried at once: err %v, %d calls", err, calls)
		}
		time.Sleep(time.Minute)
		if s, err := d.Get(t.Context()); err != nil || s.Model != "imx290" {
			t.Fatalf("after a minute: %+v %v", s, err)
		}
	})
}

func TestSensorsReportsTheCommandError(t *testing.T) {
	d := NewSensors(func(context.Context, string, ...string) ([]byte, error) {
		return nil, errors.New("exec: rpicam-still: not found")
	})
	if _, err := d.Get(t.Context()); err == nil || err.Error() != "rpicam-still --list-cameras: exec: rpicam-still: not found" {
		t.Fatalf("err %v", err)
	}
}
