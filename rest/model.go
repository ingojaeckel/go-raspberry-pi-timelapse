package rest

import (
	"encoding/json"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
)

type ListFilesResponse struct {
	Files []files.File `json:"files"`
}

type UpdateConfigurationRequest struct {
	SecondsBetweenCaptures int
	OffsetWithinHour       int
	RotateBy               int
	Quality                int
	ResolutionSetting      int

	// TODO consider adding as neccessary
	// PhotoResolutionWidth    int
	// PhotoResolutionHeight   int
	// PreviewResolutionWidth  int
	// PreviewResolutionHeight int
}

func (u UpdateConfigurationRequest) String() string {
	jsonStr, _ := json.Marshal(u)
	return string(jsonStr)
}

type MonitoringResponse struct {
	Time           string
	Uptime         string
	CpuTemperature string
	GpuTemperature string
	FreeDiskSpace  string
	// TODO add FreeMemory
}

type GetPhotosResponse struct {
	Photos []Photo
}

type Photo struct {
	Name    string
	ModTime string
	Size    string
}
