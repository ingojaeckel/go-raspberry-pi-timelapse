# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Raspberry Pi Zero W timelapse camera system: a Go HTTP server with embedded React frontend that captures images on a schedule, serves a web UI, and runs as a WiFi access point. Also includes an experimental C++ object detection application in `cpp-object-detection/`.

## Build & Test Commands

```bash
# Go: vet + test + build (development)
./build-for-development.sh

# Go: cross-compile for Pi Zero W (ARM v6)
./build-for-pi.sh

# Run individual Go tests
go test -v -cover ./timelapse/
go test -v -cover ./rest/
go test -run TestSpecificName ./timelapse/

# Frontend (from frontend/ directory)
cd frontend && npm install && npm test && npm run build

# Run the server locally
go run . -port :8080

# Load testing (requires k6)
./scripts/run-load-test.sh
```

## Architecture

**Go packages:**
- `main.go` — Entry point, HTTP routes (goji mux), middleware, embedded frontend via `//go:embed`
- `timelapse/` — Background capture goroutine (`CapturePeriodically`), camera abstraction (rpicam-still on ARM, ffmpeg on dev), scheduling with offset support
- `rest/` — HTTP handlers for photos, config, monitoring, admin, archive streaming
- `conf/` — Settings model + BoltDB persistence (`timelapse-settings.db`), validation in `conf/valid/`
- `files/` — File listing (sorted by age), ZIP/TAR streaming via `io.Pipe()`
- `admin/` — System commands (shutdown, restart, clear photos)

**Key data flow:** HTTP config changes are sent via a Go channel to the background timelapse goroutine, which adjusts its capture schedule without restart.

**Platform abstraction:** `timelapse/camera.go` uses build-time architecture detection — `rpicam-still` on ARM (production Pi), `ffmpeg` webcam capture on x86/macOS (development).

**Frontend:** React 18 + TypeScript + Vite + MUI, built into `frontend/build/` and embedded into the Go binary.

## Build Flags

Version info is injected via ldflags: `-X 'main.gitCommit=...' -X 'main.builtAt=...'`

CLI flags: `-port`, `-logToFile`, `-storageFolder`, `-secondsBetweenCaptures`, `-pprof`, `-version`

## Testing Conventions

- Uses `facebookgo/ensure` for assertions (not `testify`)
- Test files are colocated with source (e.g., `timelapse/timelapse_test.go`)
- Build scripts run `go vet ./...` and `go test -v -cover ./...` before building
- Frontend tests use Vitest
- The main package requires the frontend build to exist (`//go:embed frontend/build`). Running `go test ./...` directly will fail unless `cd frontend && npm run build` has been run first. To test only Go packages without the frontend: `go test ./admin/... ./conf/... ./files/... ./rest/... ./timelapse/...`

## CI

GitHub Actions workflow `.github/workflows/go.yml` runs on push to relevant paths: builds frontend, then runs Go vet + test + cross-compile for Pi.
