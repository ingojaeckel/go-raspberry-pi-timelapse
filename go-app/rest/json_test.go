package rest

import (
	"bytes"
	"testing"

	"github.com/facebookgo/ensure"
)

func TestToJson(t *testing.T) {
	b, err := toJSON(Photo{Name: "p.jpg", ModTime: "1", Size: "9"})
	ensure.Nil(t, err)
	ensure.DeepEqual(t, "{\"Name\":\"p.jpg\",\"ModTime\":\"1\",\"Size\":\"9\"}", string(b))
}

func TestParseInvalidJSON(t *testing.T) {
	reader := bytes.NewReader([]byte("{\"Name\":\"p.jpg\",\"ModTime\":\"1\",\"Size\":\"9"))

	var p Photo
	ensure.NotNil(t, parseJSON(reader, &p))
}

func TestParseJSON(t *testing.T) {
	reader := bytes.NewReader([]byte("{\"Name\":\"p.jpg\",\"ModTime\":\"1\",\"Size\":\"9\"}"))

	var p Photo
	ensure.Nil(t, parseJSON(reader, &p))
	ensure.DeepEqual(t, Photo{Name: "p.jpg", ModTime: "1", Size: "9"}, p)
}
