package main

import (
	"fmt"
	"os"

	"github.com/yeluyang/gopkg/fxdecorate"
	"go.uber.org/fx"
)

type (
	anyOtherDependence struct{}

	Foo           interface{ Do() }
	fooImpl       struct{}
	fooDecoratorA struct {
		i Foo
		d *anyOtherDependence
	}
	fooDecoratorB struct {
		i Foo
		d *anyOtherDependence
	}

	Bar           interface{ Magic() }
	barImpl       struct{}
	barDecoratorA struct {
		i Bar
		d *anyOtherDependence
	}
	barDecoratorB struct {
		i Bar
		d *anyOtherDependence
	}
)

func newFooImpl() *fooImpl {
	return &fooImpl{}
}

func (*fooImpl) Do() {
	fmt.Println("foo")
}

func newFooDecoratorA(d *anyOtherDependence) fxdecorate.Result[Foo] {
	return fxdecorate.With(func(f Foo) Foo {
		return &fooDecoratorA{
			i: f,
			d: d,
		}
	})
}

func (d *fooDecoratorA) Do() {
	fmt.Println("foo::decorator::A::start")
	d.i.Do()
	fmt.Println("foo::decorator::A::exit")
}

func newFooDecoratorB(d *anyOtherDependence) fxdecorate.Result[Foo] {
	return fxdecorate.With(func(f Foo) Foo {
		return &fooDecoratorB{
			i: f,
			d: d,
		}
	})
}

func (d *fooDecoratorB) Do() {
	fmt.Println("foo::decorator::B::start")
	d.i.Do()
	fmt.Println("foo::decorator::B::exit")
}

func newBarImpl() *barImpl {
	return &barImpl{}
}

func (*barImpl) Magic() {
	fmt.Println("bar")
}

func newBarDecoratorA(d *anyOtherDependence) fxdecorate.Result[Bar] {
	return fxdecorate.With(func(f Bar) Bar {
		return &barDecoratorA{
			i: f,
			d: d,
		}
	})
}

func (d *barDecoratorA) Magic() {
	fmt.Println("bar::decorator::A::start")
	d.i.Magic()
	fmt.Println("bar::decorator::A::exit")
}

func newBarDecoratorB(d *anyOtherDependence) fxdecorate.Result[Bar] {
	return fxdecorate.With(func(f Bar) Bar {
		return &barDecoratorB{
			i: f,
			d: d,
		}
	})
}

func (d *barDecoratorB) Magic() {
	fmt.Println("bar::decorator::B::start")
	d.i.Magic()
	fmt.Println("bar::decorator::B::exit")
}

func main() {
	fx.New(
		fx.Provide(func() *anyOtherDependence { return &anyOtherDependence{} }),

		fx.Module("foo",
			fx.Module("impl",
				fx.Provide(fx.Annotate(newFooImpl, fx.As(new(Foo)))),
			),
			fx.Module("decorate",
				fx.Module("A",
					fx.Provide(newFooDecoratorA),
				),
				fx.Module("B",
					fx.Provide(newFooDecoratorB),
				),
			),
		),

		fx.Module("bar",
			fx.Module("impl",
				fx.Provide(fx.Annotate(newBarImpl, fx.As(new(Bar)))),
			),
			fx.Module("decorate",
				fx.Module("A",
					fx.Provide(newBarDecoratorA),
				),
				fx.Module("B",
					fx.Provide(newBarDecoratorB),
				),
			),
		),

		fx.Decorate(fxdecorate.New[Foo]),
		fx.Decorate(fxdecorate.New[Bar]),

		fx.Invoke(func(f Foo, b Bar) {
			f.Do()
			fmt.Println("================")
			b.Magic()
			os.Exit(0)
		}),
	).Run()
}
