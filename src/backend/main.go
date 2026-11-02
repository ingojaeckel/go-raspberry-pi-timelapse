package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	_ "net/http/pprof" // registers on http.DefaultServeMux, mounted only with -pprof
	"os"
	"os/signal"
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

// frontendPath is the URL prefix the frontend is served under. It must match
// BASE in src/frontend/vite.config.ts. It is independent of where the files
// live on disk (see package web).
const frontendPath = "/static/frontend/build/"

const (
	captureTimeout  = time.Minute
	retryDelay      = time.Second
	shutdownTimeout = 10 * time.Second
	thermalFile     = "/sys/class/thermal/thermal_zone0/temp"
)

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

	store, err := openStore(opts)
	if err != nil {
		return err
	}
	defer store.Close()
	if err := os.MkdirAll(opts.storageFolder, 0o700); err != nil {
		return fmt.Errorf("create storage folder: %w", err)
	}
	listener, err := net.Listen("tcp", opts.listen)
	if err != nil {
		return err
	}

	camera := timelapse.NewCamera(backend, captureTimeout)
	server := newHTTPServer(newHandler(opts, newAPI(opts, store, camera, version)))
	var wg sync.WaitGroup
	wg.Go(func() { followDebugSetting(ctx, store, level) })
	wg.Go(func() { newScheduler(opts, store, camera).Run(ctx) })
	err = serve(ctx, server, listener, ready)
	wg.Wait()
	return err
}

func openStore(opts options) (*conf.Store, error) {
	store, err := conf.OpenStore(conf.SettingsFile, conf.Default(opts.secondsBetweenCaptures))
	if err != nil {
		return nil, err
	}
	slog.Info("settings loaded", "settings", store.Get())
	return store, nil
}

func newScheduler(opts options, store *conf.Store, camera *timelapse.Camera) *timelapse.Scheduler {
	return &timelapse.Scheduler{
		Settings:      store,
		Camera:        camera,
		Folder:        opts.storageFolder,
		Retries:       opts.captureRetries,
		RetryDelay:    retryDelay,
		BeforeCapture: func() { reclaimDiskSpace(opts.storageFolder, uint64(opts.minFreeDiskMB)<<20) },
	}
}

// reclaimDiskSpace deletes the oldest photos while less than minFree bytes are free.
func reclaimDiskSpace(folder string, minFree uint64) {
	removed, err := files.Reclaim(folder, minFree)
	if len(removed) > 0 {
		slog.Warn("disk almost full, deleted the oldest photos", "count", len(removed), "first", removed[0], "last", removed[len(removed)-1])
	}
	if err != nil {
		slog.Error("freeing disk space", "err", err)
	}
}

func newAPI(opts options, store *conf.Store, camera *timelapse.Camera, version string) *rest.Server {
	return &rest.Server{
		Settings:    store,
		Camera:      camera,
		Folder:      opts.storageFolder,
		TempFolder:  conf.TempFilesFolder,
		LogFile:     conf.LogFile, // served whenever it exists
		ThermalFile: thermalFile,
		Version:     version,
		Run:         rest.RunCommand,
	}
}

// newHandler routes the frontend, the API and optionally pprof.
func newHandler(opts options, api *rest.Server) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET "+frontendPath, web.WithCacheHeaders(http.StripPrefix(frontendPath, http.FileServerFS(web.Dist()))))
	mux.Handle("GET /{$}", http.RedirectHandler(frontendPath+"index.html", http.StatusMovedPermanently))
	api.Register(mux)
	if opts.pprof {
		slog.Info("profiling enabled at /debug/pprof/")
		mux.Handle("GET /debug/pprof/", http.DefaultServeMux)
	}
	return rest.Middleware(mux, opts.allowedHostList()...)
}

func newHTTPServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       2 * time.Minute,
		ErrorLog:          slog.NewLogLogger(slog.Default().Handler(), slog.LevelWarn),
		// No WriteTimeout: archive downloads of many photos take minutes.
	}
}

// serve runs server on listener until ctx is cancelled or serving fails,
// then shuts it down gracefully.
func serve(ctx context.Context, server *http.Server, listener net.Listener, ready func(addr string)) error {
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()
	slog.Info("listening", "address", listener.Addr().String())
	if ready != nil {
		ready(listener.Addr().String())
	}

	var err error
	select {
	case <-ctx.Done():
		slog.Info("shutting down")
		shutdown(ctx, server)
		err = <-serveErr
	case err = <-serveErr:
		shutdown(ctx, server)
	}
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func shutdown(ctx context.Context, server *http.Server) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Warn("http shutdown", "err", err)
	}
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
		level.Set(logLevel(store.Get().DebugEnabled))
		select {
		case <-ctx.Done():
			return
		case <-changed:
		}
	}
}

func logLevel(debug bool) slog.Level {
	if debug {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}
