package camera

import (
  "fmt"
  "os/exec"
  "path/filepath"
  "time"
)

const (
  STILL = "raspistill"
  HFLIP = "-hf"
  VFLIP = "-vf"
  OUTFLAG = "-o"
  FILE_TYPE =".jpg"
  TIME_STAMP = "2006-01-02_15:04::05"
)

type Camera struct {
  horizontalFlip bool
  verticalFlip bool
  savePath string
}

func New(path string) *Camera {
  if path =="" {
    return nil
  }
  return &Camera{false, false, path}
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
  args = append(args, OUTFLAG)
  fileName := time.Now().Format(TIME_STAMP) + FILE_TYPE
  fullPath := filepath.Join(c.savePath, fileName)
  args = append(args, fullPath)
  cmd := exec.Command("raspistill", "-o", fullPath)
  _, err := cmd.StdoutPipe()
  if err != nil {
    fmt.Print(err)
  }
  err = cmd.Start()
  if err != nil {
    fmt.Print(err)
  }
  cmd.Wait()
  return fullPath, nil
}


