package impl

import (
	"context"
	"testing"
	"time"
)

var ctx = context.Background()

func BenchmarkStreamRoutines(b *testing.B) {
	fn := func(b *testing.B, total int, batch, routines int) {
		benchmarkStream(b, total, batch, routines, time.Millisecond, 100*time.Microsecond, time.Millisecond)
	}
	b.Run("100w", func(b *testing.B) {
		fn := func(b *testing.B, batch, routines int) {
			fn(b, 1_000_000, batch, routines)
		}
		b.Run("200", func(b *testing.B) {
			fn := func(b *testing.B, routines int) {
				fn(b, 200, routines)
			}
			b.Run("25", func(b *testing.B) { fn(b, 25) })
			b.Run("50", func(b *testing.B) { fn(b, 50) })
			b.Run("100", func(b *testing.B) { fn(b, 100) })
		})
	})
}

func BenchmarkStreamLatency(b *testing.B) {
	fn := func(b *testing.B, sourceT, msgT, sinkT time.Duration) {
		benchmarkStream(b, 1_000_000, 200, 25, sourceT, msgT, sinkT)
	}
	b.Run("1ms", func(b *testing.B) {
		fn := func(b *testing.B, msgT, sinkT time.Duration) {
			fn(b, time.Millisecond, msgT, sinkT)
		}
		b.Run("100us", func(b *testing.B) {
			fn := func(b *testing.B, sinkT time.Duration) {
				fn(b, 100*time.Microsecond, sinkT)
			}
			b.Run("1ms", func(b *testing.B) { fn(b, time.Millisecond) })
		})
	})
	b.Run("100us", func(b *testing.B) {
		fn := func(b *testing.B, msgT, sinkT time.Duration) {
			fn(b, 100*time.Microsecond, msgT, sinkT)
		}
		b.Run("1ms", func(b *testing.B) {
			fn := func(b *testing.B, sinkT time.Duration) {
				fn(b, time.Millisecond, sinkT)
			}
			b.Run("1ms", func(b *testing.B) { fn(b, time.Millisecond) })
		})
	})
	b.Run("1ms", func(b *testing.B) {
		fn := func(b *testing.B, msgT, sinkT time.Duration) {
			fn(b, time.Millisecond, msgT, sinkT)
		}
		b.Run("1ms", func(b *testing.B) {
			fn := func(b *testing.B, sinkT time.Duration) {
				fn(b, time.Millisecond, sinkT)
			}
			b.Run("100us", func(b *testing.B) { fn(b, 100*time.Microsecond) })
		})
	})
}

func benchmarkStream(b *testing.B, total, batch, routines int, sourceT, msgT, sinkT time.Duration) {
	source := newMockSleepSource(total, batch, sourceT, msgT)
	visitor := struct{}{}
	sink := newMockSleepSink(sinkT)

	stream := New(source, visitor, routines, sink)

	b.ResetTimer()
	if err := stream.Run(ctx); err != nil {
		b.Fatal(err)
	}
}

// mockSleepMessage - message with sleep in Accept
type mockSleepMessage struct {
	data     *mockMsg
	duration time.Duration
}

func (m *mockSleepMessage) Accept(_ context.Context, _ struct{}) (*mockMsgAccepted, error) {
	time.Sleep(m.duration)
	return &mockMsgAccepted{id: m.data.id}, nil
}

// mockSleepSource - source with sleep in Next, produces sleep messages
type mockSleepSource struct {
	i              int
	datas          [][]*mockMsg
	sourceDuration time.Duration
	acceptDuration time.Duration
}

func newMockSleepSource(total, batch int, sourceDuration, acceptDuration time.Duration) *mockSleepSource {
	return &mockSleepSource{
		datas:          newBatchedData(total, batch),
		sourceDuration: sourceDuration,
		acceptDuration: acceptDuration,
	}
}

func (s *mockSleepSource) Next(_ context.Context) ([]Message[struct{}, *mockMsgAccepted], bool, error) {
	time.Sleep(s.sourceDuration)
	if s.i >= len(s.datas) {
		return nil, false, nil
	}
	batch := s.datas[s.i]
	s.i++
	msgs := make([]Message[struct{}, *mockMsgAccepted], len(batch))
	for i, d := range batch {
		msgs[i] = &mockSleepMessage{data: d, duration: s.acceptDuration}
	}
	return msgs, true, nil
}

// mockSleepSink - sink with sleep in Drain
type mockSleepSink struct {
	duration time.Duration
}

func newMockSleepSink(duration time.Duration) *mockSleepSink {
	return &mockSleepSink{duration: duration}
}

func (s *mockSleepSink) Drain(_ context.Context, _ ...*mockMsgAccepted) error {
	time.Sleep(s.duration)
	return nil
}
