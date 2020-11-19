// nolint:wrapcheck,goerr113
package util

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// TimeAsTimestamp is stored as an UNIX timestamp but used as a time.Time.
type TimeAsTimestamp time.Time

func (t TimeAsTimestamp) Value() (driver.Value, error) {
	return driver.Value(time.Time(t).Unix()), nil
}

func (t TimeAsTimestamp) Time() time.Time {
	return time.Time(t)
}

func (t *TimeAsTimestamp) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		tmp, err := strconv.ParseInt(string(src), 10, 64)
		if err != nil {
			return err
		}

		*t = TimeAsTimestamp(time.Unix(tmp, 0))
	case int64:
		tmp := TimeAsTimestamp(time.Unix(src, 0))
		*t = tmp
	default:
		return fmt.Errorf("expected []byte or int64, got %T", src)
	}

	return nil
}

func (t TimeAsTimestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time().Unix())
}

type NullTimeAsTimestamp struct {
	Time  TimeAsTimestamp
	Valid bool // Valid is true if TimeAsTimestamp is not NULL
}

// NewNullTimeAsTimestamp makes the assumption that a zero time is invalid.
func NewNullTimeAsTimestamp(t time.Time) NullTimeAsTimestamp {
	return NullTimeAsTimestamp{
		Time:  TimeAsTimestamp(t),
		Valid: !t.IsZero(),
	}
}

// Scan implements the Scanner interface.
func (ns *NullTimeAsTimestamp) Scan(value interface{}) error {
	if value == nil {
		ns.Time, ns.Valid = TimeAsTimestamp{}, false
		return nil
	}

	ns.Valid = true

	return ns.Time.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullTimeAsTimestamp) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}

	return ns.Time.Value()
}

func (ns NullTimeAsTimestamp) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(ns.Time.Time().Unix())
}
