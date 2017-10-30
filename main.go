package main

import (
	"fmt"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/rest"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
	"goji.io"
	"goji.io/pat"
	"net/http"
	"os"
	"strconv"
)

const ListenAddress = ":8080"

func main() {
	secondsBetweenCaptures := int64(60)
	offsetWithinHour := int64(0)

	if len(os.Args) == 3 {
		secondsBetweenCaptures, _ = strconv.ParseInt(os.Args[1], 10, 32)
		offsetWithinHour, _ = strconv.ParseInt(os.Args[2], 10, 32)
	}

	fmt.Printf("Seconds between captures: %d\n", secondsBetweenCaptures)
	fmt.Printf("Offset within hour:       %d\n", secondsBetweenCaptures)
	fmt.Printf("Listening on port:        %s...\n", ListenAddress)

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/"), rest.GetIndex)
	mux.HandleFunc(pat.Get("/index.html"), rest.GetIndex)
	mux.HandleFunc(pat.Get("/file"), rest.GetFiles)
	mux.HandleFunc(pat.Get("/file/last"), rest.GetMostRecentFile)
	mux.HandleFunc(pat.Get("/file/:fileName"), rest.GetFile)
	mux.HandleFunc(pat.Get("/archive"), rest.GetArchive)
	mux.HandleFunc(pat.Get("/admin/:command"), rest.Admin)
	mux.HandleFunc(pat.Get("/version"), rest.GetVersion)

	t, err := timelapse.New("timelapse-pictures", secondsBetweenCaptures, offsetWithinHour)
	if err != nil {
		fmt.Printf("Error creating new timelapse instance: %s\n", err.Error())
		// Continue starting app regardless
	} else {
		// Start capturing since there were no issues
		t.CapturePeriodically()
	}

	http.ListenAndServe(ListenAddress, mux)
}
