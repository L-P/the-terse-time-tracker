package util

import (
	"fmt"
	"strings"
	"time"
)

// FormatDuration prettifies a duration by removing useless units.
// eg. 1h20m0s -> 1h20m
// It does not round/truncate the duration, it only works on the string.
// TODO: This should be localized, also TODO: this was ported from an
// application with different requirements and another time scale.
func FormatDuration(d time.Duration) string {
	var prefix string
	if d > (24 * time.Hour) {
		prefix = fmt.Sprintf("%dd", d/(24*time.Hour))
		// Don't need minutes if its in more than a day
		d = (d % (24 * time.Hour)).Truncate(time.Hour)
	}

	ret := d.Truncate(time.Second).String()
	if strings.HasSuffix(ret, "m0s") {
		ret = strings.TrimSuffix(ret, "0s")
	}
	if strings.HasSuffix(ret, "h0m") {
		ret = strings.TrimSuffix(ret, "0m")
	}

	return prefix + ret
}

// Returns a duration in a fixed-width format.
func FormatFixedDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}

	s := int(d.Seconds())
	h := s / 3600
	m := s % 3600 / 60

	// TODO translation
	return fmt.Sprintf("%02dh%02dm", h, m)
}

func FormatSignedFixedDuration(d time.Duration) string {
	sign := '+'
	if d < 0 {
		sign = '-'
	}

	return fmt.Sprintf("%c%s", sign, FormatFixedDuration(d))
}

// GetStartOfDay does as it says in a tz-aware timeframe.
func GetStartOfDay(t time.Time) time.Time {
	return time.Date(
		t.Year(), t.Month(), t.Day(),
		0, 0, 0,
		0,
		t.Location(),
	)
}

func GetStartOfWeek(t time.Time) time.Time {
	return GetStartOfDay(t).AddDate(0, 0, -WeekdayOffset(t.Weekday()))
}

// WeekdayOffset returns the offset of a given day inside its week, week that
// starts on Monday everywhere in the world including inside the international
// date standard but not in the Go codebase.
func WeekdayOffset(wd time.Weekday) int {
	return (int(wd) + 6) % 7
}
