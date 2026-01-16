# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

This is a multi-module Go monorepo. Each package is an independent module with its own `go.mod`.

```bash
# Test all packages from root
go test ./...

# Test a specific package
go test ./contextual/...
go test ./rate/...
go test ./fxdecorate/...
go test ./routine/...
go test ./utils/...

# Run a single test
go test ./rate/... -run TestDynamicLimiter

# Build/verify all packages
go build ./...
```

## Architecture

This repository contains independent, reusable Go utility packages:

- **contextual** - Type-safe context value helpers using generics. `New[V]()` returns a pair of `With` (setter) and `From` (getter) functions that avoid string key collisions through pointer-based keys.

- **fxdecorate** - Decorator pattern integration for Uber's fx dependency injection. Uses fx groups to collect decorators and apply them in sequence. See `fxdecorate/example/main.go` for usage pattern.

- **rate** - Dynamic rate limiter that wraps `golang.org/x/time/rate` with periodic limit refresh from a callback function.

- **routine** - Panic-safe goroutine launcher that recovers panics and logs them with stack traces.

- **utils** - Miscellaneous utilities including `ForcePanic` for unrecoverable errors.

## Dependencies

- `rate` depends on `routine` (for async limiter refresh)
- `fxdecorate` depends on `go.uber.org/fx`
- Other packages are standalone
