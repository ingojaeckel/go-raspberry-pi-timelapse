package rest

import (
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/health"
)

// Response types. Field names are part of the API (src/frontend/src/api.ts).

type ListFilesResponse struct {
	Files []files.File `json:"files"`
}

// MonitoringResponse holds command output (or the error) in its string
// fields and parsed readings in the others.
type MonitoringResponse struct {
	Time           string
	Uptime         string
	CpuTemperature string
	GpuTemperature string
	FreeDiskSpace  string
	// UnixMillis is the camera's clock, to compare with the browser's.
	UnixMillis int64
	Power      PowerStatus
	CPU        CPUStatus
	Storage    StorageStatus
}

// PowerStatus lists under-voltage and throttling (health.Throttled).
type PowerStatus struct {
	Now, SinceBoot []string
	Error          string `json:",omitempty"`
}

// CPUStatus is the CPU frequency scaling state.
type CPUStatus struct {
	health.CPUFreq
	Error string `json:",omitempty"`
}

// StorageStatus describes the photo folder's file system, for estimating
// when it is full.
type StorageStatus struct {
	FreeBytes uint64
	// ReserveBytes is kept free (-minFreeDiskMB): the disk counts as full
	// when FreeBytes drops below it.
	ReserveBytes uint64
	Photos       int
	PhotoBytes   int64
	// StopWhenDiskFull is the setting: stop taking photos, or delete the
	// oldest ones (DeletedOldest since the app started).
	StopWhenDiskFull bool
	DeletedOldest    int
	Error            string `json:",omitempty"`
}

// SetTimeRequest sets the camera's clock, e.g. to the browser's.
type SetTimeRequest struct {
	UnixMillis int64
}

type SetTimeResponse struct {
	UnixMillis int64
	// RTC says whether the battery-backed clock module was updated too:
	// empty without one, "updated", or why not.
	RTC string
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
