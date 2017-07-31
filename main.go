package main

import (
	"fmt"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/rest"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
	"goji.io"
	"goji.io/pat"
	"net/http"
	"time"
)

func main() {
	addr := ":8080"
	fmt.Printf("Listening on %s...\n", addr)

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/"), rest.GetIndex)
	mux.HandleFunc(pat.Get("/index.html"), rest.GetIndex)
	mux.HandleFunc(pat.Get("/file"), rest.GetFiles)
	mux.HandleFunc(pat.Get("/file/last"), rest.GetMostRecentFile)
	mux.HandleFunc(pat.Get("/file/:fileName"), rest.GetFile)
	mux.HandleFunc(pat.Get("/archive"), rest.GetArchive)
	mux.HandleFunc(pat.Get("/admin/:command"), rest.Admin)
	mux.HandleFunc(pat.Get("/version"), rest.GetVersion)

	t, err := timelapse.New("timelapse-pictures", 60)
	if err != nil {
		fmt.Printf("Error creating new timelapse instance: %s\n", err.Error())
		// Continue starting app regardless
	} else {
		// Start capturing since there were no issues
		t.CapturePeriodically()
	}

	http.ListenAndServe(addr, mux)
}
