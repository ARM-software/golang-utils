package warden

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestWardenWaitCleanShutdown(t *testing.T) {
	w := New()
	require.True(t, w.Alive())

	require.NoError(t, w.Go(func() error { return nil }))
	require.NoError(t, w.Wait())
	assert.False(t, w.Alive())
	assert.NoError(t, w.Err())
}

func TestWardenGoErrorKillsState(t *testing.T) {
	w := New()
	expected := errors.New("boom")

	require.NoError(t, w.Go(func() error { return expected }))
	err := w.Wait()
	require.Error(t, err)
	errortest.AssertError(t, err, expected)
}

func TestWardenContextCancelledOnKill(t *testing.T) {
	w := New()
	ctx := w.Context(context.Background())

	require.NoError(t, w.Kill(context.Canceled))
	require.Eventually(t, func() bool {
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
}

func TestWardenContextReusesChildForSameParent(t *testing.T) {
	w := New()
	parent := context.Background()

	ctx1 := w.Context(parent)
	ctx2 := w.Context(parent)

	assert.Same(t, ctx1, ctx2)
}

func TestWardenContextCreatedAfterKillIsAlreadyCancelled(t *testing.T) {
	w := New()
	require.NoError(t, w.Kill(context.Canceled))

	ctx := w.Context(context.Background())

	require.Eventually(t, func() bool {
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
	errortest.AssertError(t, ctx.Err(), context.Canceled)
}

func TestWardenWithContextParentCancellationKillsState(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := WithContext(parent)
	require.NoError(t, w.Go(func() error {
		<-w.Dying()
		return nil
	}))

	cancel()
	err := w.Wait()
	require.Error(t, err)
	errortest.AssertError(t, err, context.Canceled)
}

func TestWardenGoAfterDeathFails(t *testing.T) {
	w := New()
	require.NoError(t, w.Go(func() error { return nil }))
	require.NoError(t, w.Wait())
	errortest.RequireError(t, w.Go(func() error { return nil }), ErrSpawnAfterDeath)
}

func TestGracefulWrapperKeepsDyingOpen(t *testing.T) {
	base := New()
	g := NewGracefulWrapper(base)

	require.NoError(t, g.Go(func() error {
		<-base.Dying()
		return nil
	}))
	require.NoError(t, g.Kill(nil))
	assert.Never(t, func() bool {
		select {
		case <-g.Dying():
			return true
		default:
			return false
		}
	}, 100*time.Millisecond, 10*time.Millisecond)

	require.NoError(t, g.Wait())
	require.Eventually(t, func() bool {
		select {
		case <-g.Dead():
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
}
