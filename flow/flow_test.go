package flow

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/yeluyang/gopkg/flow/interfaces"
	mockinterfaces "github.com/yeluyang/gopkg/flow/interfaces/mocks"
)

type testVisitor struct{}

func TestFlow(t *testing.T) {
	suite.Run(t, new(suiteFlow))
}

type suiteFlow struct {
	suite.Suite
}

// --- helpers ---

func makeMsg(t *testing.T, activateIDs []string, drainTo []string, acceptResult []interfaces.Message[string, *testVisitor], acceptErr error) *mockinterfaces.MockeryMessage[string, *testVisitor] {
	msg := mockinterfaces.NewMockeryMessage[string, *testVisitor](t)
	msg.EXPECT().Options().Return(nil).Maybe()
	msg.EXPECT().Activate(mock.Anything).Return(activateIDs).Maybe()
	msg.EXPECT().DrainTo(mock.Anything).Return(drainTo).Maybe()
	if drainTo == nil || len(drainTo) == 0 {
		msg.EXPECT().Accept(mock.Anything, mock.Anything).Return(acceptResult, acceptErr).Maybe()
	}
	return msg
}

func makeSource(t *testing.T, id string, msgs []interfaces.Message[string, *testVisitor]) *mockinterfaces.MockerySource[string, *testVisitor] {
	src := mockinterfaces.NewMockerySource[string, *testVisitor](t)
	src.EXPECT().ID().Return(id).Maybe()
	src.EXPECT().Activate(mock.Anything).Return(nil).Maybe()
	src.EXPECT().Close(mock.Anything).Return(nil).Maybe()

	if len(msgs) > 0 {
		src.EXPECT().Next(mock.Anything).Return(msgs, true, nil).Once()
	}
	src.EXPECT().Next(mock.Anything).Return(nil, false, nil).Maybe()

	return src
}

func makeSink(t *testing.T, id string) (*mockinterfaces.MockerySink[string, *testVisitor], *[]interfaces.Message[string, *testVisitor]) {
	sink := mockinterfaces.NewMockerySink[string, *testVisitor](t)
	sink.EXPECT().ID().Return(id).Maybe()
	sink.EXPECT().Activate(mock.Anything).Return(nil).Maybe()
	sink.EXPECT().Close(mock.Anything).Return(nil).Maybe()

	var received []interfaces.Message[string, *testVisitor]
	var mu sync.Mutex
	sink.EXPECT().Drain(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, msgs []interfaces.Message[string, *testVisitor]) error {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, msgs...)
		return nil
	}).Maybe()

	return sink, &received
}

// --- Test 1: Single source, single sink ---

func (s *suiteFlow) TestSingleSourceSingleSink() {
	msg := makeMsg(s.T(), nil, []string{"sink-1"}, nil, nil)
	msg.EXPECT().Accept(mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	src := makeSource(s.T(), "source-1", []interfaces.Message[string, *testVisitor]{msg})
	sink, received := makeSink(s.T(), "sink-1")

	f := New[string, *testVisitor](&testVisitor{}).
		ActivateSource(src).
		ActivateSink(sink).
		Build()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.Require().NoError(f.Run(ctx))
	s.Require().NoError(f.Wait())
	s.Require().Len(*received, 1)
}

// --- Test 2: Multiple sources ---

func (s *suiteFlow) TestMultipleSources() {
	msg1 := makeMsg(s.T(), nil, []string{"sink-1"}, nil, nil)
	msg1.EXPECT().Accept(mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	msg2 := makeMsg(s.T(), nil, []string{"sink-1"}, nil, nil)
	msg2.EXPECT().Accept(mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	src1 := makeSource(s.T(), "source-1", []interfaces.Message[string, *testVisitor]{msg1})
	src2 := makeSource(s.T(), "source-2", []interfaces.Message[string, *testVisitor]{msg2})
	sink, received := makeSink(s.T(), "sink-1")

	f := New[string, *testVisitor](&testVisitor{}).
		ActivateSource(src1).
		ActivateSource(src2).
		ActivateSink(sink).
		Build()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.Require().NoError(f.Run(ctx))
	s.Require().NoError(f.Wait())
	s.Require().Len(*received, 2)
}

// --- Test 3: Multiple sinks, DrainTo routes ---

func (s *suiteFlow) TestMultipleSinks() {
	msg1 := makeMsg(s.T(), nil, []string{"sink-a"}, nil, nil)
	msg1.EXPECT().Accept(mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	msg2 := makeMsg(s.T(), nil, []string{"sink-b"}, nil, nil)
	msg2.EXPECT().Accept(mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	src := makeSource(s.T(), "source-1", []interfaces.Message[string, *testVisitor]{msg1, msg2})
	sinkA, receivedA := makeSink(s.T(), "sink-a")
	sinkB, receivedB := makeSink(s.T(), "sink-b")

	f := New[string, *testVisitor](&testVisitor{}).
		ActivateSource(src).
		ActivateSink(sinkA).
		ActivateSink(sinkB).
		Build()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.Require().NoError(f.Run(ctx))
	s.Require().NoError(f.Wait())
	s.Require().Len(*receivedA, 1)
	s.Require().Len(*receivedB, 1)
}

// --- Test 4: Accept fan-out (new msgs re-enter pipeline) ---

func (s *suiteFlow) TestAcceptFanOut() {
	childMsg := makeMsg(s.T(), nil, []string{"sink-1"}, nil, nil)
	childMsg.EXPECT().Accept(mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	parentMsg := makeMsg(s.T(), nil, nil, []interfaces.Message[string, *testVisitor]{childMsg}, nil)

	src := makeSource(s.T(), "source-1", []interfaces.Message[string, *testVisitor]{parentMsg})
	sink, received := makeSink(s.T(), "sink-1")

	f := New[string, *testVisitor](&testVisitor{}).
		ActivateSource(src).
		ActivateSink(sink).
		Build()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.Require().NoError(f.Run(ctx))
	s.Require().NoError(f.Wait())
	s.Require().Len(*received, 1)
}

// --- Test 5: Lazy activation ---

func (s *suiteFlow) TestLazyActivation() {
	lazyMsg := makeMsg(s.T(), nil, []string{"sink-1"}, nil, nil)
	lazyMsg.EXPECT().Accept(mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	lazySrc := mockinterfaces.NewMockerySource[string, *testVisitor](s.T())
	lazySrc.EXPECT().ID().Return("lazy-source").Maybe()
	lazySrc.EXPECT().Activate(mock.Anything).Return(nil).Maybe()
	lazySrc.EXPECT().Close(mock.Anything).Return(nil).Maybe()
	lazySrc.EXPECT().Next(mock.Anything).Return([]interfaces.Message[string, *testVisitor]{lazyMsg}, true, nil).Once()
	lazySrc.EXPECT().Next(mock.Anything).Return(nil, false, nil).Maybe()

	triggerMsg := makeMsg(s.T(), []string{"lazy-source"}, nil, nil, nil)

	src := makeSource(s.T(), "source-1", []interfaces.Message[string, *testVisitor]{triggerMsg})
	sink, received := makeSink(s.T(), "sink-1")

	f := New[string, *testVisitor](&testVisitor{}).
		ActivateSource(src).
		ActivateSink(sink).
		Source(lazySrc).
		Build()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.Require().NoError(f.Run(ctx))
	s.Require().NoError(f.Wait())
	s.Require().Len(*received, 1)
}

// --- Test 6: Lazy activation idempotent ---

func (s *suiteFlow) TestLazyActivationIdempotent() {
	lazySrc := mockinterfaces.NewMockerySource[string, *testVisitor](s.T())
	lazySrc.EXPECT().ID().Return("lazy-source").Maybe()
	lazySrc.EXPECT().Activate(mock.Anything).Return(nil).Once()
	lazySrc.EXPECT().Close(mock.Anything).Return(nil).Maybe()
	lazySrc.EXPECT().Next(mock.Anything).Return(nil, false, nil).Maybe()

	msg1 := makeMsg(s.T(), []string{"lazy-source"}, nil, nil, nil)
	msg2 := makeMsg(s.T(), []string{"lazy-source"}, nil, nil, nil)

	src := makeSource(s.T(), "source-1", []interfaces.Message[string, *testVisitor]{msg1, msg2})

	f := New[string, *testVisitor](&testVisitor{}).
		ActivateSource(src).
		Concurrency(1).
		Source(lazySrc).
		Build()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.Require().NoError(f.Run(ctx))
	s.Require().NoError(f.Wait())
}

// --- Test 7: All lazy, no eager → immediate return ---

func (s *suiteFlow) TestAllLazyImmediate() {
	lazySrc := mockinterfaces.NewMockerySource[string, *testVisitor](s.T())
	lazySrc.EXPECT().ID().Return("lazy-source").Maybe()
	lazySrc.EXPECT().Close(mock.Anything).Return(nil).Maybe()

	f := New[string, *testVisitor](&testVisitor{}).
		Source(lazySrc).
		Build()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	s.Require().NoError(f.Run(ctx))
	s.Require().NoError(f.Wait())
}

// --- Test 8: Error from source ---

func (s *suiteFlow) TestErrorFromSource() {
	expectedErr := errors.New("source failed")

	src := mockinterfaces.NewMockerySource[string, *testVisitor](s.T())
	src.EXPECT().ID().Return("source-1").Maybe()
	src.EXPECT().Activate(mock.Anything).Return(nil).Maybe()
	src.EXPECT().Close(mock.Anything).Return(nil).Maybe()
	src.EXPECT().Next(mock.Anything).Return(nil, false, expectedErr).Once()

	f := New[string, *testVisitor](&testVisitor{}).
		ActivateSource(src).
		Build()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.Require().NoError(f.Run(ctx))
	err := f.Wait()
	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
}

// --- Test 9: Error from Accept ---

func (s *suiteFlow) TestErrorFromAccept() {
	expectedErr := errors.New("accept failed")

	msg := mockinterfaces.NewMockeryMessage[string, *testVisitor](s.T())
	msg.EXPECT().Options().Return(nil).Maybe()
	msg.EXPECT().Activate(mock.Anything).Return(nil).Maybe()
	msg.EXPECT().DrainTo(mock.Anything).Return(nil).Maybe()
	msg.EXPECT().Accept(mock.Anything, mock.Anything).Return(nil, expectedErr).Once()

	src := makeSource(s.T(), "source-1", []interfaces.Message[string, *testVisitor]{msg})

	f := New[string, *testVisitor](&testVisitor{}).
		ActivateSource(src).
		Build()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.Require().NoError(f.Run(ctx))
	err := f.Wait()
	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
}

// --- Test 10: Error from Drain ---

func (s *suiteFlow) TestErrorFromDrain() {
	expectedErr := errors.New("drain failed")

	msg := makeMsg(s.T(), nil, []string{"sink-1"}, nil, nil)
	msg.EXPECT().Accept(mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	src := makeSource(s.T(), "source-1", []interfaces.Message[string, *testVisitor]{msg})

	sink := mockinterfaces.NewMockerySink[string, *testVisitor](s.T())
	sink.EXPECT().ID().Return("sink-1").Maybe()
	sink.EXPECT().Activate(mock.Anything).Return(nil).Maybe()
	sink.EXPECT().Close(mock.Anything).Return(nil).Maybe()
	sink.EXPECT().Drain(mock.Anything, mock.Anything).Return(expectedErr).Once()

	f := New[string, *testVisitor](&testVisitor{}).
		ActivateSource(src).
		ActivateSink(sink).
		Build()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.Require().NoError(f.Run(ctx))
	err := f.Wait()
	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
}

// --- Test 11: Context cancellation ---

func (s *suiteFlow) TestContextCancellation() {
	src := mockinterfaces.NewMockerySource[string, *testVisitor](s.T())
	src.EXPECT().ID().Return("source-1").Maybe()
	src.EXPECT().Activate(mock.Anything).Return(nil).Maybe()
	src.EXPECT().Close(mock.Anything).Return(nil).Maybe()
	src.EXPECT().Next(mock.Anything).RunAndReturn(func(ctx context.Context) ([]interfaces.Message[string, *testVisitor], bool, error) {
		<-ctx.Done()
		return nil, false, nil
	}).Maybe()

	f := New[string, *testVisitor](&testVisitor{}).
		ActivateSource(src).
		Build()

	ctx, cancel := context.WithCancel(context.Background())
	s.Require().NoError(f.Run(ctx))
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()
	err := f.Wait()
	s.Require().Error(err)
	s.Require().ErrorIs(err, context.Canceled)
}

// --- Test 12: Duplex endpoint ---

func (s *suiteFlow) TestDuplex() {
	sinkMsg := makeMsg(s.T(), nil, []string{"duplex-1"}, nil, nil)
	sinkMsg.EXPECT().Accept(mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	src := makeSource(s.T(), "source-1", []interfaces.Message[string, *testVisitor]{sinkMsg})

	var duplexReceived []interfaces.Message[string, *testVisitor]
	var mu sync.Mutex

	dup := mockinterfaces.NewMockeryDuplex[string, *testVisitor](s.T())
	dup.EXPECT().ID().Return("duplex-1").Maybe()
	dup.EXPECT().Activate(mock.Anything).Return(nil).Maybe()
	dup.EXPECT().Close(mock.Anything).Return(nil).Maybe()
	dup.EXPECT().Next(mock.Anything).Return(nil, false, nil).Maybe()
	dup.EXPECT().Drain(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, msgs []interfaces.Message[string, *testVisitor]) error {
		mu.Lock()
		defer mu.Unlock()
		duplexReceived = append(duplexReceived, msgs...)
		return nil
	}).Maybe()

	f := New[string, *testVisitor](&testVisitor{}).
		ActivateSource(src).
		ActivateDuplex(dup).
		Build()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.Require().NoError(f.Run(ctx))
	s.Require().NoError(f.Wait())

	mu.Lock()
	defer mu.Unlock()
	s.Require().Len(duplexReceived, 1)
}

// --- Test 13: Actor concurrency ---

func (s *suiteFlow) TestActorConcurrency() {
	var maxConcurrent atomic.Int64
	var current atomic.Int64

	numMsgs := 10
	msgs := make([]interfaces.Message[string, *testVisitor], numMsgs)
	for i := 0; i < numMsgs; i++ {
		m := mockinterfaces.NewMockeryMessage[string, *testVisitor](s.T())
		m.EXPECT().Options().Return(nil).Maybe()
		m.EXPECT().Activate(mock.Anything).Return(nil).Maybe()
		m.EXPECT().DrainTo(mock.Anything).Return(nil).Maybe()
		m.EXPECT().Accept(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, v *testVisitor) ([]interfaces.Message[string, *testVisitor], error) {
			cur := current.Add(1)
			for {
				old := maxConcurrent.Load()
				if cur <= old || maxConcurrent.CompareAndSwap(old, cur) {
					break
				}
			}
			time.Sleep(50 * time.Millisecond)
			current.Add(-1)
			return nil, nil
		}).Once()
		msgs[i] = m
	}

	src := makeSource(s.T(), "source-1", msgs)

	f := New[string, *testVisitor](&testVisitor{}).
		ActivateSource(src).
		Concurrency(4).
		Build()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.Require().NoError(f.Run(ctx))
	s.Require().NoError(f.Wait())
	s.Require().Greater(maxConcurrent.Load(), int64(1))
}

// --- Test 14: Endpoint Close on error ---

func (s *suiteFlow) TestEndpointCloseOnError() {
	expectedErr := errors.New("boom")

	closeCalled := atomic.Bool{}

	src := mockinterfaces.NewMockerySource[string, *testVisitor](s.T())
	src.EXPECT().ID().Return("source-1").Maybe()
	src.EXPECT().Activate(mock.Anything).Return(nil).Maybe()
	src.EXPECT().Close(mock.Anything).RunAndReturn(func(ctx context.Context) error {
		closeCalled.Store(true)
		return nil
	}).Maybe()
	src.EXPECT().Next(mock.Anything).Return(nil, false, expectedErr).Once()

	f := New[string, *testVisitor](&testVisitor{}).
		ActivateSource(src).
		Build()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.Require().NoError(f.Run(ctx))
	err := f.Wait()
	s.Require().Error(err)
	s.Require().True(closeCalled.Load())
}

// --- Test 15: Message Options ---

func (s *suiteFlow) TestMessageOptions() {
	opts := []interfaces.Option{
		interfaces.WithSerial(),
		interfaces.WithSerialKey("key-1"),
	}
	cfg := interfaces.ResolveOptions(opts)
	s.Require().True(cfg.Serial)
	s.Require().Equal("key-1", cfg.SerialKey)
}
