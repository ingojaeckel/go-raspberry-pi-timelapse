package rest

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	mu  sync.Mutex
	ran [][]string
}

func (c *commands) run(_ context.Context, name string, args ...string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ran = append(c.ran, append([]string{name}, args...))
	if name == "/bin/date" {
		return []byte("Mon Nov  2 10:00:00 UTC 2026\n"), nil
	}
	return nil, errors.New(name + " not available")
}

type env struct {
	srv      *Server
	url      string
	folder   string
	store    *conf.Store
	camera   *fakeCamera
	commands *commands
	handler  http.Handler // mux + middleware, as served at url
}

func newEnv(t testing.TB) *env {
	t.Helper()
	dir := t.TempDir()
	store, err := conf.OpenStore(filepath.Join(dir, "settings.db"), conf.Default(1800))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	e := &env{store: store, folder: filepath.Join(dir, "photos"), camera: &fakeCamera{}, commands: &commands{}}
	os.Mkdir(e.folder, 0o700)
	os.Mkdir(filepath.Join(dir, "tmp"), 0o700)
	e.srv = &Server{
		Settings:    store,
		Camera:      e.camera,
		Photos:      files.NewIndex(e.folder),
		TempFolder:  filepath.Join(dir, "tmp"),
		LogFile:     filepath.Join(dir, "app.log"),
		ThermalFile: filepath.Join(dir, "temp"),
		Version:     "test-v1",
		Run:         e.commands.run,
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
	body := `{"SecondsBetweenCaptures":600,"OffsetWithinHour":300,"Quality":85,"ResolutionSetting":2,"RotateBy":180,"DebugEnabled":true,"PreviewResolutionWidth":1}`
	res, out := do(t, "POST", e.url+"/configuration", body)
	wantStatus(t, res, out, 200)
	got := e.store.Get()
	if got.SecondsBetweenCaptures != 600 || got.OffsetWithinHour != 300 || got.Quality != 85 || got.RotateBy != 180 || !got.DebugEnabled {
		t.Fatalf("stored %s", got)
	}
	if got.PhotoResolutionWidth != 1640 || got.PhotoResolutionHeight != 1232 {
		t.Fatalf("resolution not derived: %s", got)
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
		"unknown resolution":          `{"SecondsBetweenCaptures":600,"OffsetWithinHour":0,"Quality":90,"ResolutionSetting":99}`,
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
