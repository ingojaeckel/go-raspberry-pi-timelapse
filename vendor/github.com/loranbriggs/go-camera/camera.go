package camera

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	STILL      = "raspistill"
	HFLIP      = "-hf"
	VFLIP      = "-vf"
	OUTFLAG    = "-o"
	WIDTH      = "-w %d"
	HEIGHT     = "-h %d"
	FILE_TYPE  = ".jpg"
	TIME_STAMP = "2006-01-02_15:04:05"
)

type Camera struct {
	horizontalFlip bool
	verticalFlip   bool
	savePath       string
	width          int64
	height         int64
}

func New(path string, width int64, height int64) *Camera {
	if path == "" {
		return nil
	}
	return &Camera{false, false, path, width, height}
}

func (c *Camera) Hflip(b bool) {
	c.horizontalFlip = b
}

func (c *Camera) Vflip(b bool) {
	c.verticalFlip = b
}

func (c *Camera) Capture() (string, error) {
	args := make([]string, 0)
	if c.horizontalFlip {
		args = append(args, HFLIP)
	}
	if c.verticalFlip {
		args = append(args, VFLIP)
	}
	args = append(args, fmt.Sprintf(WIDTH, c.width))
	args = append(args, fmt.Sprintf(HEIGHT, c.height))
	args = append(args, OUTFLAG)
	fileName := time.Now().Format(TIME_STAMP) + FILE_TYPE
	fullPath := filepath.Join(c.savePath, fileName)
	args = append(args, fullPath)
	cmd := exec.Command(STILL, OUTFLAG, fullPath)
	_, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}
	err = cmd.Start()
	if err != nil {
		fmt.Println(err)
	}
	cmd.Wait()
	return fullPath, nil
}
