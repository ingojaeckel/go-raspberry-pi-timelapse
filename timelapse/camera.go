package timelapse

// This has been adapted from https://github.com/loranbriggs/go-camera

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	commandRaspistill = "rpicam-still"
	commandFfmpeg     = "ffmpeg"
)

type Camera struct {
	savePath                         string
	width, height                    int
	flipHorizontally, flipVertically bool
	quality                          int
}

// NewCamera Setting "rotate" to true will create a camera instance which will flip all pictures by 180 degrees. Each captured image will be flipped horizontally and vertically.
func NewCamera(path string, width, height int, rotate bool, quality int) (Camera, error) {
	if path == "" {
		return Camera{}, errors.New("invalid config: path must not be empty")
	}
	return Camera{
		savePath:         path,
		width:            width,
		height:           height,
		flipHorizontally: rotate,
		flipVertically:   rotate,
		quality:          quality,
	}, nil
}

// isDevelopment detects whether the code is running on a development system (non-ARM)
// or a deployment system (ARM/Raspberry Pi).
// Development typically happens on x86_64/AMD64 Linux/macOS, while deployment is on ARM.
func isDevelopment() bool {
	arch := runtime.GOARCH
	// ARM-based systems (arm, arm64) are considered deployment/production
	// x86_64, amd64, 386, etc. are considered development
	return arch != "arm" && arch != "arm64"
}

func (c *Camera) Capture() (string, error) {
	fullPath := c.getAbsoluteFilepath()
	
	if isDevelopment() {
		// On development systems, use ffmpeg to capture from webcam
		return c.captureWithWebcam(fullPath)
	}
	
	// On Raspberry Pi, use rpicam-still
	args := c.getRaspistillArgs(fullPath)
	log.Printf("Running command: %s %v", commandRaspistill, args)
	cmd := exec.Command(commandRaspistill, args...)

	_, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		return "", err
	}
	err = cmd.Start()
	if err != nil {
		log.Println(err)
		return "", err
	}
	cmd.Wait()
	return fullPath, nil
}

func (c *Camera) getRaspistillArgs(fullPath string) []string {
	args := []string{
		"--width", strconv.Itoa(c.width),
		"--height", strconv.Itoa(c.height),
		"--quality", strconv.Itoa(c.quality),
	}
	if c.flipVertically {
		args = append(args, "--vflip")
	}
	if c.flipHorizontally {
		args = append(args, "--hflip")
	}
	return append(args, "--output", fullPath)
}

// captureWithWebcam captures an image from the first available webcam using ffmpeg.
// This is used on development systems where raspistill is not available.
func (c *Camera) captureWithWebcam(fullPath string) (string, error) {
	// Try to capture from /dev/video0 (first webcam)
	// ffmpeg -f v4l2 -video_size WIDTHxHEIGHT -i /dev/video0 -frames:v 1 -q:v QUALITY output.jpg
	args := []string{
		"-f", "v4l2",
		"-video_size", fmt.Sprintf("%dx%d", c.width, c.height),
		"-i", "/dev/video0",
		"-frames:v", "1",
		"-q:v", strconv.Itoa(c.quality),
	}
	
	// Add flip filters if needed
	var filters []string
	if c.flipHorizontally {
		filters = append(filters, "hflip")
	}
	if c.flipVertically {
		filters = append(filters, "vflip")
	}
	if len(filters) > 0 {
		args = append(args, "-vf", strings.Join(filters, ","))
	}
	
	args = append(args, "-y", fullPath)
	
	log.Printf("Running command: %s %v", commandFfmpeg, args)
	cmd := exec.Command(commandFfmpeg, args...)
	
	// Capture stderr since ffmpeg outputs to stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("ffmpeg error: %s", string(output))
		return "", fmt.Errorf("failed to capture from webcam: %w", err)
	}
	
	return fullPath, nil
}

func getFileName(t time.Time) string {
	return fmt.Sprintf("%4d%02d%02d-%02d%02d%02d.jpg", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

func (c *Camera) getAbsoluteFilepath() string {
	return filepath.Join(c.savePath, getFileName(time.Now()))
}
