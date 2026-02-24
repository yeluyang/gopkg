package impl

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/stretchr/testify/mock"
	"github.com/yeluyang/gopkg/errorx"
	"github.com/yeluyang/gopkg/stream/errs"
	mockinterfaces "github.com/yeluyang/gopkg/stream/interfaces/mocks"
)

func (s *TestSuiteStreamer) TestFailOnSource() {
	bd := newBatchedData(_TOTAL, _BATCH)
	failed := newMockFailed(_FAILED)

	// Pre-create all message mocks in the test goroutine.
	msgBatches := make([][]Message[struct{}, *mockMsgAccepted], len(bd))
	for i, dataBatch := range bd {
		msgs := make([]Message[struct{}, *mockMsgAccepted], len(dataBatch))
		for j, d := range dataBatch {
			msg := mockinterfaces.NewMockeryMessage[struct{}, *mockMsgAccepted](s.T())
			msg.EXPECT().Accept(mock.Anything, mock.Anything).Return(&mockMsgAccepted{id: d.id}, nil).Maybe()
			msgs[j] = msg
		}
		msgBatches[i] = msgs
	}

	batchIdx := 0
	source := mockinterfaces.NewMockerySource[struct{}, *mockMsgAccepted](s.T())
	source.EXPECT().Next(mock.Anything).RunAndReturn(func(_ context.Context) ([]Message[struct{}, *mockMsgAccepted], bool, error) {
		if err := failed.isFail(); err != nil {
			return nil, false, err
		}
		if batchIdx >= len(msgBatches) {
			return nil, false, nil
		}
		batch := msgBatches[batchIdx]
		batchIdx++
		return batch, true, nil
	})

	sink := mockinterfaces.NewMockerySink[*mockMsgAccepted](s.T())
	sink.EXPECT().Drain(mock.Anything, mock.Anything).Return(nil).Maybe()

	stream := New(source, struct{}{}, _ROUTINES, sink)
	err := stream.Run(context.Background())
	s.Require().Error(err)

	// The error may be a joined error containing both CodeSource and CodeAccept
	// when concurrent workers hit Accept errors after the source fails.
	// Verify that CodeSource is present somewhere in the error chain.
	s.Require().True(
		errors.Is(err, errs.CodeSource.From(errors.New("sentinel"))),
		"expected CodeSource in error chain, got: %v", err,
	)
}

func (s *TestSuiteStreamer) TestFailOnAccept() {
	bd := newBatchedData(_TOTAL, _BATCH)
	failed := newMockFailed(_FAILED)

	// Pre-create all message mocks with failing Accept behavior.
	msgBatches := make([][]Message[struct{}, *mockMsgAccepted], len(bd))
	for i, dataBatch := range bd {
		msgs := make([]Message[struct{}, *mockMsgAccepted], len(dataBatch))
		for j, d := range dataBatch {
			msg := mockinterfaces.NewMockeryMessage[struct{}, *mockMsgAccepted](s.T())
			msg.EXPECT().Accept(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ struct{}) (*mockMsgAccepted, error) {
				if err := failed.isFail(); err != nil {
					return nil, err
				}
				return &mockMsgAccepted{id: d.id}, nil
			}).Maybe()
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
		batch := msgBatches[batchIdx]
		batchIdx++
		return batch, true, nil
	})

	sink := mockinterfaces.NewMockerySink[*mockMsgAccepted](s.T())
	sink.EXPECT().Drain(mock.Anything, mock.Anything).Return(nil).Maybe()

	stream := New(source, struct{}{}, _ROUTINES, sink)
	err := stream.Run(context.Background())
	s.Require().Error(err)
	e, ok := errorx.From(err)
	s.Require().True(ok)
	s.Require().Equal(errs.CodeAccept, e.Code())
}

func (s *TestSuiteStreamer) TestFailOnDrain() {
	bd := newBatchedData(_TOTAL, _BATCH)
	failed := newMockFailed(_FAILED)

	// Pre-create all message mocks.
	msgBatches := make([][]Message[struct{}, *mockMsgAccepted], len(bd))
	for i, dataBatch := range bd {
		msgs := make([]Message[struct{}, *mockMsgAccepted], len(dataBatch))
		for j, d := range dataBatch {
			msg := mockinterfaces.NewMockeryMessage[struct{}, *mockMsgAccepted](s.T())
			msg.EXPECT().Accept(mock.Anything, mock.Anything).Return(&mockMsgAccepted{id: d.id}, nil).Maybe()
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
		batch := msgBatches[batchIdx]
		batchIdx++
		return batch, true, nil
	}).Maybe()

	sink := mockinterfaces.NewMockerySink[*mockMsgAccepted](s.T())
	sink.EXPECT().Drain(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ ...*mockMsgAccepted) error {
		return failed.isFail()
	})

	stream := New(source, struct{}{}, _ROUTINES, sink)
	err := stream.Run(context.Background())
	s.Require().Error(err)
	e, ok := errorx.From(err)
	s.Require().True(ok)
	s.Require().Equal(errs.CodeDrain, e.Code())
}

func (s *TestSuiteStreamer) TestFailOnCtx() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	source := mockinterfaces.NewMockerySource[struct{}, int](s.T())
	sink := mockinterfaces.NewMockerySink[int](s.T())

	stream := New[struct{}, int](source, struct{}{}, _ROUTINES, sink)
	err := stream.Run(ctx)
	s.Require().Error(err)
	s.Require().ErrorIs(err, context.Canceled)
}

type mockFailed struct {
	count  atomic.Int32
	failed int
}

func newMockFailed(failed int) *mockFailed {
	return &mockFailed{failed: failed}
}

func (f *mockFailed) isFail() error {
	n := int(f.count.Add(1)) - 1
	if n >= f.failed {
		return fmt.Errorf("failed on %d", n)
	}
	return nil
}
