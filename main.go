package main

import (
	"embed"
	"flag"
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

	versionFlag := flag.Bool("version", false, "Print version and exit.")
	listenAddress := flag.String("port", conf.DefaultListenAddress, "HTTP port to listen on.")
	logToFile := flag.Bool("logToFile", conf.DefaultLogToFile, "Toggle to enable logging to a file on disk instead of stdout. Logging to a file is recommended for long term operation.")
	storageAddress := flag.String("storageFolder", conf.DefaultStorageFolder, "Folder for storage of timelapse pictures.")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		return
	}
	conf.Update(listenAddress, storageAddress, logToFile)
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
	log.Printf("Listen address:           %s\n", conf.ListenAddress)

	mux := goji.NewMux()

	// Frontend APIs
	mux.Handle(pat.Get("/static/*"), http.StripPrefix("/static/", http.FileServer(http.FS(content))))
	// Redirect to frontend. Ideally this could be built in a way it can be hosted on or closer to /.
	mux.Handle(pat.Get("/"), http.RedirectHandler("/static/frontend/build/index.html", http.StatusMovedPermanently))

	// Backend APIs (should only be called by frontend code)
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

	t, err := timelapse.New(conf.StorageFolder, s)
	if err != nil {
		log.Printf("Error creating new timelapse instance: %s\n", err.Error())
		// Continue starting app regardless
	} else {
		// Start capturing since there were no issues
		t.CapturePeriodically()
	}

	log.Println("Listening...")
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
	log.Printf("Version %s started at %s\n", version, time.Now())
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
