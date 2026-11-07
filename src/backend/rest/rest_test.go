package rest

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
)

type fakeCamera struct {
	err  error
	opts timelapse.Options
}

func (c *fakeCamera) Capture(_ context.Context, dir string, o timelapse.Options) (string, error) {
	c.opts = o
	if c.err != nil {
		return "", c.err
	}
	p := filepath.Join(dir, "preview.jpg")
	return p, os.WriteFile(p, []byte("\xff\xd8\xff preview"), 0o600)
}

type commands struct {
	mu      sync.Mutex
	ran     [][]string
	hwclock func() ([]byte, error)
}

func (c *commands) run(_ context.Context, name string, args ...string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ran = append(c.ran, append([]string{name}, args...))
	switch {
	case name == "/bin/date":
		return []byte("Mon Nov  2 10:00:00 UTC 2026\n"), nil
	case strings.HasSuffix(name, "vcgencmd") && slices.Equal(args, []string{"get_throttled"}):
		return []byte("throttled=0x50005\n"), nil
	case name == "/usr/bin/sudo" && len(args) > 1 && args[1] == "/usr/sbin/hwclock":
		return c.hwclock()
	}
	return nil, errors.New(name + " not available")
}

type fakeClock struct {
	mu   sync.Mutex
	sets []time.Time
	err  error
}

func (c *fakeClock) set(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err != nil {
		return c.err
	}
	c.sets = append(c.sets, t)
	return nil
}

type env struct {
	clock    *fakeClock
	srv      *Server
	url      string
	folder   string
	store    *conf.Store
	camera   *fakeCamera
	commands *commands
	handler  http.Handler // mux + middleware, as served at url
	// sensor is the detected camera; nil means none detected.
	sensor *timelapse.Sensor
	// plan is the scheduler's next capture; nil means none planned yet.
	plan *timelapse.Plan
}

func newEnv(t testing.TB) *env {
	t.Helper()
	dir := t.TempDir()
	store, err := conf.OpenStore(filepath.Join(dir, "settings.db"), conf.Default(1800))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	e := &env{store: store, folder: filepath.Join(dir, "photos"), camera: &fakeCamera{}, clock: &fakeClock{},
		commands: &commands{hwclock: func() ([]byte, error) { return nil, nil }}}
	os.Mkdir(e.folder, 0o700)
	os.Mkdir(filepath.Join(dir, "tmp"), 0o700)
	e.srv = &Server{
		Settings:      store,
		Camera:        e.camera,
		Photos:        files.NewIndex(e.folder),
		TempFolder:    filepath.Join(dir, "tmp"),
		LogFile:       filepath.Join(dir, "app.log"),
		ThermalFile:   filepath.Join(dir, "temp"),
		CPUFreqDir:    filepath.Join(dir, "cpufreq"),
		Version:       "test-v1",
		Run:           e.commands.run,
		FreeBytes:     func(string) (uint64, error) { return 3 << 30, nil },
		ReserveBytes:  256 << 20,
		DeletedOldest: func() int { return 7 },
		SetClock:      e.clock.set,
		RTCDevice:     filepath.Join(dir, "rtc0"), // absent unless a test creates it
	}
	e.srv.Sensor = func(context.Context) (timelapse.Sensor, error) {
		if e.sensor == nil {
			return timelapse.Sensor{}, timelapse.ErrNoCamera
		}
		return *e.sensor, nil
	}
	e.srv.NextCapture = func() (timelapse.Plan, bool) {
		if e.plan == nil {
			return timelapse.Plan{}, false
		}
		return *e.plan, true
	}
	mux := http.NewServeMux()
	e.srv.Register(mux)
	e.handler = Middleware(mux, "camera.example.org")
	ts := httptest.NewServer(e.handler)
	t.Cleanup(ts.Close)
	e.url = ts.URL
	return e
}

func (e *env) photo(t testing.TB, name, content string, age time.Duration) {
	t.Helper()
	p := filepath.Join(e.folder, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	mt := time.Now().Add(-age)
	os.Chtimes(p, mt, mt)
}

func (e *env) exists(name string) bool {
	_, err := os.Stat(filepath.Join(e.folder, name))
	return err == nil
}

func do(t *testing.T, method, url string, body string, header ...string) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i+1 < len(header); i += 2 {
		if header[i] == "Host" {
			req.Host = header[i+1]
			continue
		}
		req.Header.Set(header[i], header[i+1])
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return res, string(b)
}

func wantStatus(t *testing.T, res *http.Response, body string, status int) {
	t.Helper()
	if res.StatusCode != status {
		t.Fatalf("%s %s: status %d, want %d (body %q)", res.Request.Method, res.Request.URL.Path, res.StatusCode, status, body)
	}
}

func TestVersion(t *testing.T) {
	e := newEnv(t)
	res, body := do(t, "GET", e.url+"/version", "")
	wantStatus(t, res, body, 200)
	if !strings.Contains(body, "test-v1") {
		t.Fatalf("body %q", body)
	}
}

func TestGetConfiguration(t *testing.T) {
	e := newEnv(t)
	res, body := do(t, "GET", e.url+"/configuration", "")
	wantStatus(t, res, body, 200)
	var got conf.Settings
	if err := json.Unmarshal([]byte(body), &got); err != nil || got != e.store.Get() {
		t.Fatalf("got %s (%v)", body, err)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content type %q", ct)
	}
}

func TestUpdateConfiguration(t *testing.T) {
	e := newEnv(t)
	body := `{"SecondsBetweenCaptures":600,"OffsetWithinHour":300,"Quality":85,"PhotoResolutionWidth":1640,"PhotoResolutionHeight":1232,"RotateBy":180,"DebugEnabled":true,"StopWhenDiskFull":true,"PreviewResolutionWidth":1}`
	res, out := do(t, "POST", e.url+"/configuration", body)
	wantStatus(t, res, out, 200)
	got := e.store.Get()
	if got.SecondsBetweenCaptures != 600 || got.OffsetWithinHour != 300 || got.Quality != 85 || got.RotateBy != 180 || !got.DebugEnabled || !got.StopWhenDiskFull {
		t.Fatalf("stored %s", got)
	}
	if got.PhotoResolutionWidth != 1640 || got.PhotoResolutionHeight != 1232 {
		t.Fatalf("resolution not stored: %s", got)
	}
	if got.PreviewResolutionWidth != 640 {
		t.Fatal("preview resolution is not editable through the API")
	}
	var echoed conf.Settings
	json.Unmarshal([]byte(out), &echoed)
	if echoed != got {
		t.Fatalf("response %s, stored %s", out, got)
	}
}

func TestUpdateConfigurationRejectsInvalidSettings(t *testing.T) {
	cases := map[string]string{
		"not json":                    `not json`,
		"interval too short":          `{"SecondsBetweenCaptures":1,"OffsetWithinHour":0,"Quality":90}`,
		"interval not dividing hour":  `{"SecondsBetweenCaptures":1000,"OffsetWithinHour":0,"Quality":90}`,
		"interval longer than hour":   `{"SecondsBetweenCaptures":7200,"OffsetWithinHour":0,"Quality":90}`,
		"offset not below interval":   `{"SecondsBetweenCaptures":600,"OffsetWithinHour":900,"Quality":90}`,
		"quality out of range":        `{"SecondsBetweenCaptures":600,"OffsetWithinHour":0,"Quality":0}`,
		"photo too small":             `{"SecondsBetweenCaptures":600,"OffsetWithinHour":0,"Quality":90,"PhotoResolutionWidth":32,"PhotoResolutionHeight":24}`,
		"unsupported rotation":        `{"SecondsBetweenCaptures":600,"OffsetWithinHour":0,"Quality":90,"RotateBy":90}`,
		"body larger than any config": `{"SecondsBetweenCaptures":600,"x":"` + strings.Repeat("a", 1<<20) + `"}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			e := newEnv(t)
			before := e.store.Get()
			res, out := do(t, "POST", e.url+"/configuration", body)
			wantStatus(t, res, out, 400)
			if e.store.Get() != before {
				t.Fatal("settings changed")
			}
		})
	}
}

func TestPhotosListsFinishedPhotosOnly(t *testing.T) {
	e := newEnv(t)
	e.photo(t, "a.jpg", strings.Repeat("x", 2048), time.Hour)
	e.photo(t, ".partial-b.jpg", "half", 0)
	res, body := do(t, "GET", e.url+"/photos", "")
	wantStatus(t, res, body, 200)
	var got GetPhotosResponse
	json.Unmarshal([]byte(body), &got)
	if len(got.Photos) != 1 || got.Photos[0].Name != "a.jpg" || got.Photos[0].Size != "2 KB" {
		t.Fatalf("got %s", body)
	}
}

func TestPhotosEmptyFolderIsAnEmptyList(t *testing.T) {
	e := newEnv(t)
	_, body := do(t, "GET", e.url+"/photos", "")
	if strings.TrimSpace(body) != `{"Photos":[]}` {
		t.Fatalf("got %s", body)
	}
}

func TestListFiles(t *testing.T) {
	e := newEnv(t)
	e.photo(t, "a.jpg", "a", time.Hour)
	_, body := do(t, "GET", e.url+"/file", "")
	var got ListFilesResponse
	json.Unmarshal([]byte(body), &got)
	if len(got.Files) != 1 || got.Files[0].Name != "a.jpg" {
		t.Fatalf("got %s", body)
	}
}

func TestGetFile(t *testing.T) {
	e := newEnv(t)
	e.photo(t, "a.jpg", "\xff\xd8\xff photo", 0)
	res, body := do(t, "GET", e.url+"/file/a.jpg", "")
	wantStatus(t, res, body, 200)
	if body != "\xff\xd8\xff photo" {
		t.Fatalf("body %q", body)
	}
	if cd := res.Header.Get("Content-Disposition"); cd != "attachment; filename=a.jpg" {
		t.Fatalf("Content-Disposition %q", cd)
	}
	if ct := res.Header.Get("Content-Type"); ct != "image/jpeg" {
		t.Fatalf("Content-Type %q", ct)
	}
}

func TestGetFileSupportsRangeRequests(t *testing.T) {
	e := newEnv(t)
	e.photo(t, "a.jpg", "0123456789", 0)
	res, body := do(t, "GET", e.url+"/file/a.jpg", "", "Range", "bytes=2-4")
	wantStatus(t, res, body, http.StatusPartialContent)
	if body != "234" {
		t.Fatalf("body %q", body)
	}
}

func TestGetFileCannotEscapeTheFolder(t *testing.T) {
	e := newEnv(t)
	secret := filepath.Join(filepath.Dir(e.folder), "secret.txt")
	os.WriteFile(secret, []byte("secret"), 0o600)
	for _, p := range []string{"/file/..%2Fsecret.txt", "/file/%2E%2E%2Fsecret.txt", "/file/..%5Csecret.txt"} {
		res, body := do(t, "GET", e.url+p, "")
		if res.StatusCode == 200 || strings.Contains(body, "secret") {
			t.Errorf("%s: status %d body %q", p, res.StatusCode, body)
		}
	}
}

func TestGetFileMissingOrDirectory(t *testing.T) {
	e := newEnv(t)
	os.Mkdir(filepath.Join(e.folder, "dir"), 0o700)
	for _, name := range []string{"nope.jpg", "dir"} {
		res, body := do(t, "GET", e.url+"/file/"+name, "")
		wantStatus(t, res, body, 404)
	}
}

func TestMostRecentFile(t *testing.T) {
	e := newEnv(t)
	res, body := do(t, "GET", e.url+"/file/last", "")
	wantStatus(t, res, body, 404)

	e.photo(t, "old.jpg", "old", time.Hour)
	e.photo(t, "new.jpg", "new", time.Minute)
	e.photo(t, ".partial-newest.jpg", "half", 0)
	res, body = do(t, "GET", e.url+"/file/last", "")
	wantStatus(t, res, body, 200)
	if body != "new" {
		t.Fatalf("served %q, want the newest finished photo", body)
	}
}

func TestDeleteFiles(t *testing.T) {
	e := newEnv(t)
	e.photo(t, "a.jpg", "a", 0)
	e.photo(t, "b.jpg", "b", 0)
	e.photo(t, "c.jpg", "c", 0)
	res, body := do(t, "DELETE", e.url+"/file?f=a.jpg&f=b.jpg", "")
	wantStatus(t, res, body, 200)
	if strings.TrimSpace(body) != "2" || e.exists("a.jpg") || e.exists("b.jpg") || !e.exists("c.jpg") {
		t.Fatalf("body %s", body)
	}
}

func TestDeleteFilesRequiresNames(t *testing.T) {
	e := newEnv(t)
	res, body := do(t, "DELETE", e.url+"/file", "")
	wantStatus(t, res, body, 400)
}

func TestDeleteFilesCannotEscapeTheFolder(t *testing.T) {
	e := newEnv(t)
	victim := filepath.Join(filepath.Dir(e.folder), "victim.txt")
	os.WriteFile(victim, nil, 0o600)
	res, body := do(t, "DELETE", e.url+"/file?f=../victim.txt", "")
	if _, err := os.Stat(victim); err != nil {
		t.Fatal("deleted a file outside the photo folder")
	}
	if res.StatusCode == 200 && strings.TrimSpace(body) != "0" {
		t.Fatalf("status %d body %s", res.StatusCode, body)
	}
}

func TestGetRequestsNeverChangeState(t *testing.T) {
	e := newEnv(t)
	e.photo(t, "a.jpg", "a", 0)
	for _, p := range []string{"/file/delete?f=a.jpg", "/admin/shutdown", "/admin/restart", "/admin/clear"} {
		res, body := do(t, "GET", e.url+p, "")
		if res.StatusCode == 200 || res.StatusCode == 204 {
			t.Errorf("GET %s: status %d (%q)", p, res.StatusCode, body)
		}
	}
	if !e.exists("a.jpg") || len(e.commands.ran) != 0 {
		t.Fatalf("state changed: photo exists=%v commands=%v", e.exists("a.jpg"), e.commands.ran)
	}
}

func TestDeleteAllPhotos(t *testing.T) {
	e := newEnv(t)
	e.photo(t, "a.jpg", "a", 0)
	e.photo(t, "b.jpg", "b", 0)
	res, body := do(t, "DELETE", e.url+"/photos", "")
	wantStatus(t, res, body, 200)
	if strings.TrimSpace(body) != "2" || e.exists("a.jpg") || e.exists("b.jpg") {
		t.Fatalf("body %s", body)
	}
}

func TestAdminCommands(t *testing.T) {
	for cmd, want := range map[string][]string{
		"shutdown": {"/usr/bin/sudo", "/sbin/shutdown", "-h", "now"},
		"restart":  {"/usr/bin/sudo", "/sbin/shutdown", "-r", "now"},
	} {
		t.Run(cmd, func(t *testing.T) {
			e := newEnv(t)
			res, body := do(t, "POST", e.url+"/admin/"+cmd, "")
			wantStatus(t, res, body, 500) // the fake runner fails
			if len(e.commands.ran) != 1 || !slices.Equal(e.commands.ran[0], want) {
				t.Fatalf("ran %v", e.commands.ran)
			}
		})
	}
}

func TestAdminCommandSuccess(t *testing.T) {
	e := newEnv(t)
	e.srv.Run = func(context.Context, string, ...string) ([]byte, error) { return nil, nil }
	res, body := do(t, "POST", e.url+"/admin/restart", "")
	wantStatus(t, res, body, 204)
}

func TestUnknownAdminCommand(t *testing.T) {
	e := newEnv(t)
	res, body := do(t, "POST", e.url+"/admin/format-disk", "")
	wantStatus(t, res, body, 404)
	if len(e.commands.ran) != 0 {
		t.Fatal("ran a command")
	}
}

func TestArchiveZip(t *testing.T) {
	e := newEnv(t)
	e.photo(t, "a.jpg", "a", 2*time.Minute)
	e.photo(t, "b.jpg", "b", time.Minute)
	res, body := do(t, "GET", e.url+"/archive/zip?f=b.jpg", "")
	wantStatus(t, res, body, 200)
	if res.Header.Get("Content-Disposition") != "attachment; filename=archive.zip" {
		t.Fatal(res.Header)
	}
	zr, err := zip.NewReader(strings.NewReader(body), int64(len(body)))
	if err != nil || len(zr.File) != 1 || zr.File[0].Name != "b.jpg" {
		t.Fatalf("zip %v %v", err, zr)
	}
}

func TestArchiveTarContainsAllPhotosByDefault(t *testing.T) {
	e := newEnv(t)
	e.photo(t, "a.jpg", "a", 2*time.Minute)
	e.photo(t, "b.jpg", "b", time.Minute)
	res, body := do(t, "GET", e.url+"/archive/tar", "")
	wantStatus(t, res, body, 200)
	tr := tar.NewReader(strings.NewReader(body))
	var names []string
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, h.Name)
	}
	if !slices.Equal(names, []string{"a.jpg", "b.jpg"}) {
		t.Fatalf("names %v", names)
	}
}

func TestArchiveFailureAbortsTheDownload(t *testing.T) {
	// A truncated archive must not look like a successful download.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		streamArchive(w, func(w io.Writer) error {
			w.Write(bytes.Repeat([]byte("x"), 64<<10))
			return errors.New("disk read error")
		})
	}))
	defer ts.Close()
	res, err := http.Get(ts.URL)
	if err != nil {
		return // aborted before headers: fine
	}
	defer res.Body.Close()
	if _, err := io.ReadAll(res.Body); err == nil {
		t.Fatal("client received a complete response for a failed archive")
	}
}

func TestPreviewCapture(t *testing.T) {
	e := newEnv(t)
	e.store.Update(func(s *conf.Settings) { s.RotateBy = 180 })
	res, body := do(t, "GET", e.url+"/capture", "")
	wantStatus(t, res, body, 200)
	if !strings.Contains(body, "preview") {
		t.Fatalf("body %q", body)
	}
	if want := (timelapse.Options{Width: 640, Height: 480, Quality: 100, Rotate: true}); e.camera.opts != want {
		t.Fatalf("opts %+v, want saved settings %+v", e.camera.opts, want)
	}
	if entries, _ := os.ReadDir(e.srv.TempFolder); len(entries) != 0 {
		t.Fatalf("preview file left behind: %v", entries)
	}
}

func TestPreviewCaptureFailure(t *testing.T) {
	e := newEnv(t)
	e.camera.err = errors.New("camera busy")
	res, body := do(t, "GET", e.url+"/capture", "")
	wantStatus(t, res, body, 500)
	if !strings.Contains(body, "camera busy") {
		t.Fatalf("body %q", body)
	}
}

func TestMonitoring(t *testing.T) {
	e := newEnv(t)
	os.WriteFile(e.srv.ThermalFile, []byte("48312\n"), 0o600)
	res, body := do(t, "GET", e.url+"/monitoring", "")
	wantStatus(t, res, body, 200)
	var got MonitoringResponse
	json.Unmarshal([]byte(body), &got)
	if got.CpuTemperature != "48312\n" || !strings.HasPrefix(got.Time, "Mon Nov") || !strings.Contains(got.Uptime, "not available") {
		t.Fatalf("got %+v", got)
	}
	if d := time.Since(time.UnixMilli(got.UnixMillis)); d < 0 || d > time.Minute {
		t.Fatalf("UnixMillis %d is not the current time", got.UnixMillis)
	}
	if want := []string{"under-voltage", "throttled"}; !slices.Equal(got.Power.Now, want) || !slices.Equal(got.Power.SinceBoot, want) || got.Power.Error != "" {
		t.Fatalf("power %+v", got.Power)
	}
	if got.CPU.Error == "" {
		t.Fatalf("cpu %+v: want an error without cpufreq", got.CPU)
	}
}

func TestMonitoringReportsCPUFrequency(t *testing.T) {
	e := newEnv(t)
	os.Mkdir(e.srv.CPUFreqDir, 0o700)
	for name, v := range map[string]string{"scaling_governor": "powersave", "scaling_cur_freq": "600000", "scaling_min_freq": "600000", "scaling_max_freq": "1000000"} {
		os.WriteFile(filepath.Join(e.srv.CPUFreqDir, name), []byte(v+"\n"), 0o600)
	}
	_, body := do(t, "GET", e.url+"/monitoring", "")
	var got MonitoringResponse
	json.Unmarshal([]byte(body), &got)
	if got.CPU.Governor != "powersave" || got.CPU.CurrentMHz != 600 || got.CPU.MaxMHz != 1000 || got.CPU.Error != "" {
		t.Fatalf("cpu %+v", got.CPU)
	}
}

func TestMonitoringReportsStorage(t *testing.T) {
	e := newEnv(t)
	e.photo(t, "a.jpg", "12345", time.Hour)
	e.photo(t, "b.jpg", "123", 0)
	e.store.Update(func(s *conf.Settings) { s.StopWhenDiskFull = true })
	_, body := do(t, "GET", e.url+"/monitoring", "")
	var got MonitoringResponse
	json.Unmarshal([]byte(body), &got)
	want := StorageStatus{FreeBytes: 3 << 30, ReserveBytes: 256 << 20, Photos: 2, PhotoBytes: 8, StopWhenDiskFull: true, DeletedOldest: 7}
	if got.Storage != want {
		t.Fatalf("storage %+v, want %+v", got.Storage, want)
	}

	e.srv.FreeBytes = func(string) (uint64, error) { return 0, errors.New("statfs failed") }
	_, body = do(t, "GET", e.url+"/monitoring", "")
	json.Unmarshal([]byte(body), &got)
	if got.Storage.Error != "statfs failed" || got.Storage.Photos != 2 {
		t.Fatalf("storage %+v", got.Storage)
	}
	os.RemoveAll(e.folder)
	_, body = do(t, "GET", e.url+"/monitoring", "")
	json.Unmarshal([]byte(body), &got)
	if got.Storage.Error == "" {
		t.Fatalf("storage %+v: want an error without the photo folder", got.Storage)
	}
}

func TestMonitoringWithoutVcgencmd(t *testing.T) {
	e := newEnv(t)
	e.srv.Run = func(context.Context, string, ...string) ([]byte, error) { return nil, errors.New("no vcgencmd") }
	_, body := do(t, "GET", e.url+"/monitoring", "")
	var got MonitoringResponse
	json.Unmarshal([]byte(body), &got)
	if got.Power.Error != "no vcgencmd" || got.Power.Now == nil || got.Power.SinceBoot == nil {
		t.Fatalf("power %+v", got.Power)
	}
}

func TestSetTime(t *testing.T) {
	e := newEnv(t)
	want := time.Date(2026, 11, 7, 9, 30, 0, 0, time.UTC)
	res, body := do(t, "POST", e.url+"/admin/time", fmt.Sprintf(`{"UnixMillis":%d}`, want.UnixMilli()))
	wantStatus(t, res, body, 200)
	if len(e.clock.sets) != 1 || !e.clock.sets[0].Equal(want) {
		t.Fatalf("clock set to %v, want %v", e.clock.sets, want)
	}
	var got SetTimeResponse
	json.Unmarshal([]byte(body), &got)
	if got.RTC != "" || got.UnixMillis == 0 {
		t.Fatalf("without a clock module: %+v", got)
	}
	if slices.ContainsFunc(e.commands.ran, func(c []string) bool { return slices.Contains(c, "/usr/sbin/hwclock") }) {
		t.Fatal("ran hwclock without a clock module")
	}
}

func TestSetTimeUpdatesTheClockModule(t *testing.T) {
	e := newEnv(t)
	os.WriteFile(e.srv.RTCDevice, nil, 0o600)
	body := fmt.Sprintf(`{"UnixMillis":%d}`, time.Date(2026, 11, 7, 9, 30, 0, 0, time.UTC).UnixMilli())
	res, out := do(t, "POST", e.url+"/admin/time", body)
	wantStatus(t, res, out, 200)
	var got SetTimeResponse
	json.Unmarshal([]byte(out), &got)
	if got.RTC != "updated" {
		t.Fatalf("RTC %q", got.RTC)
	}
	// Exactly the command sudoers allows.
	wantCmd := []string{"/usr/bin/sudo", "-n", "/usr/sbin/hwclock", "--rtc=" + e.srv.RTCDevice, "--systohc", "--utc"}
	if !slices.ContainsFunc(e.commands.ran, func(c []string) bool { return slices.Equal(c, wantCmd) }) {
		t.Fatalf("ran %v, want %v", e.commands.ran, wantCmd)
	}

	e.commands.hwclock = func() ([]byte, error) { return []byte("hwclock: ioctl failed"), errors.New("exit status 1") }
	res, out = do(t, "POST", e.url+"/admin/time", body)
	wantStatus(t, res, out, 200) // the system clock was set
	json.Unmarshal([]byte(out), &got)
	if !strings.Contains(got.RTC, "not updated") || !strings.Contains(got.RTC, "ioctl failed") {
		t.Fatalf("RTC %q", got.RTC)
	}
}

func TestSetTimeRejectsBadRequests(t *testing.T) {
	e := newEnv(t)
	for _, body := range []string{"", "{", `{"UnixMillis":0}`, `{"UnixMillis":1}`, fmt.Sprintf(`{"UnixMillis":%d}`, time.Date(2200, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli())} {
		res, out := do(t, "POST", e.url+"/admin/time", body)
		wantStatus(t, res, out, 400)
	}
	if len(e.clock.sets) != 0 {
		t.Fatalf("clock set to %v", e.clock.sets)
	}
	// Like every POST, not from other websites.
	res, out := do(t, "POST", e.url+"/admin/time", `{"UnixMillis":1794000000000}`, "Origin", "https://evil.example.com", "Sec-Fetch-Site", "cross-site")
	wantStatus(t, res, out, 403)
}

func TestSetTimeFails(t *testing.T) {
	e := newEnv(t)
	e.clock.err = errors.New("operation not permitted")
	res, out := do(t, "POST", e.url+"/admin/time", `{"UnixMillis":1794000000000}`)
	wantStatus(t, res, out, 500)
	if !strings.Contains(out, "setting the clock: operation not permitted") {
		t.Fatalf("body %s", out)
	}
}

func TestLogsReturnsTheEndOfTheLog(t *testing.T) {
	e := newEnv(t)
	os.WriteFile(e.srv.LogFile, []byte(strings.Repeat("old line\n", 100000)+"latest\n"), 0o600)
	res, body := do(t, "GET", e.url+"/logs", "")
	wantStatus(t, res, body, 200)
	var got LogResponse
	json.Unmarshal([]byte(body), &got)
	if !strings.HasSuffix(got.Logs, "latest\n") || len(got.Logs) > maxLogBytes {
		t.Fatalf("got %d bytes", len(got.Logs))
	}
}

func TestLogsWithoutLogFile(t *testing.T) {
	e := newEnv(t)
	e.srv.LogFile = ""
	res, body := do(t, "GET", e.url+"/logs", "")
	wantStatus(t, res, body, 500)
}

func TestOtherWebsitesCannotReadTheAPI(t *testing.T) {
	e := newEnv(t)
	res, body := do(t, "GET", e.url+"/photos", "", "Origin", "https://evil.example.com")
	wantStatus(t, res, body, 200)
	if v := res.Header.Get("Access-Control-Allow-Origin"); v != "" {
		t.Fatalf("Access-Control-Allow-Origin %q lets other websites read photos and settings", v)
	}
}

func TestDNSRebindingIsRejected(t *testing.T) {
	// A page on a public domain that re-resolves to the camera's IP is
	// same-origin for the browser, so only the Host header gives it away.
	e := newEnv(t)
	e.photo(t, "a.jpg", "a", 0)
	for _, host := range []string{"evil.example.com", "evil.example.com:8080", "192.168.4.1.nip.io"} {
		res, body := do(t, "DELETE", e.url+"/photos", "", "Host", host, "Sec-Fetch-Site", "same-origin")
		wantStatus(t, res, body, 403)
		res, body = do(t, "GET", e.url+"/photos", "", "Host", host)
		wantStatus(t, res, body, 403)
	}
	if !e.exists("a.jpg") {
		t.Fatal("photos were deleted through a rebound host name")
	}
}

func TestLocalHostNamesAreAccepted(t *testing.T) {
	e := newEnv(t)
	for _, host := range []string{
		"192.168.4.1", "192.168.4.1:8080", "[fe80::1]:8080", "localhost:8080",
		"raspberrypi", "raspberrypi:8080", "timelapse.local", "cam.lan:8080", "cam.home.arpa", "cam.internal",
		"Camera.Example.Org:8080", // allowed explicitly (-allowedHosts)
	} {
		res, body := do(t, "GET", e.url+"/version", "", "Host", host)
		if res.StatusCode != 200 {
			t.Errorf("Host %q: status %d %q", host, res.StatusCode, body)
		}
	}
}

func TestCrossSiteRequestsCannotChangeState(t *testing.T) {
	e := newEnv(t)
	e.photo(t, "a.jpg", "a", 0)
	res, body := do(t, "DELETE", e.url+"/photos", "", "Sec-Fetch-Site", "cross-site")
	wantStatus(t, res, body, 403)
	if !e.exists("a.jpg") {
		t.Fatal("a cross-site page deleted photos")
	}
	res, body = do(t, "DELETE", e.url+"/photos", "", "Sec-Fetch-Site", "same-origin")
	wantStatus(t, res, body, 200)
}

func TestHumanReadableSize(t *testing.T) {
	for n, want := range map[int64]string{0: "0 bytes", 1023: "1023 bytes", 1024: "1 KB", 10 * 1024: "10 KB", 1 << 20: "1 MB"} {
		if got := humanReadableSize(n); got != want {
			t.Errorf("%d: got %s want %s", n, got, want)
		}
	}
}

// imx290 is a 1920x1080 sensor with a 1280x720 mode.
var imx290 = timelapse.Sensor{Model: "imx290", Width: 1920, Height: 1080, Modes: []timelapse.Size{{Width: 1280, Height: 720}, {Width: 1920, Height: 1080}}}

func TestCamera(t *testing.T) {
	e := newEnv(t)
	e.sensor = &imx290
	res, out := do(t, "GET", e.url+"/camera", "")
	wantStatus(t, res, out, 200)
	var got CameraResponse
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.Model != "imx290" || got.Width != 1920 || !slices.Equal(got.Resolutions, []timelapse.Size{{Width: 1920, Height: 1080}, {Width: 1280, Height: 720}}) || got.Error != "" {
		t.Fatalf("got %s", out)
	}
}

func TestCameraNotDetected(t *testing.T) {
	e := newEnv(t)
	res, out := do(t, "GET", e.url+"/camera", "")
	wantStatus(t, res, out, 200)
	var got CameraResponse
	json.Unmarshal([]byte(out), &got)
	if got.Error != "no camera detected" || len(got.Resolutions) != 0 {
		t.Fatalf("got %s", out)
	}
}

func TestUpdateConfigurationOffersOnlyTheCamerasResolutions(t *testing.T) {
	e := newEnv(t)
	e.sensor = &imx290
	settings := `{"SecondsBetweenCaptures":600,"OffsetWithinHour":0,"Quality":90,"PhotoResolutionWidth":%d,"PhotoResolutionHeight":%d}`

	res, out := do(t, "POST", e.url+"/configuration", fmt.Sprintf(settings, 1640, 1232))
	wantStatus(t, res, out, 400)
	if !strings.Contains(out, "1920x1080, 1280x720") {
		t.Errorf("the error does not list the supported sizes: %s", out)
	}

	res, out = do(t, "POST", e.url+"/configuration", fmt.Sprintf(settings, 1280, 720))
	wantStatus(t, res, out, 200)
	if got := e.store.Get(); got.PhotoResolutionWidth != 1280 || got.PhotoResolutionHeight != 720 {
		t.Fatalf("stored %s", got)
	}
}

func TestUpdateConfigurationKeepsAnUnsupportedSizeThatWasNotChanged(t *testing.T) {
	e := newEnv(t)
	e.sensor = &imx290 // the default 2186x1642 is not one of its sizes
	before := e.store.Get()
	body := fmt.Sprintf(`{"SecondsBetweenCaptures":600,"OffsetWithinHour":0,"Quality":70,"PhotoResolutionWidth":%d,"PhotoResolutionHeight":%d}`,
		before.PhotoResolutionWidth, before.PhotoResolutionHeight)
	res, out := do(t, "POST", e.url+"/configuration", body)
	wantStatus(t, res, out, 200)
}

func TestUpdateConfigurationAcceptsAnySizeWithoutACamera(t *testing.T) {
	e := newEnv(t)
	res, out := do(t, "POST", e.url+"/configuration", `{"SecondsBetweenCaptures":600,"OffsetWithinHour":0,"Quality":90,"PhotoResolutionWidth":800,"PhotoResolutionHeight":600}`)
	wantStatus(t, res, out, 200)
}

func TestUpdateConfigurationWithoutASizeKeepsTheSavedOne(t *testing.T) {
	e := newEnv(t)
	before := e.store.Get()
	res, out := do(t, "POST", e.url+"/configuration", `{"SecondsBetweenCaptures":600,"OffsetWithinHour":0,"Quality":90}`)
	wantStatus(t, res, out, 200)
	if got := e.store.Get(); got.PhotoResolutionWidth != before.PhotoResolutionWidth || got.PhotoResolutionHeight != before.PhotoResolutionHeight {
		t.Fatalf("stored %s", got)
	}
}

func jpegPhoto(t *testing.T, w, h, quality int) string {
	t.Helper()
	img := image.NewGray(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = uint8(i * 13)
	}
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, &jpeg.Options{Quality: quality}); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func TestMonitoringMeasuresRecentPhotos(t *testing.T) {
	e := newEnv(t)
	old := jpegPhoto(t, 64, 48, 50)
	recent := jpegPhoto(t, 64, 48, 80)
	e.photo(t, "a.jpg", old, 3*time.Hour) // taken before the quality was changed
	e.photo(t, "b.jpg", recent, 2*time.Hour)
	e.photo(t, "c.jpg", recent, time.Hour)
	e.photo(t, "d.jpg", "not a jpeg", 0) // unreadable photos are left out
	res, out := do(t, "GET", e.url+"/monitoring", "")
	wantStatus(t, res, out, 200)
	var m MonitoringResponse
	json.Unmarshal([]byte(out), &m)
	s := m.Storage
	want := float64(len(recent)) / (64 * 48)
	if s.SamplePhotos != 2 || s.SampleQuality != 80 || s.SampleBytesPerPixel < want*0.999 || s.SampleBytesPerPixel > want*1.001 {
		t.Fatalf("got %+v, want 2 photos at quality 80 with %.4f bytes per pixel", s, want)
	}
}

func TestMonitoringWithoutPhotosHasNoSample(t *testing.T) {
	e := newEnv(t)
	res, out := do(t, "GET", e.url+"/monitoring", "")
	wantStatus(t, res, out, 200)
	var m MonitoringResponse
	json.Unmarshal([]byte(out), &m)
	if m.Storage.SamplePhotos != 0 || m.Storage.SampleBytesPerPixel != 0 {
		t.Fatalf("got %+v", m.Storage)
	}
}

func TestUpdateConfigurationSavesNightSkipping(t *testing.T) {
	e := newEnv(t)
	res, out := do(t, "POST", e.url+"/configuration", `{"SecondsBetweenCaptures":600,"OffsetWithinHour":0,"Quality":90,"SkipNightPhotos":true,"Latitude":51.5074,"Longitude":-0.1278}`)
	wantStatus(t, res, out, 200)
	if got := e.store.Get(); !got.SkipNightPhotos || got.Latitude != 51.5074 || got.Longitude != -0.1278 {
		t.Fatalf("stored %s", got)
	}
}

func TestUpdateConfigurationNeedsALocationToSkipNights(t *testing.T) {
	e := newEnv(t)
	res, out := do(t, "POST", e.url+"/configuration", `{"SecondsBetweenCaptures":600,"OffsetWithinHour":0,"Quality":90,"SkipNightPhotos":true}`)
	wantStatus(t, res, out, 400)
	if !strings.Contains(out, "location") {
		t.Fatalf("got %s", out)
	}
}

func TestScheduleReportsTheNextCapture(t *testing.T) {
	e := newEnv(t)
	at := time.Now().Add(11 * time.Hour).Truncate(time.Millisecond)
	e.plan = &timelapse.Plan{At: at, Night: true}
	before := time.Now().UnixMilli()
	res, out := do(t, "GET", e.url+"/schedule", "")
	wantStatus(t, res, out, 200)
	var got ScheduleResponse
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	_, offset := time.Now().Zone()
	if got.NextMillis != at.UnixMilli() || !got.Night || got.UnixMillis < before || got.UnixMillis > time.Now().UnixMilli() || got.ZoneOffsetSeconds != offset {
		t.Fatalf("got %s", out)
	}
	if res.Header.Get("Cache-Control") != "no-store" {
		t.Errorf("Cache-Control %q", res.Header.Get("Cache-Control"))
	}
}

func TestScheduleBeforeTheFirstPlan(t *testing.T) {
	e := newEnv(t)
	res, out := do(t, "GET", e.url+"/schedule", "")
	wantStatus(t, res, out, 200)
	var got ScheduleResponse
	json.Unmarshal([]byte(out), &got)
	if got.NextMillis != 0 || got.UnixMillis == 0 {
		t.Fatalf("got %s", out)
	}
}
