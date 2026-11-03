# Go Performance Benchmarks and Long-Run Checks

Benchmarks for the hot paths of the Go backend, and the tests that check it can run unattended for a month or more. Run everything from `src/backend/` (the Go module root). The main package embeds the frontend, so run `make frontend` from the repo root first.

## Running

```bash
go test -run='^$' -bench=. -benchmem ./...             # all benchmarks
go test -run='^$' -bench=. -benchmem -cpu=1 ./rest/    # single core, like a Pi Zero
go test -run='^$' -bench=ListFiles -benchmem ./files/  # one benchmark

# Compare before/after a change
go test -run='^$' -bench=. -benchmem -count=6 ./... > old.txt   # then again > new.txt
go run golang.org/x/perf/cmd/benchstat@latest old.txt new.txt

# Profiles
go test -run='^$' -bench=Photos -cpuprofile=cpu.prof -memprofile=mem.prof ./rest/
go tool pprof -top cpu.prof
```

Benchmarks use `for b.Loop()` (Go 1.24+).

## Benchmarks

44640 photos are a month at one photo per minute; 1440 a month at the default 30 minutes.

| Benchmark | What it measures |
|---|---|
| `files` `BenchmarkListFiles/uncached/N` | `ListFiles`: read the folder, stat every photo, sort |
| `files` `BenchmarkListFiles/index-unchanged/N` | `Index.List` while the folder is unchanged (the gallery polling between captures) |
| `files` `BenchmarkListFiles/index-new-photo/N` | `Index.List` after one new photo: read the names, stat only the new one |
| `files` `BenchmarkArchive/{zip,tar}` | Writing 20 × 3 MB photos to `io.Discard` (no HTTP) |
| `files` `BenchmarkArchive/size/zip/44640` | `ZipSize` for a month of photos (computed per download for `Content-Length`) |
| `rest` `BenchmarkArchive/{zip,tar}` | Downloading 20 × 3 MB photos as one archive over loopback HTTP, through mux and middleware |
| `rest` `BenchmarkFileDownload` | Downloading one 3 MB photo |
| `rest` `BenchmarkPhotos/{full,revalidate,new-photo}/N` | `/photos`: cached response, the browser's `If-None-Match` poll (304), and the first poll after a capture |
| `rest` `BenchmarkMostRecentFile` | `/file/last` (live view) with a month of photos |
| `rest` `BenchmarkSmallRequest` | Per-request overhead of mux + middleware (`/version`) |
| `timelapse` `BenchmarkNextAligned` | Next aligned capture time (~60 ns, 0 allocs) |

### Results

Apple M1 Pro, `-cpu=1`. Expect a Pi Zero to be 15-30× slower, and the network to be the limit for downloads (see below).

| | Before (f7a3935) | After |
|---|---|---|
| `/photos`, 44640 photos | 180 ms, 54 MB, 312k allocs | 11.5 µs, 33 KB, 17 allocs (cached, gzip) |
| `/photos` poll, nothing changed, 44640 | same as above, full 4 MB body | 2.8 µs, 952 B, 304 Not Modified |
| `/photos` first poll after a capture, 44640 | same as above | 45 ms, 17 MB (re-read names, re-encode) |
| `/file/last`, 44640 photos | 170 ms, 31 MB, 268k allocs | 35 µs, 1.9 KB, 34 allocs |
| Listing 44640 photos, unchanged folder | 160 ms, 32 MB | 1.8 µs, 304 B |
| Zip download (loopback) | 1530 MB/s, 720 KB/op, chunked | 1850 MB/s, 51 KB/op, `Content-Length` |
| Tar download (loopback) | 1900 MB/s, 708 KB/op, chunked | 4080 MB/s, 110 KB/op, sendfile, `Content-Length` |
| Single photo download | 2770 MB/s, 42 KB/op (no sendfile) | 4070 MB/s, 11 KB/op (sendfile) |
| Small request overhead | 425 ns, 216 B, 8 allocs | 328 ns, 160 B, 5 allocs |

What changed:

- **Photo index** (`files.Index`): the listing is cached. The folder is re-read only when its modification time changes, and then only new names are stat'ed; photos are immutable once renamed into place. Each new folder time is re-read once more after a 3 s racy window, because file system timestamps are coarse. Elapsed time is measured on the monotonic clock, so clock jumps (NTP, a Pi without RTC) don't matter. The API and disk space reclaiming share one index.
- **`/photos`**: the JSON (and its gzip, from pooled compressors) is built once per listing change and served with an `ETag` and `Cache-Control: no-cache`, so the polling frontend gets 304s.
- **Downloads**: the debug middleware no longer hides `io.ReaderFrom`, so net/http uses `sendfile(2)` for photos. Tar is written header by header with the content handed to sendfile. Zip is written in a single pass with the CRC in a data descriptor (`zip.Writer.CreateRaw`). Both are byte-identical to `archive/tar` and `archive/zip`, and their exact size is known beforehand: `Content-Length` gives browsers a progress bar and the time left, and makes a cut connection show up as an incomplete download. `HEAD` returns the size without reading any photo.
- Reading uses pooled 128 KB buffers (the Linux read-ahead size), not a new 32 KB buffer per photo.

### Download speed on the Pi

On a Pi Zero W the WiFi radio (2.4 GHz, one stream; in practice 20-40 Mbit/s as an access point) is the limit, not the Go code: at 3-5 MB/s, tar and single photos go through sendfile with almost no CPU. Zip still copies through user space for its CRC, which is a few percent of the CPU at that rate. To get the most out of the radio:

- Prefer **tar** for big downloads on slow CPUs; zip costs an extra copy and the CRC.
- Make sure the hotspot runs 802.11n with WMM (NetworkManager does this by default for `mode=ap` when the driver supports it; check with `iw dev wlan0 info` / `iw dev wlan0 station dump` and look for MCS rates), and stay close to the Pi: rates drop quickly with distance and walls.
- A Pi Zero 2 W or USB gadget Ethernet raise the ceiling; then CPU per byte matters more and sendfile helps most.

## Long-Run Checks

These run with `go test ./...` and are skipped with `-short`.

- **`TestUnattendedOperation`** (`soak_test.go`) runs the real scheduler, camera, settings store, photo index, disk space reclaiming and HTTP handlers in a `testing/synctest` bubble, so days pass in seconds. Only the camera hardware and the disk size are simulated. It takes one photo per minute, with failures injected: every 97th attempt errors, every 211th hangs until the capture timeout, both after writing a partial file. The disk is full from the first day (a ring buffer of 600 photos). Settings change weekly, and a gallery polls `/photos` and `/file/last` every minute and downloads an archive every 6 hours. It checks that every capture slot got its photo, that no `.partial-*` file is left, that the disk stays within capacity, and that goroutines, file descriptors and the live heap are the same at the end as after day 1.

  ```bash
  go test -run TestUnattendedOperation -soak-days=35 -timeout=60m -v .   # a month: ~10 min
  ```

  35 days (Mac): 50401 photos for 50401 slots (51168 attempts), 585 photos on disk. 7 goroutines, 6 fds and 1.0 MB live heap after day 1 and at the end. The same test passes on linux/arm/v6 (emulated).

- **`TestRestartsAndTrafficReleaseEverything`** (`restart_test.go`) starts and stops the whole app 30 times in one process, as systemd does after failures. Each run serves mixed traffic: listings, photos, settings updates, and archive downloads that the client abandons after 1 MB. It also checks the systemd `READY=1`/`WATCHDOG=1` notifications. Afterwards goroutines and fds must be back at their starting values.
- **`TestStalledDownloadsAreDropped`** (`rest`): a client that stops reading, like a phone that left the WiFi, is cut off after `stallTimeout` (1 minute in production). Before this change it held a goroutine, a socket and an open photo until TCP gave up, which can take half an hour.

## Unattended Operation and Power Cuts

- **Power cuts**: photos are written to `.partial-*`, fsync'ed, renamed and the folder fsync'ed, so after a power cut a photo is either complete or absent, never empty or truncated. Leftover `.partial-*` files are hidden, so nothing else would ever delete them; they are removed at start-up. The settings database (bbolt) is crash-safe and moved aside if corrupt.
- **Hangs**: `timelapse.service` has `WatchdogSec=2min`. The app pings systemd every minute, but only while the capture loop keeps waking up (at least once a minute, and after every capture including retries). If the loop stalls, systemd restarts the app. `/etc/systemd/system.conf.d/watchdog.conf` enables the Pi's hardware watchdog, which reboots the Pi if the kernel or systemd hang.
- **Crashes and failed starts**: `Restart=always` with `StartLimitIntervalSec=0` restarts forever.
- **Stuck clients**: every response has a 2 minute `WriteTimeout`; downloads extend it while they make progress.
- **Disk full**: the oldest photos are deleted before each capture while free space is below `-minFreeDiskMB`.
