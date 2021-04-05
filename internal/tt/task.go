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
	ID          int64
	Description string
	Tags        []string
	StartedAt   time.Time
	StoppedAt   time.Time // will be a zero time for the task in progress
}

// taskProxy is the Task as stored in DB.
type taskProxy struct {
	ID          int64
	Description string
	Tags        []byte
	StartedAt   util.TimeAsTimestamp
	StoppedAt   util.NullTimeAsTimestamp
}

func taskProxyFields() string {
	// Order must match fields in taskProxy.Scan
	return `"ID", "Description", "StartedAt", "StoppedAt", "Tags"`
}

type scannable interface {
	Scan(...interface{}) error
}

func (t *taskProxy) scan(s scannable) error {
	return s.Scan(
		&t.ID,
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
		ID:          t.ID,
		Description: t.Description,
		StartedAt:   util.TimeAsTimestamp(t.StartedAt),
		StoppedAt:   util.NewNullTimeAsTimestamp(t.StoppedAt),
		Tags:        tags,
	}, nil
}

func (t taskProxy) Task() (*Task, error) {
	ret := Task{
		ID:          t.ID,
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
	if t.ID == 0 {
		return ErrInvalidTaskID
	}

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
        WHERE ID = ?`,
		proxy.Description,
		proxy.StartedAt,
		proxy.StoppedAt,
		proxy.Tags,
		t.ID,
	)
}

func (t *Task) insert(tx *sql.Tx) error {
	proxy, err := newProxyFromTask(*t)
	if err != nil {
		return err
	}

	t.ID, err = execWithLastID(
		tx,
		`INSERT INTO Task (Description, StartedAt, StoppedAt, Tags)
        VALUES (?, ?, ?, ?)`,
		proxy.Description,
		proxy.StartedAt,
		proxy.StoppedAt,
		proxy.Tags,
	)

	return err
}

func (t *Task) Duration() time.Duration {
	if t.StoppedAt.IsZero() {
		return time.Since(t.StartedAt)
	}

	return t.StoppedAt.Sub(t.StartedAt)
}

func getAllTasks(tx *sql.Tx) ([]Task, error) {
	var ret []Task

	query := fmt.Sprintf( // nolint:gosec
		`SELECT %s FROM Task ORDER BY StartedAt DESC`,
		taskProxyFields(),
	)
	rows, err := tx.Query(query)
	if err := err; err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, ErrBadQuery{err, query, nil}
	}
	defer rows.Close()

	for rows.Next() {
		var proxy taskProxy
		if err := proxy.scan(rows); err != nil {
			return nil, ErrBadQuery{err, query, nil}
		}

		task, err := proxy.Task()
		if err != nil {
			return nil, err
		}

		ret = append(ret, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, ErrBadQuery{err, query, nil}
	}

	return ret, nil
}

func getCurrentTask(tx *sql.Tx) (*Task, error) {
	var proxy taskProxy

	query := fmt.Sprintf(
		`SELECT %s FROM Task WHERE StoppedAt IS NULL LIMIT 1`,
		taskProxyFields(),
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

func deleteTask(tx *sql.Tx, id int64) error {
	if id == 0 {
		return ErrInvalidTaskID
	}

	return exec(tx, `DELETE FROM Task WHERE ID = ?`, id)
}
