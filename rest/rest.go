package rest

import (
	"encoding/json"
	"fmt"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/admin"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"github.com/loranbriggs/go-camera"
	"goji.io/pat"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

func GetVersion(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "Hello from %s on %s [version:%d]", runtime.GOARCH, runtime.GOOS, conf.Version)
}

func GetFile(w http.ResponseWriter, r *http.Request) {
	name := pat.Param(r, "fileName")

	canServe, e := files.CanServeFile(name, conf.MaxFileSizeBytes)
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
	f, _ := files.ListFiles(conf.StorageFolder, true)
	if len(f) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	mostRecentFile := f[len(f)-1]
	serveFileContent(w, fmt.Sprintf("%s/%s", conf.StorageFolder, mostRecentFile.Name))
}

func GetFiles(w http.ResponseWriter, _ *http.Request) {
	f, _ := files.ListFiles(conf.StorageFolder, true)
	resp := ListFilesResponse{f}

	b, _ := json.Marshal(resp)
	w.Header().Add(conf.HeaderContentType, conf.HeaderContentTypeJSON)
	w.Write(b)
}

func Capture(w http.ResponseWriter, _ *http.Request) {
	c := camera.New(conf.TempFilesFolder, conf.PreviewResolution.Width, conf.PreviewResolution.Height)
	path, err := c.Capture()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to take capture: %s", err.Error())
		return
	}

	serveFileContent(w, path)

	// Remove the temporary file
	os.Remove(path)
}

func GetArchive(w http.ResponseWriter, _ *http.Request) {
	f, _ := files.ListFiles(conf.StorageFolder, true)

	// Convert []File to []string
	strFiles := make([]string, len(f))
	for i, file := range f {
		strFiles[i] = fmt.Sprintf("%s/%s", conf.StorageFolder, file.Name)
	}

	// tarBytes, _ := files.Tar(strFiles)
	pr, pw := io.Pipe()

	go func() {
		files.TarWithPipes(strFiles, pw)
		defer pw.Close()
	}()

	w.Header().Add(conf.HeaderContentType, "application/tar")
	w.Header().Set(conf.HeaderContentDisposition, "attachment; filename=archive.tar")

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

	w.Header().Add(conf.HeaderContentType, http.DetectContentType(content))
	w.Header().Set(conf.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", getBasename(path)))
	w.Write(content)
}

func getBasename(path string) string {
	i := strings.LastIndex(path, "/")
	if i == -1 {
		return path
	}
	return path[i+1:]
}
