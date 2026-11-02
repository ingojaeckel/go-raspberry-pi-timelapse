package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	_ "net/http/pprof" // registers on http.DefaultServeMux, mounted only with -pprof
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/logfile"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/rest"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/web"
)

// Set by the Makefile via -ldflags; see versionString for the fallback.
var gitCommit, builtAt string

// frontendPath is the URL prefix the frontend is served under. It must match
// BASE in src/frontend/vite.config.ts. It is independent of where the files
// live on disk (see package web).
const frontendPath = "/static/frontend/build/"

const (
	captureTimeout  = time.Minute
	retryDelay      = time.Second
	shutdownTimeout = 10 * time.Second
)

type options struct {
	listen                 string
	logToFile              bool
	storageFolder          string
	secondsBetweenCaptures int
	captureRetries         int
	minFreeDiskMB          int
	pprof                  bool
	version                bool
	allowedHosts           string
}

// allowedHostList splits -allowedHosts into host names.
func (o options) allowedHostList() []string {
	var hosts []string
	for h := range strings.SplitSeq(o.allowedHosts, ",") {
		if h = strings.TrimSpace(h); h != "" {
			hosts = append(hosts, h)
		}
	}
	return hosts
}

func main() {
	opts, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if opts.version {
		fmt.Println(versionString(gitCommit, builtAt))
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, opts, timelapse.DetectBackend(), nil); err != nil {
		slog.Error("exiting", "err", err)
		os.Exit(1) // systemd restarts the service
	}
}

func parseFlags(args []string) (options, error) {
	var o options
	fs := flag.NewFlagSet("go-raspberry-pi-timelapse", flag.ContinueOnError)
	fs.BoolVar(&o.version, "version", false, "Print version and exit.")
	fs.StringVar(&o.listen, "port", conf.DefaultListenAddress, "HTTP address to listen on.")
	fs.BoolVar(&o.logToFile, "logToFile", conf.DefaultLogToFile, "Log to "+conf.LogFile+" (rotated at 10 MB) instead of stdout. Recommended for long term operation.")
	fs.StringVar(&o.storageFolder, "storageFolder", conf.DefaultStorageFolder, "Folder for storage of timelapse pictures.")
	fs.IntVar(&o.secondsBetweenCaptures, "secondsBetweenCaptures", conf.DefaultSecondsBetweenCaptures, "Seconds between captures used until settings are saved in the UI. Must divide an hour evenly, between 60 and 3600.")
	fs.IntVar(&o.captureRetries, "captureRetries", conf.DefaultCaptureRetries, "Additional attempts within the same slot when a capture fails.")
	fs.IntVar(&o.minFreeDiskMB, "minFreeDiskMB", conf.DefaultMinFreeDiskMB, "Free disk space to keep. The oldest photos are deleted before a capture to keep at least this much free.")
	fs.BoolVar(&o.pprof, "pprof", false, "Enable pprof profiling endpoints at /debug/pprof/")
	fs.StringVar(&o.allowedHosts, "allowedHosts", "", "Comma-separated public host names the UI may be reached under. IP addresses, localhost, single-label names and .local/.lan/.home.arpa/.internal names are always allowed; anything else is rejected to block DNS rebinding.")
	if err := fs.Parse(args); err != nil {
		return o, err
	}
	switch {
	case !conf.ValidInterval(o.secondsBetweenCaptures):
		return o, fmt.Errorf("-secondsBetweenCaptures must divide an hour evenly and be between %d and %d, got %d", conf.MinSecondsBetweenCaptures, conf.MaxSecondsBetweenCaptures, o.secondsBetweenCaptures)
	case o.captureRetries < 0:
		return o, fmt.Errorf("-captureRetries must not be negative")
	case o.minFreeDiskMB < 0:
		return o, fmt.Errorf("-minFreeDiskMB must not be negative")
	}
	return o, nil
}

// run starts the scheduler and the HTTP server and blocks until ctx is
// cancelled. ready, if set, receives the address the server listens on.
func run(ctx context.Context, opts options, backend timelapse.Backend, ready func(addr string)) error {
	level := new(slog.LevelVar)
	closeLog, err := setupLogging(opts.logToFile, level)
	if err != nil {
		return err
	}
	defer closeLog()
	version := versionString(gitCommit, builtAt)
	slog.Info("starting", "version", version, "options", fmt.Sprintf("%+v", opts))

	store, err := conf.OpenStore(conf.SettingsFile, conf.Default(opts.secondsBetweenCaptures))
	if err != nil {
		return err
	}
	defer store.Close()
	slog.Info("settings loaded", "settings", store.Get())

	if err := os.MkdirAll(opts.storageFolder, 0o700); err != nil {
		return fmt.Errorf("create storage folder: %w", err)
	}

	listener, err := net.Listen("tcp", opts.listen)
	if err != nil {
		return err
	}

	camera := timelapse.NewCamera(backend, captureTimeout)
	minFree := uint64(opts.minFreeDiskMB) << 20
	scheduler := &timelapse.Scheduler{
		Settings:   store,
		Camera:     camera,
		Folder:     opts.storageFolder,
		Retries:    opts.captureRetries,
		RetryDelay: retryDelay,
		BeforeCapture: func() {
			removed, err := files.Reclaim(opts.storageFolder, minFree)
			if len(removed) > 0 {
				slog.Warn("disk almost full, deleted the oldest photos", "count", len(removed), "first", removed[0], "last", removed[len(removed)-1])
			}
			if err != nil {
				slog.Error("freeing disk space", "err", err)
			}
		},
	}

	api := &rest.Server{
		Settings:    store,
		Camera:      camera,
		Folder:      opts.storageFolder,
		TempFolder:  conf.TempFilesFolder,
		LogFile:     conf.LogFile, // served whenever it exists, as before
		ThermalFile: "/sys/class/thermal/thermal_zone0/temp",
		Version:     version,
		Run:         rest.RunCommand,
	}
	mux := http.NewServeMux()
	mux.Handle("GET "+frontendPath, web.WithCacheHeaders(http.StripPrefix(frontendPath, http.FileServerFS(web.Dist()))))
	mux.Handle("GET /{$}", http.RedirectHandler(frontendPath+"index.html", http.StatusMovedPermanently))
	api.Register(mux)
	if opts.pprof {
		slog.Info("profiling enabled at /debug/pprof/")
		mux.Handle("GET /debug/pprof/", http.DefaultServeMux)
	}
	server := &http.Server{
		Handler:           rest.Middleware(mux, opts.allowedHostList()...),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       2 * time.Minute,
		ErrorLog:          slog.NewLogLogger(slog.Default().Handler(), slog.LevelWarn),
		// No WriteTimeout: archive downloads of many photos take minutes.
	}

	var wg sync.WaitGroup
	wg.Go(func() { followDebugSetting(ctx, store, level) })
	wg.Go(func() { scheduler.Run(ctx) })
	serveErr := make(chan error, 1)
	wg.Go(func() { serveErr <- server.Serve(listener) })

	slog.Info("listening", "address", listener.Addr().String())
	if ready != nil {
		ready(listener.Addr().String())
	}

	select {
	case <-ctx.Done():
		slog.Info("shutting down")
	case err = <-serveErr:
	}
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if serr := server.Shutdown(shutdownCtx); serr != nil {
		slog.Warn("http shutdown", "err", serr)
	}
	wg.Wait()
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return err
}

// setupLogging routes slog (and the standard log package) to stdout or a
// size-capped log file.
func setupLogging(toFile bool, level *slog.LevelVar) (func(), error) {
	var out io.Writer = os.Stdout
	closer := func() {}
	if toFile {
		w, err := logfile.Open(conf.LogFile, conf.LogFileMaxSize)
		if err != nil {
			return nil, fmt.Errorf("open log file: %w", err)
		}
		out, closer = w, func() { w.Close() }
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: level})))
	return closer, nil
}

// followDebugSetting switches debug logging with the DebugEnabled setting.
func followDebugSetting(ctx context.Context, store *conf.Store, level *slog.LevelVar) {
	changed := store.Subscribe()
	for {
		if store.Get().DebugEnabled {
			level.Set(slog.LevelDebug)
		} else {
			level.Set(slog.LevelInfo)
		}
		select {
		case <-ctx.Done():
			return
		case <-changed:
		}
	}
}

// versionString prefers the values injected by the Makefile and falls back
// to the VCS information Go embeds in every build.
func versionString(commit, built string) string {
	if commit != "" && built != "" {
		return fmt.Sprintf("%s built at %s", commit, built)
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		var rev, at, dirty string
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				rev = s.Value
			case "vcs.time":
				at = s.Value
			case "vcs.modified":
				if s.Value == "true" {
					dirty = "-dirty"
				}
			}
		}
		if rev != "" {
			return fmt.Sprintf("%.7s%s built at %s", rev, dirty, at)
		}
	}
	return "1"
}
