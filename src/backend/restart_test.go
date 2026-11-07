package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestRestartsAndTrafficReleaseEverything starts and stops the whole camera
// (settings database, log, listener, scheduler, watchdog) many times in one
// process, as systemd would after failures, and serves mixed traffic each
// time, including archive downloads the client abandons halfway. Afterwards
// goroutines and file descriptors must be back where they started.
func TestRestartsAndTrafficReleaseEverything(t *testing.T) {
	if testing.Short() {
		t.Skip("restarts the camera 30 times")
	}
	quietLogs(t)
	t.Chdir(t.TempDir())
	os.Mkdir("photos", 0o700)
	big := make([]byte, 1<<20)
	for i := range 30 {
		os.WriteFile(filepath.Join("photos", fmt.Sprintf("2026%04d.jpg", i)), big, 0o600)
	}
	sockDir, err := os.MkdirTemp("", "sd") // short enough for a socket path
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(sockDir)
	notified := notifySocket(t, filepath.Join(sockDir, "notify"))
	t.Setenv("NOTIFY_SOCKET", filepath.Join(sockDir, "notify"))
	t.Setenv("WATCHDOG_USEC", "100000") // ping every 50ms

	client := &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: 4}}
	baseline := measure()
	for i := range 30 {
		ctx, cancel := context.WithCancel(t.Context())
		addr := make(chan string, 1)
		done := make(chan error, 1)
		opts := options{listen: "127.0.0.1:0", storageFolder: "photos", secondsBetweenCaptures: 600, captureRetries: 2, minFreeDiskMB: 1}
		go func() { done <- run(ctx, opts, noCamera{}, func(a string) { addr <- a }) }()
		base := "http://" + <-addr
		receive(t, notified, "READY=1")
		receive(t, notified, "WATCHDOG=1")

		var wg sync.WaitGroup
		for w := range 4 {
			wg.Go(func() { traffic(t, client, base, i*4+w) })
		}
		wg.Wait()
		client.CloseIdleConnections()
		cancel()
		if err := <-done; err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
		drain(notified)
	}

	// Closed connections finish asynchronously; give them a moment.
	var after resources
	for deadline := time.Now().Add(15 * time.Second); ; time.Sleep(50 * time.Millisecond) {
		after = measure()
		if after.goroutines <= baseline.goroutines && after.fds <= baseline.fds || time.Now().After(deadline) {
			break
		}
	}
	if after.goroutines > baseline.goroutines || after.fds > baseline.fds {
		t.Fatalf("after 30 restarts: %+v, before: %+v", after, baseline)
	}
	t.Logf("before: %d goroutines, %d fds; after 30 restarts: %d goroutines, %d fds", baseline.goroutines, baseline.fds, after.goroutines, after.fds)
}

// traffic sends what the frontend sends, plus an abandoned download.
func traffic(t *testing.T, client *http.Client, base string, n int) {
	get := func(path string) {
		res, err := client.Get(base + path)
		if err != nil {
			t.Error(err)
			return
		}
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
		if res.StatusCode >= 500 {
			t.Errorf("%s: %d", path, res.StatusCode)
		}
	}
	for _, p := range []string{"/photos", "/file/last", "/file/20260003.jpg", "/configuration", "/version", "/archive/zip?f=20260001.jpg", "/nope"} {
		get(p)
	}
	body := fmt.Sprintf(`{"SecondsBetweenCaptures":600,"OffsetWithinHour":%d,"Quality":90}`, n%600)
	req, _ := http.NewRequest(http.MethodPost, base+"/configuration", strings.NewReader(body))
	if res, err := client.Do(req); err == nil {
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
	}

	// Abandon a 30 MB archive after the first megabyte.
	res, err := client.Get(base + "/archive/" + []string{"zip", "tar"}[n%2])
	if err != nil {
		t.Error(err)
		return
	}
	io.CopyN(io.Discard, res.Body, 1<<20)
	res.Body.Close()
}

func drain(c <-chan string) {
	for {
		select {
		case <-c:
		default:
			return
		}
	}
}
