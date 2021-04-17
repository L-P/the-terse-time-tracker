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
	var offset int
	switch t.Weekday() {
	case time.Monday:
		offset = 0
	case time.Tuesday:
		offset = 1
	case time.Wednesday:
		offset = 2
	case time.Thursday:
		offset = 3
	case time.Friday:
		offset = 4
	case time.Saturday:
		offset = 5
	case time.Sunday:
		offset = 6
	}

	return GetStartOfDay(t).AddDate(0, 0, -offset)
}
