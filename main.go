package main

import (
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/rest"
	"goji.io"
	"goji.io/pat"
	"net/http"
)

func main() {
	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/version"), rest.GetVersion)
	mux.HandleFunc(pat.Get("/file/:fileName"), rest.GetFile)

	http.ListenAndServe("localhost:8080", mux)
}
