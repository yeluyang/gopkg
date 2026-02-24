package stream

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yeluyang/gopkg/stream/interfaces"
	mockinterfaces "github.com/yeluyang/gopkg/stream/interfaces/mocks"
)

func TestBuilder(t *testing.T) {
	suite.Run(t, new(BuilderSuite))
}

type BuilderSuite struct {
	suite.Suite
}

func (s *BuilderSuite) TestBuildMissingSink() {
	source := mockinterfaces.NewMockerySource[struct{}, int](s.T())
	_, err := From[struct{}, int](source).Build()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "sink")
}

func (s *BuilderSuite) TestBuildValid() {
	source := mockinterfaces.NewMockerySource[struct{}, int](s.T())
	sink := mockinterfaces.NewMockerySink[int](s.T())
	stream, err := From[struct{}, int](source).
		Visitor(struct{}{}).
		To(sink).
		Build()
	s.Require().NoError(err)
	s.Require().NotNil(stream)
}

func (s *BuilderSuite) TestBuildDefaultRoutines() {
	source, sink, results := s.setupPipeline()
	stream, err := From[struct{}, int](source).
		Visitor(struct{}{}).
		To(sink).
		Build()
	s.Require().NoError(err)

	s.Require().NoError(stream.Run(context.Background()))
	s.Require().Len(*results, 1)
	s.Require().Equal(1, (*results)[0])
}

func (s *BuilderSuite) TestBuildRoutinesClamp() {
	source, sink, results := s.setupPipeline()
	stream, err := From[struct{}, int](source).
		Visitor(struct{}{}).
		To(sink).
		Routines(0).
		Build()
	s.Require().NoError(err)

	s.Require().NoError(stream.Run(context.Background()))
	s.Require().Len(*results, 1)
	s.Require().Equal(1, (*results)[0])
}

func (s *BuilderSuite) setupPipeline() (
	*mockinterfaces.MockerySource[struct{}, int],
	*mockinterfaces.MockerySink[int],
	*[]int,
) {
	msg := mockinterfaces.NewMockeryMessage[struct{}, int](s.T())
	msg.EXPECT().Accept(mock.Anything, struct{}{}).Return(1, nil).Once()

	source := mockinterfaces.NewMockerySource[struct{}, int](s.T())
	source.EXPECT().Next(mock.Anything).
		Return([]interfaces.Message[struct{}, int]{msg}, true, nil).Once()
	source.EXPECT().Next(mock.Anything).
		Return(nil, false, nil).Once()

	results := &[]int{}
	sink := mockinterfaces.NewMockerySink[int](s.T())
	sink.EXPECT().Drain(mock.Anything, mock.Anything).
		Run(func(_ context.Context, msgs ...int) {
			*results = append(*results, msgs...)
		}).
		Return(nil)

	return source, sink, results
}
