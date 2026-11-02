package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
)

func toJSON(r interface{}) ([]byte, error) {
	val, err := json.Marshal(r)
	if err != nil {
		return []byte{}, err
	}
	return val, nil
}

func parseJSON(reader io.Reader, val interface{}) error {
	body, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, &val)
}

func writeJSON(w http.ResponseWriter, status int, r interface{}) error {
	val, err := toJSON(r)

	if err != nil {
		return err
	}

	w.Header().Set(conf.HeaderContentType, conf.HeaderContentTypeJSON)
	w.WriteHeader(status)
	io.WriteString(w, string(val))
	return nil
}
