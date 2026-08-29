// Package businesshours evaluates whether "now" (or a supplied time)
// falls inside an owner's configured business hours. Pure functions —
// no DB, no imports of settings, so the caller (main.go inbound hook)
// passes the config in.
package businesshours

import (
	"strconv"
	"strings"
	"time"
)

// Config is what the settings package stores; kept as a plain struct
// so we don't depend on settings.PrefsData (that would cause an
// import cycle with the inbound hook wiring).
type Config struct {
	Enabled   bool
	Days      []int  // 1=Mon .. 7=Sun; empty => all days
	OpenTime  string // "09:00"
	CloseTime string // "18:00"
	Timezone  string // "Asia/Kolkata"; blank => UTC
}

// IsOpen returns true when the given instant falls inside the config's
// day + time window. Nil-safe: an empty/disabled config returns true
// so callers who don't opt in are never blocked.
func IsOpen(cfg Config, now time.Time) bool {
	if !cfg.Enabled {
		return true
	}
	loc := loadLocation(cfg.Timezone)
	local := now.In(loc)

	if !dayAllowed(cfg.Days, local.Weekday()) {
		return false
	}

	openMin, ok1 := parseHM(cfg.OpenTime)
	closeMin, ok2 := parseHM(cfg.CloseTime)
	if !ok1 || !ok2 {
		// Misconfigured — fail open so we don't silently block traffic.
		return true
	}
	nowMin := local.Hour()*60 + local.Minute()
	// Support overnight windows (e.g. 22:00 → 06:00) too.
	if openMin <= closeMin {
		return nowMin >= openMin && nowMin < closeMin
	}
	return nowMin >= openMin || nowMin < closeMin
}

func loadLocation(tz string) *time.Location {
	if tz == "" {
		tz = "Asia/Kolkata"
	}
	if loc, err := time.LoadLocation(tz); err == nil {
		return loc
	}
	return time.UTC
}

// dayAllowed accepts an empty list as "every day" and otherwise
// checks weekday membership (Go's Weekday: Sunday=0, ISO uses 7).
func dayAllowed(days []int, wd time.Weekday) bool {
	if len(days) == 0 {
		return true
	}
	iso := int(wd)
	if iso == 0 {
		iso = 7 // Sunday
	}
	for _, d := range days {
		if d == iso {
			return true
		}
	}
	return false
}

func parseHM(s string) (int, bool) {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) != 2 {
		return 0, false
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, false
	}
	return h*60 + m, true
}
