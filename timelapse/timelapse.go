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
		for {
			t.WaitForCapture()

			beforeCapture := time.Now()
			s, err := t.Camera.Capture()
			if err != nil {
				fmt.Errorf("Error during capture: %s\n", err.Error())
			}

			timeToCaptureSeconds := time.Now().Unix() - beforeCapture.Unix()
			fmt.Printf("Photo stored in '%s'. Capturing took %d seconds\n", s, timeToCaptureSeconds)
		}
	}()
}

func (t Timelapse) WaitForCapture() {
	secondsUntilFirstCapture := t.SecondsToSleepUntilOffset(time.Now())
	sleepDuration := time.Duration(secondsUntilFirstCapture) * time.Second
	nextCaptureAt := time.Now().Add(sleepDuration )

	fmt.Printf("Will take the next picture in %d seconds at %v.\n", secondsUntilFirstCapture, nextCaptureAt)
	time.Sleep(sleepDuration)
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


