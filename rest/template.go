package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/admin"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
)

func GetMonitoring(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(conf.HeaderContentType, conf.HeaderContentTypeJSON)
	json.NewEncoder(w).Encode(&MonitoringResponse{
		Time:           admin.RunCommandOrFallback("/bin/date"),
		GpuTemperature: admin.RunCommandOrFallback("/opt/vc/bin/vcgencmd", "measure_temp"),
		CpuTemperature: admin.RunCommandOrFallback("/bin/cat", "/sys/class/thermal/thermal_zone0/temp"),
		Uptime:         admin.RunCommandOrFallback("/usr/bin/uptime"),
		FreeDiskSpace:  admin.RunCommandOrFallback("/bin/df", "-h"),
	})
}

func GetLogs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(conf.HeaderContentType, conf.HeaderContentTypeJSON)

	logFileContent, err := os.ReadFile(conf.LogFile)
	if err == nil {
		json.NewEncoder(w).Encode(LogResponse{Logs: string(logFileContent)})
	} else {
		writeJSON(w, http.StatusInternalServerError, err.Error())
	}
}

func GetPhotos(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(conf.HeaderContentType, conf.HeaderContentTypeJSON)
	json.NewEncoder(w).Encode(getPhotosFrom(conf.StorageFolder))
}

func getPhotosFrom(folder string) GetPhotosResponse {
	files, _ := files.ListFiles(folder, true)
	photos := make([]Photo, len(files))
	for i, f := range files {
		photos[i] = Photo{
			Name:    f.Name,
			ModTime: f.ModTime,
			Size:    getHumanReadableSize(f.Bytes),
		}
	}
	return GetPhotosResponse{Photos: photos}
}

func getHumanReadableSize(sizeInBytes int64) string {
	if sizeInBytes >= 1024*1024 {
		return fmt.Sprintf("%d MB", sizeInBytes/1024/1024)
	}
	if sizeInBytes >= 1024 {
		return fmt.Sprintf("%d KB", sizeInBytes/1024)
	}
	return fmt.Sprintf("%d bytes", sizeInBytes)
}
