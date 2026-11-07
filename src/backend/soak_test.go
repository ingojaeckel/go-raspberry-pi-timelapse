package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
)

var soakDays = flag.Int("soak-days", 2, "days of unattended operation TestUnattendedOperation simulates (35 for a month)")

// simCamera writes a small photo for every capture. Every 97th attempt fails
// and every 211th hangs until the camera's timeout, both after writing part
// of the file, like a camera that loses its connection.
type simCamera struct {
	attempts, ok atomic.Int64
}

var simPhoto = make([]byte, 2048)

func (c *simCamera) Capture(ctx context.Context, path string, _ timelapse.Options) error {
	n := c.attempts.Add(1)
	if err := os.WriteFile(path, simPhoto, 0o600); err != nil {
		return err
	}
	switch {
	case n%211 == 0:
		<-ctx.Done()
		return ctx.Err()
	case n%97 == 0:
		return errors.New("simulated camera error")
	}
	c.ok.Add(1)
	return nil
}

// sim is the camera running inside a synctest bubble: the real scheduler,
// camera, settings store, photo index, disk space reclaiming and HTTP
// handlers. Only the camera hardware and the disk size are simulated.
type sim struct {
	app    *app
	camera *simCamera
	cancel context.CancelFunc
	wg     sync.WaitGroup
	gallery
}

// capacity is how many photos fit on the simulated disk; 64 KB each, with
// -minFreeDiskMB=1 reclaiming starts 16 photos before it is full.
const (
	capacity  = 600
	photoSize = 64 << 10
)

func startSim(t testing.TB) *sim {
	dir := t.TempDir()
	store, err := conf.OpenStore(filepath.Join(dir, "settings.db"), conf.Default(60))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Update(func(s *conf.Settings) { s.OffsetWithinHour = 0 }); err != nil {
		t.Fatal(err)
	}
	s := &sim{camera: &simCamera{}}
	photos := filepath.Join(dir, "photos")
	os.Mkdir(photos, 0o700)
	s.app = &app{
		opts:   options{storageFolder: photos, captureRetries: 2, minFreeDiskMB: 1},
		level:  new(slog.LevelVar),
		store:  store,
		camera: timelapse.NewCamera(s.camera, captureTimeout),
		photos: files.NewIndex(photos),
		// Health readings of a Pi that had an under-voltage once.
		runCommand: func(context.Context, string, ...string) ([]byte, error) { return []byte("throttled=0x50000\n"), nil },
		started:    time.Now(),
	}
	s.app.freeBytes = func(string) (uint64, error) {
		list, err := s.app.photos.List()
		return uint64(max(capacity-len(list), 0)) * photoSize, err
	}
	s.gallery = gallery{t: t, h: s.app.handler()}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.wg.Go(func() { followDebugSetting(ctx, store, s.app.level) })
	s.wg.Go(func() { s.app.scheduler().Run(ctx) })
	s.wg.Go(func() { s.app.watchHealth(ctx) })
	s.wg.Go(func() { s.gallery.poll(ctx) })
	return s
}

func (s *sim) stop() {
	s.cancel()
	s.wg.Wait()
	s.app.store.Close()
}

// gallery behaves like the frontend left open on a phone: it polls the
// listing and the live view every minute and downloads an archive of the
// latest photos every six hours.
type gallery struct {
	t                     testing.TB
	h                     http.Handler
	etag                  string
	polls                 int
	notModified, archives int
}

func (g *gallery) get(path string, header ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "http://localhost"+path, nil)
	for i := 0; i+1 < len(header); i += 2 {
		req.Header.Set(header[i], header[i+1])
	}
	rec := httptest.NewRecorder()
	g.h.ServeHTTP(rec, req)
	return rec
}

func (g *gallery) poll(ctx context.Context) {
	time.Sleep(30 * time.Second) // between captures
	for ctx.Err() == nil {
		g.polls++
		res := g.get("/photos", "If-None-Match", g.etag, "Accept-Encoding", "gzip")
		switch res.Code {
		case http.StatusNotModified:
			g.notModified++
		case http.StatusOK:
			g.etag = res.Header().Get("ETag")
		default:
			g.t.Errorf("/photos: %d", res.Code)
		}
		if res := g.get("/file/last"); res.Code != http.StatusOK {
			g.t.Errorf("/file/last: %d", res.Code)
		}
		if g.polls%360 == 0 {
			g.archive()
		}
		select {
		case <-ctx.Done():
		case <-time.After(time.Minute):
		}
	}
}

func (g *gallery) archive() {
	g.archives++
	var list struct{ Files []files.File }
	if err := json.NewDecoder(g.get("/file").Body).Decode(&list); err != nil || len(list.Files) == 0 {
		g.t.Errorf("/file: %v, %d photos", err, len(list.Files))
		return
	}
	var q []string
	for _, f := range list.Files[max(len(list.Files)-10, 0):] {
		q = append(q, "f="+f.Name)
	}
	format := []string{"zip", "tar"}[g.archives%2]
	res := g.get("/archive/" + format + "?" + strings.Join(q, "&"))
	if res.Code != http.StatusOK || res.Header().Get("Content-Length") != fmt.Sprint(res.Body.Len()) {
		g.t.Errorf("/archive/%s: %d, Content-Length %s, body %d", format, res.Code, res.Header().Get("Content-Length"), res.Body.Len())
	}
}

// resources is what a leak would make grow, plus what the work cost.
type resources struct {
	goroutines, fds int
	heap            uint64 // live heap after a GC
	allocs, bytes   uint64 // allocated so far
	at              time.Time
}

func measure() resources {
	runtime.GC()
	runtime.GC() // twice, so sync.Pools are empty too
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fds := -1
	if entries, err := os.ReadDir("/dev/fd"); err == nil {
		fds = len(entries)
	}
	return resources{goroutines: runtime.NumGoroutine(), fds: fds, heap: m.HeapAlloc, allocs: m.Mallocs, bytes: m.TotalAlloc, at: realNow()}
}

// realNow is the wall clock, also inside a synctest bubble, from the
// modification time of a new file.
func realNow() time.Time {
	f, err := os.CreateTemp("", "now")
	if err != nil {
		return time.Time{}
	}
	defer os.Remove(f.Name())
	defer f.Close()
	info, _ := f.Stat()
	return info.ModTime()
}

func quietLogs(t testing.TB) {
	old := slog.Default()
	slog.SetDefault(slog.New(slog.DiscardHandler))
	t.Cleanup(func() { slog.SetDefault(old) })
}

// TestUnattendedOperation runs the camera for -soak-days (default 2, use 35
// for a month) of simulated time at one photo per minute, with camera
// failures and hangs, a full disk from the first day on, settings changes and
// a gallery polling all the time. It checks that no capture slot is lost and
// that goroutines, file descriptors and memory stay flat.
func TestUnattendedOperation(t *testing.T) {
	if testing.Short() {
		t.Skip("simulates days of operation")
	}
	quietLogs(t)
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		s := startSim(t)
		day := 24 * time.Hour

		// After a day the disk is full and every cache has its final size.
		time.Sleep(day)
		synctest.Wait()
		warm := measure()
		for d := 1; d < *soakDays; d++ {
			if d%7 == 0 { // weekly settings changes, which re-plan the schedule
				s.app.store.Update(func(c *conf.Settings) { c.Quality = 90 + d%10; c.DebugEnabled = d%14 == 0 })
			}
			time.Sleep(day)
			synctest.Wait()
		}
		time.Sleep(30 * time.Second) // past the last slot
		synctest.Wait()
		end := measure()
		s.stop()

		slots := int64(time.Since(start)/time.Minute) + 1
		if got := s.camera.ok.Load(); got != slots {
			t.Errorf("%d photos for %d capture slots (%d attempts)", got, slots, s.camera.attempts.Load())
		}
		entries, _ := os.ReadDir(s.app.photos.Dir())
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), ".") {
				t.Errorf("left behind %s", e.Name())
			}
		}
		if len(entries) > capacity {
			t.Errorf("%d photos on a disk for %d: disk space was not reclaimed", len(entries), capacity)
		}
		if end.goroutines != warm.goroutines || end.fds != warm.fds {
			t.Errorf("after day 1: %+v, after day %d: %+v", warm, *soakDays, end)
		}
		if end.heap > warm.heap+1<<20 {
			t.Errorf("heap grew from %d to %d bytes", warm.heap, end.heap)
		}
		if s.gallery.notModified == 0 || s.gallery.archives == 0 {
			t.Errorf("gallery: %+v", s.gallery)
		}
		days := uint64(*soakDays - 1)
		t.Logf("%d days: %d photos (%d attempts), %d on disk, %d polls (%d not modified), %d archives",
			*soakDays, s.camera.ok.Load(), s.camera.attempts.Load(), len(entries), s.gallery.polls, s.gallery.notModified, s.gallery.archives)
		t.Logf("after day 1: %d goroutines, %d fds, %d KB heap; at the end: %d goroutines, %d fds, %d KB heap",
			warm.goroutines, warm.fds, warm.heap>>10, end.goroutines, end.fds, end.heap>>10)
		if days > 0 {
			t.Logf("per simulated day (Mac, real time): %v, %d allocs, %d KB allocated",
				end.at.Sub(warm.at)/time.Duration(days), (end.allocs-warm.allocs)/days, (end.bytes-warm.bytes)/days>>10)
		}
	})
}
