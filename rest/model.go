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
}
