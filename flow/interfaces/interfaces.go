package interfaces

import (
	"context"
)

type Option interface{ apply(*MsgConfig) }

type MsgConfig struct {
	Serial    bool
	SerialKey string
}

type OptionFunc func(*MsgConfig)

func (f OptionFunc) apply(c *MsgConfig) { f(c) }

func WithSerial() Option {
	return OptionFunc(func(c *MsgConfig) { c.Serial = true })
}

func WithSerialKey(key string) Option {
	return OptionFunc(func(c *MsgConfig) { c.Serial = true; c.SerialKey = key })
}

func ResolveOptions(opts []Option) MsgConfig {
	var cfg MsgConfig
	for _, o := range opts {
		o.apply(&cfg)
	}
	return cfg
}

type (
	Message[E comparable, V any] interface {
		Activate(ctx context.Context) []E
		DrainTo(ctx context.Context) []E
		Accept(ctx context.Context, visitor V) ([]Message[E, V], error)
		Options() []Option
	}

	Source[E comparable, V any] interface {
		Endpoint[E]
		inlet[Message[E, V]]
	}

	Sink[E comparable, V any] interface {
		Endpoint[E]
		outlet[Message[E, V]]
	}

	Duplex[E comparable, V any] interface {
		Endpoint[E]
		inlet[Message[E, V]]
		outlet[Message[E, V]]
	}
)

type (
	Endpoint[ID comparable] interface {
		ID() ID
		Activate(ctx context.Context) error
		Close(ctx context.Context) error
	}
	inlet[Msg any] interface {
		Next(ctx context.Context) ([]Msg, bool, error)
	}
	outlet[Msg any] interface {
		Drain(ctx context.Context, msg []Msg) error
	}
)
