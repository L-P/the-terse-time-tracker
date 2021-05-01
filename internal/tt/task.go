package tt

import (
	"database/sql"
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

func (t *Task) IsStopped() bool {
	return !t.StoppedAt.IsZero()
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
	return queryTasks(tx, fmt.Sprintf(
		`SELECT %s FROM Task ORDER BY StartedAt DESC`,
		taskProxyFields(),
	))
}

func (tt *TT) GetFirstTask() (Task, error) {
	var tasks []Task

	if err := tt.transaction(func(tx *sql.Tx) (err error) {
		tasks, err = queryTasks(tx, fmt.Sprintf(
			`SELECT %s FROM Task ORDER BY StartedAt ASC LIMIT 1`,
			taskProxyFields(),
		))
		return err
	}); err != nil {
		return Task{}, err
	}

	if len(tasks) == 0 {
		return Task{}, ErrNoTasks
	}

	return tasks[0], nil
}

func getTasksInRange(tx *sql.Tx, startTime, endTime time.Time) ([]Task, error) {
	start, end := startTime.Unix(), endTime.Unix()

	query := fmt.Sprintf(
		`SELECT %s FROM Task
        WHERE (StartedAt >= ? AND StartedAt < ?)
           OR (StoppedAt > ? AND StoppedAt < ?) OR StoppedAt IS NULL
        ORDER BY StartedAt DESC`,
		taskProxyFields(),
	)

	return queryTasks(tx, query, start, end, start, end)
}

func queryTasks(tx *sql.Tx, query string, params ...interface{}) ([]Task, error) {
	rows, err := tx.Query(query, params...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, ErrBadQuery{err, query, nil}
	}
	defer rows.Close()

	var ret []Task
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

func stopTask(tx *sql.Tx, id int64) error {
	return exec(
		tx,
		`UPDATE Task SET StoppedAt = ? WHERE ID = ?`,
		util.NewNullTimeAsTimestamp(time.Now()),
		id,
	)
}

func deleteTask(tx *sql.Tx, id int64) error {
	if id == 0 {
		return ErrInvalidTaskID
	}

	return exec(tx, `DELETE FROM Task WHERE ID = ?`, id)
}
