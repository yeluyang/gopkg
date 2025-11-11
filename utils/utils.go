package utils

import (
	"fmt"
	"runtime/debug"
)

func ForcePanic(err error) {
	if err != nil {
		stack := debug.Stack()
		go func() {
			panic(fmt.Errorf("%+v\n%s", err, string(stack)))
		}()
	}
}
