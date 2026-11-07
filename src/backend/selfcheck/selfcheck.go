// Package selfcheck tests a camera end to end once the app is installed: it
// takes a real photo with the configured settings, checks that the file is a
// usable JPEG that is not black or blank, and looks for the setup problems
// that make captures fail on a Raspberry Pi (no camera detected, the camera
// held by another program, a full card, bad power, overheating).
package selfcheck

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/health"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/sun"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
	berrors "go.etcd.io/bbolt/errors"
)

// Status is the outcome of one check. Only Fail makes the self-check fail.
type Status int

const (
	Pass Status = iota
	Skip        // does not apply here, e.g. no vcgencmd off a Pi
	Warn
	Fail
)

func (s Status) String() string {
	return [...]string{"PASS", "SKIP", "WARN", "FAIL"}[s]
}

// Result is the outcome of one check. Hint says what to change.
type Result struct {
	Check  string
	Status Status
	Detail string
	Hint   string
}

// Config is what the self-check needs. The system access fields are
// replaced in tests.
type Config struct {
	// Backend takes the photo. Its String() is "rpicam-still" on a Pi.
	Backend        timelapse.Backend
	CaptureTimeout time.Duration
	StorageFolder  string
	// SettingsFile is read, never written; DefaultSettings apply when
	// nothing is saved or the running app holds the file.
	SettingsFile    string
	DefaultSettings conf.Settings
	MinFreeBytes    uint64
	// MinBrightness is the lowest mean brightness (0..255) of a usable photo.
	MinBrightness float64
	// Keep leaves the test photo on disk and reports its path.
	Keep bool

	FreeBytes   func(dir string) (uint64, error)
	Run         health.Run
	ThermalFile string
	CPUFreqDir  string
	// ProcDir is /proc, scanned for other programs using the camera; empty
	// skips the scan.
	ProcDir string
	// Now is the time of the test photo, to tell night at the saved location
	// from a covered lens; nil means time.Now.
	Now func() time.Time
}

const rpicamStill = "rpicam-still"

// minContrast is the lowest standard deviation of brightness in a usable
// photo; a camera that returns a flat frame (all black, white or grey) is
// below it.
const minContrast = 2

// Run runs every check in order and returns their results.
func Run(ctx context.Context, cfg Config) []Result {
	pi := fmt.Sprint(cfg.Backend) == rpicamStill
	sensor, detectErr := detect(ctx, cfg, pi)
	if detectErr == nil {
		// Like the app on its first run: the camera's default size.
		d := sensor.DefaultResolution()
		cfg.DefaultSettings.PhotoResolutionWidth, cfg.DefaultSettings.PhotoResolutionHeight = d.Width, d.Height
	}
	settings, settingsResult := readSettings(cfg)
	opts := timelapse.PhotoOptions(settings)
	results := []Result{settingsResult, checkStorage(cfg), checkDetected(pi, sensor, detectErr, opts), checkCameraFree(cfg, pi)}
	if why := cannotCapture(results); why != "" {
		results = append(results, skipped("capture", why), skipped("photo", "no photo taken"))
	} else {
		results = append(results, testPhoto(ctx, cfg, settings, opts)...)
	}
	return append(results, checkPower(ctx, cfg), checkTemperature(cfg), checkCPU(cfg))
}

// captureNeeds are the checks a test photo needs, with what to say when one
// fails.
var captureNeeds = map[string]string{
	"storage":         "no storage folder to write to",
	"camera detected": "no camera",
	"camera free":     "the camera is in use",
}

// cannotCapture says why the results rule out a test photo, or "".
func cannotCapture(results []Result) string {
	for _, r := range results {
		if why, needed := captureNeeds[r.Check]; needed && r.Status == Fail {
			return why
		}
	}
	return ""
}

// testPhoto takes a photo and checks it: the "capture" and "photo" results.
func testPhoto(ctx context.Context, cfg Config, settings conf.Settings, opts timelapse.Options) []Result {
	photo, captured := capture(ctx, cfg, opts)
	if photo == "" {
		return []Result{captured, skipped("photo", "no photo taken")}
	}
	if !cfg.Keep {
		defer os.RemoveAll(filepath.Dir(photo))
	}
	return []Result{captured, checkPhoto(cfg, settings, photo, opts)}
}

// Write prints the results and a summary line, and reports whether no check
// failed.
func Write(w io.Writer, results []Result) bool {
	var failed, warned int
	for _, r := range results {
		fmt.Fprintf(w, "%s  %-16s %s\n", r.Status, r.Check, r.Detail)
		if r.Hint != "" {
			fmt.Fprintf(w, "      %-16s → %s\n", "", r.Hint)
		}
		switch r.Status {
		case Fail:
			failed++
		case Warn:
			warned++
		}
	}
	switch {
	case failed > 0:
		fmt.Fprintf(w, "\nSelf-check failed: %s, %s.\n", plural(failed, "problem"), plural(warned, "warning"))
	case warned > 0:
		fmt.Fprintf(w, "\nThe camera works, with %s.\n", plural(warned, "warning"))
	default:
		fmt.Fprintln(w, "\nAll checks passed.")
	}
	return failed == 0
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return strconv.Itoa(n) + " " + word + "s"
}

func skipped(check, why string) Result {
	return Result{Check: check, Status: Skip, Detail: why}
}

func describe(s conf.Settings) string {
	d := fmt.Sprintf("%dx%d, quality %d", s.PhotoResolutionWidth, s.PhotoResolutionHeight, s.Quality)
	if s.RotateBy == 180 {
		d += ", rotated 180°"
	}
	return d
}

func readSettings(cfg Config) (conf.Settings, Result) {
	r := Result{Check: "settings"}
	s, err := conf.ReadSettings(cfg.SettingsFile)
	switch {
	case err == nil:
		r.Detail = "saved settings: " + describe(s)
		return s, r
	case errors.Is(err, os.ErrNotExist):
		r.Detail = "nothing saved yet"
	case errors.Is(err, berrors.ErrTimeout):
		r.Status = Warn
		r.Detail = "the timelapse app is running (it holds " + cfg.SettingsFile + ")"
		r.Hint = "a scheduled capture can collide with this test photo; stop the app for a clean check (in the image: sudo systemctl stop timelapse)"
	default:
		r.Status = Warn
		r.Detail = fmt.Sprintf("cannot read %s (%v)", cfg.SettingsFile, err)
	}
	r.Detail += ", using the defaults: " + describe(cfg.DefaultSettings)
	return cfg.DefaultSettings, r
}

func checkStorage(cfg Config) Result {
	r := Result{Check: "storage"}
	if err := writable(cfg.StorageFolder); err != nil {
		r.Status, r.Detail = Fail, fmt.Sprintf("cannot write to %s: %v", cfg.StorageFolder, err)
		r.Hint = "check -storageFolder, its owner and that the card is not read-only"
		return r
	}
	free, err := cfg.FreeBytes(cfg.StorageFolder)
	if err != nil {
		r.Status, r.Detail = Warn, fmt.Sprintf("%s is writable, free space unknown: %v", cfg.StorageFolder, err)
		return r
	}
	r.Detail = fmt.Sprintf("%s is writable, %s free", cfg.StorageFolder, formatBytes(free))
	if free < cfg.MinFreeBytes {
		r.Status = Warn
		r.Hint = fmt.Sprintf("below the %s the app keeps free: it deletes the oldest photos, or stops taking photos (Settings)", formatBytes(cfg.MinFreeBytes))
	}
	return r
}

// writable creates dir if needed, and a file in it, which it deletes again.
func writable(dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".selfcheck-")
	if err != nil {
		return err
	}
	f.Close()
	return os.Remove(f.Name())
}

func formatBytes(b uint64) string {
	if b >= 1<<30 {
		return fmt.Sprintf("%.1f GB", float64(b)/(1<<30))
	}
	return fmt.Sprintf("%d MB", b>>20)
}

// detect lists the cameras on a Pi; errNotPi elsewhere.
func detect(ctx context.Context, cfg Config, pi bool) (timelapse.Sensor, error) {
	if !pi {
		return timelapse.Sensor{}, errNotPi
	}
	ctx, cancel := context.WithTimeout(ctx, cfg.CaptureTimeout)
	defer cancel()
	return timelapse.NewSensors(timelapse.ListCommand(cfg.Run)).Get(ctx)
}

var errNotPi = errors.New("not a Raspberry Pi camera")

func checkDetected(pi bool, sensor timelapse.Sensor, err error, o timelapse.Options) Result {
	r := Result{Check: "camera detected"}
	if !pi {
		return skipped(r.Check, "not a Raspberry Pi camera")
	}
	if err != nil {
		r.Status = Fail
		r.Detail = "rpicam-still --list-cameras finds no camera"
		if !errors.Is(err, timelapse.ErrNoCamera) {
			r.Detail += fmt.Sprintf(" (%v)", err)
		}
		r.Hint = "check that the ribbon cable is seated firmly at both ends, contacts facing the board. " +
			"Raspberry Pi camera modules need camera_auto_detect=1 in /boot/firmware/config.txt; " +
			"third-party sensors such as Arducam need their overlay, e.g. dtoverlay=imx290 or dtoverlay=imx519, and a reboot"
		return r
	}
	r.Detail = fmt.Sprintf("%s (%dx%d)", sensor.Model, sensor.Width, sensor.Height)
	if !sensor.Supports(o.Width, o.Height) {
		r.Status = Warn
		r.Detail += fmt.Sprintf(": photos of %dx%d are not one of its sizes (%s) and get scaled", o.Width, o.Height, sensor.ResolutionList())
		r.Hint = "choose one of the camera's resolutions in Settings"
	}
	return r
}

// cameraDevice is what libcamera opens: the sensor's media and video nodes.
var cameraDevice = regexp.MustCompile(`^/dev/(media|video)\d+$`)

func checkCameraFree(cfg Config, pi bool) Result {
	r := Result{Check: "camera free"}
	if !pi || cfg.ProcDir == "" {
		return skipped(r.Check, "not a Raspberry Pi camera")
	}
	holders := cameraHolders(cfg.ProcDir)
	if len(holders) == 0 {
		r.Detail = "no other program has the camera open"
		return r
	}
	r.Status = Fail
	var described, units []string
	for _, h := range holders {
		described = append(described, h.String())
		if h.unit != "" {
			units = append(units, h.unit)
		}
	}
	r.Detail = "in use by " + strings.Join(described, "; ")
	r.Hint = "only one program can use the camera; stop the other one"
	if len(units) > 0 {
		r.Hint += ": sudo systemctl disable --now " + strings.Join(units, " ")
	}
	return r
}

type holder struct {
	pid, cmdline, unit string
}

func (h holder) String() string {
	s := "pid " + h.pid + " " + h.cmdline
	if h.unit != "" {
		s += " (" + h.unit + ")"
	}
	return s
}

// cameraHolders lists the other processes with a camera device open. It sees
// only the processes the current user may inspect (all of them as root).
func cameraHolders(procDir string) []holder {
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return nil
	}
	self := strconv.Itoa(os.Getpid())
	var holders []holder
	for _, e := range entries {
		pid := e.Name()
		if _, err := strconv.Atoi(pid); err != nil || pid == self {
			continue
		}
		if holdsCamera(filepath.Join(procDir, pid, "fd")) {
			holders = append(holders, readHolder(procDir, pid))
		}
	}
	return holders
}

// readHolder reads the command line and systemd unit of process pid.
func readHolder(procDir, pid string) holder {
	h := holder{pid: pid}
	if b, err := os.ReadFile(filepath.Join(procDir, pid, "cmdline")); err == nil {
		h.cmdline = strings.TrimSpace(strings.ReplaceAll(string(b), "\x00", " "))
	}
	if b, err := os.ReadFile(filepath.Join(procDir, pid, "cgroup")); err == nil {
		h.unit = serviceUnit(string(b))
	}
	return h
}

func holdsCamera(fdDir string) bool {
	fds, err := os.ReadDir(fdDir)
	if err != nil {
		return false
	}
	for _, fd := range fds {
		if target, err := os.Readlink(filepath.Join(fdDir, fd.Name())); err == nil && cameraDevice.MatchString(target) {
			return true
		}
	}
	return false
}

// serviceUnit finds the systemd service in /proc/<pid>/cgroup, e.g.
// "0::/system.slice/surveillance-sender.service".
func serviceUnit(cgroup string) string {
	for part := range strings.SplitSeq(strings.TrimSpace(cgroup), "/") {
		if strings.HasSuffix(part, ".service") {
			return part
		}
	}
	return ""
}

func capture(ctx context.Context, cfg Config, o timelapse.Options) (string, Result) {
	r := Result{Check: "capture"}
	dir, err := os.MkdirTemp(cfg.StorageFolder, ".selfcheck-")
	if err != nil {
		r.Status, r.Detail = Fail, err.Error()
		return "", r
	}
	start := time.Now()
	path, err := timelapse.NewCamera(cfg.Backend, cfg.CaptureTimeout).Capture(ctx, dir, o)
	if err != nil {
		os.RemoveAll(dir)
		r.Status, r.Detail, r.Hint = Fail, lastLines(err.Error(), 3), captureHint(err.Error())
		return "", r
	}
	r.Detail = fmt.Sprintf("%v took a photo in %.1f s", cfg.Backend, time.Since(start).Seconds())
	if fi, err := os.Stat(path); err == nil {
		r.Detail += fmt.Sprintf(" (%d KB)", fi.Size()>>10)
	}
	return path, r
}

// lastLines keeps the end of a command's output, where the error is.
func lastLines(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, " | ")
}

func captureHint(out string) string {
	switch {
	case strings.Contains(out, "in use by another process"), strings.Contains(out, "Device or resource busy"):
		return "another program has the camera open; run the self-check with sudo to see which"
	case strings.Contains(out, "no cameras available"):
		return "no camera detected: check the ribbon cable, and for third-party sensors the dtoverlay= line in /boot/firmware/config.txt"
	case strings.Contains(out, "deadline exceeded"), strings.Contains(out, "signal: killed"):
		return "the capture did not finish in time: check the cable, power supply and that no other program uses the camera"
	case strings.Contains(out, "executable file not found"):
		return "install the camera tools: sudo apt install rpicam-apps-lite"
	}
	return ""
}

func checkPhoto(cfg Config, set conf.Settings, path string, o timelapse.Options) Result {
	r := Result{Check: "photo"}
	img, err := decodeJPEG(path)
	if err != nil {
		r.Status, r.Detail = Fail, err.Error()
		return r
	}
	size := img.Bounds().Size()
	mean, sd := brightness(img)
	r.Detail = fmt.Sprintf("%dx%d JPEG, brightness %.0f of 255, contrast %.0f", size.X, size.Y, mean, sd)
	switch {
	case mean < cfg.MinBrightness:
		r = tooDark(cfg, set, mean)
	case sd < minContrast:
		r.Status = Fail
		r.Detail = fmt.Sprintf("every pixel has the same brightness (%.0f of 255): the sensor returns a blank frame", mean)
		r.Hint = "check the ribbon cable and that the dtoverlay= in /boot/firmware/config.txt matches the sensor"
	case size.X != o.Width || size.Y != o.Height:
		r.Status = Warn
		r.Detail += fmt.Sprintf("; %dx%d was requested", o.Width, o.Height)
	}
	if cfg.Keep {
		r.Detail += "; kept at " + path
	}
	return r
}

func decodeJPEG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := jpeg.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("the file is not a JPEG: %v", err)
	}
	return img, nil
}

// tooDark fails a photo below MinBrightness, unless it is night at the
// saved location: then the lens may well be fine.
func tooDark(cfg Config, set conf.Settings, mean float64) Result {
	r := Result{Check: "photo", Status: Fail,
		Detail: fmt.Sprintf("too dark: brightness %.0f of 255 (at least %.0f expected)", mean, cfg.MinBrightness),
		Hint: "remove the lens cap, point the camera at something lit, and check the lens focus; " +
			"testing in the dark? lower -selfcheckMinBrightness"}
	if rise, night := nightAt(cfg, set); night {
		r.Status = Warn
		r.Detail += "; it is night at the camera's location"
		r.Hint = "run the self-check again after sunrise" + rise
	}
	return r
}

// brightness returns the mean and standard deviation of the luma (0..255)
// over a grid of about 512x512 samples, which is plenty to tell a usable
// photo from a black or blank one.
func brightness(img image.Image) (mean, sd float64) {
	b := img.Bounds()
	stepX, stepY := max(1, b.Dx()/512), max(1, b.Dy()/512)
	luma := func(x, y int) uint8 { return color.GrayModel.Convert(img.At(x, y)).(color.Gray).Y }
	if yc, ok := img.(*image.YCbCr); ok {
		luma = func(x, y int) uint8 { return yc.Y[yc.YOffset(x, y)] }
	}
	var n, sum, sumSq float64
	for y := b.Min.Y; y < b.Max.Y; y += stepY {
		for x := b.Min.X; x < b.Max.X; x += stepX {
			v := float64(luma(x, y))
			n++
			sum += v
			sumSq += v * v
		}
	}
	if n == 0 {
		return 0, 0
	}
	mean = sum / n
	return mean, math.Sqrt(max(0, sumSq/n-mean*mean))
}

func checkPower(ctx context.Context, cfg Config) Result {
	r := Result{Check: "power"}
	t, err := health.ReadThrottled(ctx, cfg.Run)
	if err != nil {
		return skipped(r.Check, "vcgencmd not available: not a Raspberry Pi")
	}
	switch now, since := t.Now(), t.SinceBoot(); {
	case len(now) > 0:
		r.Status, r.Detail = Fail, strings.Join(now, ", ")+" now"
		r.Hint = "use a 5 V supply that delivers at least 2.5 A and a short, thick cable; under-voltage corrupts SD cards"
	case len(since) > 0:
		r.Status, r.Detail = Warn, strings.Join(since, ", ")+" since boot"
		r.Hint = "check the power supply and cable"
	default:
		r.Detail = "no under-voltage or throttling since boot"
	}
	return r
}

func checkTemperature(cfg Config) Result {
	r := Result{Check: "temperature"}
	c, err := health.ReadCelsius(cfg.ThermalFile)
	if err != nil {
		return skipped(r.Check, "no CPU temperature reading")
	}
	r.Detail = fmt.Sprintf("%.1f °C", c)
	switch {
	case c >= 80:
		r.Status, r.Hint = Fail, "the Pi throttles at 80 °C: add shade, ventilation or a heatsink"
	case c >= 70:
		r.Status, r.Hint = Warn, "close to the 80 °C throttling limit: add shade, ventilation or a heatsink"
	}
	return r
}

func checkCPU(cfg Config) Result {
	c, err := health.ReadCPUFreq(cfg.CPUFreqDir)
	if err != nil {
		return skipped("cpu", "no CPU clock reading")
	}
	return Result{Check: "cpu", Detail: fmt.Sprintf("%d MHz (%s, %d–%d MHz)", c.CurrentMHz, c.Governor, c.MinMHz, c.MaxMHz)}
}

// nightAt reports whether the sun is down at the saved location (only known
// when one is set, for SkipNightPhotos), and " (HH:MM)" of the next sunrise.
func nightAt(cfg Config, set conf.Settings) (string, bool) {
	if set.Latitude == 0 && set.Longitude == 0 {
		return "", false
	}
	now := time.Now()
	if cfg.Now != nil {
		now = cfg.Now()
	}
	if sun.Up(now, set.Latitude, set.Longitude) {
		return "", false
	}
	if rise, ok := sun.NextSunrise(now, set.Latitude, set.Longitude); ok {
		return " (" + rise.Local().Format("15:04") + ")", true
	}
	return "", true
}
