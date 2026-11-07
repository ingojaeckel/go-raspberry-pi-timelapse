package selfcheck

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
)

// scene draws a test photo: noise around a mean brightness.
func scene(w, h int, mean, spread float64) image.Image {
	img := image.NewGray(image.Rect(0, 0, w, h))
	r := rand.New(rand.NewPCG(1, 2))
	for i := range img.Pix {
		v := mean + (r.Float64()*2-1)*spread
		img.Pix[i] = uint8(max(0, min(255, v)))
	}
	return img
}

// fakeCamera writes img as a JPEG (a scene of the requested size if nil), or
// fails with err.
type fakeCamera struct {
	img  image.Image
	err  error
	got  timelapse.Options
	name string
}

func (f *fakeCamera) Capture(_ context.Context, path string, o timelapse.Options) error {
	f.got = o
	if f.err != nil {
		return f.err
	}
	img := f.img
	if img == nil {
		img = scene(o.Width, o.Height, 120, 60)
	}
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, nil); err != nil {
		return err
	}
	return os.WriteFile(path, b.Bytes(), 0o600)
}

func (f *fakeCamera) String() string {
	if f.name != "" {
		return f.name
	}
	return "rpicam-still"
}

// commands fakes rpicam-still --list-cameras and vcgencmd.
type commands map[string]string

func (c commands) run(_ context.Context, name string, args ...string) ([]byte, error) {
	key := filepath.Base(name) + " " + strings.Join(args, " ")
	out, ok := c[key]
	if !ok {
		return nil, errors.New(key + ": not found")
	}
	return []byte(out), nil
}

const imx290 = `Available cameras
-----------------
0 : imx290 [1920x1080 12-bit RGGB] (/base/soc/i2c0mux/i2c@1/imx290@1a)
    Modes: 'SRGGB10_CSI2P' : 1280x720 [30.00 fps - (65535, 65535)/65535x65535 crop]
                             1920x1080 [30.00 fps - (65535, 65535)/65535x65535 crop]
`

const imx219 = `Available cameras
-----------------
0 : imx219 [3280x2464 10-bit RGGB] (/base/soc/i2c0mux/i2c@1/imx219@10)
    Modes: 'SRGGB10_CSI2P' : 640x480 [206.65 fps - (1000, 752)/1280x960 crop]
                             1640x1232 [83.70 fps - (0, 0)/3280x2464 crop]
                             3280x2464 [21.19 fps - (0, 0)/3280x2464 crop]
`

// healthyPi is a working Pi: a camera, good power, cool, powersave.
func healthyPi(t *testing.T) (Config, *fakeCamera) {
	t.Helper()
	dir := t.TempDir()
	thermal := filepath.Join(dir, "temp")
	write(t, thermal, "48312\n")
	cpu := filepath.Join(dir, "cpufreq")
	write(t, filepath.Join(cpu, "scaling_governor"), "powersave\n")
	write(t, filepath.Join(cpu, "scaling_cur_freq"), "600000\n")
	write(t, filepath.Join(cpu, "scaling_min_freq"), "600000\n")
	write(t, filepath.Join(cpu, "scaling_max_freq"), "1000000\n")
	if err := os.MkdirAll(filepath.Join(dir, "proc"), 0o700); err != nil {
		t.Fatal(err)
	}
	cam := &fakeCamera{}
	defaults := conf.Default(1800)
	defaults.PhotoResolutionWidth, defaults.PhotoResolutionHeight = 1640, 1232
	return Config{
		Backend:         cam,
		CaptureTimeout:  time.Minute,
		StorageFolder:   filepath.Join(dir, "photos"),
		SettingsFile:    filepath.Join(dir, "settings.db"),
		DefaultSettings: defaults,
		MinFreeBytes:    256 << 20,
		MinBrightness:   10,
		FreeBytes:       func(string) (uint64, error) { return 8 << 30, nil },
		Run: commands{
			"rpicam-still --list-cameras": imx219,
			"vcgencmd get_throttled":      "throttled=0x0\n",
		}.run,
		ThermalFile: thermal,
		CPUFreqDir:  cpu,
		ProcDir:     filepath.Join(dir, "proc"),
	}, cam
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func find(t *testing.T, results []Result, check string) Result {
	t.Helper()
	for _, r := range results {
		if r.Check == check {
			return r
		}
	}
	t.Fatalf("no %q check in %+v", check, results)
	return Result{}
}

func report(results []Result) string {
	var b strings.Builder
	Write(&b, results)
	return b.String()
}

func TestAHealthyCameraPassesEveryCheck(t *testing.T) {
	cfg, cam := healthyPi(t)
	results := Run(t.Context(), cfg)
	for _, r := range results {
		if r.Status != Pass {
			t.Errorf("%s: %s %s", r.Check, r.Status, r.Detail)
		}
	}
	var b strings.Builder
	if !Write(&b, results) {
		t.Errorf("Write reported a failure:\n%s", b.String())
	}
	out := b.String()
	for _, want := range []string{"imx219 (3280x2464)", "brightness", "48.3 °C", "600 MHz (powersave", "All checks passed"} {
		if !strings.Contains(out, want) {
			t.Errorf("report lacks %q:\n%s", want, out)
		}
	}
	if cam.got.Width != 1640 || cam.got.Height != 1232 {
		t.Errorf("captured %dx%d, want the settings' 1640x1232", cam.got.Width, cam.got.Height)
	}
	entries, err := os.ReadDir(cfg.StorageFolder)
	if err != nil || len(entries) != 0 {
		t.Errorf("the test photo was left behind: %v %v", entries, err)
	}
}

func TestABlackPhotoFails(t *testing.T) {
	cfg, cam := healthyPi(t)
	cam.img = scene(64, 48, 3, 3) // lens cap on, or no light
	r := find(t, Run(t.Context(), cfg), "photo")
	if r.Status != Fail || !strings.Contains(r.Detail, "too dark") {
		t.Fatalf("got %s %q", r.Status, r.Detail)
	}
}

func TestTheDarknessThresholdIsConfigurable(t *testing.T) {
	cfg, cam := healthyPi(t)
	cam.img = scene(64, 48, 3, 3)
	cfg.MinBrightness = 0 // testing at night
	if r := find(t, Run(t.Context(), cfg), "photo"); r.Status == Fail && strings.Contains(r.Detail, "too dark") {
		t.Fatalf("got %s %q", r.Status, r.Detail)
	}
}

func TestADarkPhotoAtNightIsAWarning(t *testing.T) {
	cfg, cam := healthyPi(t)
	cam.img = scene(64, 48, 3, 3)
	cfg.DefaultSettings.SkipNightPhotos, cfg.DefaultSettings.Latitude, cfg.DefaultSettings.Longitude = true, 51.5074, -0.1278
	cfg.Now = func() time.Time { return time.Date(2026, 11, 7, 22, 0, 0, 0, time.UTC) } // night in London
	r := find(t, Run(t.Context(), cfg), "photo")
	if r.Status != Warn || !strings.Contains(r.Detail, "too dark") || !strings.Contains(r.Detail, "night at the camera") {
		t.Fatalf("got %s %q", r.Status, r.Detail)
	}
	if !strings.Contains(r.Hint, "after sunrise") {
		t.Errorf("hint %q should say when to check again", r.Hint)
	}
}

func TestADarkPhotoInDaylightFailsEvenWithALocation(t *testing.T) {
	cfg, cam := healthyPi(t)
	cam.img = scene(64, 48, 3, 3)
	cfg.DefaultSettings.SkipNightPhotos, cfg.DefaultSettings.Latitude, cfg.DefaultSettings.Longitude = true, 51.5074, -0.1278
	cfg.Now = func() time.Time { return time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC) }
	if r := find(t, Run(t.Context(), cfg), "photo"); r.Status != Fail {
		t.Fatalf("got %s %q", r.Status, r.Detail)
	}
}

func TestAUniformPhotoFails(t *testing.T) {
	cfg, cam := healthyPi(t)
	cam.img = scene(64, 48, 128, 0) // a sensor that returns a flat grey frame
	r := find(t, Run(t.Context(), cfg), "photo")
	if r.Status != Fail || !strings.Contains(r.Detail, "same brightness") {
		t.Fatalf("got %s %q", r.Status, r.Detail)
	}
}

func TestAFileThatIsNotAJPEGFails(t *testing.T) {
	cfg, _ := healthyPi(t)
	cfg.Backend = writer("jpeg\n")
	r := find(t, Run(t.Context(), cfg), "photo")
	if r.Status != Fail || !strings.Contains(r.Detail, "not a JPEG") {
		t.Fatalf("got %s %q", r.Status, r.Detail)
	}
}

type writer string

func (w writer) Capture(_ context.Context, path string, _ timelapse.Options) error {
	return os.WriteFile(path, []byte(w), 0o600)
}

func TestKeepLeavesThePhotoForInspection(t *testing.T) {
	cfg, _ := healthyPi(t)
	cfg.Keep = true
	r := find(t, Run(t.Context(), cfg), "photo")
	path := r.Detail[strings.LastIndex(r.Detail, " ")+1:]
	if _, err := jpeg.DecodeConfig(must(os.Open(path))); err != nil {
		t.Fatalf("kept photo %q: %v (%s)", path, err, r.Detail)
	}
}

func must(f *os.File, err error) *os.File {
	if err != nil {
		panic(err)
	}
	return f
}

func TestAPhotoOfTheWrongSizeIsAWarning(t *testing.T) {
	cfg, cam := healthyPi(t)
	cam.img = scene(64, 48, 120, 60)
	cfg.DefaultSettings.PhotoResolutionWidth, cfg.DefaultSettings.PhotoResolutionHeight = 640, 480
	r := find(t, Run(t.Context(), cfg), "photo")
	if r.Status != Warn || !strings.Contains(r.Detail, "64x48") {
		t.Fatalf("got %s %q", r.Status, r.Detail)
	}
}

func TestACaptureErrorExplainsTheLikelyCause(t *testing.T) {
	for _, tc := range []struct{ out, hint string }{
		{"INFO Camera camera.cpp:1021 Pipeline handler in use by another process\nERROR: *** failed to acquire camera ***", "another program"},
		{"ERROR: *** no cameras available ***", "ribbon cable"},
		{"context deadline exceeded", "did not finish"},
	} {
		cfg, cam := healthyPi(t)
		cam.err = errors.New("rpicam-still: exit status 255: " + tc.out)
		results := Run(t.Context(), cfg)
		r := find(t, results, "capture")
		if r.Status != Fail || !strings.Contains(r.Hint, tc.hint) {
			t.Errorf("%q: got %s %q hint %q", tc.out, r.Status, r.Detail, r.Hint)
		}
		if p := find(t, results, "photo"); p.Status != Skip {
			t.Errorf("photo check after a failed capture: %s", p.Status)
		}
		if out := report(results); !strings.Contains(out, "Self-check failed") {
			t.Errorf("report:\n%s", out)
		}
	}
}

func TestACameraHeldByAnotherProgramIsNamed(t *testing.T) {
	cfg, cam := healthyPi(t)
	proc := filepath.Join(cfg.ProcDir, "698")
	if err := os.MkdirAll(filepath.Join(proc, "fd"), 0o700); err != nil {
		t.Fatal(err)
	}
	for fd, dev := range map[string]string{"20": "/dev/media0", "21": "/dev/video0", "3": "/home/ingo/log.txt"} {
		if err := os.Symlink(dev, filepath.Join(proc, "fd", fd)); err != nil {
			t.Fatal(err)
		}
	}
	write(t, filepath.Join(proc, "cmdline"), "venv/bin/python3\x00video_sender.py\x00")
	write(t, filepath.Join(proc, "cgroup"), "0::/system.slice/surveillance-sender.service\n")

	results := Run(t.Context(), cfg)
	r := find(t, results, "camera free")
	if r.Status != Fail || !strings.Contains(r.Detail, "pid 698 venv/bin/python3 video_sender.py") ||
		!strings.Contains(r.Detail, "surveillance-sender.service") {
		t.Fatalf("got %s %q", r.Status, r.Detail)
	}
	if !strings.Contains(r.Hint, "sudo systemctl disable --now surveillance-sender.service") {
		t.Errorf("hint %q", r.Hint)
	}
	if c := find(t, results, "capture"); c.Status != Skip {
		t.Errorf("captured although the camera is in use: %s", c.Status)
	}
	if cam.got != (timelapse.Options{}) {
		t.Error("the camera was used")
	}
}

func TestNoCameraDetected(t *testing.T) {
	cfg, _ := healthyPi(t)
	cfg.Run = commands{"rpicam-still --list-cameras": "No cameras available!\n", "vcgencmd get_throttled": "throttled=0x0"}.run
	r := find(t, Run(t.Context(), cfg), "camera detected")
	if r.Status != Fail || !strings.Contains(r.Hint, "dtoverlay=") || !strings.Contains(r.Hint, "ribbon cable") {
		t.Fatalf("got %s %q hint %q", r.Status, r.Detail, r.Hint)
	}
}

func TestAPhotoSizeTheCameraLacksIsAWarning(t *testing.T) {
	cfg, _ := healthyPi(t)
	cfg.Run = commands{"rpicam-still --list-cameras": imx290, "vcgencmd get_throttled": "throttled=0x0"}.run
	s, err := conf.OpenStore(cfg.SettingsFile, conf.Default(1800))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(func(v *conf.Settings) { v.PhotoResolutionWidth, v.PhotoResolutionHeight = 1640, 1232 }); err != nil {
		t.Fatal(err)
	}
	s.Close()
	r := find(t, Run(t.Context(), cfg), "camera detected")
	if r.Status != Warn || !strings.Contains(r.Detail, "1640x1232") || !strings.Contains(r.Detail, "1920x1080, 1280x720") {
		t.Fatalf("got %s %q", r.Status, r.Detail)
	}
}

func TestAWebcamSkipsTheRaspberryPiCameraChecks(t *testing.T) {
	cfg, cam := healthyPi(t)
	cam.name = "ffmpeg"
	results := Run(t.Context(), cfg)
	if r := find(t, results, "camera detected"); r.Status != Skip {
		t.Errorf("camera detected: %s %q", r.Status, r.Detail)
	}
	if r := find(t, results, "capture"); r.Status != Pass {
		t.Errorf("capture: %s %q", r.Status, r.Detail)
	}
}

func TestSavedSettingsAreUsed(t *testing.T) {
	cfg, cam := healthyPi(t)
	s, err := conf.OpenStore(cfg.SettingsFile, conf.Default(1800))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(func(v *conf.Settings) { v.PhotoResolutionWidth, v.PhotoResolutionHeight, v.RotateBy = 1640, 1232, 180 }); err != nil {
		t.Fatal(err)
	}
	s.Close()
	r := find(t, Run(t.Context(), cfg), "settings")
	if r.Status != Pass || !strings.Contains(r.Detail, "saved") || !cam.got.Rotate || cam.got.Width != 1640 {
		t.Fatalf("got %s %q, captured with %+v", r.Status, r.Detail, cam.got)
	}
}

func TestARunningAppIsAWarning(t *testing.T) {
	cfg, _ := healthyPi(t)
	s, err := conf.OpenStore(cfg.SettingsFile, conf.Default(1800)) // stays open like the app
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	r := find(t, Run(t.Context(), cfg), "settings")
	if r.Status != Warn || !strings.Contains(r.Detail, "running") {
		t.Fatalf("got %s %q", r.Status, r.Detail)
	}
}

func TestStorage(t *testing.T) {
	cfg, _ := healthyPi(t)
	cfg.FreeBytes = func(string) (uint64, error) { return 100 << 20, nil }
	if r := find(t, Run(t.Context(), cfg), "storage"); r.Status != Warn || !strings.Contains(r.Detail, "100 MB free") {
		t.Errorf("low space: %s %q", r.Status, r.Detail)
	}

	cfg, _ = healthyPi(t)
	write(t, cfg.StorageFolder, "a file, not a folder")
	results := Run(t.Context(), cfg)
	if r := find(t, results, "storage"); r.Status != Fail {
		t.Errorf("unwritable: %s %q", r.Status, r.Detail)
	}
	if r := find(t, results, "capture"); r.Status != Skip {
		t.Errorf("capture without storage: %s", r.Status)
	}
}

func TestPowerAndTemperature(t *testing.T) {
	for _, tc := range []struct {
		throttled, celsius string
		check              string
		want               Status
		detail             string
	}{
		{"throttled=0x50005", "48000", "power", Fail, "under-voltage"},
		{"throttled=0x50000", "48000", "power", Warn, "since boot"},
		{"throttled=0x0", "74000", "temperature", Warn, "74.0 °C"},
		{"throttled=0x0", "83000", "temperature", Fail, "83.0 °C"},
	} {
		cfg, _ := healthyPi(t)
		cfg.Run = commands{"rpicam-still --list-cameras": imx219, "vcgencmd get_throttled": tc.throttled}.run
		write(t, cfg.ThermalFile, tc.celsius)
		r := find(t, Run(t.Context(), cfg), tc.check)
		if r.Status != tc.want || !strings.Contains(r.Detail, tc.detail) {
			t.Errorf("%s %s: got %s %q", tc.throttled, tc.celsius, r.Status, r.Detail)
		}
	}
}

func TestReadingsThatAreMissingOffAPiAreSkipped(t *testing.T) {
	cfg, cam := healthyPi(t)
	cam.name = "ffmpeg"
	cfg.Run = commands{}.run
	cfg.ThermalFile = filepath.Join(t.TempDir(), "missing")
	cfg.CPUFreqDir = filepath.Join(t.TempDir(), "missing")
	cfg.ProcDir = ""
	results := Run(t.Context(), cfg)
	for _, check := range []string{"camera detected", "camera free", "power", "temperature", "cpu"} {
		if r := find(t, results, check); r.Status != Skip {
			t.Errorf("%s: %s %q", check, r.Status, r.Detail)
		}
	}
	if !Write(new(strings.Builder), results) {
		t.Error("skipped checks made the self-check fail")
	}
}

func TestWriteSummarizesWarnings(t *testing.T) {
	out := report([]Result{{Check: "power", Status: Warn, Detail: "under-voltage since boot", Hint: "check the cable"}})
	for _, want := range []string{"WARN  power", "under-voltage since boot", "→ check the cable", "1 warning"} {
		if !strings.Contains(out, want) {
			t.Errorf("lacks %q:\n%s", want, out)
		}
	}
}

func TestBrightnessOfAColorPhoto(t *testing.T) {
	img := image.NewYCbCr(image.Rect(0, 0, 40, 30), image.YCbCrSubsampleRatio420)
	for i := range img.Y {
		img.Y[i] = uint8(i % 200)
	}
	mean, sd := brightness(img)
	if mean < 80 || mean > 120 || sd < 30 {
		t.Fatalf("mean %v sd %v", mean, sd)
	}
	gray := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for x := range 4 {
		for y := range 4 {
			gray.Set(x, y, color.RGBA{200, 200, 200, 255})
		}
	}
	if mean, sd := brightness(gray); mean < 199 || mean > 201 || sd != 0 {
		t.Fatalf("rgba: mean %v sd %v", mean, sd)
	}
}

func TestWithoutSavedSettingsTheCamerasDefaultSizeIsUsed(t *testing.T) {
	cfg, cam := healthyPi(t)
	cfg.Run = commands{"rpicam-still --list-cameras": imx290, "vcgencmd get_throttled": "throttled=0x0"}.run
	results := Run(t.Context(), cfg)
	if cam.got.Width != 1920 || cam.got.Height != 1080 {
		t.Fatalf("captured %dx%d", cam.got.Width, cam.got.Height)
	}
	if r := find(t, results, "camera detected"); r.Status != Pass {
		t.Fatalf("%s %q", r.Status, r.Detail)
	}
}
