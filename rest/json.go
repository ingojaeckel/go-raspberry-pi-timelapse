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

	w.Header().Set("content-type", "application/json")

	// Allow CORS
	// TODO Limit to dev mode
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	w.WriteHeader(status)
	io.WriteString(w, string(val))
	return nil
}
