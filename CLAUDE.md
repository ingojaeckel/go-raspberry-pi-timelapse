# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Raspberry Pi Zero W timelapse camera system: a Go HTTP server with embedded React frontend that captures images on a schedule, serves a web UI, and runs as a WiFi access point. 
## Build & Test Commands

```bash
# From the repo root (Makefile is the entry point):
make            # frontend build + tests, Go vet + tests, dev build
make pi         # same, then cross-compile for Pi Zero W (ARM v6) into src/backend/bin/arm/
make test       # frontend tests + Go vet/tests
make frontend   # build frontend into src/backend/web/dist
make run        # build frontend, run server on :8080
make loadtest   # k6 load test (requires k6)

# Individual Go tests (Go module root is src/backend; frontend must be built first)
cd src/backend && go test -race -cover ./timelapse/
cd src/backend && go test -run TestSpecificName ./timelapse/

# Frontend only
cd src/frontend && npm install && npm test && npm run build
```

## Repository Layout

- `src/backend/` — Go module (`go.mod`, `main.go`, packages below). Committed Pi binary at `src/backend/bin/arm/`.
- `src/frontend/` — React app; Vite builds into `src/backend/web/dist` (gitignored).
- `src/installer/custompios/` — CustomPiOS Pi image build; `src/installer/device/` — on-device scripts (`setup.sh`, `rtc.sh`).
- `src/perf-testing/k6/` — k6 load tests + `run-load-test.sh`.
- `docs/` — hardware, installer, load-testing, benchmarks docs; images in `docs/images/`.

**Path couplings to keep in sync when moving things:**
- `//go:embed` cannot use `..`: `src/backend/web/embed.go` embeds `dist`, and `src/frontend/vite.config.ts` `OUT_DIR` writes there.
- URL prefix `/static/frontend/build/` = `frontendPath` in `main.go` = Vite `BASE` (independent of disk layout).
- `timelapse.service` (`src/installer/custompios/modules/timelapsepi/filesystem/root/etc/systemd/system/`) runs `/home/pi/go-raspberry-pi-timelapse/src/backend/bin/arm/go-raspberry-pi-timelapse` from a git clone of the default branch — moving the binary requires updating `ExecStart` in the same commit.

## Architecture

All Go paths below are relative to `src/backend/`.

**Go packages:**
- `main.go` — Entrypoint: flags, signal handling, `run` (= `newApp` + `app.run`)
- `app.go` — `app` owns the resources (log, settings store, camera, listener) and releases them in reverse order; wires scheduler, API, `http.ServeMux` (frontend, API, pprof) and `http.Server` (timeouts, graceful shutdown on SIGINT/SIGTERM)
- `flags.go` — CLI options; `logging.go` — slog setup and debug level; `version.go` — version string
- `web/` — `//go:embed dist` of the frontend build
- `timelapse/` — `Scheduler.Run` (capture loop: aligned or unaligned schedule, retries within the slot, re-plans immediately on settings changes), `NextAligned`, `Camera` (one capture at a time, timeout, writes `.partial-*` then renames, never overwrites)
- `rest/` — `Server` with the HTTP handlers (`config.go`, `photos.go`, `system.go`), `Middleware` (Host allowlist against DNS rebinding, `http.CrossOriginProtection`, no CORS headers, debug request log). GET never changes state: `DELETE /file?f=`, `DELETE /photos`, `POST /admin/{shutdown,restart}`
- `conf/` — `Settings` + `Validate`/`Normalized`; `Store` keeps bbolt (`timelapse-settings.db`) open, serves settings from memory, notifies subscribers, recovers from a corrupt file
- `files/` — Listing (oldest first, skips hidden/partial files), streaming ZIP/TAR, `RemoveFile` and `Reclaim` (delete oldest photos when free space < `-minFreeDiskMB`, ring buffer) via `os.Root`
- `logfile/` — Size-capped rotating log (`timelapse.log` → `.1` at 10 MB) and `Tail` for `/logs`

**Key data flow:** `POST /configuration` → `conf.Store.Update` (validates, persists, notifies). The scheduler and the debug log level subscribe to the store and read the latest settings with `Get()`; nothing is passed through buffered channels that could drop updates.

**Schedule rules:** `SecondsBetweenCaptures` must divide 3600 and be 60..3600; `OffsetWithinHour` is -1 (unaligned, every N seconds from start-up) or 0..interval-1. Anything else is a 400. Settings saved by older versions are repaired on load by `Settings.Normalized` (same capture times where possible).

**Platform abstraction:** `timelapse.DetectBackend` uses `rpicam-still` when it is on `PATH` (the Pi), otherwise `ffmpeg` webcam capture (Linux v4l2 / macOS avfoundation). It does not look at GOARCH, so Apple Silicon Macs use ffmpeg.

**Frontend:** React 19 + TypeScript + Vite, no UI framework (plain CSS, `fetch`), built into `src/backend/web/dist/` and embedded into the Go binary. Hash-based navigation; API types in `src/frontend/src/api.ts` mirror the Go models. `src/frontend/src/schedule.ts` ports `timelapse.NextAligned` and `conf.ValidInterval`; both sides are tested against `src/frontend/src/schedule-cases.json`, so change that file together with either implementation.

## Build Flags

Version info is injected via ldflags: `-X 'main.gitCommit=...' -X 'main.builtAt=...'`

If ldflags are not set, the version comes from the VCS info Go embeds (`debug.ReadBuildInfo`).

CLI flags: `-port`, `-logToFile`, `-storageFolder`, `-secondsBetweenCaptures` (first-run default only; once settings are saved the database wins), `-captureRetries` (default 2), `-minFreeDiskMB` (default 256), `-allowedHosts` (extra public host names; IPs, localhost, single-label and .local/.lan/.home.arpa/.internal names are always allowed), `-pprof`, `-version`

Go version: `go.mod` pins `go 1.27.0` / `toolchain go1.27.3`. Dependencies: `go.etcd.io/bbolt`, `golang.org/x/sys` only.

## Testing Conventions

- Plain `testing` (no assertion library). Time-dependent code is tested with `testing/synctest` (fake clock: hours of scheduling run instantly); HTTP handlers through the real mux + middleware with `httptest.NewServer`; file system code against `t.TempDir()`
- Tests must not depend on the working directory or wall clock; use `t.TempDir()`, `t.Chdir()`, `synctest`
- Test files are colocated with source (e.g., `timelapse/scheduler_test.go`)
- `make` / `make pi` run frontend tests, `go vet ./...` and `go test -race -shuffle=on -cover ./...` before building
- Frontend tests use Vitest
- The `web` package (and thus `main`) requires the frontend build to exist (`//go:embed dist`). Running `go test ./...` in `src/backend` fails unless `make frontend` has been run first. To test only Go packages without the frontend: `cd src/backend && go test -race ./conf/... ./files/... ./logfile/... ./rest/... ./timelapse/...`

## CI

GitHub Actions workflow `.github/workflows/go.yml` runs `make pi` on push to `src/backend/**`, `src/frontend/**`, `Makefile`: builds + tests frontend, then Go vet + test + cross-compile for Pi.
