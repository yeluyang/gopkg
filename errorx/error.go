package errorx

import (
	"errors"
	"fmt"
	"io"
	"runtime"
)

type Error struct {
	code   Code
	err    error
	reason string
	stack  []uintptr
}

func New(code Code, err error, stack []uintptr) *Error {
	return &Error{code: code, err: err, stack: stack}
}

func From(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

// WithReason adds context to the error.
func (e *Error) WithReason(reason string) *Error {
	e.reason = reason
	return e
}

// Code returns the application error code.
func (e *Error) Code() Code {
	if e == nil || e.err == nil {
		return 0
	}
	return e.code
}

// Error implements the built-in error interface.
func (e *Error) Error() string {
	if e == nil || e.err == nil {
		return "<nil>"
	}
	if len(e.reason) == 0 {
		return fmt.Sprintf("[code=%d] %s", e.code, e.err)
	}
	return fmt.Sprintf("[code=%d] %s => %s", e.code, e.reason, e.err)
}

// Unwrap exposes the underlying error to errors.Is, errors.As, and errors.Unwrap.
func (e *Error) Unwrap() error {
	if e == nil || e.err == nil {
		return nil
	}
	return e.err
}

// Format implements fmt.Formatter, including the stack trace for the %v verb.
func (e *Error) Format(s fmt.State, verb rune) {
	if e == nil || e.err == nil {
		io.WriteString(s, "<nil>")
		return
	}
	switch verb {
	case 'v':
		io.WriteString(s, e.Error())
		if len(e.stack) > 0 {
			io.WriteString(s, "\nStack Trace:")
			frames := runtime.CallersFrames(e.stack)
			for {
				frame, more := frames.Next()
				if frame.Function == "" {
					break
				}
				fmt.Fprintf(s, "\n  %s\n    %s:%d", frame.Function, frame.File, frame.Line)
				if !more {
					break
				}
			}
		}
	case 's':
		io.WriteString(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}

// Is lets errors.Is compare Error and Code values by application code; Unwrap
// alone can only match the underlying errors.
func (e *Error) Is(target error) bool {
	if e == nil || e.err == nil {
		return false
	}
	switch target := target.(type) {
	case Code:
		return target == e.code
	case *Error:
		return target != nil && target.code == e.code
	default:
		return false
	}
}
