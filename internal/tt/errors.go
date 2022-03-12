package tt

import (
	"errors"
	"fmt"
)

var ErrContinue = errors.New("continuing identical running task")
var ErrNoCurrentTask = errors.New("there is no running task")
var ErrInvalidTaskID = errors.New("invalid task ID")
var ErrInvalidTaskDesc = errors.New("invalid task description")
var ErrNotConfigured = errors.New("missing configuration for this feature")
var ErrNoTasks = errors.New("no tasks are present in the specified range")

type IOError struct {
	msg, path string
	wrapped   error
}

func (e IOError) Error() string {
	if e.wrapped != nil {
		return fmt.Sprintf("I/O error: %s (%s): %s", e.msg, e.path, e.wrapped)
	}

	return fmt.Sprintf("I/O error: %s (%s)", e.msg, e.path)
}

type BadQueryError struct {
	wrapped error
	query   string
	params  []interface{}
}

func (e BadQueryError) Error() string {
	return fmt.Sprintf("bad query: %s\n%s", e.wrapped, e.query)
}

type (
	InvalidInputError string // user provided invalid data
	DatabaseError     string // database driver error
	RuntimeError      string // generic runtime error
	ExitCodeError     int    // normal error that needs to bubble up to the shell
)

func (e InvalidInputError) Error() string { return string(e) }
func (e DatabaseError) Error() string     { return string(e) }
func (e RuntimeError) Error() string      { return string(e) }
func (e ExitCodeError) Error() string     { return fmt.Sprintf("exit code %d", e.Code()) }
func (e ExitCodeError) Code() int         { return int(e) }
