package tt

import (
	"fmt"
)

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

type ErrInvalidInput string

func (e ErrInvalidInput) Error() string { return string(e) }
