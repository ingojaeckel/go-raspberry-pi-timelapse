package rest

import (
	"net/http"
	"fmt"
	"goji.io/pat"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"runtime"
)

func GetVersion(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from %s on %s", runtime.GOARCH, runtime.GOOS)
}

func GetFile(w http.ResponseWriter, r *http.Request) {
	name := pat.Param(r, "fileName")
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