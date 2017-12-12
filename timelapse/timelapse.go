package timelapse

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type Timelapse struct {
	Camera                  Camera
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
	c := NewCamera(folder, res.Width, res.Height)
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
	picturesPerHour := 3600 / t.SecondsBetweenCapture

	secondsIntoCurrentHour := int64(currentTime.Minute() * 60 + currentTime.Second())

	for i := 0; i<int(picturesPerHour); i++ {
		if i == 0 {
			if 0 <= secondsIntoCurrentHour && secondsIntoCurrentHour <= t.OffsetWithinHourSeconds {
				return int(t.OffsetWithinHourSeconds - secondsIntoCurrentHour)
			}
		}

		lowerBoundary := t.OffsetWithinHourSeconds + int64(i-1) * t.SecondsBetweenCapture
		upperBoundary := int64(t.OffsetWithinHourSeconds + int64(i) * t.SecondsBetweenCapture)

		if lowerBoundary <= secondsIntoCurrentHour && secondsIntoCurrentHour <= upperBoundary {
			return int(upperBoundary - secondsIntoCurrentHour)
		}
	}

	return int(3600 - secondsIntoCurrentHour + t.OffsetWithinHourSeconds)
}


