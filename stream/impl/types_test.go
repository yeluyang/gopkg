package impl

import (
	"slices"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	_TOTAL    = 500
	_BATCH    = 10
	_ROUTINES = 25
	_FAILED   = 26
)

type mockMsg struct {
	id       int
	accessed int32
}

func (d *mockMsg) access(t *testing.T) {
	require.Equal(t, int32(1), atomic.AddInt32(&d.accessed, 1), "race")
}

type mockMsgAccepted struct {
	id int
}

// newBatchedData generates mockMsg items chunked into batches for tests and benchmarks.
func newBatchedData(total, batch int) [][]*mockMsg {
	datas := make([]*mockMsg, total)
	for i := range total {
		datas[i] = &mockMsg{id: i}
	}
	return slices.Collect(slices.Chunk(datas, batch))
}
