package timelapse

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"
	"strconv"
)

const (
	STILL      = "raspistill"
	OUTFLAG    = "-o"
	WIDTH      = "-w"
	HEIGHT     = "-h"
	FILE_TYPE  = ".jpg"
	TIME_STAMP = "2006-01-02_15:04:05"
)

type Camera struct {
	savePath string
	width    int64
	height   int64
}

func NewCamera(path string, width int64, height int64) *Camera {
	if path == "" {
		return nil
	}
	return &Camera{path, width, height}
}

func (c *Camera) Capture() (string, error) {
	fullPath := getAbsoluteFilepath(c.savePath)
	args := getRaspistillArgs(c.width, c.height, fullPath)
	cmd := exec.Command(STILL, args...)

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

func getAbsoluteFilepath(savePath string) string {
	fileName := time.Now().Format(TIME_STAMP) + FILE_TYPE
	return filepath.Join(savePath, fileName)
}

func getRaspistillArgs(width int64, height int64, fullPath string) []string {
	args := make([]string, 0)
	args = append(args, WIDTH)
	args = append(args, strconv.Itoa(int(width)))
	args = append(args, HEIGHT)
	args = append(args, strconv.Itoa(int(height)))
	args = append(args, OUTFLAG)
	args = append(args, fullPath)
	return args
}