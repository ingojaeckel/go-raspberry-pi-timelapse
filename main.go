package main

import (
	"fmt"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/rest"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
	"goji.io"
	"goji.io/pat"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	if err := initLogging(); err != nil {
		fmt.Errorf("Failed to initialize logging. Unable to start.")
		return
	}

	secondsBetweenCaptures := int64(60)
	offsetWithinHour := int64(0)
	width := int64(conf.PhotoResolution.Width)
	height := int64(conf.PhotoResolution.Height)

	if len(os.Args) >= 3 {
		secondsBetweenCaptures, _ = strconv.ParseInt(os.Args[1], 10, 32)
		offsetWithinHour, _ = strconv.ParseInt(os.Args[2], 10, 32)
	}
	if len(os.Args) == 5 {
		width, _ = strconv.ParseInt(os.Args[3], 10, 32)
		height, _ = strconv.ParseInt(os.Args[4], 10, 32)
	}

	log.Printf("Seconds between captures: %d\n", secondsBetweenCaptures)
	log.Printf("Offset within hour:       %d\n", offsetWithinHour)
	log.Printf("Resolution:               %d x %d\n", width, height)
	log.Printf("Listening on port:        %s...\n", conf.ListenAddress)

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/"), rest.GetIndex)
	mux.HandleFunc(pat.Get("/capture"), rest.Capture)
	mux.HandleFunc(pat.Get("/index.html"), rest.GetIndex)
	mux.HandleFunc(pat.Get("/file"), rest.GetFiles)
	mux.HandleFunc(pat.Get("/file/last"), rest.GetMostRecentFile)
	mux.HandleFunc(pat.Get("/file/:fileName"), rest.GetFile)
	mux.HandleFunc(pat.Get("/archive/zip"), rest.GetArchiveZip)
	mux.HandleFunc(pat.Get("/admin/:command"), rest.Admin)
	mux.HandleFunc(pat.Get("/version"), rest.GetVersion)

	t, err := timelapse.New("timelapse-pictures", secondsBetweenCaptures, offsetWithinHour, timelapse.Resolution{width, height})
	if err != nil {
		log.Printf("Error creating new timelapse instance: %s\n", err.Error())
		// Continue starting app regardless
	} else {
		// Start capturing since there were no issues
		t.CapturePeriodically()
	}

	http.ListenAndServe(conf.ListenAddress, mux)
}

func initLogging() error {
	f, err := os.OpenFile(conf.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	log.SetOutput(f)
	log.Printf("Started at %s\n", time.Now())
	return nil
}
