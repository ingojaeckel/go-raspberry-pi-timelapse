# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Raspberry Pi Zero W timelapse camera system: a Go HTTP server with embedded React frontend that captures images on a schedule, serves a web UI, and runs as a WiFi access point. 
## Build & Test Commands

```bash
# From the repo root (Makefile is the entry point):
make            # frontend build + tests, Go vet + tests, dev build
make pi         # same, then cross-compile for Pi Zero W (ARM v6) into the image module (src/installer/custompios/modules/timelapsepi/filesystem/root/opt/timelapse/, gitignored)
make pi64       # same tests, then cross-compile for 64-bit Raspberry Pi OS (arm64: Zero 2 W, 3, 4, 5) into src/backend/timelapse-linux-arm64 (gitignored; copy to a Pi by hand)
make image      # make pi, then the SD card image via CustomPiOS in Docker
make test       # frontend tests + Go vet/tests (-short)
make test-long  # all Go tests, incl. the long soak/zip64/restart tests (CI runs it)
make lint-sh    # shellcheck the installer and helper scripts (CI runs it)
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

- `src/backend/` — Go module (`go.mod`, `main.go`, packages below). No binaries are committed; the Pi build goes into the image.
- `src/frontend/` — React app; Vite builds into `src/backend/web/dist` (gitignored).
- `src/installer/custompios/` — CustomPiOS Pi image build (`make image`, see "Pi image" below).
- `src/perf-testing/k6/` — k6 load tests + `run-load-test.sh`.
- `docs/` — hardware, installer, load-testing, benchmarks docs; images in `docs/images/`.

**Path couplings to keep in sync when moving things:**
- `//go:embed` cannot use `..`: `src/backend/web/embed.go` embeds `dist`, and `src/frontend/vite.config.ts` `OUT_DIR` writes there.
- URL prefix `/static/frontend/build/` = `frontendPath` in `app.go` = Vite `BASE` (independent of disk layout).
- `timelapse.service` (`src/installer/custompios/modules/timelapsepi/filesystem/root/etc/systemd/system/`) runs `/opt/timelapse/timelapse`, which `make pi` builds into the module's `filesystem/root/opt/timelapse/` (gitignored) for the image build.

## Architecture

All Go paths below are relative to `src/backend/`.

**Go packages:**
- `main.go` — Entrypoint: flags, signal handling, `run` (= `newApp` + `app.run`)
- `app.go` — `app` owns the resources (log, run marker, settings store, camera, photo index, listener) and releases them in reverse order; at start-up logs how the previous run ended and deletes `.partial-*` files left by a power cut; wires scheduler, API, `http.ServeMux` (frontend, API, pprof) and `http.Server` (timeouts incl. `WriteTimeout`, graceful shutdown on SIGINT/SIGTERM). Before each capture `prepareDisk` deletes the oldest photos, or with `StopWhenDiskFull` skips the slot. `watchHealth` samples temperature and `vcgencmd get_throttled` every 10 min and logs a `daily summary` every 24 h of uptime
- `watchdog.go` — systemd notify (`READY=1`, `WATCHDOG=1` every `WatchdogSec/2` while the scheduler heartbeat is fresh, measured on the monotonic clock so setting the wall clock does not matter), so systemd restarts a hung camera (`WatchdogSec=` in `timelapse.service`)
- `flags.go` — CLI options; `logging.go` — slog setup and debug level; `version.go` — version string
- `web/` — `//go:embed dist` of the frontend build
- `timelapse/` — `Scheduler.Run` (capture loop: aligned or unaligned schedule, retries within the slot, re-plans immediately on settings changes, one catch-up capture rather than a burst after the clock jumps forward; with `SkipNightPhotos` a slot at night moves to sunrise (unaligned) or the first aligned slot after it, logged once per night; in the polar night (no sunrise within `longestNight`, 24 h) nights are not skipped, since searching weeks ahead at every wake-up costs more than the photos saved (`polarNight`/`skipsNights` in the frontend follow the same rule); `BeforeCapture` can skip a slot, `Captured` reports each slot's outcome; `Next` returns the planned capture as a `Plan` with `Night`, for the API), `NextAligned`, `Camera` (one capture at a time, timeout, writes `.partial-*`, fsyncs, renames, fsyncs the folder, never overwrites), `RemovePartials`; `Sensors` (detects the camera with `rpicam-still --list-cameras` once, retries a failure at most once a minute; backends that can list cameras implement `SensorLister`), `ParseCameras`, `Sensor.Resolutions` (its modes and full size, largest first) and `DefaultResolution` (largest full-view mode up to 4 MP)
- `rest/` — `Server` with the HTTP handlers (`config.go`, `photos.go`, `system.go`), `Middleware` (Host allowlist against DNS rebinding, `http.CrossOriginProtection`, no CORS headers, debug request log only when debug is on; forwards `ReadFrom` so sendfile works). `/photos` is cached per index generation (JSON + gzip, ETag/304). Downloads go through `deadlineWriter` (write deadline extended while making progress, 256 KB sendfile chunks). GET never changes state: `DELETE /file?f=`, `DELETE /photos`, `POST /admin/{shutdown,restart,time}`. `/monitoring` returns command output strings plus the camera's `UnixMillis`, `Power` (throttling), `CPU` (governor, MHz) and `Storage` (free/reserved bytes, photo count and size, policy, deleted count, and `SampleBytesPerPixel`/`SampleQuality` of the newest 20 photos for the size estimate in Settings). `GET /schedule` returns the running scheduler's next capture (`NextMillis`, `Night`) with the camera's clock and UTC offset; `GET /camera` returns the detected camera and its `Resolutions`; `POST /configuration` rejects a new photo size the detected camera lacks (an unchanged one is kept, any size is accepted without detection). At start-up the app detects the camera and, on the first run, uses its `DefaultResolution`. `POST /admin/time` sets the clock (`SetSystemClock`, needs `CAP_SYS_TIME`) and then a clock module via `sudo -n hwclock` (fixed arguments, see sudoers)
- `conf/` — `Settings` + `Validate`/`Normalized` (the photo size is stored as `PhotoResolutionWidth`/`Height`; settings of older versions also had a `ResolutionSetting` index, ignored on load); `ReadSettings` reads without locking for writing (self-check); `SkipNightPhotos` (off by default) needs `Latitude`/`Longitude` (not both 0); `Store` keeps bbolt (`timelapse-settings.db`) open, serves settings from memory, notifies subscribers, recovers from a corrupt file (moved aside, also when bbolt panics reading its pages)
- `files/` — `Index` (cached listing, oldest first, skips hidden/partial files; re-reads the folder only when its mtime changed and stats only new names, so it relies on photos never being modified in place), streaming ZIP/TAR with exact `TarSize`/`ZipSize` for `Content-Length` (tar content via sendfile; zip single pass with CRC in a data descriptor; byte-identical to `archive/tar`/`archive/zip`), `RemoveFile(s)` and `Index.Reclaim` (delete oldest photos when free space < `-minFreeDiskMB`, ring buffer) via `os.Root`; `ReadJPEGInfo` (size and libjpeg quality from the header)
- `health/` — `Throttled` (decodes `vcgencmd get_throttled`), `ReadCPUFreq`, `ReadCelsius`, `StartRun` (run marker with the boot ID: stopped cleanly / app died / Pi restarted), `Recorder` (daily summary counts; logs power problems when they appear or end)
- `selfcheck/` — `-selfcheck` (`main.selfCheck`): one test photo with the saved settings (`conf.ReadSettings`, read-only, so it runs next to the app), JPEG decode plus brightness/contrast (black or blank fails; too dark while the sun is down at the saved location is only a warning), `rpicam-still --list-cameras`, other processes holding `/dev/media*`/`/dev/video*` (from `/proc`, with their systemd unit), storage, power, temperature, CPU. PASS/SKIP/WARN/FAIL lines with hints; exit 1 only on FAIL
- `sun/` — offline sunrise/sunset: `Elevation` (NOAA solar position), `Up` (above -0.833°), `NextSunrise`/`NextSunset` (10-minute steps, then bisection; a year's search covers the polar night), `NextSunriseWithin` (bounded search, for the scheduler), `ValidLocation`. `src/frontend/src/sun.ts` is a port; both are tested against `src/frontend/src/sun-cases.json` (generated with the Python astral library, ±2 min), so change them together
- `logfile/` — Size-capped rotating log (`timelapse.log` → `.1` at 10 MB) and `Tail` for `/logs`

**Key data flow:** `POST /configuration` → `conf.Store.Update` (validates, persists, notifies). The scheduler and the debug log level subscribe to the store and read the latest settings with `Get()`; nothing is passed through buffered channels that could drop updates.

**Schedule rules:** `SecondsBetweenCaptures` must divide 3600 and be 60..3600; `OffsetWithinHour` is -1 (unaligned, every N seconds from start-up) or 0..interval-1. Anything else is a 400. Settings saved by older versions are repaired on load by `Settings.Normalized` (same capture times where possible).

**Platform abstraction:** `timelapse.DetectBackend` uses `rpicam-still` when it is on `PATH` (the Pi), otherwise `ffmpeg` webcam capture (Linux v4l2 / macOS avfoundation). It does not look at GOARCH, so Apple Silicon Macs use ffmpeg.

**Frontend:** React 19 + TypeScript + Vite, no UI framework (plain CSS, `fetch`), built into `src/backend/web/dist/` and embedded into the Go binary. Hash-based navigation; API types in `src/frontend/src/api.ts` mirror the Go models. Settings estimates the photo size from the sampled size per pixel and a measured quality→size curve (`qualitySizeFactor` in `format.ts`), and how long the card lasts at the drafted schedule (with nights skipped, scaled by the share of daylight in the coming 24 hours: `photosPerDay` in `schedule.ts`, also used by the Status page's runway). With "Skip night photos" the location is picked from `src/frontend/src/places.ts` (countries at their center from Google's public country table; large countries as a few cities instead) or entered as exact coordinates; only `Latitude`/`Longitude` are saved, and a saved location matching a place exactly shows as that place. The Photos page loads no image until a photo is opened; only Preview takes a (640×480) capture. The top bar's "Next shot" and `HourDial` show the camera's own plan from `GET /schedule` rather than predicting it (`useNextShot` in `nextShot.ts`: converted to the browser's clock so a clock difference doesn't skew the countdown; asked again after settings are saved, every minute, and every 5 s once the planned time has passed); after a skipped night they show "at HH:MM" / "Night, photos resume at …", and the dial's hand uses the camera's UTC offset. `src/frontend/src/schedule.ts` ports `conf.ValidInterval`; both sides are tested against `src/frontend/src/schedule-cases.json` (whose `cases` also test `timelapse.NextAligned`), so change that file together with either implementation.

## Build Flags

Version info is injected via ldflags: `-X 'main.gitCommit=...' -X 'main.builtAt=...'`

If ldflags are not set, the version comes from the VCS info Go embeds (`debug.ReadBuildInfo`).

CLI flags: `-port`, `-logToFile`, `-storageFolder`, `-secondsBetweenCaptures` (first-run default only; once settings are saved the database wins), `-captureRetries` (default 2), `-minFreeDiskMB` (default 256), `-allowedHosts` (extra public host names; IPs, localhost, single-label and .local/.lan/.home.arpa/.internal names are always allowed), `-pprof`, `-version`, `-selfcheck` (with `-selfcheckKeep`, `-selfcheckMinBrightness`, default 10)

Go version: `go.mod` pins `go 1.27.0` / `toolchain go1.27.3`. Dependencies: `go.etcd.io/bbolt`, `golang.org/x/sys` only.

## Testing Conventions

- Plain `testing` (no assertion library). Time-dependent code is tested with `testing/synctest` (fake clock: hours of scheduling run instantly); HTTP handlers through the real mux + middleware with `httptest.NewServer`; file system code against `t.TempDir()`
- Tests must not depend on the working directory or wall clock; use `t.TempDir()`, `t.Chdir()`, `synctest`
- Test files are colocated with source (e.g., `timelapse/scheduler_test.go`)
- `make` / `make pi` / `make pi64` run frontend tests, `go vet ./...` and `go test -short -race -shuffle=on -cover ./...` before building; `make test-long` runs the Go tests without `-short`
- API contract: `rest.TestAPIShapesMatchTheFrontendFixtures` records the JSON shape (field names and types) of each endpoint the frontend reads into `src/frontend/src/test/api-shapes.json` and fails when the handlers change it (regenerate with `UPDATE_API_SHAPES=1 go test ./rest -run APIShapes`); `src/frontend/src/test/contract.test.ts` checks that `mockServer.ts` sends only those fields, with those types. Change a JSON field → regenerate, then fix `api.ts` and `mockServer.ts`
- `service_test.go` parses `timelapse.service`'s `ExecStart` flags with the real flag parser and checks `WatchdogSec`/`TimeoutStopSec` against the app's timings
- Frontend: Vitest with jsdom against the fake server in `src/frontend/src/test/mockServer.ts`; a Playwright smoke test against the real binary runs with `npm run test:e2e`
- Long-run checks (skipped with `-short`, so by `make`; run by `make test-long` and CI): `TestUnattendedOperation` (`soak_test.go`) simulates days of operation in a synctest bubble (`-soak-days=35` for a month; default 2) and asserts every slot is captured and goroutines/fds/heap stay flat; `TestRestartsAndTrafficReleaseEverything` restarts the app 30× under traffic; `files.TestZip64ForManyPhotos` writes 65,536 photos. See `docs/benchmarks.md`
- The `web` package (and thus `main`) requires the frontend build to exist (`//go:embed dist`). Running `go test ./...` in `src/backend` fails unless `make frontend` has been run first. To test only Go packages without the frontend: `cd src/backend && go test -race ./conf/... ./files/... ./logfile/... ./rest/... ./timelapse/...`

## Pi image

`make image` → `src/installer/custompios/build-image.sh`: downloads the Raspberry Pi OS Lite armhf base image pinned (URL + SHA-256) in `src/installer/custompios/config`, checks that Docker can run 32-bit ARM programs (host binfmt_misc qemu-arm; Docker Desktop has it, CI uses `docker/setup-qemu-action`), then runs CustomPiOS from a digest-pinned `ghcr.io/guysoft/custompios` container with `--privileged --init` (loop mounts, chroot via `qemu-arm-static`; Ctrl-C reaches the build). It detaches leftover loop devices and chowns `workspace/` afterwards. Works on macOS with Docker Desktop and on Linux. Output: `src/installer/custompios/workspace/*.zip`. The `timelapsepi` module (`modules/timelapsepi/start_chroot_script`, settings in `modules/timelapsepi/config`) installs `rpicam-apps-lite` and `hwclock`, adds the (commented-out unless `TIMELAPSEPI_RTC` is set) RTC overlay to `config.txt`, applies the power profile (`TIMELAPSEPI_PROFILE`: `remote` turns HDMI off, `bench` keeps it; both turn off audio and Bluetooth and mask unused services), sets the time zone (`TIMELAPSEPI_TIMEZONE`, default `Etc/UTC`), installs each file of `filesystem/root` (root-owned, 644, then `chmod` for executables; not CustomPiOS's `unpack`, which would also reset the modes of existing directories like `/etc/sudoers.d`) and enables `timelapse.service` and `timelapse-cpufreq.service` (powersave governor: lowest CPU clock). `filesystem/root` also holds the hardware watchdog (`etc/systemd/system.conf.d/watchdog.conf`), the RAM-only journal (`etc/systemd/journald.conf.d/timelapse.conf`) and the sudoers rule (shutdown, restart, `hwclock` to `/dev/rtc0`). `timelapse.service` has `AmbientCapabilities=CAP_SYS_TIME` for the web UI's clock button. See `docs/installer.md`.

## CI

GitHub Actions workflow `.github/workflows/go.yml` runs `make pi` on push to `src/backend/**`, `src/frontend/**`, `src/installer/**`, `Makefile`: builds + tests frontend, then Go vet + test (-short) + cross-compile for Pi, then `make test-long` and `make lint-sh` (shellcheck). `.github/workflows/image.yml` runs `make image` on `ubuntu-24.04-arm` on demand and for tags; a separate job with `contents: write` attaches the zipped image to the tag's GitHub release.
