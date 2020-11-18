package tt

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"tt/internal/util"
)

type Task struct {
	Description string
	Tags        []string
	StartedAt   time.Time
	StoppedAt   time.Time // will be a zero time for the task in progress
}

// taskProxy is the Task as stored in DB.
type taskProxy struct {
	Description string
	Tags        []byte
	StartedAt   util.TimeAsTimestamp
	StoppedAt   util.NullTimeAsTimestamp
}

func (t taskProxy) fields() string {
	// Order must match fields in scan
	return `"Description", "StartedAt", "StoppedAt", "Tags"`
}

func (t *taskProxy) scan(row *sql.Row) error {
	return row.Scan(
		&t.Description,
		&t.StartedAt,
		&t.StoppedAt,
		&t.Tags,
	)
}

func newProxyFromTask(t Task) (taskProxy, error) {
	tags := []byte("[]")
	if t.Tags != nil {
		var err error
		tags, err = json.Marshal(t.Tags)
		if err != nil {
			return taskProxy{}, ErrRuntime(fmt.Sprintf("unable to encode tags: %s", err))
		}
	}

	return taskProxy{
		Description: t.Description,
		StartedAt:   util.TimeAsTimestamp(t.StartedAt),
		StoppedAt:   util.NewNullTimeAsTimestamp(t.StoppedAt),
		Tags:        tags,
	}, nil
}

func (t taskProxy) Task() (*Task, error) {
	ret := Task{
		Description: t.Description,
		StartedAt:   t.StartedAt.Time(),
		StoppedAt:   t.StoppedAt.Time.Time(),
	}

	if err := json.Unmarshal(t.Tags, &ret.Tags); err != nil {
		return nil, ErrRuntime(fmt.Sprintf("unable to parse tags array: %s", err))
	}

	return &ret, nil
}

func NewTask(desc string, tags []string) *Task {
	return &Task{
		Description: desc,
		Tags:        tags,
		StartedAt:   time.Now(),
	}
}

func (t *Task) update(tx *sql.Tx) error {
	proxy, err := newProxyFromTask(*t)
	if err != nil {
		return err
	}

	return exec(
		tx,
		`UPDATE Task
        SET Description = ?,
            StartedAt = ?,
            StoppedAt = ?,
            Tags = ?
        WHERE StartedAt = ?`,
		proxy.Description,
		proxy.StartedAt,
		proxy.StoppedAt,
		proxy.Tags,
		proxy.StartedAt,
	)
}

func (t *Task) insert(tx *sql.Tx) error {
	proxy, err := newProxyFromTask(*t)
	if err != nil {
		return err
	}

	return exec(
		tx,
		`INSERT INTO Task (Description, StartedAt, StoppedAt, Tags)
        VALUES (?, ?, ?, ?)`,
		proxy.Description,
		proxy.StartedAt,
		proxy.StoppedAt,
		proxy.Tags,
	)
}

func (t *Task) Duration() time.Duration {
	if t.StoppedAt.IsZero() {
		return time.Since(t.StartedAt)
	}

	return t.StoppedAt.Sub(t.StartedAt)
}

func getCurrentTask(tx *sql.Tx) (*Task, error) {
	var proxy taskProxy

	query := fmt.Sprintf(
		`SELECT %s FROM Task WHERE StoppedAt IS NULL LIMIT 1`,
		proxy.fields(),
	)
	if err := proxy.scan(tx.QueryRow(query)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, ErrBadQuery{err, query, nil}
	}

	return proxy.Task()
}

func stopCurrentTask(tx *sql.Tx) error {
	return exec(
		tx,
		`UPDATE Task SET StoppedAt = ? WHERE StoppedAt IS NULL`,
		util.NewNullTimeAsTimestamp(time.Now()),
	)
}
