package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/rest"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
	"goji.io"
	"goji.io/pat"
)

var gitCommit string
var builtAt string
var version string

//go:embed frontend/build
var content embed.FS

func main() {
	initVersion()
	if err := initLogging(); err != nil {
		log.Fatalf("Failed to initialize logging. Unable to start. Cause: %s", err.Error())
		return
	}
	s, err := conf.LoadConfiguration()
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err.Error())
		return
	}

	log.Println("Version: ", version)
	log.Printf("Seconds between captures: %d\n", s.SecondsBetweenCaptures)
	log.Printf("Offset within hour:       %d\n", s.OffsetWithinHour)
	log.Printf("Resolution:               %d x %d\n", s.PhotoResolutionWidth, s.PhotoResolutionHeight)
	log.Printf("Listening on port:        %s...\n", conf.ListenAddress)

	mux := goji.NewMux()
	// TODO redirect to new frontend
	// mux.HandleFunc(pat.Get("/"), rest.GetIndex)
	mux.HandleFunc(pat.Get("/capture"), func(w http.ResponseWriter, _ *http.Request) {
		rest.Capture(w, s)
	})

	mux.HandleFunc(pat.Get("/photos"), rest.GetPhotos)
	mux.HandleFunc(pat.Get("/monitoring"), rest.GetMonitoring)

	mux.HandleFunc(pat.Get("/file"), rest.GetFiles)
	mux.HandleFunc(pat.Get("/file/last"), rest.GetMostRecentFile)
	mux.HandleFunc(pat.Get("/file/:fileName"), rest.GetFile)
	mux.HandleFunc(pat.Get("/archive/zip"), rest.GetArchiveZip)
	mux.HandleFunc(pat.Get("/admin/:command"), rest.Admin)
	mux.HandleFunc(pat.Get("/configuration"), rest.GetConfiguration)
	mux.HandleFunc(pat.Post("/configuration"), rest.UpdateConfiguration)
	mux.HandleFunc(pat.Get("/version"), rest.MakeGetVersionFn(version))

	mux.Handle(pat.Get("/static/*"), http.StripPrefix("/static/", http.FileServer(http.FS(content))))

	t, err := timelapse.New(conf.StorageFolder, s)
	if err != nil {
		log.Printf("Error creating new timelapse instance: %s\n", err.Error())
		// Continue starting app regardless
	} else {
		// Start capturing since there were no issues
		t.CapturePeriodically()
	}

	if err := http.ListenAndServe(conf.ListenAddress, mux); err != nil {
		log.Fatal("Failed start: ", err.Error())
	}
}

func initLogging() error {
	if conf.LogToFile {
		f, err := os.OpenFile(conf.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		log.SetOutput(f)
	}
	log.Printf("Started at %s\n", time.Now())
	return nil
}

func initVersion() {
	if len(gitCommit) == 0 || len(builtAt) == 0 {
		// Fallback to v1 if either the commit or build timestamp is not set
		version = "1"
	} else {
		version = fmt.Sprintf("%s built at %s", gitCommit, builtAt)
	}
}
