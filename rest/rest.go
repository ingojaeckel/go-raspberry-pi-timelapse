package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/admin"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
	"goji.io/pat"
)

func MakeGetVersionFn(version string) func(w http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "Hello from %s on %s [version:%s]", runtime.GOARCH, runtime.GOOS, version)
	}
}

func GetConfiguration(w http.ResponseWriter, _ *http.Request) {
	c, _ := conf.LoadConfiguration()
	writeJSON(w, 200, c)
}

func UpdateConfiguration(w http.ResponseWriter, r *http.Request) {
	var request UpdateConfigurationRequest
	if err := parseJSON(r.Body, &request); err != nil {
		writeJSON(w, 400, err.Error())
		return
	}
	updatedSettings, err := UpdatePartialConfiguration(request)
	if err != nil {
		writeJSON(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, updatedSettings)
}

func GetFile(w http.ResponseWriter, r *http.Request) {
	name := pat.Param(r, "fileName")
	fullyQualifiedPath := conf.StorageFolder + "/" + name

	canServe, e := files.CanServeFile(fullyQualifiedPath, conf.MaxFileSizeBytes)
	if !canServe {
		w.WriteHeader(http.StatusNotFound)
		if e != nil {
			fmt.Fprint(w, e.Error())
		}
		return
	}

	serveFileContent(w, fullyQualifiedPath)
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

func Capture(w http.ResponseWriter, s *conf.Settings) {
	log.Printf("Capturing preview picture inside of %s at resolution: %d x %d\n", conf.TempFilesFolder, s.PreviewResolutionWidth, s.PreviewResolutionHeight)
	c, err := timelapse.NewCamera(conf.TempFilesFolder, s.PreviewResolutionWidth, s.PreviewResolutionHeight, s.RotateBy == 180, s.Quality)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to instantiate camera: %s", err.Error())
		return
	}

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

// GetArchiveZip Reply with ZIP file containing all timelapse pictures
func GetArchiveZip(w http.ResponseWriter, r *http.Request) {
	filteredFiles := r.URL.Query()["f"]
	filterProvided := len(filteredFiles) > 0
	fileNamesIncluded := make(map[string]bool)

	if filterProvided {
		log.Printf("Limit archive to files: %v\n", filteredFiles)

		for _, name := range filteredFiles {
			fileNamesIncluded[name] = true
		}
	}

	filesToArchive, _ := files.ListFiles(conf.StorageFolder, true) // TODO handle error
	if filterProvided {
		var filteredFileList []files.File
		for _, file := range filesToArchive {
			if fileNamesIncluded[file.Name] {
				filteredFileList = append(filteredFileList, file)
			}
		}
		log.Printf("Reduced number of files in archive from %d to %d based on user provided filter.\n", len(filesToArchive), len(filteredFileList))
		filesToArchive = filteredFileList
	}

	// Convert []File to []string
	strFiles := make([]string, len(filesToArchive))
	for i, file := range filesToArchive {
		strFiles[i] = fmt.Sprintf("%s/%s", conf.StorageFolder, file.Name)
	}

	pr, pw := io.Pipe()
	go func() {
		files.ZipWithPipes(strFiles, pw)
		defer pw.Close()
	}()

	w.Header().Add(conf.HeaderContentType, "application/zip")
	w.Header().Set(conf.HeaderContentDisposition, "attachment; filename=archive.zip")

	// read 1MB from pr and call w.Write()
	buf := make([]byte, 1024*1024)
	for {
		log.Println("Reading...")
		n, err := pr.Read(buf)
		log.Printf("Read %d bytes\n", n)
		if err == io.EOF {
			log.Println("reached EOF")
			break
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Error: ", err.Error())
			break
		}
		if n == 0 {
			log.Println("no bytes left")
			return
		}
		log.Printf("Writing %d bytes..\n", n)
		w.Write(buf[0:n])
	}
}

func DeleteFiles(w http.ResponseWriter, r *http.Request) {
	filesToDelete := r.URL.Query()["f"]
	filesToDeleteAreProvided := len(filesToDelete) > 0

	if !filesToDeleteAreProvided {
		writeJSON(w, http.StatusBadRequest, "no files to delete provided.")
		return
	}

	log.Printf("Files to delete: %v\n", filesToDelete)

	for _, name := range filesToDelete {
		path := fmt.Sprintf("%s/%s", conf.StorageFolder, name)
		if err := files.RemoveFile(path); err != nil {
			log.Printf("Failed to delete %s: %v\n", path, err)
		}
	}
	writeJSON(w, http.StatusOK, len(filesToDelete))
}

func Admin(_ http.ResponseWriter, r *http.Request) {
	command := pat.Param(r, "command")
	admin.HandleCommand(command)
}

func serveFileContent(w http.ResponseWriter, path string) {
	content, err := files.GetFile(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Print(err.Error())
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

func UpdatePartialConfiguration(updateRequest UpdateConfigurationRequest) (*conf.Settings, error) {
	// TODO validate new config coming in via updateRequest
	log.Printf("Received new configuration: %v\n", updateRequest)

	s, err := conf.LoadConfiguration()
	log.Printf("Updating old configuration (%v)\nwith new configuration (%v)\n", s, updateRequest)

	if err != nil {
		return nil, err
	}

	s.Quality = updateRequest.Quality
	s.RotateBy = updateRequest.RotateBy
	s.OffsetWithinHour = updateRequest.OffsetWithinHour
	s.ResolutionSetting = updateRequest.ResolutionSetting
	s.SecondsBetweenCaptures = updateRequest.SecondsBetweenCaptures
	switch s.ResolutionSetting {
	case 2:
		s.PhotoResolutionWidth, s.PhotoResolutionHeight = 1640, 1232
	case 1:
		s.PhotoResolutionWidth, s.PhotoResolutionHeight = 2186, 1642
	case 0:
	default:
		s.PhotoResolutionWidth, s.PhotoResolutionHeight = 3280, 2464
	}

	log.Printf("New configuration: %v\n", s)
	return conf.WriteConfiguration(*s)
}
