package routine

import (
	"fmt"
	"log/slog"
	"runtime/debug"
)

func Go(fn func()) {
	stack := debug.Stack()
	go func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%+v", r)
			if v, ok := r.(error); ok {
				err = v
			}
			slog.Error("catch panic in go routine", slog.String("stack", string(stack)), slog.Any("error", err))
		}
		fn()
	}()
}
