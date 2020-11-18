package tt

import (
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3" // SQL driver
)

type TT struct {
	db *sql.DB
}

func New() (TT, error) {
	dsn, err := getDBDSN()
	if err != nil {
		return TT{}, err
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return TT{}, ErrIO{"unable to open database", dsn, err}
	}

	return TT{db: db}, nil
}

func (tt *TT) Close() error {
	return tt.db.Close()
}

func getDBDSN() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", ErrIO{"unable to fetch config dir", "os.UserConfigDir", nil}
	}

	path := filepath.Join(dir, "the-terse-time-tracker.db")
	return "sqlite3://" + path, nil
}

func (tt *TT) Start(raw string) (Task, error) {
	desc, tags := parseRawDesc(raw)
	return NewTask(desc, tags), nil
}

// parseRawDesc splits the description and tags from user input.
// tags can only be provided at the end of the string
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
