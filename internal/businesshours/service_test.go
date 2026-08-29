package businesshours

import (
	"testing"
	"time"
)

func TestIsOpen(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Kolkata")

	weekday := time.Date(2026, 3, 4, 10, 0, 0, 0, loc) // Wednesday 10:00 IST
	weekend := time.Date(2026, 3, 7, 10, 0, 0, 0, loc) // Saturday 10:00 IST
	afterHrs := time.Date(2026, 3, 4, 20, 0, 0, 0, loc) // Wed 20:00

	cases := []struct {
		name   string
		cfg    Config
		at     time.Time
		want   bool
	}{
		{"disabled always open", Config{Enabled: false}, weekday, true},
		{"weekday inside window", Config{Enabled: true, Days: []int{1, 2, 3, 4, 5}, OpenTime: "09:00", CloseTime: "18:00", Timezone: "Asia/Kolkata"}, weekday, true},
		{"weekday after hours", Config{Enabled: true, Days: []int{1, 2, 3, 4, 5}, OpenTime: "09:00", CloseTime: "18:00", Timezone: "Asia/Kolkata"}, afterHrs, false},
		{"weekend excluded", Config{Enabled: true, Days: []int{1, 2, 3, 4, 5}, OpenTime: "09:00", CloseTime: "18:00", Timezone: "Asia/Kolkata"}, weekend, false},
		{"empty days = every day", Config{Enabled: true, OpenTime: "09:00", CloseTime: "18:00"}, weekday, true},
		{"overnight window", Config{Enabled: true, Days: []int{1, 2, 3, 4, 5, 6, 7}, OpenTime: "22:00", CloseTime: "06:00"}, afterHrs, false},
		{"overnight window inside", Config{Enabled: true, Days: []int{1, 2, 3, 4, 5, 6, 7}, OpenTime: "22:00", CloseTime: "06:00", Timezone: "Asia/Kolkata"}, time.Date(2026, 3, 4, 23, 0, 0, 0, loc), true},
		{"malformed times fail open", Config{Enabled: true, OpenTime: "nope", CloseTime: "1800"}, weekday, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsOpen(tc.cfg, tc.at); got != tc.want {
				t.Errorf("IsOpen: got %v, want %v", got, tc.want)
			}
		})
	}
}
