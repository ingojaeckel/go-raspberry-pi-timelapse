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
	secondsBetweenCaptures := flag.Int("secondsBetweenCaptures", conf.DefaultSecondsBetweenCaptures, "Number of seconds between captures")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		return
	}
	
	// Validate CLI flags before applying them
	if err := validateCLIFlags(secondsBetweenCaptures); err != nil {
		log.Fatalf("Invalid CLI flags: %s", err.Error())
		return
	}
	
	conf.OverrideDefaultConfig(listenAddress, storageAddress, logToFile, secondsBetweenCaptures)
	if err := initLogging(); err != nil {
		log.Fatalf("Failed to initialize logging. Unable to start. Cause: %s", err.Error())
		return
	}

	initialSettings, err := conf.LoadConfiguration()
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err.Error())
		return
	}
	
	// Apply CLI overrides with proper priority: CLI flags override persisted settings
	*initialSettings = initialSettings.ApplyCLIOverrides(secondsBetweenCaptures)
	
	log.Printf("Settings:       %s\n", *initialSettings)
	log.Printf("Listen address: %s\n", conf.ListenAddress)

	mux := goji.NewMux()
	mux.Use(func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			before := time.Now()
			inner.ServeHTTP(w, r)
			duration := time.Now().Sub(before)
			log.Printf("[%s] %s took %v\n", r.Method, r.RequestURI, duration)
		})
	})

	// Frontend APIs
	mux.Handle(pat.Get("/static/*"), http.StripPrefix("/static/", http.FileServer(http.FS(content))))
	// Redirect to frontend. Ideally this could be built in a way it can be hosted on or closer to /.
	mux.Handle(pat.Get("/"), http.RedirectHandler("/static/frontend/build/index.html", http.StatusMovedPermanently))

	// Backend APIs (should only be called by frontend code)
	mux.HandleFunc(pat.Get("/capture"), func(w http.ResponseWriter, _ *http.Request) {
		rest.Capture(w, initialSettings)
	})

	configUpdatedChan := make(chan conf.Settings)

	mux.HandleFunc(pat.Get("/logs"), rest.GetLogs)
	mux.HandleFunc(pat.Get("/photos"), rest.GetPhotos)
	mux.HandleFunc(pat.Get("/monitoring"), rest.GetMonitoring)
	mux.HandleFunc(pat.Get("/file"), rest.GetFiles)
	mux.HandleFunc(pat.Get("/file/delete"), rest.DeleteFiles)
	mux.HandleFunc(pat.Get("/file/last"), rest.GetMostRecentFile)
	mux.HandleFunc(pat.Get("/file/:fileName"), rest.GetFile)
	mux.HandleFunc(pat.Get("/archive/zip"), rest.GetArchiveZip)
	mux.HandleFunc(pat.Get("/archive/tar"), rest.GetArchiveTar)
	mux.HandleFunc(pat.Get("/admin/:command"), rest.Admin)
	mux.HandleFunc(pat.Get("/configuration"), rest.GetConfiguration)
	mux.HandleFunc(pat.Options("/configuration"), rest.GetConfiguration)
	mux.HandleFunc(pat.Post("/configuration"), rest.MakeUpdateConfigurationFn(configUpdatedChan))
	mux.HandleFunc(pat.Get("/version"), rest.MakeGetVersionFn(version))

	t, err := timelapse.New(conf.StorageFolder, *initialSettings, configUpdatedChan)
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
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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

func validateCLIFlags(secondsBetweenCaptures *int) error {
	if secondsBetweenCaptures != nil && *secondsBetweenCaptures < conf.MinSecondsBetweenCaptures {
		return fmt.Errorf("secondsBetweenCaptures must be at least %d seconds to allow sufficient exposure time (got %d)", conf.MinSecondsBetweenCaptures, *secondsBetweenCaptures)
	}
	return nil
}
