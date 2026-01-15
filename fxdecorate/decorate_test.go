package fxdecorate

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestFxDecorate(t *testing.T) {
	suite.Run(t, new(TestSuiteFxDecorate))
}

type TestSuiteFxDecorate struct {
	suite.Suite
}

type Counter interface {
	Value() int
	Add(int)
}

type counterImpl struct {
	value int
}

func (c *counterImpl) Value() int { return c.value }
func (c *counterImpl) Add(v int)  { c.value += v }

func newCounterImpl() *counterImpl {
	return &counterImpl{value: 0}
}

func (s *TestSuiteFxDecorate) TestWithNoDecorators() {
	var counter Counter

	app := fxtest.New(s.T(),
		fx.Provide(fx.Annotate(newCounterImpl, fx.As(new(Counter)))),
		fx.Decorate(New[Counter]),
		fx.Populate(&counter),
	)

	app.RequireStart()
	defer app.RequireStop()

	s.NotNil(counter)
	s.Equal(0, counter.Value())
}

func (s *TestSuiteFxDecorate) TestWithSingleDecorator() {
	var counter Counter

	app := fxtest.New(s.T(),
		fx.Provide(fx.Annotate(newCounterImpl, fx.As(new(Counter)))),
		fx.Provide(func() Result[Counter] {
			return With(func(c Counter) Counter {
				c.Add(10)
				return c
			})
		}),
		fx.Decorate(New[Counter]),
		fx.Populate(&counter),
	)

	app.RequireStart()
	defer app.RequireStop()

	s.NotNil(counter)
	s.Equal(10, counter.Value())
}

func (s *TestSuiteFxDecorate) TestWithMultipleDecorators() {
	var counter Counter

	app := fxtest.New(s.T(),
		fx.Provide(fx.Annotate(newCounterImpl, fx.As(new(Counter)))),
		fx.Provide(func() Result[Counter] {
			return With(func(c Counter) Counter {
				c.Add(10)
				return c
			})
		}),
		fx.Provide(func() Result[Counter] {
			return With(func(c Counter) Counter {
				c.Add(5)
				return c
			})
		}),
		fx.Provide(func() Result[Counter] {
			return With(func(c Counter) Counter {
				c.Add(3)
				return c
			})
		}),
		fx.Decorate(New[Counter]),
		fx.Populate(&counter),
	)

	app.RequireStart()
	defer app.RequireStop()

	s.NotNil(counter)
	s.Equal(18, counter.Value())
}

func (s *TestSuiteFxDecorate) TestDecoratorCanWrapImplementation() {
	type wrappedCounter struct {
		Counter
		calls int
	}

	var counter Counter

	app := fxtest.New(s.T(),
		fx.Provide(fx.Annotate(newCounterImpl, fx.As(new(Counter)))),
		fx.Provide(func() Result[Counter] {
			return With(func(c Counter) Counter {
				return &wrappedCounter{Counter: c, calls: 0}
			})
		}),
		fx.Decorate(New[Counter]),
		fx.Populate(&counter),
	)

	app.RequireStart()
	defer app.RequireStop()

	_, ok := counter.(*wrappedCounter)
	s.True(ok)
}

func (s *TestSuiteFxDecorate) TestDecoratorsFromModule() {
	var counter Counter

	decoratorModule := fx.Module("decorators",
		fx.Provide(func() Result[Counter] {
			return With(func(c Counter) Counter {
				c.Add(100)
				return c
			})
		}),
	)

	app := fxtest.New(s.T(),
		fx.Provide(fx.Annotate(newCounterImpl, fx.As(new(Counter)))),
		decoratorModule,
		fx.Decorate(New[Counter]),
		fx.Populate(&counter),
	)

	app.RequireStart()
	defer app.RequireStop()

	s.Equal(100, counter.Value())
}
