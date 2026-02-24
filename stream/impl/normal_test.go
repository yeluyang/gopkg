package impl

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	mockinterfaces "github.com/yeluyang/gopkg/stream/interfaces/mocks"
)

func TestStreamer(t *testing.T) {
	suite.Run(t, new(TestSuiteStreamer))
}

type TestSuiteStreamer struct {
	suite.Suite
}

func (s *TestSuiteStreamer) TestNormal() {
	bd := newBatchedData(_TOTAL, _BATCH)
	sent := make(map[int]bool, _TOTAL)
	for i := range _TOTAL {
		sent[i] = false
	}
	var accepted sync.Map

	// Pre-create all message mocks in the test goroutine to avoid cleanup races.
	msgBatches := make([][]Message[struct{}, *mockMsgAccepted], len(bd))
	for i, dataBatch := range bd {
		msgs := make([]Message[struct{}, *mockMsgAccepted], len(dataBatch))
		for j, d := range dataBatch {
			msg := mockinterfaces.NewMockeryMessage[struct{}, *mockMsgAccepted](s.T())
			msg.EXPECT().Accept(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ struct{}) (*mockMsgAccepted, error) {
				d.access(s.T())
				_, loaded := accepted.LoadOrStore(d.id, struct{}{})
				require.False(s.T(), loaded, d.id)
				return &mockMsgAccepted{id: d.id}, nil
			}).Once()
			msgs[j] = msg
		}
		msgBatches[i] = msgs
	}

	batchIdx := 0
	source := mockinterfaces.NewMockerySource[struct{}, *mockMsgAccepted](s.T())
	source.EXPECT().Next(mock.Anything).RunAndReturn(func(_ context.Context) ([]Message[struct{}, *mockMsgAccepted], bool, error) {
		if batchIdx >= len(msgBatches) {
			return nil, false, nil
		}
		for _, d := range bd[batchIdx] {
			sent[d.id] = true
		}
		batch := msgBatches[batchIdx]
		batchIdx++
		return batch, true, nil
	})

	result := make(map[int]struct{}, _TOTAL)
	sink := mockinterfaces.NewMockerySink[*mockMsgAccepted](s.T())
	sink.EXPECT().Drain(mock.Anything, mock.Anything).Run(func(_ context.Context, datas ...*mockMsgAccepted) {
		for _, d := range datas {
			require.NotContains(s.T(), result, d.id)
			result[d.id] = struct{}{}
		}
	}).Return(nil)

	stream := New(source, struct{}{}, _ROUTINES, sink)
	s.Require().NoError(stream.Run(context.Background()))

	_, ok := <-stream.actorPool.C()
	s.Require().False(ok)

	for _, batch := range bd {
		for _, d := range batch {
			s.Require().True(sent[d.id], d.id)
			_, ok := accepted.Load(d.id)
			s.Require().True(ok, d.id)
			s.Require().Contains(result, d.id)
		}
	}
}

func (s *TestSuiteStreamer) TestEmptySource() {
	source := mockinterfaces.NewMockerySource[struct{}, int](s.T())
	source.EXPECT().Next(mock.Anything).Return(nil, false, nil).Once()

	sink := mockinterfaces.NewMockerySink[int](s.T())

	stream := New[struct{}, int](source, struct{}{}, _ROUTINES, sink)
	s.Require().NoError(stream.Run(context.Background()))

	_, ok := <-stream.actorPool.C()
	s.Require().False(ok)
}

func (s *TestSuiteStreamer) TestCloseWhileRunning() {
	var msgID atomic.Int32
	source := mockinterfaces.NewMockerySource[struct{}, *mockMsgAccepted](s.T())
	source.EXPECT().Next(mock.Anything).RunAndReturn(func(_ context.Context) ([]Message[struct{}, *mockMsgAccepted], bool, error) {
		msgs := make([]Message[struct{}, *mockMsgAccepted], _BATCH)
		for i := range msgs {
			id := int(msgID.Add(1)) - 1
			// Use bare mock (no cleanup registration) to avoid races with infinite creation.
			msg := &mockinterfaces.MockeryMessage[struct{}, *mockMsgAccepted]{}
			msg.On("Accept", mock.Anything, mock.Anything).Return(&mockMsgAccepted{id: id}, nil)
			msgs[i] = msg
		}
		return msgs, true, nil
	}).Maybe()

	var count atomic.Int32
	started := make(chan struct{})
	var once sync.Once
	sink := mockinterfaces.NewMockerySink[*mockMsgAccepted](s.T())
	sink.EXPECT().Drain(mock.Anything, mock.Anything).Run(func(_ context.Context, datas ...*mockMsgAccepted) {
		count.Add(int32(len(datas)))
		once.Do(func() { close(started) })
	}).Return(nil).Maybe()

	stream := New(source, struct{}{}, _ROUTINES, sink)

	done := make(chan error, 1)
	go func() { done <- stream.Run(context.Background()) }()

	// Wait until data has flowed through the entire pipeline (source -> accept -> sink).
	<-started

	// Close while pipeline is actively processing.
	_ = stream.Close()

	// Run should complete.
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		s.Fail("Run did not return after Close")
	}

	// Channel should be closed (coordinator goroutine cleaned up).
	_, ok := <-stream.actorPool.C()
	s.Require().False(ok)

	// Some data should have been processed.
	s.Require().Greater(int(count.Load()), 0)
}

func (s *TestSuiteStreamer) TestClose() {
	// Use a blocking source so Run() doesn't complete on its own.
	source := mockinterfaces.NewMockerySource[struct{}, int](s.T())
	source.EXPECT().Next(mock.Anything).RunAndReturn(func(ctx context.Context) ([]Message[struct{}, int], bool, error) {
		<-ctx.Done()
		return nil, false, ctx.Err()
	}).Maybe()

	sink := mockinterfaces.NewMockerySink[int](s.T())

	stream := New[struct{}, int](source, struct{}{}, _ROUTINES, sink)

	// Run in background
	done := make(chan error, 1)
	go func() { done <- stream.Run(context.Background()) }()

	// Close should cause Run to return.
	_ = stream.Close()

	// Run should also complete
	select {
	case err := <-done:
		// err may be nil or context.Canceled — both acceptable
		_ = err
	case <-time.After(5 * time.Second):
		s.Fail("Run did not return after Close")
	}
}
