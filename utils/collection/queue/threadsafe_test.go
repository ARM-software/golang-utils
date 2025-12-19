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

	tests := []struct {
		details     string
		constructor func() IQueue[int]
	}{
		{
			details:     "mutex-based thread-safe queue",
			constructor: NewThreadSafeQueue[int],
		},
		{
			details: "channel-based queue",
			// Capacity must be >= total because we enqueue everything before we start dequeuing.
			constructor: func() IQueue[int] {
				c, err := NewChannelQueue[int](total)
				require.NoError(t, err)
				return c
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.details, func(t *testing.T) {
			for rep := 0; rep < repetitions; rep++ {
				q := tc.constructor()

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

				var popped int64
				seen := make(map[int]struct{}, total)
				var mu sync.Mutex

				var wgDeq sync.WaitGroup
				wgDeq.Add(dequeuers)

				for i := 0; i < dequeuers; i++ {
					go func() {
						defer wgDeq.Done()

						for {
							// Stop once we've successfully dequeued total items.
							if atomic.LoadInt64(&popped) >= int64(total) {
								return
							}

							v := q.Dequeue()
							if v == 0 {
								// Non-blocking empty read; retry.
								continue
							}

							// Count this as a successful dequeue (cap at total to avoid overshoot races).
							n := atomic.AddInt64(&popped, 1)
							if n > int64(total) {
								return
							}

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

				assert.Equal(t, int64(total), atomic.LoadInt64(&popped), "rep=%d", rep)
				assert.Len(t, seen, total, "rep=%d", rep)
				assert.True(t, q.IsEmpty(), "rep=%d", rep)
				assert.Equal(t, 0, q.Len(), "rep=%d", rep)
			}
		})
	}
}
