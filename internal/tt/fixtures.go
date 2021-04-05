//+build fixture

// nolint:gosec
package tt

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"
)

func (tt *TT) Fixture() error {
	words := loadTags()

	return tt.transaction(func(tx *sql.Tx) error {
		tasks, err := tt.GetTasks()
		if err != nil {
			return err
		}

		for _, v := range tasks {
			if err := deleteTask(tx, v.ID); err != nil {
				return nil
			}
		}

		var i int
		now := time.Now()
		min := now.Add(24 * 7 * -time.Hour)

		for now.After(min) {
			next := now.Add(time.Duration(rand.Intn(3*3600)) * -time.Second)
			task := Task{
				Description: fmt.Sprintf("%03d - dev task", i),
				StartedAt:   next,
				StoppedAt:   now,
				Tags:        randomPick(words, rand.Intn(4)),
			}

			if err := task.insert(tx); err != nil {
				return err
			}

			now = next
			if now.Hour() < 8 {
				now = now.Add(12 * -time.Hour)
			}

			i++
		}

		return nil
	})
}

func randomPick(words []string, count int) []string {
	ret := make([]string, 0, count)
	for i := 0; i < count; i++ {
		idx := rand.Intn(len(words))
		ret = append(ret, words[idx])
	}

	return ret
}

func loadTags() []string {
	max := 30
	ret := make([]string, 0, max)
	for i := 0; i < max; i++ {
		ret = append(ret, fmt.Sprintf("@tag-%02d", i))
	}

	return ret
}
