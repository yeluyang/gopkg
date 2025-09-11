package routine

import "fmt"

func Recover() error {
	r := recover()
	if r == nil {
		return nil
	}
	if err, ok := r.(error); ok {
		return err
	} else {
		return fmt.Errorf("recover: %+v", r)
	}
}
