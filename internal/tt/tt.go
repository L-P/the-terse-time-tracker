package tt

import (
	"database/sql"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQL driver

	"tt/internal/util"
)

type TT struct {
	db     *sql.DB
	config Config
}

func New(dsn string) (*TT, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, ErrIO{"unable to open database", dsn, err}
	}

	if err := migrate(db); err != nil {
		return nil, ErrIO{"unable to migrate DB", dsn, err}
	}

	config, err := loadConfig(db)
	if err != nil {
		return nil, ErrRuntime(fmt.Sprintf("unable to load config: %s", err))
	}

	return &TT{
		db:     db,
		config: config,
	}, nil
}

func (tt *TT) Close() error {
	return tt.db.Close()
}

// Start can start a new task and stop the current one, update a current task,
// or do nothing.
// It returns the current task (if any) and next task (always). If the task is
// the same, the data might differ.
func (tt *TT) Start(raw string) (*Task, *Task, error) {
	desc, tags := ParseRawDesc(raw)
	var current, next *Task

	err := tt.transaction(func(tx *sql.Tx) (err error) {
		current, err = getCurrentTask(tx)
		if err != nil {
			return err
		}

		// Same task, maybe update tags.
		if current != nil && (current.Description == desc || desc == "") {
			next = &Task{}
			*next = *current

			if reflect.DeepEqual(current.Tags, tags) {
				return ErrContinue
			}

			// Tags differ, update current task with new tags.
			next.Tags = tags
			if err := next.update(tx); err != nil {
				return err
			}
			return nil
		}

		if current != nil {
			if err := stopTask(tx, current.ID); err != nil {
				return err
			}
		}

		if len(tags) == 0 && current != nil {
			tags = current.Tags
		}

		if desc == "" {
			return ErrInvalidTaskDesc
		}

		next = NewTask(desc, tags)
		if err := next.insert(tx); err != nil {
			return err
		}

		return nil
	})

	return current, next, err
}

// Stop stops the current task if any.
func (tt *TT) Stop() (*Task, error) {
	var cur *Task

	if err := tt.transaction(func(tx *sql.Tx) (err error) {
		cur, err = getCurrentTask(tx)
		if err != nil {
			return err
		}

		if cur == nil {
			return ErrNoCurrentTask
		}

		cur.StoppedAt = time.Now()
		if err := stopTask(tx, cur.ID); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return cur, nil
}

// ParseRawDesc splits the description and tags from user input.
// tags can only be provided at the end of the string.
func ParseRawDesc(raw string) (string, []string) {
	reverse := func(v []string) {
		for i, j := 0, len(v)-1; i < j; i, j = i+1, j-1 {
			v[i], v[j] = v[j], v[i]
		}
	}

	var (
		desc string
		tags []string
	)

	parts := strings.Split(raw, " ")
	reverse(parts)
	for k := range parts {
		if len(parts[k]) == 0 {
			continue
		}

		if parts[k][0] == '@' {
			tags = append(tags, parts[k])
			continue
		}

		parts = parts[k:]
		reverse(parts)
		desc = strings.Join(parts, " ")
		break
	}

	sort.Strings(tags)
	return desc, tags
}

func (tt *TT) GetTasks() ([]Task, error) {
	var ret []Task

	if err := tt.transaction(func(tx *sql.Tx) (err error) {
		ret, err = getAllTasks(tx)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return ret, nil
}

func (tt *TT) DeleteTask(taskID int64) error {
	return tt.transaction(func(tx *sql.Tx) error {
		return deleteTask(tx, taskID)
	})
}

func (tt *TT) UpdateTask(t Task) error {
	return tt.transaction(t.update)
}

func (tt *TT) CurrentTask() (*Task, error) {
	var cur *Task

	if err := tt.transaction(func(tx *sql.Tx) (err error) {
		cur, err = getCurrentTask(tx)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return cur, nil
}

func (tt *TT) GetConfig() Config {
	return tt.config
}

func (tt *TT) SetConfig(config Config) error {
	if err := config.Validate(); err != nil {
		return err
	}

	tt.config = config
	return tt.transaction(tt.config.save)
}

func (tt *TT) GetDurationLeft() (time.Duration, time.Duration, error) {
	if tt.config.WeeklyHours <= 0 {
		return 0, 0, ErrNotConfigured
	}

	now := time.Now()

	dayStart := util.GetStartOfDay(now)
	dayEnd := dayStart.AddDate(0, 0, 1)
	daily, err := tt.getAggregatedTime(dayStart, dayEnd)
	if err != nil {
		return 0, 0, err
	}

	weekStart := util.GetStartOfWeek(now)
	weekEnd := weekStart.AddDate(0, 0, 7)
	weekly, err := tt.getAggregatedTime(weekStart, weekEnd)
	if err != nil {
		return 0, 0, err
	}

	return (tt.config.WeeklyHours / 5) - daily, tt.config.WeeklyHours - weekly, nil
}

func (tt *TT) getAggregatedTime(start, end time.Time) (time.Duration, error) {
	var tasks []Task
	if err := tt.transaction(func(tx *sql.Tx) (err error) {
		tasks, err = getTasksInRange(tx, start, end)
		return err
	}); err != nil {
		return 0, err
	}

	var acc time.Duration
	for _, task := range tasks {
		clampedStart := task.StartedAt
		if clampedStart.Before(start) {
			clampedStart = start
		}

		clampedEnd := task.StoppedAt
		if clampedEnd.IsZero() {
			clampedEnd = time.Now()
		} else if clampedEnd.After(end) {
			clampedEnd = end
		}

		acc += clampedEnd.Sub(clampedStart)
	}

	return acc, nil
}
