package routine

import (
	"runtime"
	"sync"
	"testing"
)

// BenchmarkGo vs bare go: measure the overhead of runtime.Callers + closure escape
func BenchmarkGo(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(b.N)
	b.ResetTimer()
	for range b.N {
		Go(func() { wg.Done() })
	}
	wg.Wait()
}

func BenchmarkBareGo(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(b.N)
	b.ResetTimer()
	for range b.N {
		go func() { wg.Done() }()
	}
	wg.Wait()
}

// Isolate runtime.Callers cost
func BenchmarkRuntimeCallers(b *testing.B) {
	for range b.N {
		var pcs [32]uintptr
		runtime.Callers(2, pcs[:])
	}
}
