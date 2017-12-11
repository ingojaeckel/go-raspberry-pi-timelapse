package timelapse

import (
	"errors"
	"fmt"
	"github.com/loranbriggs/go-camera"
	"os"
	"time"
)

type Timelapse struct {
	Camera                  camera.Camera
	Folder                  string
	SecondsBetweenCapture   int64
	OffsetWithinHourSeconds int64
	Res                     Resolution
}

type Resolution struct {
	Width  int64
	Height int64
}

func New(folder string, secondsBetweenCapture int64, offsetWithinHourSeconds int64, res Resolution) (*Timelapse, error) {
	_, err := os.Stat(folder)
	createFolder := err != nil && os.IsNotExist(err)

	if createFolder {
		if err := os.Mkdir(folder, 0700); err != nil {
			return nil, err
		}
	}
	// Assume folder exists
	c := camera.New(folder, res.Width, res.Height)
	if c == nil {
		return nil, errors.New("Failed to instantiate camera")
	}

	return &Timelapse{
		Camera:                  *c,
		Folder:                  folder,
		SecondsBetweenCapture:   secondsBetweenCapture,
		OffsetWithinHourSeconds: offsetWithinHourSeconds,
		Res: res,
	}, nil
}

func (t Timelapse) CapturePeriodically() {
	go func() {
		t.WaitForFirstCapture()

		for {
			beforeCapture := time.Now()
			s, err := t.Camera.Capture()
			if err != nil {
				fmt.Errorf("Error during capture: %s\n", err.Error())
			}

			fmt.Printf("captured picture in %s\n", s)
			timeToCaptureSeconds := time.Now().Unix() - beforeCapture.Unix()

			sleepTime := t.SecondsBetweenCapture - timeToCaptureSeconds
			fmt.Printf("capture took %d seconds, will sleep for %d seconds\n", timeToCaptureSeconds, sleepTime)
			time.Sleep(time.Duration(sleepTime) * time.Second)
		}
	}()
}

func (t Timelapse) WaitForFirstCapture() {
	secondsUntilFirstCapture := t.SecondsToSleepUntilOffset(time.Now())
	fmt.Printf("Waiting %d seconds before first capture\n", secondsUntilFirstCapture)
	time.Sleep(time.Duration(secondsUntilFirstCapture) * time.Second)
}

func (t Timelapse) SecondsToSleepUntilOffset(currentTime time.Time) int {
	if t.OffsetWithinHourSeconds <= 0 {
		return 0
	}

	offsetWithinHourMinutes := int(t.OffsetWithinHourSeconds / 60)
	offsetWithinHourSeconds := int(t.OffsetWithinHourSeconds % 60)

	nextOffset := time.Date(
		currentTime.Year(),
		currentTime.Month(),
		currentTime.Day(),
		currentTime.Hour(),
		offsetWithinHourMinutes,
		offsetWithinHourSeconds,
		0,
		currentTime.Location())

	d := nextOffset.Sub(currentTime)
	durationInSeconds := int(d.Seconds())

	if durationInSeconds < 0 {
		// We just missed the window to wait for the offset.
		// Wait until we reach the next window for: time between captures - the time by which we missed the offset.
		// (Using + since durationInSeconds is already negative..)
		return int(t.SecondsBetweenCapture) + durationInSeconds
	}
	return durationInSeconds
}
