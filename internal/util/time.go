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
