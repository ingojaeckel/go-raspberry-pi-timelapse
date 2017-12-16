package main

import (
	"fmt"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/rest"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
	"goji.io"
	"goji.io/pat"
	"net/http"
	"os"
	"strconv"
)

func main() {
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

	fmt.Printf("Seconds between captures: %d\n", secondsBetweenCaptures)
	fmt.Printf("Offset within hour:       %d\n", offsetWithinHour)
	fmt.Printf("Resolution:               %d x %d\n", width, height)
	fmt.Printf("Listening on port:        %s...\n", conf.ListenAddress)

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/"), rest.GetIndex)
	mux.HandleFunc(pat.Get("/capture"), rest.Capture)
	mux.HandleFunc(pat.Get("/index.html"), rest.GetIndex)
	mux.HandleFunc(pat.Get("/file"), rest.GetFiles)
	mux.HandleFunc(pat.Get("/file/last"), rest.GetMostRecentFile)
	mux.HandleFunc(pat.Get("/file/:fileName"), rest.GetFile)
	mux.HandleFunc(pat.Get("/archive/tar"), rest.GetArchive)
	mux.HandleFunc(pat.Get("/archive/zip"), rest.GetArchiveZip)
	mux.HandleFunc(pat.Get("/admin/:command"), rest.Admin)
	mux.HandleFunc(pat.Get("/version"), rest.GetVersion)

	t, err := timelapse.New("timelapse-pictures", secondsBetweenCaptures, offsetWithinHour, timelapse.Resolution{width, height})
	if err != nil {
		fmt.Printf("Error creating new timelapse instance: %s\n", err.Error())
		// Continue starting app regardless
	} else {
		// Start capturing since there were no issues
		t.CapturePeriodically()
	}

	http.ListenAndServe(conf.ListenAddress, mux)
}
