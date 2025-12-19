package stack

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {

	tests := []struct {
		details     string
		constructor func() IStack[int]
	}{
		{
			details:     "unsafe stack",
			constructor: NewStack[int],
		},
		{
			details:     "thread stack",
			constructor: NewThreadSafeStack[int],
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.details, func(t *testing.T) {
			t.Run("new stack is empty", func(t *testing.T) {
				s := test.constructor()
				assert.Zero(t, s.Len())
				assert.True(t, s.IsEmpty())
				assert.Zero(t, s.Pop())
				assert.Zero(t, s.Peek())
			})

			t.Run("push then peek does not remove", func(t *testing.T) {
				s := test.constructor()
				s.Push(1)
				assert.False(t, s.IsEmpty())
				assert.Equal(t, 1, s.Len())
				assert.Equal(t, 1, s.Peek())
				assert.False(t, s.IsEmpty())
				assert.Equal(t, 1, s.Len())
			})

			t.Run("push then pop removes (LIFO)", func(t *testing.T) {
				s := test.constructor()
				s.Push(1)
				assert.False(t, s.IsEmpty())
				assert.Equal(t, 1, s.Len())
				assert.Equal(t, 1, s.Pop())
				assert.True(t, s.IsEmpty())
			})

			t.Run("multiple push and pop", func(t *testing.T) {
				s := test.constructor()
				s.Push(1, 2, 3, 4)
				assert.False(t, s.IsEmpty())
				assert.Equal(t, 4, s.Len())
				assert.Equal(t, 4, s.Pop())
				assert.Equal(t, 3, s.Pop())
				assert.Equal(t, 2, s.Pop())
				assert.Equal(t, 1, s.Pop())
				assert.True(t, s.IsEmpty())
			})

			t.Run("various actions", func(t *testing.T) {
				s := test.constructor()
				assert.Zero(t, s.Len())
				s.Push(1)

				assert.Equal(t, 1, s.Len())
				assert.Equal(t, 1, s.Peek())
				assert.Equal(t, 1, s.Pop())
				assert.Zero(t, s.Len())
				s.Push(1)
				s.Push(2)

				assert.Equal(t, 2, s.Len())
				assert.Equal(t, 2, s.Peek())
			})

			t.Run("clear empties the stack", func(t *testing.T) {
				s := test.constructor()
				s.Push(1, 1, 1, 1)
				assert.False(t, s.IsEmpty())
				s.Clear()
				assert.True(t, s.IsEmpty())
			})

			t.Run("Clear then reuse", func(t *testing.T) {
				s := test.constructor()
				s.Push(10)
				s.Push(20)

				s.Clear()
				assert.True(t, s.IsEmpty())

				s.Push(30)
				assert.False(t, s.IsEmpty())
				assert.Equal(t, 30, s.Peek())
			})

			t.Run("values drains the stack", func(t *testing.T) {
				s := test.constructor()
				s.Push(1, 2, 3, 4)
				assert.False(t, s.IsEmpty())
				values := slices.Collect(s.Values())
				assert.True(t, s.IsEmpty())
				assert.Len(t, values, 4)
				assert.Equal(t, []int{4, 3, 2, 1}, values)
			})

		})
	}

}
