package routine

import (
	"fmt"
	"runtime"
)

// captureStack captures a formatted stack trace, skipping the specified
// number of frames above the caller. skip=0 starts at the caller of captureStack.
func captureStack(skip int) string {
	var pcs [32]uintptr
	// +2: skip runtime.Callers + captureStack itself
	n := runtime.Callers(2+skip, pcs[:])
	return formatFrames(pcs[:n])
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
