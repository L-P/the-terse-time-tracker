package tt

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

type Config struct {
	WeeklyHours, MonthlyHours float64
}

func (c Config) Validate() error {
	if c.WeeklyHours < 0 {
		return errors.New("WeeklyHours cannot be negative")
	}

	if c.MonthlyHours < 0 {
		return errors.New("MonthlyHours cannot be negative")
	}

	return nil
}

// nolint:wrapcheck,goerr113
func (c *Config) scan(s scannable) error {
	var key, value string
	if err := s.Scan(&key, &value); err != nil {
		return err
	}

	var err error
	switch key {
	case "WeeklyHours":
		c.WeeklyHours, err = strconv.ParseFloat(value, 64)
	case "MonthlyHours":
		c.MonthlyHours, err = strconv.ParseFloat(value, 64)
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
