package rest

import (
	"encoding/json"
	"fmt"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/admin"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"goji.io/pat"
	"net/http"
	"runtime"
	"strings"
)

const MaxFileSizeBytes = 100485760 // 100 MB
const Version = 1

func GetVersion(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "Hello from %s on %s [version:%d]", runtime.GOARCH, runtime.GOOS, Version)
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

	serveFileContent(w, name)
}

func GetMostRecentFile(w http.ResponseWriter, _ *http.Request) {
	f, _ := files.ListFiles("timelapse-pictures", true)
	if len(f) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	mostRecentFile := f[len(f)-1]
	serveFileContent(w, "timelapse-pictures/"+mostRecentFile.Name)
}

func GetFiles(w http.ResponseWriter, _ *http.Request) {
	f, _ := files.ListFiles("timelapse-pictures", true)
	resp := ListFilesResponse{f}

	b, _ := json.Marshal(resp)
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}

func GetArchive(w http.ResponseWriter, _ *http.Request) {
	f, _ := files.ListFiles("timelapse-pictures", true)

	// Convert []File to []string
	strFiles := make([]string, len(f))
	for i, file := range f {
		strFiles[i] = "timelapse-pictures/" + file.Name
	}
	tarBytes, _ := files.Tar(strFiles)

	w.Header().Add("Content-Type", "application/tar")
	w.Header().Set("Content-Disposition", "attachment; filename=archive.tar")
	w.Write(tarBytes)
}

func Admin(_ http.ResponseWriter, r *http.Request) {
	command := pat.Param(r, "command")
	admin.HandleCommand(command)
}

func serveFileContent(w http.ResponseWriter, path string) {
	content, e := files.GetFile(path)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Print(e.Error())
		return
	}

	w.Header().Add("Content-Type", http.DetectContentType(content))
	w.Header().Set("Content-Disposition", "attachment; filename="+getBasename(path))
	w.Write(content)
}

func getBasename(path string) string {
	i := strings.LastIndex(path, "/")
	if i == -1 {
		return path
	}
	return path[i+1:]
}
