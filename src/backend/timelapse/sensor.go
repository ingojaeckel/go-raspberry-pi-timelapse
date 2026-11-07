package timelapse

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Size is a photo size in pixels.
type Size struct{ Width, Height int }

func (s Size) String() string { return fmt.Sprintf("%dx%d", s.Width, s.Height) }

func (s Size) pixels() int { return s.Width * s.Height }

// Sensor is a camera as listed by `rpicam-still --list-cameras`.
type Sensor struct {
	Model         string
	Width, Height int
	// Modes are the sizes the sensor reads out natively (binned or
	// cropped), in the order listed.
	Modes []Size
}

// defaultMaxPixels caps the default photo size: larger photos fill the card
// faster and take longer to write, with little gain for a timelapse.
const defaultMaxPixels = 4_000_000

var (
	cameraHeader = regexp.MustCompile(`^\s*\d+ : (\S+) \[(\d+)x(\d+)`)
	cameraMode   = regexp.MustCompile(`(\d+)x(\d+) \[`)
)

// ParseCameras reads the output of `rpicam-still --list-cameras`:
//
//	0 : imx290 [1920x1080 12-bit RGGB] (/base/soc/i2c0mux/i2c@1/imx290@1a)
//	    Modes: 'SRGGB10_CSI2P' : 1280x720 [30.00 fps - (65535, 65535)/65535x65535 crop]
//	                             1920x1080 [30.00 fps - ...]
func ParseCameras(out string) []Sensor {
	var sensors []Sensor
	for line := range strings.Lines(out) {
		if m := cameraHeader.FindStringSubmatch(line); m != nil {
			full := parseSize(m[2], m[3])
			sensors = append(sensors, Sensor{Model: m[1], Width: full.Width, Height: full.Height})
			continue
		}
		if len(sensors) == 0 {
			continue
		}
		if m := cameraMode.FindStringSubmatch(line); m != nil {
			sensor, mode := &sensors[len(sensors)-1], parseSize(m[1], m[2])
			if !slices.Contains(sensor.Modes, mode) {
				sensor.Modes = append(sensor.Modes, mode)
			}
		}
	}
	return sensors
}

// parseSize parses digits matched by the regular expressions above.
func parseSize(width, height string) Size {
	w, _ := strconv.Atoi(width)
	h, _ := strconv.Atoi(height)
	return Size{w, h}
}

// Resolutions are the photo sizes offered for this sensor: its modes and its
// full size, largest first.
func (s Sensor) Resolutions() []Size {
	out := slices.Clone(s.Modes)
	if full := (Size{s.Width, s.Height}); !slices.Contains(out, full) {
		out = append(out, full)
	}
	slices.SortStableFunc(out, func(a, b Size) int { return cmp.Compare(b.pixels(), a.pixels()) })
	return out
}

// ResolutionList is the Resolutions for messages, e.g. "1920x1080, 1280x720".
func (s Sensor) ResolutionList() string {
	var sizes []string
	for _, r := range s.Resolutions() {
		sizes = append(sizes, r.String())
	}
	return strings.Join(sizes, ", ")
}

// Supports reports whether width x height is one of the Resolutions.
func (s Sensor) Supports(width, height int) bool {
	return slices.Contains(s.Resolutions(), Size{width, height})
}

// DefaultResolution is the largest resolution of at most 4 MP that shows the
// sensor's whole field of view (its aspect ratio), else the largest of at
// most 4 MP, else the smallest.
func (s Sensor) DefaultResolution() Size {
	all := s.Resolutions()
	fullView := func(r Size) bool {
		return abs(r.Width*s.Height-r.Height*s.Width)*50 <= s.Width*s.Height // within 2%
	}
	small := func(r Size) bool { return r.pixels() <= defaultMaxPixels }
	for _, ok := range []func(Size) bool{
		func(r Size) bool { return small(r) && fullView(r) },
		small,
	} {
		if i := slices.IndexFunc(all, ok); i >= 0 {
			return all[i]
		}
	}
	return all[len(all)-1]
}

func abs(v int) int { return max(v, -v) }

// SensorLister is implemented by backends that can list their cameras
// (rpicam-still).
type SensorLister interface {
	Sensors() *Sensors
}

// ErrNoCamera means rpicam-still lists no camera.
var ErrNoCamera = errors.New("no camera detected")

// ListCommand runs a command and returns its output (CommandOutput).
type ListCommand func(ctx context.Context, name string, args ...string) ([]byte, error)

// Sensors detects the connected camera once and remembers it. A failed
// detection (camera not connected yet, rpicam-still missing) is retried at
// most once a minute. It is safe for concurrent use.
type Sensors struct {
	run ListCommand

	mu      sync.Mutex
	sensor  *Sensor
	err     error
	checked time.Time
}

// sensorRetry is how long a failed detection is remembered.
const sensorRetry = time.Minute

// NewSensors detects cameras with run (`rpicam-still --list-cameras`).
func NewSensors(run ListCommand) *Sensors { return &Sensors{run: run} }

// Get returns the first camera.
func (d *Sensors) Get(ctx context.Context) (Sensor, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.sensor != nil {
		return *d.sensor, nil
	}
	if d.err != nil && time.Since(d.checked) < sensorRetry {
		return Sensor{}, d.err
	}
	d.checked = time.Now()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	out, err := d.run(ctx, commandRpicamStill, "--list-cameras")
	sensors := ParseCameras(string(out))
	switch {
	case len(sensors) > 0:
		d.sensor, d.err = &sensors[0], nil
		return sensors[0], nil
	case err != nil:
		d.err = fmt.Errorf("%s --list-cameras: %w", commandRpicamStill, err)
	default:
		d.err = ErrNoCamera
	}
	return Sensor{}, d.err
}
