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
