package rest

import (
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
)

type ListFilesResponse struct {
	Files []files.File `json:"files"`
}

type UpdateConfigurationRequest struct {
	TimeBetween   int `json:"timeBetween"`
	InitialOffset int `json:"initialOffset"`
	Resolution    int `json:"resolution"`
	RotateBy      int `json:"rotateBy"`
	Quality       int `json:"quality"`
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
