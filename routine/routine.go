package routine

import (
	"fmt"
	"os"
	"runtime"
)

// Recover returns a function that calls fn with panic recovery.
// If fn panics, the returned function returns a formatted error
// containing both the panic stack and the caller stack.
// Must be called before launching the goroutine to capture the caller stack.
func Recover(fn func() error, options ...Option) func() error {
	var cfg config
	for _, opt := range options {
		opt(&cfg)
	}

	var callerFrames []uintptr
	if !cfg.noCallerStack {
		var callerPCs [32]uintptr
		// skip: runtime.Callers + Wrap + user-requested extra
		n := runtime.Callers(2+cfg.callerSkip, callerPCs[:])
		callerFrames = callerPCs[:n]
	}

	capturePanicStack := !cfg.noPanicStack

	return func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				var panicErr error
				if e, ok := r.(error); ok {
					panicErr = e
				} else {
					panicErr = fmt.Errorf("%+v", r)
				}

				msg := fmt.Appendf(nil, "panic: %s", panicErr)
				if capturePanicStack {
					// skip: captureStack + this defer + runtime.gopanic
					msg = fmt.Appendf(msg, "\n\n%s", captureStack(2))
				}
				if len(callerFrames) > 0 {
					msg = fmt.Appendf(msg, "\ncalled from:\n\n%s", formatFrames(callerFrames))
				}
				if err != nil {
					msg = fmt.Appendf(msg, "\noriginal error: %+v", err)
				}
				err = fmt.Errorf("%s", msg)
			}
		}()
		return fn()
	}
}

// Go launches fn in a new goroutine with panic recovery.
// Panics are recovered and the resulting error is passed to the error handler.
// If no WithErrorHandler option is provided, errors are written to stderr.
func Go(fn func(), options ...Option) {
	var cfg config
	for _, opt := range options {
		opt(&cfg)
	}

	errHandler := cfg.errorHandler
	if errHandler == nil {
		errHandler = func(err error) { fmt.Fprintln(os.Stderr, err.Error()) }
	}

	// WithCallerSkip(1) is the baseline to skip Go itself; user options follow
	// and can override it (e.g. WithCallerSkip(2) to also skip a wrapper).
	opts := append([]Option{WithCallerSkip(1)}, options...)
	wrapped := Recover(func() error { fn(); return nil }, opts...)
	go func() {
		if err := wrapped(); err != nil {
			errHandler(err)
		}
	}()
}
