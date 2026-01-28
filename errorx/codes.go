package errorx

import (
	"errors"
	"fmt"
	"runtime"
)

type Code int64

func (c Code) Format(format string, a ...any) *Error {
	return c.from(fmt.Errorf(format, a...))
}

func (c Code) With(msg string) *Error {
	return c.from(errors.New(msg))
}

func (c Code) From(err error) *Error {
	return c.from(err)
}

func (c Code) from(err error) *Error {
	pc := make([]uintptr, 32)
	return New(c, err, pc[:runtime.Callers(3, pc)])
}
