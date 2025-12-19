package stack

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/safecast"
)

func TestThreadSafeStack(t *testing.T) {
	const (
		pushers     = 16
		poppers     = 16
		perPusher   = 10_000
		total       = pushers * perPusher
		repetitions = 50
	)

	for rep := 0; rep < repetitions; rep++ {
		s := NewThreadSafeStack[int]()

		var id uint64
		var wgPush sync.WaitGroup
		wgPush.Add(pushers)

		for p := 0; p < pushers; p++ {
			go func() {
				defer wgPush.Done()
				for i := 0; i < perPusher; i++ {
					v := safecast.ToInt(atomic.AddUint64(&id, 1))
					s.Push(v)
				}
			}()
		}
		wgPush.Wait()

		assert.False(t, s.IsEmpty())
		assert.Equal(t, total, s.Len())

		var remaining int64 = total
		seen := make(map[int]struct{}, total)
		var mu sync.Mutex

		var wgPop sync.WaitGroup
		wgPop.Add(poppers)

		for i := 0; i < poppers; i++ {
			go func() {
				defer wgPop.Done()
				for {
					n := atomic.AddInt64(&remaining, -1)
					if n < 0 {
						return // done
					}
					v := s.Pop()

					mu.Lock()
					if _, exists := seen[v]; exists {
						mu.Unlock()
						require.Fail(t, "poped duplicate value", "rep", rep, "value", v)
						return
					}
					seen[v] = struct{}{}
					mu.Unlock()
				}
			}()
		}
		wgPop.Wait()

		assert.Len(t, seen, total)
		assert.True(t, s.IsEmpty())
	}
}
