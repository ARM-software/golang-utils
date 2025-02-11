package supervisor

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/logs/logstest"
	"github.com/ARM-software/golang-utils/utils/subprocess"
)

func TestSupervisor(t *testing.T) {
	t.Run("with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		logger, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "Test")
		require.NoError(t, err)

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test")

		cmd, err := subprocess.New(ctx, logger, "starting", "success", "failed", "sed", "-i", `$a test123`, testFile)
		require.NoError(t, err)

		runner := NewSupervisor(func(ctx context.Context) *subprocess.Subprocess {
			return cmd
		})

		require.False(t, filesystem.Exists(testFile))
		err = filesystem.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)

		err = runner.Run(ctx)
		errortest.AssertError(t, err, commonerrors.ErrTimeout)

		written, err := filesystem.ReadFile(testFile)
		require.NoError(t, err)
		assert.NotEmpty(t, written)
		assert.Contains(t, string(written), "test\ntest123\ntest123")
	})

	t.Run("with pre run", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		logger, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "Test")
		require.NoError(t, err)

		counter := atomic.Uint64{}
		assert.Zero(t, counter.Load())

		cmd, err := subprocess.New(ctx, logger, "starting", "success", "failed", "echo", "123")
		require.NoError(t, err)

		runner := NewSupervisor(func(ctx context.Context) *subprocess.Subprocess {
			return cmd
		}, WithPreStart(func(_ context.Context) error {
			_ = counter.Inc()
			return nil
		}))

		err = runner.Run(ctx)
		errortest.AssertError(t, err, commonerrors.ErrTimeout)

		assert.NotZero(t, counter.Load())
	})

	t.Run("with post run", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		logger, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "Test")
		require.NoError(t, err)

		counter := atomic.Uint64{}
		assert.Zero(t, counter.Load())

		cmd, err := subprocess.New(ctx, logger, "starting", "success", "failed", "echo", "123")
		require.NoError(t, err)

		runner := NewSupervisor(func(ctx context.Context) *subprocess.Subprocess {
			return cmd
		}, WithPostStart(func(_ context.Context) error {
			_ = counter.Inc()
			return nil
		}))

		err = runner.Run(ctx)
		errortest.AssertError(t, err, commonerrors.ErrTimeout)

		assert.NotZero(t, counter.Load())
	})

	t.Run("with post stop", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		logger, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "Test")
		require.NoError(t, err)

		counter := atomic.Uint64{}
		assert.Zero(t, counter.Load())

		cmd, err := subprocess.New(ctx, logger, "starting", "success", "failed", "echo", "123")
		require.NoError(t, err)

		runner := NewSupervisor(func(ctx context.Context) *subprocess.Subprocess {
			return cmd
		}, WithPostStop(func(_ error) error {
			_ = counter.Inc()
			return nil
		}))

		err = runner.Run(ctx)
		errortest.AssertError(t, err, commonerrors.ErrTimeout)

		assert.NotZero(t, counter.Load())
	})

	t.Run("with pre and post start", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		logger, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "Test")
		require.NoError(t, err)

		counter1 := atomic.Uint64{}
		assert.Zero(t, counter1.Load())
		counter2 := atomic.Uint64{}
		assert.Zero(t, counter2.Load())

		cmd, err := subprocess.New(ctx, logger, "starting", "success", "failed", "echo", "123")
		require.NoError(t, err)

		runner := NewSupervisor(func(ctx context.Context) *subprocess.Subprocess {
			return cmd
		}, WithPreStart(func(_ context.Context) error {
			_ = counter1.Inc()
			return nil
		}), WithPostStart(func(_ context.Context) error {
			_ = counter2.Inc()
			return nil
		}))

		err = runner.Run(ctx)
		errortest.AssertError(t, err, commonerrors.ErrTimeout)

		assert.NotZero(t, counter1.Load())
		assert.NotZero(t, counter2.Load())
		assert.Equal(t, counter1.Load(), counter2.Load())
	})

	t.Run("with cancel", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		logger, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "Test")
		require.NoError(t, err)

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test")

		cmd, err := subprocess.New(ctx, logger, "starting", "success", "failed", "sed", "-i", `$a test123`, testFile)
		require.NoError(t, err)

		runner := NewSupervisor(func(ctx context.Context) *subprocess.Subprocess {
			return cmd
		})

		cancel()
		err = runner.Run(ctx)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
	})

	t.Run("with ignore cancelled", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		logger, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "Test")
		require.NoError(t, err)

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test")

		failMessage := "failed"

		cmd, err := subprocess.New(ctx, logger, "starting", "success", failMessage, "sed", "-i", `$a test123`, testFile)
		require.NoError(t, err)

		runner := NewSupervisor(func(ctx context.Context) *subprocess.Subprocess {
			return cmd
		}, WithIgnoreErrors(fmt.Errorf("%v %v", failMessage, commonerrors.ErrCancelled)))

		require.False(t, filesystem.Exists(testFile))
		err = filesystem.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)

		go func() {
			time.Sleep(50 * time.Millisecond)
			cmd.Cancel()
		}()

		err = runner.Run(ctx)
		errortest.AssertError(t, err, commonerrors.ErrTimeout)

		written, err := filesystem.ReadFile(testFile)
		require.NoError(t, err)
		assert.NotEmpty(t, written)
		assert.Contains(t, string(written), "test\ntest123\ntest123")
	})

	t.Run("with delay", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		logger, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "Test")
		require.NoError(t, err)

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test")

		cmd, err := subprocess.New(ctx, logger, "starting", "success", "failed", "sed", "-i", `$a test123`, testFile)
		require.NoError(t, err)

		runner := NewSupervisor(func(ctx context.Context) *subprocess.Subprocess {
			return cmd
		}, WithRestartDelay(time.Hour)) // won't have time to restart

		require.False(t, filesystem.Exists(testFile))
		err = filesystem.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)

		err = runner.Run(ctx)
		errortest.AssertError(t, err, commonerrors.ErrTimeout)

		written, err := filesystem.ReadFile(testFile)
		require.NoError(t, err)
		assert.NotEmpty(t, written)
		assert.Equal(t, string(written), "test\ntest123\n")
	})
}
