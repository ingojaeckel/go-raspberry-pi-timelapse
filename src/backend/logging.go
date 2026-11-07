package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/logfile"
)

// setupLogging routes slog (and the standard log package) to stdout or a
// size-capped log file.
func setupLogging(toFile bool, level *slog.LevelVar) (func(), error) {
	var out io.Writer = os.Stdout
	closer := func() {}
	if toFile {
		w, err := logfile.Open(logFile, logFileMaxSize)
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
