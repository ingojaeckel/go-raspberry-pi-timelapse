package rest

import "github.com/ingojaeckel/go-raspberry-pi-timelapse/files"

// Response types. Field names are part of the API (src/frontend/src/api.ts).

type ListFilesResponse struct {
	Files []files.File `json:"files"`
}

type MonitoringResponse struct {
	Time           string
	Uptime         string
	CpuTemperature string
	GpuTemperature string
	FreeDiskSpace  string
}

type GetPhotosResponse struct {
	Photos []Photo
}

type LogResponse struct {
	Logs string
}

type Photo struct {
	Name    string
	ModTime string
	Size    string
}
