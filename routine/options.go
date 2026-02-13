package routine

// Option configures Wrap behavior.
type Option func(*config)

type config struct {
	callerSkip    int
	noCallerStack bool
	noPanicStack  bool
}

// WithCallerSkip adds additional frames to skip when capturing the caller
// stack. Use this when Wrap is called through an intermediate helper, so
// the caller stack starts from the real call site instead of the helper.
func WithCallerSkip(n int) Option {
	return func(c *config) {
		c.callerSkip = n
	}
}

// WithCallerStack controls whether to capture the caller stack.
func WithCallerStack(on bool) Option {
	return func(c *config) {
		c.noCallerStack = !on
	}
}

// WithPanicStack controls whether to capture the panic stack.
func WithPanicStack(on bool) Option {
	return func(c *config) {
		c.noPanicStack = !on
	}
}
