package rest

import (
	"encoding/json"
	"errors"
	"net/http"

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
	updated, err := s.Settings.Update(func(c *conf.Settings) { applyEditable(c, req) })
	if err != nil {
		writeError(w, updateErrorStatus(err), err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// applyEditable copies the schedule and photo fields of req into c. The
// preview size is not editable through the API.
func applyEditable(c *conf.Settings, req conf.Settings) {
	c.SecondsBetweenCaptures = req.SecondsBetweenCaptures
	c.OffsetWithinHour = req.OffsetWithinHour
	c.Quality = req.Quality
	c.RotateBy = req.RotateBy
	c.ResolutionSetting = req.ResolutionSetting
	c.DebugEnabled = req.DebugEnabled
}

// updateErrorStatus is 400 for invalid settings and 500 for anything else.
func updateErrorStatus(err error) int {
	if errors.Is(err, conf.ErrInvalid) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}
