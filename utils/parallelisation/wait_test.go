package parallelisation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

type mockErrorWaiter struct {
	waitFunc func() error
}

func (m *mockErrorWaiter) Wait() error {
	if m.waitFunc != nil {
		return m.waitFunc()
	}
	return nil
}

type mockWaiter struct {
	waitFunc func()
}

func (m *mockWaiter) Wait() {
	if m.waitFunc != nil {
		m.waitFunc()
	}
}

func TestWaitWithContextAndError(t *testing.T) {
	t.Run("wait completes successfully", func(t *testing.T) {
		waiter := &mockErrorWaiter{
			waitFunc: func() error {
				time.Sleep(50 * time.Millisecond)
				return nil
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err := WaitWithContextAndError(ctx, waiter)
		assert.NoError(t, err)
	})

	t.Run("wait returns error", func(t *testing.T) {
		expectedErr := commonerrors.ErrUnexpected
		waiter := &mockErrorWaiter{
			waitFunc: func() error {
				time.Sleep(10 * time.Millisecond)
				return expectedErr
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err := WaitWithContextAndError(ctx, waiter)
		errortest.AssertError(t, err, expectedErr)
	})

	t.Run("context canceled before wait returns", func(t *testing.T) {
		waiter := &mockErrorWaiter{
			waitFunc: func() error {
				time.Sleep(500 * time.Millisecond)
				return nil
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := WaitWithContextAndError(ctx, waiter)
		elapsed := time.Since(start)
		assert.Error(t, err)
		assert.Less(t, elapsed, 100*time.Millisecond) // should return almost immediately
	})

	t.Run("wait returns after context canceled, should return context error", func(t *testing.T) {
		waiter := &mockErrorWaiter{
			waitFunc: func() error {
				time.Sleep(100 * time.Millisecond)
				return nil
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err := WaitWithContextAndError(ctx, waiter)
		assert.Error(t, err)
	})
}

func TestWaitWithContext(t *testing.T) {
	t.Run("wait completes successfully", func(t *testing.T) {
		waiter := &mockWaiter{
			waitFunc: func() {
				time.Sleep(50 * time.Millisecond)
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err := WaitWithContext(ctx, waiter)
		assert.NoError(t, err)
	})

	t.Run("context canceled before wait returns", func(t *testing.T) {
		waiter := &mockWaiter{
			waitFunc: func() {
				time.Sleep(500 * time.Millisecond)
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := WaitWithContext(ctx, waiter)
		elapsed := time.Since(start)
		assert.Error(t, err)
		assert.Less(t, elapsed, 100*time.Millisecond) // should return almost immediately
	})

	t.Run("wait returns after context canceled, should return context error", func(t *testing.T) {
		waiter := &mockWaiter{
			waitFunc: func() {
				time.Sleep(100 * time.Millisecond)
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err := WaitWithContext(ctx, waiter)
		assert.Error(t, err)
	})
}
