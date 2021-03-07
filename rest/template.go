package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/admin"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
)

// TODO process should be auto restarted (e.g. by supervisord)
// TODO add monitoring for main memory

func GetMonitoring(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // TODO limit to dev mode
	json.NewEncoder(w).Encode(&MonitoringResponse{
		Time:           admin.RunCommandOrPanic("/bin/date"),
		GpuTemperature: admin.RunCommandOrPanic("/opt/vc/bin/vcgencmd", "measure_temp"),
		CpuTemperature: admin.RunCommandOrPanic("/bin/cat", "/sys/class/thermal/thermal_zone0/temp"),
		Uptime:         admin.RunCommandOrPanic("/usr/bin/uptime"),
		FreeDiskSpace:  admin.RunCommandOrPanic("/bin/df", "-h"),
	})
}

func GetPhotos(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // TODO limit to dev mode
	json.NewEncoder(w).Encode(getPhotosFrom(conf.StorageFolder))
}

func getPhotosFrom(folder string) GetPhotosResponse {
	files, _ := files.ListFiles(folder, true) // TODO handle error
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
		// At least 1MB of files
		return fmt.Sprintf("%d MB", sizeInBytes/1024/1024)
	}
	if sizeInBytes >= 1024 {
		// Size is between 1KB - 1MB
		return fmt.Sprintf("%d KB", sizeInBytes/1024)
	}
	// Size is between 0 - 1KB
	return fmt.Sprintf("%d bytes", sizeInBytes)
}
