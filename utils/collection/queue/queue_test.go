package queue

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	tests := []struct {
		details     string
		constructor func() IQueue[int]
	}{
		{
			details:     "unsafe queue",
			constructor: NewQueue[int],
		},
		{
			details:     "thread queue",
			constructor: NewThreadSafeQueue[int],
		},
		{
			details: "channel-based queue",
			// Capacity must be >= total because we enqueue everything before we start dequeuing.
			constructor: func() IQueue[int] {
				c, err := NewChannelQueue[int](10)
				require.NoError(t, err)
				return c
			},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.details, func(t *testing.T) {
			t.Run("new queue is empty", func(t *testing.T) {
				q := test.constructor()
				assert.Zero(t, q.Len())
				assert.True(t, q.IsEmpty())
				assert.Zero(t, q.Dequeue())
				assert.Zero(t, q.Peek())
			})

			t.Run("enqueue then peek does not remove", func(t *testing.T) {
				q := test.constructor()
				q.Enqueue(1)
				assert.False(t, q.IsEmpty())
				assert.Equal(t, 1, q.Len())
				assert.Equal(t, 1, q.Peek())
				assert.False(t, q.IsEmpty())
				assert.Equal(t, 1, q.Len())
			})

			t.Run("enqueue then dequeue removes (FIFO)", func(t *testing.T) {
				q := test.constructor()
				q.Enqueue(1)
				assert.False(t, q.IsEmpty())
				assert.Equal(t, 1, q.Len())
				assert.Equal(t, 1, q.Dequeue())
				assert.True(t, q.IsEmpty())
			})

			t.Run("multiple enqueue and dequeue", func(t *testing.T) {
				q := test.constructor()
				q.Enqueue(1, 2, 3, 4)
				assert.False(t, q.IsEmpty())
				assert.Equal(t, 4, q.Len())
				assert.Equal(t, 1, q.Dequeue())
				assert.Equal(t, 2, q.Dequeue())
				assert.Equal(t, 3, q.Dequeue())
				assert.Equal(t, 4, q.Dequeue())
				assert.True(t, q.IsEmpty())
			})

			t.Run("various actions", func(t *testing.T) {
				q := test.constructor()
				assert.Zero(t, q.Len())
				q.Enqueue(1)

				assert.Equal(t, 1, q.Len())
				assert.Equal(t, 1, q.Peek())
				assert.Equal(t, 1, q.Dequeue())
				assert.Zero(t, q.Len())
				q.Enqueue(1)
				q.Enqueue(2)

				assert.Equal(t, 2, q.Len())
				assert.Equal(t, 1, q.Peek())
			})

			t.Run("clear empties the queue", func(t *testing.T) {
				q := test.constructor()
				q.Enqueue(1, 1, 1, 1)
				assert.False(t, q.IsEmpty())
				q.Clear()
				assert.True(t, q.IsEmpty())
			})

			t.Run("Clear then reuse", func(t *testing.T) {
				q := test.constructor()
				q.Enqueue(10)
				q.Enqueue(20)

				q.Clear()
				assert.True(t, q.IsEmpty())

				q.Enqueue(30)
				assert.False(t, q.IsEmpty())
				assert.Equal(t, 30, q.Peek())
			})

			t.Run("values drains the queue", func(t *testing.T) {
				q := test.constructor()
				q.Enqueue(1, 2, 3, 4)
				assert.False(t, q.IsEmpty())
				values := slices.Collect(q.Values())
				assert.True(t, q.IsEmpty())
				assert.Len(t, values, 4)
				assert.Equal(t, []int{1, 2, 3, 4}, values) // FIFO drain
			})
		})
	}
}
