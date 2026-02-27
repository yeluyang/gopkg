package routine

import "fmt"

// Panic triggers an unrecoverable panic that cannot be caught by any recover().
// It panics in a separate goroutine where no deferred recover can intercept it,
// and blocks the caller so execution cannot continue.
// The panic message includes the caller's stack trace for diagnostics.
func Panic(v any) {
	// skip: captureStack + Panic
	stack := captureStack(1)

	go func() {
		panic(fmt.Sprintf("%v\n\n%s", v, stack))
	}()

	select {}
}
