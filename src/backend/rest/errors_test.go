package rest

import (
	"context"
	"errors"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
)

// Error paths of the handlers and the small helpers behind them.

func TestUpdateConfigurationStorageErrorIs500(t *testing.T) {
	e := newEnv(t)
	os.RemoveAll(filepath.Dir(e.settingsFile)) // the card went away: writes fail from now on
	res, body := do(t, "POST", e.url+"/configuration", `{"SecondsBetweenCaptures":600,"OffsetWithinHour":0,"Quality":90}`)
	wantStatus(t, res, body, 500)
}

func TestUpdateErrorStatus(t *testing.T) {
	if got := updateErrorStatus(conf.ErrInvalid); got != 400 {
		t.Errorf("invalid: %d", got)
	}
	if got := updateErrorStatus(errors.New("disk")); got != 500 {
		t.Errorf("other: %d", got)
	}
}

func TestMissingPhotoFolder(t *testing.T) {
	e := newEnv(t)
	os.RemoveAll(e.folder)
	res, body := do(t, "GET", e.url+"/photos", "")
	if res.StatusCode != 200 || strings.TrimSpace(body) != `{"Photos":[]}` {
		t.Errorf("/photos: %d %s", res.StatusCode, body)
	}
	for _, req := range [][2]string{{"DELETE", "/photos"}, {"GET", "/archive/zip"}} {
		res, body := do(t, req[0], e.url+req[1], "")
		wantStatus(t, res, body, 500)
	}
}

// vanishingCamera reports a photo that is gone before it can be served.
type vanishingCamera struct{}

func (vanishingCamera) Capture(context.Context, string, timelapse.Options) (string, error) {
	return "/nonexistent/preview.jpg", nil
}

func TestPreviewFileVanished(t *testing.T) {
	e := newEnv(t)
	e.srv.Camera = vanishingCamera{}
	res, body := do(t, "GET", e.url+"/capture", "")
	wantStatus(t, res, body, 500)
}

func TestLogsMissingFile(t *testing.T) {
	e := newEnv(t)
	res, body := do(t, "GET", e.url+"/logs", "") // LogFile set but never written
	wantStatus(t, res, body, 500)
}

func TestMonitoringWithoutThermalSensor(t *testing.T) {
	e := newEnv(t)
	_, body := do(t, "GET", e.url+"/monitoring", "") // ThermalFile does not exist
	if !strings.Contains(body, "no such file") {
		t.Fatalf("body %s", body)
	}
}

func TestWriteJSONLogsEncodingErrors(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, 200, make(chan int)) // channels cannot be encoded
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("status %d headers %v", rec.Code, rec.Header())
	}
}

func TestStatusRecorderUnwrap(t *testing.T) {
	rec := httptest.NewRecorder()
	if (&statusRecorder{ResponseWriter: rec}).Unwrap() != http.ResponseWriter(rec) {
		t.Fatal("Unwrap must return the wrapped writer")
	}
}

func TestSelected(t *testing.T) {
	all := []files.File{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	if got := files.Names(selected(all, nil)); strings.Join(got, ",") != "a,b,c" {
		t.Errorf("no filter: %v", got)
	}
	if got := files.Names(selected(all, []string{"c", "a", "x"})); strings.Join(got, ",") != "a,c" {
		t.Errorf("filter keeps listing order: %v", got)
	}
}
