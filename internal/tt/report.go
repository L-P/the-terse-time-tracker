package tt

import (
	"database/sql"
	"fmt"
	"time"
	"tt/internal/util"
)

// Report entry for a time slice. By day here, by week in CLI output.
type ReportEntry struct {
	Day                time.Time
	WorkStart, WorkEnd time.Time

	// Raw total durations disregarding any rules
	WorkDuration   time.Duration
	OnCallDuration time.Duration

	// {{{ Rule-computed.
	// Either paid or given back as 1:1 time off.
	Overtime time.Duration

	// Non-financial compensation, time off only.
	InLieu time.Duration

	// Time taken off work.
	Taken time.Duration
	// }}}
}

// Add sums the durations, it does not changes the dates.
func (e *ReportEntry) Add(v ReportEntry) {
	if v.Day.Before(e.Day) || e.Day.IsZero() {
		e.Day = v.Day
	}
	if v.WorkStart.Before(e.WorkStart) || e.WorkStart.IsZero() {
		e.WorkStart = v.WorkStart
	}
	if v.WorkEnd.After(e.WorkEnd) || e.WorkEnd.IsZero() {
		e.WorkEnd = v.WorkEnd
	}

	e.WorkDuration += v.WorkDuration
	e.OnCallDuration += v.OnCallDuration
	e.Overtime += v.Overtime
	e.Taken += v.Taken
	e.InLieu += v.InLieu
}

func (e *ReportEntry) Reset() {
	*e = ReportEntry{}
}

const (
	tagOnCall = "@oncall"
	tagOff    = "@off" // Neither counts as overtime nor time taken.
)

//nolint: cyclop
func newReportEntry(day time.Time, tasks []Task, rules rulesSnapshot) ReportEntry {
	e := ReportEntry{Day: day}
	// not counted as work time but doesn't create an overtime deficit, eg. paid half-day leave
	var offTime time.Duration

	for _, task := range tasks {
		switch {
		case task.HasTag(tagOff):
			offTime += task.Duration()
		case task.HasTag(tagOnCall):
			e.OnCallDuration += task.Duration()
		default:
			e.WorkDuration += task.Duration()
		}

		if e.WorkStart.IsZero() || task.StartedAt.Before(e.WorkStart) {
			e.WorkStart = task.StartedAt
		}
		if e.WorkEnd.IsZero() || task.StoppedAt.After(e.WorkEnd) {
			e.WorkEnd = task.StoppedAt
		}
		if task.StoppedAt.IsZero() {
			task.StoppedAt = time.Now()
		}
	}

	if isOffDay(day.Weekday()) {
		e.InLieu += time.Duration(float64(e.WorkDuration) * rules[ruleHolidayFactor])
	} else if e.WorkDuration > 0 { // non-worked weekdays are considered off
		dailyWork := time.Hour * time.Duration((rules[ruleWeeklyHours])) / 5
		delta := (e.WorkDuration + offTime) - dailyWork
		if delta > 0 {
			e.Overtime += delta
		} else if delta < 0 {
			e.Taken += -delta
		}
	}

	e.InLieu += time.Duration(float64(e.OnCallDuration) * rules[ruleOnCallFactor])

	return e
}

func isOffDay(wd time.Weekday) bool {
	return wd == time.Saturday || wd == time.Sunday
}

type Report struct {
	Daily []ReportEntry

	Accumulated ReportEntry
}

func (r *Report) Add(v ReportEntry) {
	r.Daily = append(r.Daily, v)
	r.Accumulated.Add(v)
}

func (r *Report) computeAggregates() {
	for _, v := range r.Daily {
		r.Accumulated.Add(v)
	}
}

func (tt *TT) GetReport() (Report, error) {
	var report Report
	firstTask, err := tt.GetFirstTask()
	if err != nil {
		return Report{}, fmt.Errorf("unable to fetch first task: %w", err)
	}

	err = tt.transaction(func(tx *sql.Tx) (err error) {
		report, err = tt.getReportTx(tx, firstTask.StartedAt, time.Now())
		return err
	})

	return report, err
}

func (tt *TT) getReportTx(tx *sql.Tx, start, end time.Time) (Report, error) {
	var (
		report   Report
		timeline = staticRulesTimeline()
		day      = util.GetStartOfDay(start)
	)

	for day.Before(end) {
		var (
			rules   = timeline.forDay(day)
			nextDay = day.AddDate(0, 0, 1)
		)

		tasks, err := getTasksInRange(tx, day, nextDay)
		if err != nil {
			return Report{}, fmt.Errorf("unable to fetch tasks for range %s-%s: %w", day, nextDay, err)
		}

		report.Daily = append(report.Daily, newReportEntry(
			day,
			clampTasks(tasks, day, nextDay),
			rules,
		))

		day = nextDay
	}

	report.computeAggregates()

	return report, nil
}

// Remove tasks or cut them if they don't fit in the given start/end.
func clampTasks(tasks []Task, start, end time.Time) []Task {
	ret := make([]Task, 0, len(tasks))

	for _, task := range tasks {
		if task.StartedAt.After(end) {
			continue
		}

		if task.StartedAt.Before(start) {
			if !task.StoppedAt.IsZero() && task.StoppedAt.Before(start) {
				continue
			}
		}

		if task.StoppedAt.After(end) {
			task.StoppedAt = end
		}

		ret = append(ret, task)
	}

	return ret
}

func (tt *TT) GetTagReport() (map[string]time.Duration, error) {
	ret := map[string]time.Duration{}
	tasks, err := tt.GetTasks()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch tasks: %w", err)
	}

	for _, task := range tasks {
		for _, tag := range task.Tags {
			ret[tag] += task.Duration()
		}
	}

	return ret, nil
}
