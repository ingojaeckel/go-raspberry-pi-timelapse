package timelapse

import "time"

// NextAligned returns the first capture time at or after now (truncated to
// the second) when captures happen at offset seconds past every local hour
// and every interval seconds after that. interval must divide 3600 and
// 0 <= offset < interval (see conf.Settings.Validate), so every hour has the
// same capture times.
//
// src/frontend/src/schedule.ts implements the same function; both are tested
// against src/frontend/src/schedule-cases.json.
func NextAligned(now time.Time, interval, offset int) time.Time {
	now = now.Truncate(time.Second)
	hour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
	base := hour.Add(time.Duration(offset) * time.Second)
	step := time.Duration(interval) * time.Second
	// Since offset < interval, base-step is before now, so k >= 0.
	k := now.Sub(base) / step
	next := base.Add(k * step)
	if next.Before(now) {
		next = next.Add(step)
	}
	return next
}
