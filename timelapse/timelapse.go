package timelapse

import (
	"log"
	"os"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
)

func New(folder string, initialSettings conf.Settings, configUpdatedChan <-chan conf.Settings) (*Timelapse, error) {
	_, err := os.Stat(folder)
	createFolder := err != nil && os.IsNotExist(err)

	if createFolder {
		if err := os.Mkdir(folder, 0700); err != nil {
			return nil, err
		}
	}
	// Assume folder exists

	return &Timelapse{
		Folder:              folder,
		Settings:            initialSettings,
		ConfigUpdateChannel: configUpdatedChan,
	}, nil
}

func (t Timelapse) CapturePeriodically() {
	offsetDisabled := t.Settings.OffsetWithinHour == -1

	if offsetDisabled {
		log.Println("Offset is disabled. Will start taking pictures immediately.")
		go func() {
			for {
				camera, err := NewCamera(t.Folder, t.Settings.PhotoResolutionWidth, t.Settings.PhotoResolutionHeight, t.Settings.RotateBy == 180, t.Settings.Quality)
				if err != nil {
					log.Printf("Error instantiating camera: %s\n", err)
					// Sleep for a bit and create a new camera instance on the next iteration.
				} else {
					s, err := camera.Capture()
					if err != nil {
						log.Printf("Error during capture: %s\n", err.Error())
					}
					log.Printf("Photo stored in '%s'. Will sleep for %d seconds.\n", s, t.Settings.SecondsBetweenCaptures)
				}
				time.Sleep(time.Duration(t.Settings.SecondsBetweenCaptures) * time.Second)
			}
		}()
	} else {
		log.Println("Offset is enabled. Will wait before taking first picture.")
		go func() {
			for {
				t.waitForCapture()

				beforeCapture := time.Now()

				camera, err := NewCamera(t.Folder, t.Settings.PhotoResolutionWidth, t.Settings.PhotoResolutionHeight, t.Settings.RotateBy == 180, t.Settings.Quality)
				if err != nil {
					log.Printf("Error instantiating camera: %s\n", err)
					// Sleep for a bit and create a new camera instance on the next iteration.
				} else {
					photoPath, err := camera.Capture()
					if err != nil {
						log.Printf("Error during capture: %s\n", err.Error())

						// Sleep for 1s after an error to ensure time changed sufficiently before next invocation of WaitForCapture
						time.Sleep(time.Duration(1 * time.Second))
						continue
					}
					log.Printf("Photo stored in '%s'\n", photoPath)
				}
				timeToCaptureSeconds := time.Now().Unix() - beforeCapture.Unix()
				log.Printf("Capture took %d seconds\n", timeToCaptureSeconds)
			}
		}()
	}
}

func (t *Timelapse) waitForCapture() {
	secondsUntilFirstCapture := t.secondsToSleepUntilOffset(time.Now())
	sleepDuration := time.Duration(secondsUntilFirstCapture) * time.Second
	nextCaptureAt := time.Now().Add(sleepDuration)

	log.Printf("Will take the next picture in %d seconds at %v.\n", secondsUntilFirstCapture, nextCaptureAt)

	for {
		secondsUntilFirstCapture := t.secondsToSleepUntilOffset(time.Now())
		if secondsUntilFirstCapture == 0 {
			// Game time!
			break
		}
		if t.Settings.DebugEnabled {
			log.Printf("Sleeping for 1 second. Seconds left: %d. Time: %s.\n", secondsUntilFirstCapture, time.Now())
		}
		select {
		case newConfig := <-t.ConfigUpdateChannel:
			log.Printf("Received new configuration: %s\n", newConfig)
			t.Settings = newConfig
			break
		case <-time.After(time.Duration(1 * time.Second)):
			break
		}

	}
}

func (t Timelapse) secondsToSleepUntilOffset(currentTime time.Time) int {
	picturesPerHour := 3600 / t.Settings.SecondsBetweenCaptures

	secondsIntoCurrentHour := currentTime.Minute()*60 + currentTime.Second()

	for i := 0; i < int(picturesPerHour); i++ {
		if i == 0 {
			if 0 <= secondsIntoCurrentHour && secondsIntoCurrentHour <= t.Settings.OffsetWithinHour {
				return t.Settings.OffsetWithinHour - secondsIntoCurrentHour
			}
		}

		lowerBoundary := t.Settings.OffsetWithinHour + (i-1)*t.Settings.SecondsBetweenCaptures
		upperBoundary := t.Settings.OffsetWithinHour + (i)*t.Settings.SecondsBetweenCaptures

		if lowerBoundary <= secondsIntoCurrentHour && secondsIntoCurrentHour <= upperBoundary {
			return upperBoundary - secondsIntoCurrentHour
		}
	}

	return 3600 - secondsIntoCurrentHour + t.Settings.OffsetWithinHour
}
