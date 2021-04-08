package tt

import (
	"encoding/json"
	"fmt"
	"tt/internal/util"
)

// taskProxy is the Task as stored in DB.
type taskProxy struct {
	ID          int64
	Description string
	Tags        []byte
	StartedAt   util.TimeAsTimestamp
	StoppedAt   util.NullTimeAsTimestamp
}

type scannable interface {
	Scan(...interface{}) error
}

func taskProxyFields() string {
	// Order must match fields in taskProxy.Scan
	return `"ID", "Description", "StartedAt", "StoppedAt", "Tags"`
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
