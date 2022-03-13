package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"tt/internal/tt"
	"tt/internal/util"
)

const dateFormat = "2006-01-02"

func report(app *tt.TT, out output) error {
	report, err := app.GetReport()
	if err != nil {
		return fmt.Errorf("unable to generate report: %w", err)
	}

	if out.json {
		enc := json.NewEncoder(out.w)
		return enc.Encode(struct { // nolint:wrapcheck
			WorkDuration, OnCallDuration time.Duration
			Overtime, InLieu, Taken      time.Duration
		}{
			report.Accumulated.WorkDuration,
			report.Accumulated.OnCallDuration,
			report.Accumulated.Overtime,
			report.Accumulated.InLieu,
			report.Accumulated.Taken,
		})
	}

	var (
		total       tt.ReportEntry
		weekly      tt.Report
		_, lastWeek = report.Accumulated.Day.ISOWeek()
	)

	for _, daily := range report.Daily {
		if _, curWeek := daily.Day.ISOWeek(); curWeek != lastWeek {
			lastWeek = curWeek
			printWeeklyReport(out, weekly, total)
			weekly = tt.Report{}
		}

		weekly.Add(daily)
		total.Add(daily)
	}

	printWeeklyReport(out, weekly, total)

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
func printWeeklyReport(out output, r tt.Report, total tt.ReportEntry) {
	var (
		b          strings.Builder
		showDays   = 7
		blank      = "         "
		_, isoWeek = r.Accumulated.Day.ISOWeek()
	)

	fmt.Fprintf(&b,
		"Week #%d from %s to %s\n",
		isoWeek,
		r.Accumulated.Day.Format(dateFormat),
		r.Accumulated.Day.AddDate(0, 0, 6).Format(dateFormat),
	)

	if r.Accumulated.WorkDuration == 0 {
		b.WriteString("No work done this week.\n\n")
		fmt.Fprint(out.w, b.String())
		return
	}

	var dailyReport [7]tt.ReportEntry
	for _, v := range r.Daily {
		dailyReport[(v.Day.Weekday()+6)%7] = v
	}

	b.Grow(300) // a little over expected exact length

	// Start hour
	for i := 0; i < showDays; i++ {
		dr := dailyReport[i]
		if dr.WorkDuration <= 0 {
			b.WriteString(blank)
			continue
		}

		fmt.Fprintf(&b, "  %s  ", dr.WorkStart.Format("15:04"))
	}
	b.WriteRune('\n')

	// End hour
	for i := 0; i < showDays; i++ {
		dr := dailyReport[i]
		if dr.WorkDuration <= 0 {
			b.WriteString(blank)
			continue
		}

		fmt.Fprintf(&b, "  %s  ", dr.WorkEnd.Format("15:04"))
	}
	b.WriteRune('\n')

	// Duration
	for i := 0; i < showDays; i++ {
		dr := dailyReport[i]
		if dr.WorkDuration <= 0 {
			b.WriteString(blank)
			continue
		}

		fmt.Fprintf(&b, "  %s ", util.FormatFixedDuration(dr.WorkDuration))
	}
	fmt.Fprintf(&b, " %7s ", strings.TrimSuffix(r.Accumulated.WorkDuration.Truncate(time.Minute).String(), "0s"))
	fmt.Fprintf(&b, " (%7s)", util.FormatDuration(total.WorkDuration))
	b.WriteRune('\n')

	// Daily over/under time
	for i := 0; i < showDays; i++ {
		dr := dailyReport[i]
		over := dr.Overtime + dr.InLieu
		switch {
		case over > 0:
			fmt.Fprintf(&b, " %s ", util.FormatSignedFixedDuration(over))
		case dr.Taken > 0:
			fmt.Fprintf(&b, " %s ", util.FormatSignedFixedDuration(-dr.Taken))
		default:
			b.WriteString(blank)
			continue
		}
	}
	weeklyDeltaSum := (r.Accumulated.Overtime + r.Accumulated.InLieu) - r.Accumulated.Taken
	totalDeltaSum := (total.Overtime + total.InLieu) - total.Taken
	fmt.Fprintf(&b, " %7s ", util.FormatSignedFixedDuration(weeklyDeltaSum))
	fmt.Fprintf(&b, " (%7s)", util.FormatSignedFixedDuration(totalDeltaSum))
	b.WriteRune('\n')

	b.WriteString("   Mon.     Tue.     Wed.     Thu.     Fri.     Sat.     Sun.     Total\n\n")

	fmt.Fprint(out.w, b.String())
}
