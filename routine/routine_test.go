package routine

import (
	"bytes"
	"log/slog"
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
	var buf bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(old)

	done := make(chan struct{})

	s.callerL1()

	// Wait for the recover + slog.Error to complete
	s.Eventually(func() bool {
		return buf.Len() > 0
	}, time.Second, 10*time.Millisecond)

	close(done)

	logOutput := buf.String()
	s.T().Logf("captured slog output:\n%s", logOutput)
	s.Contains(logOutput, "panic: test panic")
	s.Contains(logOutput, "routine_test.go")
}

func (s *TestSuiteRoutine) callerL1() { s.callerL2() }
func (s *TestSuiteRoutine) callerL2() { s.callerL3() }
func (s *TestSuiteRoutine) callerL3() { Go(s.panicL1) }
func (s *TestSuiteRoutine) panicL1()  { s.panicL2() }
func (s *TestSuiteRoutine) panicL2()  { s.panicL3() }
func (s *TestSuiteRoutine) panicL3()  { panic("test panic") }
