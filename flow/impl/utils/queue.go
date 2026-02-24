package utils

import "sync"

// Queue is a concurrent-safe FIFO queue backed by a mutex and slice.
type Queue[T any] struct {
	mu    sync.Mutex
	items []T
}

func (q *Queue[T]) Push(items []T) {
	q.mu.Lock()
	q.items = append(q.items, items...)
	q.mu.Unlock()
}

func (q *Queue[T]) Pop() (item T, ok bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return
	}
	item = q.items[0]
	q.items[0] = *new(T) // allow GC
	q.items = q.items[1:]
	ok = true
	return
}
