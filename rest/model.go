package rest

import (
	"encoding/json"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
)

type ListFilesResponse struct {
	Files []files.File `json:"files"`
}

type UpdateConfigurationRequest struct {
	// For now match the settings type exactly. All settings can be updated. Consider changing this in the future.
	conf.Settings
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
