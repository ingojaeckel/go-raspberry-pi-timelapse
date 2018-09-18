package timelapse

import (
	"errors"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"log"
	"os"
	"time"
)

var failedToInitCamera = errors.New("failed to instantiate camera")

type Timelapse struct {
	Camera                  Camera
	Folder                  string
	SecondsBetweenCapture   int
	OffsetWithinHourSeconds int
	Res                     Resolution
	DebugEnabled            bool
}

type Resolution struct {
	Width  int
	Height int
}

func New(folder string, s *conf.Settings) (*Timelapse, error) {
	_, err := os.Stat(folder)
	createFolder := err != nil && os.IsNotExist(err)

	if createFolder {
		if err := os.Mkdir(folder, 0700); err != nil {
			return nil, err
		}
	}
	// Assume folder exists

	c, err := NewCamera(folder, s.PhotoResolutionWidth, s.PhotoResolutionHeight, s.RotateBy == 180)
	if err != nil {
		return nil, failedToInitCamera
	}

	return &Timelapse{
		Camera:                  c,
		Folder:                  folder,
		SecondsBetweenCapture:   s.SecondsBetweenCaptures,
		OffsetWithinHourSeconds: s.OffsetWithinHour,
		Res:                     Resolution{Width: s.PhotoResolutionWidth, Height: s.PhotoResolutionHeight},
		DebugEnabled:            s.DebugEnabled,
	}, nil
}

func (t Timelapse) CapturePeriodically() {
	offsetDisabled := t.OffsetWithinHourSeconds == -1

	if offsetDisabled {
		log.Println("Offset is disabled. Will start taking pictures immediately.")
		go func() {
			for {
				s, err := t.Camera.Capture()
				if err != nil {
					log.Printf("Error during capture: %s\n", err.Error())
				}
				log.Printf("Photo stored in '%s'. Will sleep for %d seconds.\n", s, t.SecondsBetweenCapture)
				time.Sleep(time.Duration(t.SecondsBetweenCapture) * time.Second)
			}
		}()
	} else {
		log.Println("Offset is enabled. Will wait before taking first picture.")
		go func() {
			for {
				t.WaitForCapture()

				beforeCapture := time.Now()
				s, err := t.Camera.Capture()
				if err != nil {
					log.Printf("Error during capture: %s\n", err.Error())
				}

				timeToCaptureSeconds := time.Now().Unix() - beforeCapture.Unix()
				log.Printf("Photo stored in '%s'. Capturing took %d seconds\n", s, timeToCaptureSeconds)
			}
		}()
	}
}

func (t Timelapse) WaitForCapture() {
	secondsUntilFirstCapture := t.SecondsToSleepUntilOffset(time.Now())
	sleepDuration := time.Duration(secondsUntilFirstCapture) * time.Second
	nextCaptureAt := time.Now().Add(sleepDuration)

	log.Printf("Will take the next picture in %d seconds at %v.\n", secondsUntilFirstCapture, nextCaptureAt)

	for {
		secondsUntilFirstCapture := t.SecondsToSleepUntilOffset(time.Now())
		if secondsUntilFirstCapture == 0 {
			// Game time!
			break
		}
		if t.DebugEnabled {
			log.Printf("Sleeping for 1 second. Seconds left: %d. Time: %s.\n", secondsUntilFirstCapture, time.Now())
		}
		time.Sleep(time.Duration(1 * time.Second))
	}
}

func (t Timelapse) SecondsToSleepUntilOffset(currentTime time.Time) int {
	picturesPerHour := 3600 / t.SecondsBetweenCapture

	secondsIntoCurrentHour := currentTime.Minute()*60 + currentTime.Second()

	for i := 0; i < int(picturesPerHour); i++ {
		if i == 0 {
			if 0 <= secondsIntoCurrentHour && secondsIntoCurrentHour <= t.OffsetWithinHourSeconds {
				return t.OffsetWithinHourSeconds - secondsIntoCurrentHour
			}
		}

		lowerBoundary := t.OffsetWithinHourSeconds + (i-1)*t.SecondsBetweenCapture
		upperBoundary := t.OffsetWithinHourSeconds + (i)*t.SecondsBetweenCapture

		if lowerBoundary <= secondsIntoCurrentHour && secondsIntoCurrentHour <= upperBoundary {
			return upperBoundary - secondsIntoCurrentHour
		}
	}

	return 3600 - secondsIntoCurrentHour + t.OffsetWithinHourSeconds
}
