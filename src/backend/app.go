package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	_ "net/http/pprof" // registers on http.DefaultServeMux, mounted only with -pprof
	"os"
	"slices"
	"sync"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
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

// app owns the resources of a running camera: log, settings, camera and the
// listening socket.
type app struct {
	opts     options
	version  string
	level    *slog.LevelVar
	store    *conf.Store
	camera   *timelapse.Camera
	listener net.Listener
	closers  []func()
}

// newApp acquires all resources in order. On failure, those acquired so far
// are released again.
func newApp(opts options, backend timelapse.Backend) (*app, error) {
	a := &app{opts: opts, version: versionString(gitCommit, builtAt), level: new(slog.LevelVar)}
	for _, step := range []func() error{a.startLogging, a.openSettings, a.createStorageFolder, a.listen} {
		if err := step(); err != nil {
			a.close()
			return nil, err
		}
	}
	a.camera = timelapse.NewCamera(backend, captureTimeout)
	return a, nil
}

func (a *app) startLogging() error {
	closeLog, err := setupLogging(a.opts.logToFile, a.level)
	if err != nil {
		return err
	}
	a.closers = append(a.closers, closeLog)
	slog.Info("starting", "version", a.version, "options", fmt.Sprintf("%+v", a.opts))
	return nil
}

func (a *app) openSettings() error {
	store, err := conf.OpenStore(conf.SettingsFile, conf.Default(a.opts.secondsBetweenCaptures))
	if err != nil {
		return err
	}
	a.store = store
	a.closers = append(a.closers, func() { store.Close() })
	slog.Info("settings loaded", "settings", store.Get())
	return nil
}

func (a *app) createStorageFolder() error {
	if err := os.MkdirAll(a.opts.storageFolder, 0o700); err != nil {
		return fmt.Errorf("create storage folder: %w", err)
	}
	return nil
}

func (a *app) listen() (err error) {
	a.listener, err = net.Listen("tcp", a.opts.listen)
	return err
}

// close releases the resources in reverse order of acquisition.
func (a *app) close() {
	for _, c := range slices.Backward(a.closers) {
		c()
	}
}

// run starts the background loops and serves HTTP until ctx is cancelled.
func (a *app) run(ctx context.Context, ready func(addr string)) error {
	server := a.httpServer()
	var wg sync.WaitGroup
	wg.Go(func() { followDebugSetting(ctx, a.store, a.level) })
	wg.Go(func() { a.scheduler().Run(ctx) })
	err := serve(ctx, server, a.listener, ready)
	wg.Wait()
	return err
}

func (a *app) scheduler() *timelapse.Scheduler {
	return &timelapse.Scheduler{
		Settings:      a.store,
		Camera:        a.camera,
		Folder:        a.opts.storageFolder,
		Retries:       a.opts.captureRetries,
		RetryDelay:    retryDelay,
		BeforeCapture: func() { reclaimDiskSpace(a.opts.storageFolder, uint64(a.opts.minFreeDiskMB)<<20) },
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

func (a *app) httpServer() *http.Server {
	return &http.Server{
		Handler:           a.handler(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       2 * time.Minute,
		ErrorLog:          slog.NewLogLogger(slog.Default().Handler(), slog.LevelWarn),
		// No WriteTimeout: archive downloads of many photos take minutes.
	}
}

// handler routes the frontend, the API and optionally pprof.
func (a *app) handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET "+frontendPath, web.WithCacheHeaders(http.StripPrefix(frontendPath, http.FileServerFS(web.Dist()))))
	mux.Handle("GET /{$}", http.RedirectHandler(frontendPath+"index.html", http.StatusMovedPermanently))
	a.api().Register(mux)
	if a.opts.pprof {
		slog.Info("profiling enabled at /debug/pprof/")
		mux.Handle("GET /debug/pprof/", http.DefaultServeMux)
	}
	return rest.Middleware(mux, a.opts.allowedHostList()...)
}

func (a *app) api() *rest.Server {
	return &rest.Server{
		Settings:    a.store,
		Camera:      a.camera,
		Folder:      a.opts.storageFolder,
		TempFolder:  conf.TempFilesFolder,
		LogFile:     conf.LogFile, // served whenever it exists
		ThermalFile: thermalFile,
		Version:     a.version,
		Run:         rest.RunCommand,
	}
}

// serve runs server on listener until ctx is cancelled or serving fails,
// then shuts it down gracefully.
func serve(ctx context.Context, server *http.Server, listener net.Listener, ready func(addr string)) error {
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()
	announce(listener, ready)

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

func announce(listener net.Listener, ready func(addr string)) {
	addr := listener.Addr().String()
	slog.Info("listening", "address", addr)
	if ready != nil {
		ready(addr)
	}
}

func shutdown(ctx context.Context, server *http.Server) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Warn("http shutdown", "err", err)
	}
}
