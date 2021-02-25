package main

import (
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

var GitCommit string
var BuiltAt string

func main() {

	version := fmt.Sprintf("%s built at %s", GitCommit, BuiltAt)
	log.Println("version: ", version)

	if err := initLogging(); err != nil {
		log.Fatalf("Failed to initialize logging. Unable to start. Cause: %s", err.Error())
		return
	}
	s, err := conf.LoadConfiguration()
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err.Error())
		return
	}

	log.Printf("Seconds between captures: %d\n", s.SecondsBetweenCaptures)
	log.Printf("Offset within hour:       %d\n", s.OffsetWithinHour)
	log.Printf("Resolution:               %d x %d\n", s.PhotoResolutionWidth, s.PhotoResolutionHeight)
	log.Printf("Listening on port:        %s...\n", conf.ListenAddress)

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/"), rest.GetIndex)
	mux.HandleFunc(pat.Get("/capture"), func(w http.ResponseWriter, _ *http.Request) {
		rest.Capture(w, s)
	})

	mux.HandleFunc(pat.Get("/photos"), rest.GetPhotos)
	mux.HandleFunc(pat.Get("/monitoring"), rest.GetMonitoring)

	mux.HandleFunc(pat.Get("/index.html"), rest.GetIndex)
	mux.HandleFunc(pat.Get("/file"), rest.GetFiles)
	mux.HandleFunc(pat.Get("/file/last"), rest.GetMostRecentFile)
	mux.HandleFunc(pat.Get("/file/:fileName"), rest.GetFile)
	mux.HandleFunc(pat.Get("/archive/zip"), rest.GetArchiveZip)
	mux.HandleFunc(pat.Get("/admin/:command"), rest.Admin)
	mux.HandleFunc(pat.Get("/configuration"), rest.GetConfiguration)
	mux.HandleFunc(pat.Post("/configuration"), rest.UpdateConfiguration)
	mux.HandleFunc(pat.Get("/version"), rest.MakeGetVersionFn(version))

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
