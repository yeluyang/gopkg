package impl

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestShutdown(t *testing.T) {
	suite.Run(t, new(TestSuiteShutdown))
}

type TestSuiteShutdown struct{ suite.Suite }

func (s *TestSuiteShutdown) TestAbortNil() {
	sd := newShutdown()
	sd.WithContext(context.Background())

	sd.Abort(nil)

	select {
	case <-sd.Done():
		s.Fail("Done() should still block after Abort(nil)")
	default:
		// expected: channel is not closed
	}
	s.Require().NoError(sd.Err())
}

func (s *TestSuiteShutdown) TestAbortSingle() {
	sd := newShutdown()
	sd.WithContext(context.Background())

	sd.Abort(errors.New("x"))

	select {
	case <-sd.Done():
		// expected: channel is closed
	default:
		s.Fail("Done() should be closed after Abort")
	}

	err := sd.Err()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "x")
}

func (s *TestSuiteShutdown) TestAbortMultiple() {
	sd := newShutdown()
	sd.WithContext(context.Background())

	sd.Abort(errors.New("first"))
	sd.Abort(errors.New("second"))

	err := sd.Err()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "first")
	s.Require().Contains(err.Error(), "second")
}

func (s *TestSuiteShutdown) TestAbortConcurrent() {
	sd := newShutdown()
	sd.WithContext(context.Background())

	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)
	for i := range n {
		go func(id int) {
			defer wg.Done()
			sd.Abort(fmt.Errorf("err-%d", id))
		}(i)
	}
	wg.Wait()

	err := sd.Err()
	s.Require().Error(err)
	for i := range n {
		s.Require().Contains(err.Error(), fmt.Sprintf("err-%d", i))
	}
}

func (s *TestSuiteShutdown) TestStop() {
	sd := newShutdown()
	sd.WithContext(context.Background())

	sd.Stop()

	select {
	case <-sd.Done():
		// expected: channel is closed
	default:
		s.Fail("Done() should be closed after Stop()")
	}
	s.Require().NoError(sd.Err())
}

func (s *TestSuiteShutdown) TestWithContextPreCanceled() {
	sd := newShutdown()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before WithContext

	sd.WithContext(ctx)

	select {
	case <-sd.Done():
		// expected: channel fires immediately
	default:
		s.Fail("Done() should fire immediately with pre-canceled context")
	}
}

func (s *TestSuiteShutdown) TestErrParentCanceled() {
	sd := newShutdown()

	ctx, cancel := context.WithCancel(context.Background())
	sd.WithContext(ctx)

	sd.Abort(errors.New("internal"))
	cancel()

	err := sd.Err()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "internal")
	s.Require().ErrorIs(err, context.Canceled)
}

func (s *TestSuiteShutdown) TestNewShutdownInitialState() {
	sd := newShutdown()

	s.Require().NoError(sd.Err())

	select {
	case <-sd.Done():
		s.Fail("Done() should still block on a freshly created shutdown")
	default:
		// expected: channel is not closed
	}
}

func (s *TestSuiteShutdown) TestDoubleStop() {
	sd := newShutdown()
	sd.WithContext(context.Background())

	sd.Stop()
	sd.Stop()

	select {
	case <-sd.Done():
		// expected: channel is closed
	default:
		s.Fail("Done() should be closed after Stop()")
	}
	s.Require().NoError(sd.Err())
}

func (s *TestSuiteShutdown) TestStopAfterAbort() {
	sd := newShutdown()
	sd.WithContext(context.Background())

	sd.Abort(errors.New("x"))
	sd.Stop()

	select {
	case <-sd.Done():
		// expected: channel is closed
	default:
		s.Fail("Done() should be closed after Abort+Stop")
	}

	err := sd.Err()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "x")
}

func (s *TestSuiteShutdown) TestWithContextAfterStop() {
	sd := newShutdown()
	sd.WithContext(context.Background())

	sd.Stop()
	sd.WithContext(context.Background())

	select {
	case <-sd.Done():
		// expected: channel is closed immediately because stopped flag is set
	default:
		s.Fail("Done() should be closed immediately after WithContext on a stopped shutdown")
	}
}
