package timelapse

// Originally adapted from https://github.com/loranbriggs/go-camera

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	commandRpicamStill = "rpicam-still"
	commandFfmpeg      = "ffmpeg"
)

// Options describe a single photo.
type Options struct {
	Width, Height, Quality int
	// Rotate flips the image horizontally and vertically (180 degrees).
	Rotate bool
}

// Backend takes a photo and writes it to path.
type Backend interface {
	Capture(ctx context.Context, path string, o Options) error
}

// DetectBackend returns rpicam-still when it is installed (the Raspberry Pi),
// and ffmpeg webcam capture otherwise (development machines).
func DetectBackend() Backend {
	if _, err := exec.LookPath(commandRpicamStill); err == nil {
		return rpicamStill{}
	}
	return ffmpeg{goos: runtime.GOOS}
}

type rpicamStill struct{}

func (rpicamStill) Capture(ctx context.Context, path string, o Options) error {
	return runCommand(ctx, commandRpicamStill, rpicamStillArgs(path, o)...)
}

func rpicamStillArgs(path string, o Options) []string {
	args := []string{
		"--width", strconv.Itoa(o.Width),
		"--height", strconv.Itoa(o.Height),
		"--quality", strconv.Itoa(o.Quality),
	}
	if o.Rotate {
		args = append(args, "--vflip", "--hflip")
	}
	return append(args, "--output", path)
}

type ffmpeg struct{ goos string }

func (f ffmpeg) Capture(ctx context.Context, path string, o Options) error {
	return runCommand(ctx, commandFfmpeg, ffmpegArgs(f.goos, path, o)...)
}

// ffmpegArgs captures one frame from the first webcam: avfoundation device 0
// on macOS, /dev/video0 (v4l2) elsewhere.
func ffmpegArgs(goos, path string, o Options) []string {
	size := fmt.Sprintf("%dx%d", o.Width, o.Height)
	var args []string
	if goos == "darwin" {
		args = []string{"-f", "avfoundation", "-framerate", "30", "-i", "0", "-frames:v", "1", "-s", size}
	} else {
		args = []string{"-f", "v4l2", "-video_size", size, "-i", "/dev/video0", "-frames:v", "1", "-q:v", strconv.Itoa(o.Quality)}
	}
	if o.Rotate {
		args = append(args, "-vf", "hflip,vflip")
	}
	return append(args, "-y", path)
}

// runCommand runs a command, killing it when ctx ends. Its combined output is
// included in the error.
func runCommand(ctx context.Context, name string, args ...string) error {
	slog.Debug("running command", "command", name, "args", args)
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = 2 * time.Second // don't wait forever for orphaned pipes
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Run(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			err = fmt.Errorf("%w (%v)", ctxErr, err)
		}
		return fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(out.String()))
	}
	return nil
}

// Camera serializes access to the camera hardware, bounds every capture by a
// timeout and writes photos atomically: the backend writes to a hidden file
// which is renamed once complete, so a half-written photo is never listed or
// served.
type Camera struct {
	backend Backend
	timeout time.Duration
	lock    chan struct{}
}

// NewCamera returns a Camera that gives up on a capture after timeout.
func NewCamera(backend Backend, timeout time.Duration) *Camera {
	return &Camera{backend: backend, timeout: timeout, lock: make(chan struct{}, 1)}
}

// Capture takes a photo into dir (created if missing) and returns its path.
// It waits while another capture is using the camera.
func (c *Camera) Capture(ctx context.Context, dir string, o Options) (string, error) {
	release, err := c.acquire(ctx)
	if err != nil {
		return "", err
	}
	defer release()

	final, err := newPhotoPath(dir)
	if err == nil {
		err = c.captureTo(ctx, final, o)
	}
	if err != nil {
		return "", err
	}
	return final, nil
}

// newPhotoPath creates dir if needed and returns an unused path for a photo
// taken now.
func newPhotoPath(dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return freePath(dir, time.Now())
}

// acquire waits for exclusive use of the camera.
func (c *Camera) acquire(ctx context.Context) (release func(), err error) {
	select {
	case c.lock <- struct{}{}:
		return func() { <-c.lock }, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("waiting for the camera: %w", ctx.Err())
	}
}

// captureTo has the backend write a hidden file next to final, then renames
// it. Nothing is left behind on failure.
func (c *Camera) captureTo(ctx context.Context, final string, o Options) error {
	partial := filepath.Join(filepath.Dir(final), ".partial-"+filepath.Base(final))
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	err := c.backend.Capture(ctx, partial, o)
	if err == nil {
		err = os.Rename(partial, final)
	}
	if err != nil {
		os.Remove(partial)
	}
	return err
}

// freePath returns the path for a photo taken at t that does not exist yet.
// Names are second-resolution timestamps; if the clock moved backwards a
// numeric suffix keeps the earlier photo from being overwritten.
func freePath(dir string, t time.Time) (string, error) {
	name := fileName(t)
	stem := strings.TrimSuffix(name, ".jpg")
	for i := 0; ; i++ {
		if i > 0 {
			name = fmt.Sprintf("%s-%d.jpg", stem, i)
		}
		p := filepath.Join(dir, name)
		if _, err := os.Lstat(p); errors.Is(err, fs.ErrNotExist) {
			return p, nil
		} else if err != nil {
			return "", err
		}
	}
}

func fileName(t time.Time) string {
	return t.Format("20060102-150405") + ".jpg"
}
