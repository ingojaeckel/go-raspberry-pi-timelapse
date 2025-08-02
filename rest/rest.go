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
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/admin"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf/valid"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/detection"
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

func MakeUpdateConfigurationFn(configUpdatedChan chan<- conf.Settings) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request UpdateConfigurationRequest
		if err := parseJSON(r.Body, &request); err != nil {
			writeJSON(w, 400, err.Error())
			return
		}
		// Validate configuration before updating it.
		if err := valid.New().Validate(request.Settings); err != nil {
			writeJSON(w, 400, err.Error())
			return
		}
		updatedSettings, err := updatePartialConfiguration(request)
		if err != nil {
			writeJSON(w, 400, err.Error())
			return
		}
		writeJSON(w, 200, updatedSettings)
		configUpdatedChan <- *updatedSettings
	}
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

// GetDetection performs object detection on the most recent photo and returns results
func GetDetection(w http.ResponseWriter, s *conf.Settings) {
	// Load current settings to get the latest configuration
	currentSettings, err := conf.LoadConfiguration()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": fmt.Sprintf("Failed to load configuration: %s", err.Error())})
		return
	}
	
	if !currentSettings.ObjectDetectionEnabled {
		writeJSON(w, 200, DetectionResponse{&detection.DetectionResult{
			Summary: "Object detection is disabled",
		}})
		return
	}

	// Get the most recent photo
	files, err := files.ListFiles(conf.StorageFolder, true)
	if err != nil || len(files) == 0 {
		writeJSON(w, 404, map[string]string{"error": "No photos found"})
		return
	}

	mostRecentFile := files[len(files)-1]
	photoPath := fmt.Sprintf("%s/%s", conf.StorageFolder, mostRecentFile.Name)

	// Run object detection with current settings
	config := &detection.DetectionConfig{
		UseOpenCV: currentSettings.UseOpenCVDetection,
		Timeout:   time.Duration(currentSettings.DetectionTimeout) * time.Second,
	}
	result, err := detection.AnalyzePhotoWithConfig(photoPath, config)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": fmt.Sprintf("Object detection failed: %s", err.Error())})
		return
	}

	writeJSON(w, 200, DetectionResponse{result})
}

// GetArchiveZip Reply with ZIP file containing all timelapse pictures
func GetArchiveZip(w http.ResponseWriter, r *http.Request) {
	strFiles, _ := requestedFilesToRelativePaths(r.URL.Query()["f"]) // TODO handle error

	pr, pw := io.Pipe()
	go func() {
		if err := files.ZipWithPipes(strFiles, pw); err != nil {
			log.Println("failed to create archive", err.Error())
		}
		defer pw.Close()
	}()

	w.Header().Add(conf.HeaderContentType, "application/zip")
	w.Header().Set(conf.HeaderContentDisposition, "attachment; filename=archive.zip")

	writePipeContent(w, pr)
}

// GetArchiveTar Reply with TAR file containing all timelapse pictures
func GetArchiveTar(w http.ResponseWriter, r *http.Request) {
	strFiles, _ := requestedFilesToRelativePaths(r.URL.Query()["f"]) // TODO handle error

	pr, pw := io.Pipe()
	go func() {
		if err := files.TarWithPipes(strFiles, pw); err != nil {
			log.Println("failed to create archive", err.Error())
		}
		defer pw.Close()
	}()

	w.Header().Add(conf.HeaderContentType, "application/tar")
	w.Header().Set(conf.HeaderContentDisposition, "attachment; filename=archive.tar")

	writePipeContent(w, pr)
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

func updatePartialConfiguration(updateRequest UpdateConfigurationRequest) (*conf.Settings, error) {
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
	s.ObjectDetectionEnabled = updateRequest.ObjectDetectionEnabled
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

func writePipeContent(w http.ResponseWriter, pr *io.PipeReader) {
	emptyReadMaxAttemps := 100 // The first few read attempts may not get any data yet as another go routine produces the data. Tweak this to adjust the number of attempts used before giving up.
	buf := make([]byte, 1024*1024)

	for i := 0; ; i++ {
		// read the next (up to) 1MB from the pipe to send to response writer
		n, err := pr.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Error: ", err.Error())
			break
		}
		if n == 0 {
			if i < emptyReadMaxAttemps {
				continue
			}
			return
		}
		w.Write(buf[0:n])
	}
}

// Converts the front-end requested file names (provided via query param) to paths relative to the storage folder. If no filter was provided, assume all files are requested.
func requestedFilesToRelativePaths(filteredFiles []string) ([]string, error) {
	filterProvided := len(filteredFiles) > 0
	fileNamesIncluded := make(map[string]bool)

	if filterProvided {
		log.Printf("Limit archive to files: %v\n", filteredFiles)

		for _, name := range filteredFiles {
			fileNamesIncluded[name] = true
		}
	}

	filesToArchive, err := files.ListFiles(conf.StorageFolder, true)
	if err != nil {
		return nil, err
	}
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

	return strFiles, nil
}
