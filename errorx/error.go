package errorx

import (
	"errors"
	"fmt"
	"io"
	"runtime"
)

type Error struct {
	code  Code
	err   error
	stack []uintptr
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

func MustFrom(err error) *Error {
	if e, ok := From(err); ok {
		return e
	}
	if err != nil {
		return New(CodeUnknown, err, nil)
	}
	return New(CodeOK, nil, nil)
}

func (e *Error) Error() string {
	if e == nil || e.err == nil {
		return "<nil>"
	}
	return fmt.Sprintf("[%d] %s", e.code, e.err)
}

func (e *Error) Code() Code {
	if e == nil || e.err == nil {
		return 0
	}
	return e.code
}

func (e *Error) Unwrap() error {
	if e == nil || e.err == nil {
		return nil
	}
	return e.err
}

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

func (e *Error) Is(target error) bool {
	if e == nil || e.err == nil {
		return false
	}
	err, ok := target.(*Error)
	if !ok {
		return false
	}
	return err.code == e.code
}
