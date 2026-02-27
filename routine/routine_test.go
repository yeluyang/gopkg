package routine

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func TestRoutine(t *testing.T) {
	suite.Run(t, new(TestSuiteRoutine))
}

type TestSuiteRoutine struct {
	suite.Suite
}

func (s *TestSuiteRoutine) TestGoExecutesFunction() {
	var executed atomic.Bool
	var wg sync.WaitGroup
	wg.Add(1)

	Go(func() {
		executed.Store(true)
		wg.Done()
	})

	wg.Wait()
	s.True(executed.Load())
}

func (s *TestSuiteRoutine) TestGoExecutesConcurrently() {
	const numGoroutines = 10
	var counter atomic.Int32
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		Go(func() {
			counter.Add(1)
			wg.Done()
		})
	}

	wg.Wait()
	s.Equal(int32(numGoroutines), counter.Load())
}

func (s *TestSuiteRoutine) TestGoStartsImmediately() {
	started := make(chan struct{})

	Go(func() {
		close(started)
	})

	select {
	case <-started:
		// success
	case <-time.After(time.Second):
		s.Fail("goroutine did not start within timeout")
	}
}

func (s *TestSuiteRoutine) TestGoRecoversPanic() {
	errCh := make(chan error, 1)
	handler := WithErrorHandler(func(err error) { errCh <- err })

	s.callerL1(handler)

	var captured error
	select {
	case captured = <-errCh:
	case <-time.After(time.Second):
		s.Fail("goroutine did not panic within timeout")
		return
	}

	s.T().Logf("captured error:\n%s", captured.Error())
	s.Contains(captured.Error(), "panic: test panic")
	s.Contains(captured.Error(), "routine_test.go")
}

func (s *TestSuiteRoutine) callerL1(opts ...Option) { s.callerL2(opts...) }
func (s *TestSuiteRoutine) callerL2(opts ...Option) { s.callerL3(opts...) }
func (s *TestSuiteRoutine) callerL3(opts ...Option) { Go(s.panicL1, opts...) }
func (s *TestSuiteRoutine) panicL1()  { s.panicL2() }
func (s *TestSuiteRoutine) panicL2()  { s.panicL3() }
func (s *TestSuiteRoutine) panicL3()  { panic("test panic") }
