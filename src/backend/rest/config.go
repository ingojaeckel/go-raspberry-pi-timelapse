package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
)

const maxSettingsBytes = 64 << 10

func (s *Server) getConfiguration(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.Settings.Get())
}

func (s *Server) updateConfiguration(w http.ResponseWriter, r *http.Request) {
	var req conf.Settings
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxSettingsBytes)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.checkPhotoSize(r.Context(), req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	updated, err := s.Settings.Update(func(c *conf.Settings) { applyEditable(c, req) })
	if err != nil {
		writeError(w, updateErrorStatus(err), err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// checkPhotoSize rejects a new photo size the detected camera does not
// offer. A size saved earlier (e.g. before the camera was swapped) can be
// kept; without a detected camera any valid size is accepted.
func (s *Server) checkPhotoSize(ctx context.Context, req conf.Settings) error {
	cur := s.Settings.Get()
	if req.PhotoResolutionWidth == 0 && req.PhotoResolutionHeight == 0 ||
		req.PhotoResolutionWidth == cur.PhotoResolutionWidth && req.PhotoResolutionHeight == cur.PhotoResolutionHeight ||
		s.Sensor == nil {
		return nil
	}
	sensor, err := s.Sensor(ctx)
	if err != nil || sensor.Supports(req.PhotoResolutionWidth, req.PhotoResolutionHeight) {
		return nil
	}
	var sizes []string
	for _, r := range sensor.Resolutions() {
		sizes = append(sizes, r.String())
	}
	return fmt.Errorf("%w: the %s camera does not support %dx%d; it supports %s", conf.ErrInvalid, sensor.Model,
		req.PhotoResolutionWidth, req.PhotoResolutionHeight, strings.Join(sizes, ", "))
}

// applyEditable copies the schedule and photo fields of req into c. The
// preview size is not editable through the API; a request without a photo
// size keeps the saved one.
func applyEditable(c *conf.Settings, req conf.Settings) {
	c.SecondsBetweenCaptures = req.SecondsBetweenCaptures
	c.OffsetWithinHour = req.OffsetWithinHour
	c.Quality = req.Quality
	c.RotateBy = req.RotateBy
	if req.PhotoResolutionWidth != 0 || req.PhotoResolutionHeight != 0 {
		c.PhotoResolutionWidth, c.PhotoResolutionHeight = req.PhotoResolutionWidth, req.PhotoResolutionHeight
	}
	c.DebugEnabled = req.DebugEnabled
	c.StopWhenDiskFull = req.StopWhenDiskFull
}

// updateErrorStatus is 400 for invalid settings and 500 for anything else.
func updateErrorStatus(err error) int {
	if errors.Is(err, conf.ErrInvalid) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

// camera describes the detected camera and the photo sizes it offers.
func (s *Server) camera(w http.ResponseWriter, r *http.Request) {
	var res CameraResponse
	if s.Sensor == nil {
		res.Error = "camera detection is not available"
	} else if sensor, err := s.Sensor(r.Context()); err != nil {
		res.Error = err.Error()
	} else {
		res = CameraResponse{Model: sensor.Model, Width: sensor.Width, Height: sensor.Height, Resolutions: sensor.Resolutions()}
	}
	writeJSON(w, http.StatusOK, res)
}
