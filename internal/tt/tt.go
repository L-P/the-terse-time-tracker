package tt

import (
	"database/sql"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQL driver
)

type TT struct {
	db *sql.DB
}

func New(dsn string) (*TT, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, ErrIO{"unable to open database", dsn, err}
	}

	if err := migrate(db); err != nil {
		return nil, ErrIO{"unable to migrate DB", dsn, err}
	}

	return &TT{db: db}, nil
}

func (tt *TT) Close() error {
	return tt.db.Close()
}

// Start can start a new task and stop the current one, update a current task,
// or do nothing.
// It returns the created task if any, and the updated (current) task if any.
func (tt *TT) Start(raw string) (*Task, *Task, error) {
	desc, tags := parseRawDesc(raw)
	var cur, created *Task

	if err := tt.transaction(func(tx *sql.Tx) (err error) {
		cur, err = getCurrentTask(tx)
		if err != nil {
			return err
		}

		if cur != nil && cur.Description == desc {
			if reflect.DeepEqual(cur.Tags, tags) {
				cur = nil
				return nil
			}

			// Tags differ, update current task with new tags.
			cur.Tags = tags
			if err := cur.update(tx); err != nil {
				return err
			}
			return nil
		}

		if err := stopCurrentTask(tx); err != nil {
			return err
		}

		if tags == nil && cur != nil {
			tags = cur.Tags
		}
		created = NewTask(desc, tags)
		if err := created.insert(tx); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	return created, cur, nil
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
		if err := stopCurrentTask(tx); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return cur, nil
}

// parseRawDesc splits the description and tags from user input.
// tags can only be provided at the end of the string.
func parseRawDesc(raw string) (string, []string) {
	reverse := func(v []string) {
		sort.Slice(v, func(i, j int) bool { return true })
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

func migrate(db *sql.DB) (err error) {
	cur := getVersion(db)
	if cur < 0 {
		return ErrDatabase(fmt.Sprintf("invalid database version: %d", cur))
	}

	switch cur {
	case 0:
		err = doInitialMigration(db)
		fallthrough
	case 1:
		break // current version
	default:
		return ErrDatabase(fmt.Sprintf("database is at version %d which is not compatible with your local tt version", cur))
	}

	return err
}

func doInitialMigration(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE "Config" (
            "Key" text NOT NULL,
            "Value" text COLLATE 'BINARY' NOT NULL,
            PRIMARY KEY ("Key")
        );`,

		`CREATE TABLE "Task" (
            "Description" text NOT NULL,
            "StartedAt" integer NOT NULL,
            "StoppedAt" integer NULL,
            "Tags" text COLLATE 'BINARY' NOT NULL,
            PRIMARY KEY ("StartedAt")
        );`,

		`INSERT INTO "Config" ("Key", "Value") VALUES (
            'MigrationVersion', 1
        )`,
	}

	for k := range queries {
		_, err := db.Exec(queries[k])
		if err != nil {
			return ErrBadQuery{err, queries[k], nil}
		}
	}

	return nil
}

func getVersion(db *sql.DB) int {
	var version int
	if err := db.QueryRow(
		`SELECT Value FROM Config WHERE Key = 'MigrationVersion' LIMIT 1`,
	).Scan(&version); err != nil {
		return 0
	}

	return version
}
