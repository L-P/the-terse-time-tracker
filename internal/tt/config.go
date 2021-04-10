package tt

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type Config struct {
	WeeklyHours, MonthlyHours time.Duration
}

func (c Config) Validate() error {
	if c.WeeklyHours < 0 {
		return ErrInvalidInput("WeeklyHours cannot be negative")
	}
	if c.WeeklyHours > (7 * 24 * time.Hour) {
		return ErrInvalidInput("WeeklyHours must fit in a week")
	}

	if c.MonthlyHours < 0 {
		return ErrInvalidInput("MonthlyHours cannot be negative")
	}
	if c.MonthlyHours > (31 * 7 * 24 * time.Hour) {
		return ErrInvalidInput("MonthlyHours must fit in a month")
	}

	if c.WeeklyHours != 0 && c.MonthlyHours != 0 && c.WeeklyHours > c.MonthlyHours {
		return ErrInvalidInput("WeeklyHours must be < MonthlyHours")
	}

	return nil
}

// nolint:wrapcheck,goerr113
func (c *Config) scan(s scannable) error {
	var key, value string
	if err := s.Scan(&key, &value); err != nil {
		return err
	}

	asDuration := func(str string) (time.Duration, error) {
		i, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return 0, err
		}

		return time.Duration(i), nil
	}

	var err error

	switch key {
	case "WeeklyHours":
		c.WeeklyHours, err = asDuration(value)
	case "MonthlyHours":
		c.MonthlyHours, err = asDuration(value)
	case "MigrationVersion": // NOP
	default:
		err = fmt.Errorf("unknown configuration key: %s", key)
	}

	return err
}

func loadConfig(db *sql.DB) (Config, error) {
	query := `SELECT Key, Value FROM Config` // order must match scan()
	rows, err := db.Query(query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Config{}, nil
		}

		return Config{}, ErrBadQuery{err, query, nil}
	}
	defer rows.Close()

	var c Config

	for rows.Next() {
		if err := c.scan(rows); err != nil {
			return Config{}, ErrBadQuery{err, query, nil}
		}
	}

	if err := rows.Err(); err != nil {
		return Config{}, ErrBadQuery{err, query, nil}
	}

	return c, nil
}

func (c Config) save(tx *sql.Tx) error {
	for k, v := range map[string]interface{}{
		"WeeklyHours":  c.WeeklyHours,
		"MonthlyHours": c.MonthlyHours,
	} {
		if err := exec(
			tx,
			`INSERT INTO Config (Key, Value) VALUES (?, ?)
            ON CONFLICT(Key) DO UPDATE SET Value=excluded.Value
            `,
			k, v,
		); err != nil {
			return err
		}
	}

	return nil
}
