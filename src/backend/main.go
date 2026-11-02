package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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
