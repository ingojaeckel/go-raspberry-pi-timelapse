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

func GetArchive(w http.ResponseWriter, r *http.Request) {
	// TODO let client pass in directory name / timestamp to control which files are downloaded. for now, tar everything in the CWD.
	d, _ := os.Getwd()
	allFiles, _ := files.ListFiles(d)

	actualFiles := make([]string, len(allFiles))
	i := 0

	for _, file := range allFiles {
		if !file.IsDir {
			actualFiles[i] = file.Name
			i = i + 1
		}
	}

	tarBytes, _ := files.Tar(actualFiles[:i])

	w.Header().Add("Content-Type", "application/tar")
	w.Write(tarBytes)
}