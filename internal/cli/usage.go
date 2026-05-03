package cli

import (
	"errors"
	"fmt"
)

// ErrUsage is returned for invalid CLI usage (unknown command, bad flags, bad arguments).
// The main package exits with code 2 when errors.Is(err, ErrUsage).
// Wrap with usageErrf so Error() is a clear user-facing message only.
var ErrUsage = errors.New("usage")

type usageError struct {
	msg string
}

func (e *usageError) Error() string { return e.msg }

func (e *usageError) Unwrap() error { return ErrUsage }

func usageErrf(format string, args ...any) error {
	return &usageError{msg: fmt.Sprintf(format, args...)}
}
