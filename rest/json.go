package rest

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

func toJSON(r interface{}) ([]byte, error) {
	val, err := json.Marshal(r)
	if err != nil {
		return []byte{}, err
	}
	return val, nil
}

func parseJSON(reader io.Reader, val interface{}) error {
	body, err := ioutil.ReadAll(reader)
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

	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	io.WriteString(w, string(val))
	return nil
}
