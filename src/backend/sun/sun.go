// Package sun tells day from night anywhere on Earth without a network: the
// sun's elevation follows from the time and the location (NOAA's solar
// position equations, accurate to about a minute for sunrise and sunset).
//
// src/frontend/src/sun.ts is a port; both are tested against
// src/frontend/src/sun-cases.json.
package sun

import (
	"math"
	"time"
)

// horizon is the sun's elevation at sunrise and sunset: its centre is below
// the horizon by its radius plus atmospheric refraction.
const horizon = -0.833

const (
	// step is how far apart the search for the next sunrise or sunset
	// samples the elevation; a day shorter than this (at the edge of the
	// polar night) may be missed.
	step = 10 * time.Minute
	// searchLimit covers the longest polar night (at the poles).
	searchLimit = 366 * 24 * time.Hour
)

// ValidLocation reports whether lat and lon are degrees on Earth: latitude
// -90 (south) to 90 (north), longitude -180 (west) to 180 (east).
func ValidLocation(lat, lon float64) bool {
	return lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}

// Up reports whether the sun is above the horizon at t at the location.
func Up(t time.Time, lat, lon float64) bool {
	return Elevation(t, lat, lon) > horizon
}

// NextSunrise is the first time after t when the sun rises, and false if it
// does not rise within a year. If the sun is up at t, it is the sunrise after
// the coming sunset.
func NextSunrise(t time.Time, lat, lon float64) (time.Time, bool) {
	return next(t, lat, lon, true)
}

// NextSunset is the first time after t when the sun sets, and false if it
// does not set within a year.
func NextSunset(t time.Time, lat, lon float64) (time.Time, bool) {
	return next(t, lat, lon, false)
}

// next finds the next change of Up to up: sampling every step, then
// bisecting to the second.
func next(t time.Time, lat, lon float64, up bool) (time.Time, bool) {
	before := t
	wasUp := Up(before, lat, lon)
	for elapsed := time.Duration(0); elapsed < searchLimit; elapsed += step {
		after := before.Add(step)
		isUp := Up(after, lat, lon)
		if isUp == up && wasUp != up {
			for after.Sub(before) > time.Second {
				mid := before.Add(after.Sub(before) / 2)
				if Up(mid, lat, lon) == up {
					after = mid
				} else {
					before = mid
				}
			}
			return after.Truncate(time.Second), true
		}
		before, wasUp = after, isUp
	}
	return time.Time{}, false
}

// Elevation is the sun's angle above the horizon in degrees at t, without
// refraction.
func Elevation(t time.Time, lat, lon float64) float64 {
	unix := float64(t.UnixNano()) / 1e9
	jd := unix/86400 + 2440587.5
	jc := (jd - 2451545) / 36525 // Julian centuries since J2000

	meanLong := math.Mod(280.46646+jc*(36000.76983+jc*0.0003032), 360)
	meanAnom := 357.52911 + jc*(35999.05029-0.0001537*jc)
	ecc := 0.016708634 - jc*(0.000042037+0.0000001267*jc)
	center := sin(meanAnom)*(1.914602-jc*(0.004817+0.000014*jc)) +
		sin(2*meanAnom)*(0.019993-0.000101*jc) + sin(3*meanAnom)*0.000289
	omega := 125.04 - 1934.136*jc
	appLong := meanLong + center - 0.00569 - 0.00478*sin(omega)
	meanObliq := 23 + (26+(21.448-jc*(46.815+jc*(0.00059-jc*0.001813)))/60)/60
	obliq := meanObliq + 0.00256*cos(omega)
	decl := deg(math.Asin(sin(obliq) * sin(appLong)))

	y := math.Pow(math.Tan(rad(obliq/2)), 2)
	eqTime := 4 * deg(y*sin(2*meanLong)-2*ecc*sin(meanAnom)+4*ecc*y*sin(meanAnom)*cos(2*meanLong)-
		0.5*y*y*sin(4*meanLong)-1.25*ecc*ecc*sin(2*meanAnom)) // minutes

	minutes := math.Mod(unix, 86400) / 60 // UTC minutes of the day
	trueSolar := math.Mod(minutes+eqTime+4*lon, 1440)
	if trueSolar < 0 {
		trueSolar += 1440
	}
	hourAngle := trueSolar/4 - 180
	cosZenith := sin(lat)*sin(decl) + cos(lat)*cos(decl)*cos(hourAngle)
	return 90 - deg(math.Acos(max(-1, min(1, cosZenith))))
}

func rad(d float64) float64 { return d * math.Pi / 180 }
func deg(r float64) float64 { return r * 180 / math.Pi }
func sin(d float64) float64 { return math.Sin(rad(d)) }
func cos(d float64) float64 { return math.Cos(rad(d)) }
