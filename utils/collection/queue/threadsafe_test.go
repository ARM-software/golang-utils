package queue

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/safecast"
)

func TestThreadSafeQueue(t *testing.T) {
	const (
		enqueuers   = 16
		dequeuers   = 16
		perEnqueuer = 10_000
		total       = enqueuers * perEnqueuer
		repetitions = 50
	)

	for rep := 0; rep < repetitions; rep++ {
		q := NewThreadSafeQueue[int]()

		var id uint64
		var wgEnq sync.WaitGroup
		wgEnq.Add(enqueuers)

		for e := 0; e < enqueuers; e++ {
			go func() {
				defer wgEnq.Done()
				for i := 0; i < perEnqueuer; i++ {
					v := safecast.ToInt(atomic.AddUint64(&id, 1))
					q.Enqueue(v)
				}
			}()
		}
		wgEnq.Wait()

		assert.False(t, q.IsEmpty())
		assert.Equal(t, total, q.Len())

		var remaining int64 = total
		seen := make(map[int]struct{}, total)
		var mu sync.Mutex

		var wgDeq sync.WaitGroup
		wgDeq.Add(dequeuers)

		for i := 0; i < dequeuers; i++ {
			go func() {
				defer wgDeq.Done()
				for {
					n := atomic.AddInt64(&remaining, -1)
					if n < 0 {
						return // done
					}
					v := q.Dequeue()

					mu.Lock()
					if _, exists := seen[v]; exists {
						mu.Unlock()
						require.Fail(t, "dequeued duplicate value", "rep", rep, "value", v)
						return
					}
					seen[v] = struct{}{}
					mu.Unlock()
				}
			}()
		}
		wgDeq.Wait()

		assert.Len(t, seen, total)
		assert.True(t, q.IsEmpty())
	}
}
