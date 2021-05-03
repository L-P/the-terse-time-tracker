package tt

import (
	"database/sql"
	"errors"
	"time"
	"tt/internal/util"
)

type DailyReport struct {
	Start, End time.Time
	Total      time.Duration
}

type WeeklyReport struct {
	Start, End time.Time
	Daily      [7]DailyReport // index is time.Weekday
	Total      time.Duration
	Overtime   time.Duration
}

func (tt *TT) GetWeeklyReport(t time.Time) (WeeklyReport, error) {
	report := WeeklyReport{Start: util.GetStartOfWeek(t)}
	report.End = report.Start.AddDate(0, 0, len(report.Daily))
	weeklyHours := tt.GetConfig().WeeklyHours

	if err := tt.transaction(func(tx *sql.Tx) error {
		for i := 0; i < len(report.Daily); i++ {
			dayStart := report.Start.AddDate(0, 0, i)
			dayEnd := dayStart.AddDate(0, 0, 1)

			agg, err := tt.getAggregatedTime(tx, dayStart, dayEnd)
			if err != nil {
				return err
			}

			start, end, err := tt.getWorkedHoursBounds(tx, dayStart)
			if err == nil {
				report.Overtime += agg - (weeklyHours / 5)
				report.Total += agg
				report.Daily[dayStart.Weekday()] = DailyReport{
					Start: start,
					End:   end,
					Total: agg,
				}
			} else if !errors.Is(err, ErrNoTasks) {
				return err
			}
		}

		return nil
	}); err != nil {
		return WeeklyReport{}, err
	}

	return report, nil
}
