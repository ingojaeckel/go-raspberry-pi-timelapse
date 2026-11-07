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
	"sync/atomic"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/health"
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
	// rtcDevice is the clock module; sudoers allows hwclock for it only.
	rtcDevice = "/dev/rtc0"

	// The Pi's health is sampled every healthSampleInterval (the firmware
	// remembers power problems since boot, so short ones are not missed)
	// and summarized in the log every summaryInterval of uptime: the clock
	// may be wrong, so not at midnight.
	healthSampleInterval = 10 * time.Minute
	summaryInterval      = 24 * time.Hour
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
	// runCommand executes system commands (rest.RunCommand); thermalFile and
	// cpuFreqDir are read for the health readings. Empty or nil in tests.
	runCommand  health.Run
	thermalFile string
	cpuFreqDir  string
	started     time.Time
	record      health.Recorder // for the daily summary
	// sensors detects the camera; nil for backends that cannot list cameras.
	sensors *timelapse.Sensors
	deleted atomic.Int64 // photos deleted to free disk space
	// sched is the running scheduler, for the API's next capture.
	sched atomic.Pointer[timelapse.Scheduler]
}

// newApp acquires all resources in order. On failure, those acquired so far
// are released again.
func newApp(opts options, backend timelapse.Backend) (*app, error) {
	a := &app{
		opts: opts, version: versionString(gitCommit, builtAt), level: new(slog.LevelVar), freeBytes: files.FreeBytes,
		runCommand: rest.RunCommand, thermalFile: thermalFile, cpuFreqDir: health.CPUFreqDir, started: time.Now(),
	}
	if l, ok := backend.(timelapse.SensorLister); ok {
		a.sensors = l.Sensors()
	}
	for _, step := range []func() error{a.startLogging, a.markRun, a.openSettings, a.createStorageFolder, a.listen} {
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

// markRun logs how the previous run ended, and marks this one as running
// until close. Without a marker (read-only folder) the camera still runs.
func (a *app) markRun() error {
	previous, done, err := health.StartRun(conf.RunMarkerFile, health.BootIDFile)
	if err != nil {
		slog.Warn("cannot tell how the previous run ended", "err", err)
		return nil
	}
	a.closers = append(a.closers, func() {
		if err := done(); err != nil {
			slog.Warn("removing the run marker", "err", err)
		}
	})
	if previous == health.StoppedCleanly {
		slog.Info("previous run stopped cleanly")
	} else {
		slog.Warn("previous run did not stop cleanly: "+previous.String(), "previous", int(previous))
	}
	return nil
}

// openSettings loads the settings. On the first run the photo size is the
// detected camera's default; a saved size the camera does not offer (e.g.
// after swapping the camera) is kept, but logged.
func (a *app) openSettings() error {
	defaults := conf.Default(a.opts.secondsBetweenCaptures)
	sensor, detected := a.detectCamera()
	if detected {
		d := sensor.DefaultResolution()
		defaults.PhotoResolutionWidth, defaults.PhotoResolutionHeight = d.Width, d.Height
	}
	store, err := conf.OpenStore(conf.SettingsFile, defaults)
	if err != nil {
		return err
	}
	a.store = store
	a.closers = append(a.closers, func() { store.Close() })
	s := store.Get()
	slog.Info("settings loaded", "settings", s)
	if detected && !sensor.Supports(s.PhotoResolutionWidth, s.PhotoResolutionHeight) {
		slog.Warn("the photo size is not one of the camera's sizes, so photos are scaled; choose one in Settings",
			"size", fmt.Sprintf("%dx%d", s.PhotoResolutionWidth, s.PhotoResolutionHeight), "supported", sensor.Resolutions())
	}
	return nil
}

// cameraDetectTimeout bounds camera detection at start-up, which must not
// hold up the web UI for long.
const cameraDetectTimeout = 15 * time.Second

func (a *app) detectCamera() (timelapse.Sensor, bool) {
	if a.sensors == nil {
		return timelapse.Sensor{}, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), cameraDetectTimeout)
	defer cancel()
	sensor, err := a.sensors.Get(ctx)
	if err != nil {
		slog.Warn("camera not detected", "err", err)
		return sensor, false
	}
	slog.Info("camera detected", "model", sensor.Model, "resolutions", sensor.Resolutions())
	return sensor, true
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
	sched := a.scheduler()
	a.sched.Store(sched) // before serving: /schedule reports its plan
	server := a.httpServer()
	var wg sync.WaitGroup
	wg.Go(func() { followDebugSetting(ctx, a.store, a.level) })
	a.startWatchdog(ctx, &wg, sched)
	wg.Go(func() { sched.Run(ctx) })
	wg.Go(func() { a.watchHealth(ctx) })
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

// startWatchdog pings systemd's watchdog while sched keeps waking up, if
// systemd asks for it.
func (a *app) startWatchdog(ctx context.Context, wg *sync.WaitGroup, sched *timelapse.Scheduler) {
	interval, ok := watchdogInterval()
	if !ok {
		return
	}
	w := newWatchdog(a.maxStall())
	sched.Heartbeat = w.beat
	wg.Go(func() { w.run(ctx, interval) })
	slog.Info("systemd watchdog enabled", "interval", interval, "maxStall", w.maxStall)
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
		BeforeCapture: a.prepareDisk,
		Captured:      a.record.Captured,
	}
}

var errDiskFull = errors.New("disk full")

// prepareDisk keeps -minFreeDiskMB free before every capture: by deleting
// the oldest photos or, with StopWhenDiskFull, by not taking the photo.
func (a *app) prepareDisk() error {
	if !a.store.Get().StopWhenDiskFull {
		a.reclaimDiskSpace()
		return nil
	}
	free, err := a.freeBytes(a.photos.Dir())
	if err != nil {
		slog.Error("checking free disk space", "err", err)
		return nil // try anyway
	}
	if free < a.minFreeBytes() {
		return fmt.Errorf("%w: %d MB free, keeping %d MB (-minFreeDiskMB); delete photos or turn off \"stop when full\"", errDiskFull, free>>20, a.opts.minFreeDiskMB)
	}
	return nil
}

func (a *app) minFreeBytes() uint64 { return uint64(a.opts.minFreeDiskMB) << 20 }

// reclaimDiskSpace deletes the oldest photos while less than -minFreeDiskMB
// are free.
func (a *app) reclaimDiskSpace() {
	removed, err := a.photos.Reclaim(a.minFreeBytes(), func() (uint64, error) {
		return a.freeBytes(a.photos.Dir())
	})
	a.deleted.Add(int64(len(removed)))
	a.record.Deleted(len(removed))
	if len(removed) > 0 {
		slog.Warn("disk almost full, deleted the oldest photos", "count", len(removed), "first", removed[0], "last", removed[len(removed)-1])
	}
	if err != nil {
		slog.Error("freeing disk space", "err", err)
	}
}

// watchHealth samples the Pi's health and logs a summary every day until ctx
// ends.
func (a *app) watchHealth(ctx context.Context) {
	sample := time.NewTicker(healthSampleInterval)
	defer sample.Stop()
	summary := time.NewTicker(summaryInterval)
	defer summary.Stop()
	a.sampleHealth(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-sample.C:
			a.sampleHealth(ctx)
		case <-summary.C:
			a.logSummary()
		}
	}
}

func (a *app) sampleHealth(ctx context.Context) {
	var s health.Sample
	if c, err := health.ReadCelsius(a.thermalFile); err == nil {
		s.Celsius = &c
	}
	if a.runCommand != nil {
		if t, err := health.ReadThrottled(ctx, a.runCommand); err == nil {
			s.Throttled = &t
		}
	}
	a.record.Sample(s)
}

// logSummary logs what happened since the previous summary: what someone
// visiting the camera after weeks wants to know first.
func (a *app) logSummary() {
	s := a.record.Summary()
	attrs := []any{
		"captured", s.Captured, "failed", s.Failed, "deletedOldest", s.DeletedOldest,
		"throttledSinceBoot", s.ThrottledSinceBoot, "uptime", time.Since(a.started).Round(time.Minute),
	}
	if s.LastError != "" {
		attrs = append(attrs, "lastError", s.LastError)
	}
	if s.MinCelsius != nil {
		attrs = append(attrs, "minCelsius", *s.MinCelsius, "maxCelsius", *s.MaxCelsius)
	}
	if list, err := a.photos.List(); err == nil {
		attrs = append(attrs, "photos", len(list))
	}
	if free, err := a.freeBytes(a.photos.Dir()); err == nil {
		attrs = append(attrs, "freeMB", free>>20)
	}
	level := slog.LevelInfo
	if s.Failed > 0 || len(s.ThrottledSinceBoot) > 0 {
		level = slog.LevelWarn
	}
	slog.Log(context.Background(), level, "daily summary", attrs...)
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
		Settings:      a.store,
		Camera:        a.camera,
		Photos:        a.photos,
		TempFolder:    conf.TempFilesFolder,
		LogFile:       conf.LogFile, // served whenever it exists
		ThermalFile:   a.thermalFile,
		CPUFreqDir:    a.cpuFreqDir,
		Version:       a.version,
		Run:           a.runCommand,
		FreeBytes:     a.freeBytes,
		ReserveBytes:  a.minFreeBytes(),
		DeletedOldest: func() int { return int(a.deleted.Load()) },
		SetClock:      rest.SetSystemClock,
		RTCDevice:     rtcDevice,
		Sensor:        a.sensorFunc(),
		NextCapture:   a.nextCapture,
	}
}

// nextCapture is the running scheduler's next capture.
func (a *app) nextCapture() (timelapse.Plan, bool) {
	if s := a.sched.Load(); s != nil {
		return s.Next()
	}
	return timelapse.Plan{}, false
}

func (a *app) sensorFunc() func(context.Context) (timelapse.Sensor, error) {
	if a.sensors == nil {
		return nil
	}
	return a.sensors.Get
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
