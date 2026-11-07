package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
)

// apiShapesFile records the JSON shape (field names and types) of each API
// response. The frontend's fake server (mockServer.ts) is tested against it,
// so a renamed or retyped Go field fails a test on both sides.
const apiShapesFile = "../../frontend/src/test/api-shapes.json"

// contractRequests are the JSON endpoints the frontend reads.
var contractRequests = []struct{ method, path, body string }{
	{"GET", "/configuration", ""},
	{"GET", "/monitoring", ""},
	{"GET", "/camera", ""},
	{"GET", "/schedule", ""},
	{"GET", "/photos", ""},
	{"GET", "/logs", ""},
	{"POST", "/admin/time", `{"UnixMillis":1800000000000}`},
}

func TestAPIShapesMatchTheFrontendFixtures(t *testing.T) {
	got := recordAPIShapes(t)
	if os.Getenv("UPDATE_API_SHAPES") != "" {
		if err := os.WriteFile(apiShapesFile, got, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(apiShapesFile)
	if err != nil {
		t.Fatalf("%v (create it with UPDATE_API_SHAPES=1 go test ./rest -run APIShapes)", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("the API changed shape; update %s with UPDATE_API_SHAPES=1 go test ./rest -run APIShapes, then fix api.ts and mockServer.ts to match.\ngot:\n%s", apiShapesFile, got)
	}
}

// recordAPIShapes calls each contract endpoint on a camera with a photo, a
// detected sensor and a planned capture, so lists and objects are filled.
func recordAPIShapes(t *testing.T) []byte {
	e := newEnv(t)
	e.photo(t, "20261101-081500.jpg", "jpeg", time.Hour)
	e.sensor = &timelapse.Sensor{Model: "imx219", Width: 3280, Height: 2464, Modes: []timelapse.Size{{Width: 1640, Height: 1232}}}
	e.plan = &timelapse.Plan{At: time.Now().Add(time.Hour)}
	os.WriteFile(e.srv.LogFile, []byte("started\n"), 0o600)
	shapes := map[string]any{}
	for _, c := range contractRequests {
		res, body := do(t, c.method, e.url+c.path, c.body)
		wantStatus(t, res, body, 200)
		var v any
		if err := json.Unmarshal([]byte(body), &v); err != nil {
			t.Fatalf("%s %s: %v", c.method, c.path, err)
		}
		shapes[c.method+" "+c.path] = shapeOf(v)
	}
	out, err := json.MarshalIndent(shapes, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return append(out, '\n')
}

// shapeOf replaces each value by its JSON type name; an array becomes a list
// of its first element's shape, or [] when empty (any element type).
func shapeOf(v any) any {
	switch v := v.(type) {
	case map[string]any:
		shape := map[string]any{}
		for key, field := range v {
			shape[key] = shapeOf(field)
		}
		return shape
	case []any:
		if len(v) == 0 {
			return []any{}
		}
		return []any{shapeOf(v[0])}
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case nil:
		return "null"
	}
	panic(fmt.Sprintf("unexpected JSON value %T", v))
}
