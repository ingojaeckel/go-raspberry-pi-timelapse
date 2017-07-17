package main

import (
	"fmt"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/rest"
	"goji.io"
	"goji.io/pat"
	"net/http"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
	"time"
)

func main() {
	addr := ":8080"
	fmt.Printf("Listening on %s...\n", addr)

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/version"), rest.GetVersion)

	mux.HandleFunc(pat.Get("/file"), rest.GetFiles)
	mux.HandleFunc(pat.Get("/file/:fileName"), rest.GetFile)

	mux.HandleFunc(pat.Get("/archive"), rest.GetArchive)

	t, err := timelapse.New("timelapse-pictures", 1 * time.Minute)
	if err != nil {
		fmt.Printf("Error creating new timelapse instance: %s\n", err.Error())
		// Continue starting app regardless
	} else {
		// Start capturing since there were no issues
		t.CapturePeriodically()
	}

	http.ListenAndServe(addr, mux)
}
