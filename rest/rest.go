package rest

import (
	"encoding/json"
	"fmt"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/admin"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"github.com/loranbriggs/go-camera"
	"goji.io/pat"
	"io"
	"net/http"
	"runtime"
	"strings"
)

const (
	Version                  = 1
	MaxFileSizeBytes         = 100485760 // 100 MB
	HeaderContentType        = "Content-Type"
	HeaderContentDisposition = "Content-Disposition"
	HeaderContentTypeJSON    = "application/json"
	StorageFolder            = "timelapse-pictures"
	TempFilesFolder          = "/tmp"
)

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
	f, _ := files.ListFiles(StorageFolder, true)
	if len(f) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	mostRecentFile := f[len(f)-1]
	serveFileContent(w, fmt.Sprintf("%s/%s", StorageFolder, mostRecentFile.Name))
}

func GetFiles(w http.ResponseWriter, _ *http.Request) {
	f, _ := files.ListFiles(StorageFolder, true)
	resp := ListFilesResponse{f}

	b, _ := json.Marshal(resp)
	w.Header().Add(HeaderContentType, HeaderContentTypeJSON)
	w.Write(b)
}

func Capture(w http.ResponseWriter, _ *http.Request) {
	c := camera.New(TempFilesFolder)
	path, err := c.Capture()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to take capture: %s", err.Error())
		return
	}

	serveFileContent(w, path)

	// Remove the temporary file
	// os.Remove(path)
}

func GetArchive(w http.ResponseWriter, _ *http.Request) {
	f, _ := files.ListFiles(StorageFolder, true)

	// Convert []File to []string
	strFiles := make([]string, len(f))
	for i, file := range f {
		strFiles[i] = fmt.Sprintf("%s/%s", StorageFolder, file.Name)
	}

	// tarBytes, _ := files.Tar(strFiles)
	pr, pw := io.Pipe()

	go func() {
		files.TarWithPipes(strFiles, pw)
		defer pw.Close()
	}()

	w.Header().Add(HeaderContentType, "application/tar")
	w.Header().Set(HeaderContentDisposition, "attachment; filename=archive.tar")

	// read 1MB from pr and call w.Write()
	buf := make([]byte, 1024*1024)
	for {
		fmt.Println("Reading...")
		n, err := pr.Read(buf)
		fmt.Printf("Read %d bytes\n", n)
		if err == io.EOF {
			fmt.Println("reached EOF")
			break
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Errorf("Error: %s", err.Error())
			break
		}
		if n == 0 {
			fmt.Println("no bytes left")
			return
		}
		fmt.Printf("Writing %d bytes..\n", n)
		w.Write(buf[0:n])
	}
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

	w.Header().Add(HeaderContentType, http.DetectContentType(content))
	w.Header().Set(HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", getBasename(path)))
	w.Write(content)
}

func getBasename(path string) string {
	i := strings.LastIndex(path, "/")
	if i == -1 {
		return path
	}
	return path[i+1:]
}
