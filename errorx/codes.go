package errorx

import (
	"errors"
	"fmt"
	"runtime"
)

type Code int64

// Error implements error so Code can be used directly with errors.Is.
func (c Code) Error() string {
	return fmt.Sprintf("code=%d", c)
}

func (c Code) With(msg string, a ...any) *Error {
	if len(a) > 0 {
		msg = fmt.Sprintf(msg, a...)
	}
	return c.from(errors.New(msg), "")
}

func (c Code) From(err error) *Error {
	return c.from(err, "")
}

func (c Code) Fromf(err error, format string, a ...any) *Error {
	return c.from(err, fmt.Sprintf(format, a...))
}

func (c Code) from(err error, reason string) *Error {
	pc := make([]uintptr, 32)
	return New(c, err, pc[:runtime.Callers(3, pc)]).WithReason(reason)
}
