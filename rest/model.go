package rest

import (
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
)

type ListFilesResponse struct {
	Files []files.File `json:"files"`
}
