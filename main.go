package main

import (
	"fmt"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/rest"
	"goji.io"
	"goji.io/pat"
	"net/http"
)

func main() {
	addr := "localhost:8080"
	fmt.Printf("Listening on %s...\n", addr)

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/version"), rest.GetVersion)

	mux.HandleFunc(pat.Get("/file"), rest.GetFiles)
	mux.HandleFunc(pat.Get("/file/:fileName"), rest.GetFile)

	mux.HandleFunc(pat.Get("/archive"), rest.GetArchive)

	http.ListenAndServe(addr, mux)
}
