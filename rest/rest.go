package rest

import (
	"encoding/json"
	"fmt"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"goji.io/pat"
	"net/http"
	"os"
	"runtime"
)

const MaxFileSizeBytes = 100485760

func GetVersion(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from %s on %s", runtime.GOARCH, runtime.GOOS)
}

func GetFile(w http.ResponseWriter, r *http.Request) {
	name := pat.Param(r, "fileName")

	canServe, e := files.CanServeFile(name, MaxFileSizeBytes)
	if !canServe {
		w.WriteHeader(http.StatusNotFound)
		if e != nil {
			fmt.Fprint(w, e.Error())
		}
		return
	}

	content, e := files.GetFile(name)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Print(e.Error())
		return
	}

	contentType := http.DetectContentType(content)
	w.Header().Add("Content-Type", contentType)
	w.Write(content)
}

func GetFiles(w http.ResponseWriter, r *http.Request) {
	d, _ := os.Getwd()
	f, _ := files.ListFiles(d)
	resp := ListFilesResponse{f}

	b, _ := json.Marshal(resp)
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}
