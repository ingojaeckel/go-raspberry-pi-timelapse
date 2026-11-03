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
	// writeTimeout bounds every response; downloads extend it while they
	// make progress (see rest.deadlineWriter).
	writeTimeout = 2 * time.Minute
	thermalFile  = "/sys/class/thermal/thermal_zone0/temp"
)

// app owns the resources of a running camera: log, settings, camera and the
// listening socket.
type app struct {
	opts     options
	version  string
	level    *slog.LevelVar
	store    *conf.Store
	camera   *timelapse.Camera
	photos   *files.Index // shared by the API and disk space reclaiming
	listener net.Listener
	closers  []func()
	// freeBytes reports the free space of the photo folder's file system;
	// tests simulate a disk.
	freeBytes func(dir string) (uint64, error)
}

// newApp acquires all resources in order. On failure, those acquired so far
// are released again.
func newApp(opts options, backend timelapse.Backend) (*app, error) {
	a := &app{opts: opts, version: versionString(gitCommit, builtAt), level: new(slog.LevelVar), freeBytes: files.FreeBytes}
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

// createStorageFolder creates the photo folder if needed and deletes photos
// that were still being written when the camera last stopped (power cut).
func (a *app) createStorageFolder() error {
	if err := os.MkdirAll(a.opts.storageFolder, 0o700); err != nil {
		return fmt.Errorf("create storage folder: %w", err)
	}
	for _, dir := range []string{a.opts.storageFolder, conf.TempFilesFolder} {
		if removed, err := timelapse.RemovePartials(dir); err != nil {
			slog.Warn("removing unfinished photos", "folder", dir, "err", err)
		} else if len(removed) > 0 {
			slog.Warn("removed unfinished photos from before the last stop", "folder", dir, "count", len(removed))
		}
	}
	a.photos = files.NewIndex(a.opts.storageFolder)
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

// run starts the background loops and serves HTTP until ctx is cancelled or
// serving fails; either way the background loops are stopped before it returns.
func (a *app) run(ctx context.Context, ready func(addr string)) error {
	ctx, stop := context.WithCancel(ctx)
	defer stop()
	server := a.httpServer()
	sched := a.scheduler()
	var wg sync.WaitGroup
	wg.Go(func() { followDebugSetting(ctx, a.store, a.level) })
	if interval, ok := watchdogInterval(); ok {
		w := newWatchdog(a.maxStall())
		sched.Heartbeat = w.beat
		wg.Go(func() { w.run(ctx, interval) })
		slog.Info("systemd watchdog enabled", "interval", interval, "maxStall", w.maxStall)
	}
	wg.Go(func() { sched.Run(ctx) })
	err := serve(ctx, server, a.listener, func(addr string) {
		sdNotify("READY=1") // fails harmlessly when not run by systemd
		if ready != nil {
			ready(addr)
		}
	})
	stop()
	wg.Wait()
	return err
}

// maxStall is the longest the scheduler loop may go without waking up: all
// attempts of one capture timing out, plus time to free disk space.
func (a *app) maxStall() time.Duration {
	attempts := time.Duration(a.opts.captureRetries + 1)
	return attempts*(captureTimeout+retryDelay+timelapse.KillDelay) + 5*time.Minute
}

func (a *app) scheduler() *timelapse.Scheduler {
	return &timelapse.Scheduler{
		Settings:      a.store,
		Camera:        a.camera,
		Folder:        a.opts.storageFolder,
		Retries:       a.opts.captureRetries,
		RetryDelay:    retryDelay,
		BeforeCapture: a.reclaimDiskSpace,
	}
}

// reclaimDiskSpace deletes the oldest photos while less than -minFreeDiskMB
// are free.
func (a *app) reclaimDiskSpace() {
	removed, err := a.photos.Reclaim(uint64(a.opts.minFreeDiskMB)<<20, func() (uint64, error) {
		return a.freeBytes(a.photos.Dir())
	})
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
		WriteTimeout:      writeTimeout,
		IdleTimeout:       2 * time.Minute,
		ErrorLog:          slog.NewLogLogger(slog.Default().Handler(), slog.LevelWarn),
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
		Photos:      a.photos,
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
