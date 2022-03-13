package tt

import (
	"time"
	"tt/internal/util"
)

type rule int

const (
	ruleWeeklyHours   = iota
	ruleHolidayFactor // worked holiday = (factor * worktime) of in lieu
	ruleOnCallFactor  // on call worktime = (factor * worktime) of in lieu
)

type ruleEntry struct {
	start time.Time  // can be a zero value to indicate the first applicable rule
	end   *time.Time // can be nil for an ongoing rule
	rule  rule
	value float64
}

// Sorted by start time.
type rulesTimeline []ruleEntry

// Applicable rules at a specific time.
type rulesSnapshot map[rule]float64

// forDay returns the applicable rules for a given day.
// TODO: this is called for every day, should be memoized away.
func (timeline rulesTimeline) forDay(t time.Time) rulesSnapshot {
	day := util.GetStartOfDay(t)

	ret := make(map[rule]float64, 3)
	for _, v := range timeline {
		if v.start.After(day) {
			// timeline is ordered, nothing for us past this point
			break
		}

		if v.end == nil || day.Before(*v.end) {
			ret[v.rule] = v.value
		}
	}

	return ret
}

// TODO in DB, configurable at runtime.
func staticRulesTimeline() rulesTimeline {
	return rulesTimeline{
		ruleEntry{rule: ruleWeeklyHours, value: 39},
		ruleEntry{rule: ruleHolidayFactor, value: 2},
		ruleEntry{rule: ruleOnCallFactor, value: 1.5},
	}
}
