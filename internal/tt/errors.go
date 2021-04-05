package tt

import (
	"errors"
	"fmt"
)

var ErrContinue = errors.New("continuing identical running task")
var ErrNoCurrentTask = errors.New("there is no running task")
var ErrInvalidTaskID = errors.New("invalid task ID")

type ErrIO struct {
	msg, path string
	wrapped   error
}

func (e ErrIO) Error() string {
	if e.wrapped != nil {
		return fmt.Sprintf("I/O error: %s (%s): %s", e.msg, e.path, e.wrapped)
	}

	return fmt.Sprintf("I/O error: %s (%s)", e.msg, e.path)
}

type ErrBadQuery struct {
	wrapped error
	query   string
	params  []interface{}
}

func (e ErrBadQuery) Error() string {
	return fmt.Sprintf("bad query: %s\n%s", e.wrapped, e.query)
}

type (
	ErrInvalidInput string // user provided invalid data
	ErrDatabase     string // database driver error
	ErrRuntime      string // generic runtime error
)

func (e ErrInvalidInput) Error() string { return string(e) }
func (e ErrDatabase) Error() string     { return string(e) }
func (e ErrRuntime) Error() string      { return string(e) }
