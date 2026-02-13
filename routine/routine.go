package routine

import (
	"fmt"
	"log/slog"
	"runtime"
)

func Go(fn func()) {
	var callerPCs [32]uintptr
	n := runtime.Callers(2, callerPCs[:])
	callerFrames := callerPCs[:n]

	go func() {
		defer func() {
			if r := recover(); r != nil {
				var panicPCs [32]uintptr
				// skip: runtime.Callers, this defer, panic()
				pn := runtime.Callers(3, panicPCs[:])

				var err error
				if v, ok := r.(error); ok {
					err = v
				} else {
					err = fmt.Errorf("%+v", r)
				}
				slog.Error(fmt.Sprintf("panic: %s\n\n%s\ncall from: \n\n%s", err,
					formatFrames(panicPCs[:pn]),
					formatFrames(callerFrames),
				))
			}
		}()
		fn()
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
