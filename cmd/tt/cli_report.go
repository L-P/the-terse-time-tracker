// nolint:wrapcheck
package main

import (
	"fmt"
	"strings"
	"time"
	"tt/internal/tt"
	"tt/internal/util"
)

const dateFormat = "2006-01-02"

func report(app *tt.TT, out output) error {
	firstTask, err := app.GetFirstTask()
	if err != nil {
		return err
	}

	var (
		max         = util.GetStartOfWeek(time.Now()).AddDate(0, 0, 7)
		min         = util.GetStartOfWeek(firstTask.StartedAt)
		weeklyHours = app.GetConfig().WeeklyHours
		delta       time.Duration
	)

	for cur := min; cur.Before(max); cur = cur.AddDate(0, 0, 7) {
		r, err := app.GetWeeklyReport(cur)
		if err != nil {
			return err
		}

		delta += r.Total - weeklyHours
		printWeeklyReport(app, out, r, delta)

		fmt.Fprint(out.w, "\n")
	}

	return nil
}

// nolint:funlen // no need to split/abstract too much over this, embrace the
// spaghetti (uncooked).
// Example output:
// Week #17 from 2021-04-26 to 2021-05-02
//  00:00    00:00    00:00    00:00    00:00
//  00:00    00:00    00:00    00:00    00:00
//  07h48m   07h48m   07h48m   07h48m   07h48m   39h00m
// +01h20m  +00h00m  +00h00m  +00h00m  +00h00m  +00h00m (+00h00m)
//   Mon.     Tue.     Wed.     Thu.     Fri.    Total
func printWeeklyReport(app *tt.TT, out output, r tt.WeeklyReport, runningDelta time.Duration) {
	var (
		b           strings.Builder
		showDays    = 5 // show only worked days
		startOffset = 1 // start on monday (offset on time.Weekday)
		weeklyHours = app.GetConfig().WeeklyHours
		blank       = "         "
		_, isoWeek  = r.Start.ISOWeek()
	)

	fmt.Fprintf(&b,
		"Week #%d from %s to %s\n",
		isoWeek,
		r.Start.Format(dateFormat),
		r.End.AddDate(0, 0, -1).Format(dateFormat),
	)

	if r.Total == 0 {
		b.WriteString("No work done this week.\n")
		fmt.Fprint(out.w, b.String())
		return
	}

	dailyReport := func(i int) (tt.DailyReport, bool) {
		// start on monday
		dr := r.Daily[(i+startOffset)%len(r.Daily)]
		return dr, dr.Total != 0
	}

	b.Grow(300) // a little over expected exact length

	// Start hour
	for i := 0; i < showDays; i++ {
		dr, ok := dailyReport(i)
		if !ok {
			b.WriteString(blank)
			continue
		}

		fmt.Fprintf(&b, "  %s  ", dr.Start.Format("15:04"))
	}
	b.WriteRune('\n')

	// End hour
	for i := 0; i < showDays; i++ {
		dr, ok := dailyReport(i)
		if !ok {
			b.WriteString(blank)
			continue
		}

		fmt.Fprintf(&b, "  %s  ", dr.End.Format("15:04"))
	}
	b.WriteRune('\n')

	// Duration
	for i := 0; i < showDays; i++ {
		dr, ok := dailyReport(i)
		if !ok {
			b.WriteString(blank)
			continue
		}

		fmt.Fprintf(&b, "  %s ", util.FormatFixedDuration(dr.Total))
	}
	b.WriteRune('\n')

	if weeklyHours > 0 {
		var totalOvertime time.Duration
		// Daily over/under time
		for i := 0; i < showDays; i++ {
			dr, ok := dailyReport(i)
			if !ok {
				b.WriteString(blank)
				continue
			}

			delta := dr.Total - (weeklyHours / 5)
			totalOvertime += delta
			fmt.Fprintf(&b, " %s ", util.FormatSignedFixedDuration(delta))
		}

		fmt.Fprintf(&b, " %s ", util.FormatSignedFixedDuration(totalOvertime))
		fmt.Fprintf(&b, " (%s)", util.FormatSignedFixedDuration(runningDelta))

		b.WriteRune('\n')
	}

	// HARDCODED, will shift if startOffset changes.
	b.WriteString("   Mon.     Tue.     Wed.     Thu.     Fri.    Total\n")
	fmt.Fprint(out.w, b.String())
}
