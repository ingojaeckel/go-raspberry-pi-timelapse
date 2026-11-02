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
cd src/backend && go test -v -cover ./timelapse/
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
- `main.go` — Entry point, HTTP routes (goji mux), middleware, serves embedded frontend
- `web/` — `//go:embed dist` of the frontend build
- `timelapse/` — Background capture goroutine (`CapturePeriodically`), camera abstraction (rpicam-still on ARM, ffmpeg on dev), scheduling with offset support
- `rest/` — HTTP handlers for photos, config, monitoring, admin, archive streaming
- `conf/` — Settings model + BoltDB persistence (`timelapse-settings.db`), validation in `conf/valid/`
- `files/` — File listing (sorted by age), ZIP/TAR streaming via `io.Pipe()`
- `admin/` — System commands (shutdown, restart, clear photos)

**Key data flow:** HTTP config changes are sent via a Go channel to the background timelapse goroutine, which adjusts its capture schedule without restart.

**Platform abstraction:** `timelapse/camera.go` uses build-time architecture detection — `rpicam-still` on ARM (production Pi), `ffmpeg` webcam capture on x86/macOS (development).

**Frontend:** React 19 + TypeScript + Vite, no UI framework (plain CSS, `fetch`), built into `src/backend/web/dist/` and embedded into the Go binary. Hash-based navigation; API types in `src/frontend/src/api.ts` mirror the Go models. `src/frontend/src/schedule.ts` ports `getSecondsToFirstCapture`, so keep both in sync.

## Build Flags

Version info is injected via ldflags: `-X 'main.gitCommit=...' -X 'main.builtAt=...'`

CLI flags: `-port`, `-logToFile`, `-storageFolder`, `-secondsBetweenCaptures`, `-pprof`, `-version`

## Testing Conventions

- Uses `facebookgo/ensure` for assertions (not `testify`)
- Test files are colocated with source (e.g., `timelapse/timelapse_test.go`)
- `make` / `make pi` run frontend tests, `go vet ./...` and `go test -v -cover ./...` before building
- Frontend tests use Vitest
- The `web` package (and thus `main`) requires the frontend build to exist (`//go:embed dist`). Running `go test ./...` in `src/backend` fails unless `make frontend` has been run first. To test only Go packages without the frontend: `cd src/backend && go test ./admin/... ./conf/... ./files/... ./rest/... ./timelapse/...`

## CI

GitHub Actions workflow `.github/workflows/go.yml` runs `make pi` on push to `src/backend/**`, `src/frontend/**`, `Makefile`: builds + tests frontend, then Go vet + test + cross-compile for Pi.
