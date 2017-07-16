package main

import (
	"fmt"
	"goji.io"
	"goji.io/pat"
	"net/http"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
)

func getVersion(w http.ResponseWriter, r *http.Request) {
	fmt.Print(w, "Hello!")
}

func getFile(w http.ResponseWriter, r *http.Request) {
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

func main() {
	fmt.Println("Hello World from ARM")

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/version"), getVersion)
	mux.HandleFunc(pat.Get("/file/:fileName"), getFile)

	http.ListenAndServe("localhost:8080", mux)
}
