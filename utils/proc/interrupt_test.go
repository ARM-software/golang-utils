package proc

import (
	"context"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

func TestTerminateGracefully(t *testing.T) {
	for _, test := range []struct {
		name     string
		testFunc func(ctx context.Context, pid int, gracePeriod time.Duration) error
	}{
		{
			name:     "TerminateGracefully",
			testFunc: TerminateGracefully,
		},
		{
			name:     "TerminateGracefullyWithChildren",
			testFunc: TerminateGracefullyWithChildren,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			t.Run("single process", func(t *testing.T) {
				cmd := exec.Command("sleep", "50")
				require.NoError(t, cmd.Start())
				defer func() {
					p, _ := FindProcess(context.Background(), cmd.Process.Pid)
					if p != nil && (p.IsRunning() || p.IsAZombie()) {
						_ = cmd.Wait()
					}
				}()
				process, err := FindProcess(context.Background(), cmd.Process.Pid)
				require.NoError(t, err)
				require.True(t, process.IsRunning())

				now := time.Now()
				gracePeriod := 10 * time.Second
				require.NoError(t, test.testFunc(context.Background(), cmd.Process.Pid, gracePeriod))
				assert.Less(t, time.Since(now), gracePeriod) // this indicates that the process was closed by INT/SIG not KILL

				time.Sleep(500 * time.Millisecond)
				process, err = FindProcess(context.Background(), cmd.Process.Pid)
				if err == nil {
					require.NotEmpty(t, process)
					assert.False(t, process.IsRunning())
				} else {
					errortest.AssertError(t, err, commonerrors.ErrNotFound)
					assert.Empty(t, process)
				}
			})
			t.Run("process with children", func(t *testing.T) {
				if runtime.GOOS == "windows" {
					t.Skip("test with bash")
				}
				// see https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
				// https://forum.golangbridge.org/t/killing-child-process-on-timeout-in-go-code/995/16
				cmd := exec.Command("bash", "-c", "watch date > date.txt 2>&1")
				require.NoError(t, cmd.Start())
				defer func() {
					p, _ := FindProcess(context.Background(), cmd.Process.Pid)
					if p != nil && (p.IsRunning() || p.IsAZombie()) {
						_ = cmd.Wait()
					}
				}()
				time.Sleep(500 * time.Millisecond)
				require.NotNil(t, cmd.Process)

				timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				require.NoError(t, parallelisation.WaitUntil(timeoutCtx, func(fCtx context.Context) (bool, error) {
					p, fErr := FindProcess(fCtx, cmd.Process.Pid)
					if fErr != nil {
						return false, fErr
					}
					return p.IsRunning() || p.IsAZombie(), nil
				}, 200*time.Millisecond))
				p, err := FindProcess(context.Background(), cmd.Process.Pid)
				require.NoError(t, err)
				require.True(t, p.IsRunning() || p.IsAZombie())
				children, err := p.Children(timeoutCtx)
				require.NoError(t, err)
				require.NotZero(t, len(children))

				now := time.Now()
				gracePeriod := 10 * time.Second
				require.NoError(t, test.testFunc(context.Background(), cmd.Process.Pid, gracePeriod))
				assert.Less(t, time.Since(now), gracePeriod) // this indicates that the process was closed by INT/SIG not KILL

				p, err = FindProcess(context.Background(), cmd.Process.Pid)
				if err == nil {
					require.NotEmpty(t, p)
					assert.False(t, p.IsRunning() || p.IsAZombie())
				} else {
					errortest.AssertError(t, err, commonerrors.ErrNotFound)
					assert.Empty(t, p)
				}
			})
			t.Run("no process", func(t *testing.T) {
				random, err := faker.RandomInt(9000, 20000, 1)
				require.NoError(t, err)
				require.NoError(t, test.testFunc(context.Background(), random[0], 100*time.Millisecond))
			})
			t.Run("cancelled", func(t *testing.T) {
				random, err := faker.RandomInt(9000, 20000, 1)
				require.NoError(t, err)
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				errortest.AssertError(t, test.testFunc(ctx, random[0], 100*time.Millisecond), commonerrors.ErrCancelled)
			})
		})
	}
}
