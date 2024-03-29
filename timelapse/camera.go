package timelapse

// This has been adapted from https://github.com/loranbriggs/go-camera

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

const (
	commandRaspistill = "raspistill"
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

func (c *Camera) Capture() (string, error) {
	fullPath := c.getAbsoluteFilepath()
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
		"-w", strconv.Itoa(c.width),
		"-h", strconv.Itoa(c.height),
		"-q", strconv.Itoa(c.quality),
	}
	if c.flipVertically {
		args = append(args, "-vf")
	}
	if c.flipHorizontally {
		args = append(args, "-hf")
	}
	return append(args, "-o", fullPath)
}

func getFileName(t time.Time) string {
	return fmt.Sprintf("%4d%02d%02d-%02d%02d%02d.jpg", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

func (c *Camera) getAbsoluteFilepath() string {
	return filepath.Join(c.savePath, getFileName(time.Now()))
}
