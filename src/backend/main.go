package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/health"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/rest"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/selfcheck"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
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
	if opts.selfCheck {
		code := selfCheck(ctx, opts, timelapse.DetectBackend(), os.Stdout)
		stop()
		os.Exit(code)
	}
	if err := run(ctx, opts, timelapse.DetectBackend(), nil); err != nil {
		slog.Error("exiting", "err", err)
		os.Exit(1) // systemd restarts the service
	}
}

// run starts the scheduler and the HTTP server and blocks until ctx is
// cancelled. ready, if set, receives the address the server listens on.
func run(ctx context.Context, opts options, backend timelapse.Backend, ready func(addr string)) error {
	a, err := newApp(opts, backend)
	if err != nil {
		return err
	}
	defer a.close()
	return a.run(ctx, ready)
}

// selfCheck takes a test photo and checks the Pi, prints the results to w
// and returns the exit code: 0 when no check failed. It only reads the
// settings, so it can run next to the app.
func selfCheck(ctx context.Context, opts options, backend timelapse.Backend, w io.Writer) int {
	fmt.Fprintf(w, "timelapse self-check, %s\n\n", versionString(gitCommit, builtAt))
	results := selfcheck.Run(ctx, selfcheck.Config{
		Backend:         backend,
		CaptureTimeout:  captureTimeout,
		StorageFolder:   opts.storageFolder,
		SettingsFile:    conf.SettingsFile,
		DefaultSettings: conf.Default(opts.secondsBetweenCaptures),
		MinFreeBytes:    uint64(opts.minFreeDiskMB) << 20,
		MinBrightness:   float64(opts.selfCheckMinBrightness),
		Keep:            opts.selfCheckKeep,
		FreeBytes:       files.FreeBytes,
		Run:             rest.RunCommand,
		ThermalFile:     thermalFile,
		CPUFreqDir:      health.CPUFreqDir,
		ProcDir:         "/proc",
	})
	if !selfcheck.Write(w, results) {
		return 1
	}
	return 0
}
