package routine

import (
	"fmt"
	"log/slog"
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
					var panicPCs [32]uintptr
					// skip: runtime.Callers + this defer + runtime.gopanic
					pn := runtime.Callers(3, panicPCs[:])
					msg = fmt.Appendf(msg, "\n\n%s", formatFrames(panicPCs[:pn]))
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
// Panics are recovered and logged via slog.Error with both stacks.
func Go(fn func()) {
	wrapped := Recover(func() error { fn(); return nil }, WithCallerSkip(1)) // skip Go itself
	go func() {
		if err := wrapped(); err != nil {
			slog.Error(err.Error())
		}
	}()
}

func formatFrames(pcs []uintptr) string {
	frames := runtime.CallersFrames(pcs)
	var buf []byte
	for {
		frame, more := frames.Next()
		buf = fmt.Appendf(buf, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	return string(buf)
}
