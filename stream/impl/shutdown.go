package impl

import (
	"context"
	"errors"
	"sync"
)

type shutdown struct {
	ctx     context.Context
	cancel  context.CancelFunc
	parent  context.Context
	errs    []error
	mu      sync.Mutex
	stopped bool
}

func newShutdown() *shutdown {
	ctx := context.Background()
	child, cancel := context.WithCancel(ctx)
	return &shutdown{
		ctx:    child,
		cancel: cancel,
		parent: ctx,
	}
}

func (s *shutdown) WithContext(ctx context.Context) {
	s.mu.Lock()
	s.parent = ctx
	s.ctx, s.cancel = context.WithCancel(ctx)
	if s.stopped {
		s.cancel()
	}
	s.mu.Unlock()
}

func (s *shutdown) Done() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ctx.Done()
}

func (s *shutdown) Abort(err error) {
	if err == nil {
		return
	}
	s.mu.Lock()
	s.errs = append(s.errs, err)
	cancel := s.cancel
	s.mu.Unlock()
	cancel()
}

func (s *shutdown) Stop() {
	s.mu.Lock()
	s.stopped = true
	cancel := s.cancel
	s.mu.Unlock()
	cancel()
}

func (s *shutdown) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	errs := s.errs
	if pErr := s.parent.Err(); pErr != nil {
		errs = append(append([]error{}, s.errs...), pErr)
	}
	return errors.Join(errs...)
}
