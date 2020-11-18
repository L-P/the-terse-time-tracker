package tt

import "time"

type Task struct {
	Description          string
	Tags                 []string
	StartedAt, StoppedAt time.Time
}

func NewTask(desc string, tags []string) Task {
	return Task{
		Description: desc,
		Tags:        tags,
		StartedAt:   time.Now(),
	}
}
