package safeio

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContextualReadCloser(t *testing.T) {
	t.Run("Normal contextual reader blocks even after cancel", func(t *testing.T) {
		r, w, err := os.Pipe()
		require.NoError(t, err)
		defer func() { _ = r.Close(); _ = w.Close() }()

		ctx, cancel := context.WithCancel(context.Background())
		reader := NewContextualReader(ctx, r)

		done := make(chan struct{})
		go func() {
			_, _ = io.Copy(io.Discard, reader) // will block in read(2) https://man7.org/linux/man-pages/man2/read.2.html
			close(done)
		}()

		// Allow io.Copy to enter kernel read then try to cancel
		time.Sleep(50 * time.Millisecond)
		cancel()

		select {
		case <-done:
			assert.FailNow(t, "cancelling context shouldn't unblock a blocking Read in io.Copy")
		case <-time.After(200 * time.Millisecond):
			// Expected case: still blocked
		}
	})

	t.Run("Contextual read closer does not block even on long running copies", func(t *testing.T) {
		r, w, err := os.Pipe()
		require.NoError(t, err)
		defer func() { _ = w.Close() }()

		ctx, cancel := context.WithCancel(context.Background())
		rc := NewContextualReadCloser(ctx, r)

		done := make(chan struct{})
		go func() {
			_, _ = io.Copy(io.Discard, rc) // will block in read(2) https://man7.org/linux/man-pages/man2/read.2.html
			close(done)
		}()

		time.Sleep(50 * time.Millisecond)
		cancel()

		select {
		case <-done:
			// Expected case: successfully unblocked
		case <-time.After(2 * time.Second):
			assert.FailNow(t, "copy should have been unblocked by context cancel")
		}
	})
}
