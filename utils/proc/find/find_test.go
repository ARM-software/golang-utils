package find

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestFindProcessByName(t *testing.T) {
	if runtime.GOOS != "linux" {
		defer goleak.VerifyNone(t)
	}
	tests := []struct {
		cmdWindows         *exec.Cmd
		cmdOther           *exec.Cmd
		processNameWindows string
		processNameOther   string
	}{
		{
			cmdWindows:         exec.Command("cmd.exe", "/c", fmt.Sprintf("ping localhost -n %v > nul", time.Second.Seconds())), //nolint: gosec // G204 Subprocess launched with a potential tainted input or cmd arguments (gosec)
			cmdOther:           exec.Command("sh", "-c", fmt.Sprintf("sleep %v", time.Second.Seconds())),                        //nolint: gosec // G204 Subprocess launched with a potential tainted input or cmd arguments (gosec)
			processNameWindows: "ping",
			processNameOther:   "sleep",
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run("subtest", func(t *testing.T) {
			ctx := context.Background()
			cmd := test.cmdOther
			toFind := test.processNameOther
			if runtime.GOOS == "windows" {
				cmd = test.cmdWindows
				toFind = test.processNameWindows
			}
			ps, err := FindProcessByName(ctx, toFind)
			require.NoError(t, err)
			assert.Empty(t, ps)
			require.NoError(t, cmd.Start())
			defer func() { _ = cmd.Process.Kill() }()
			ps, err = FindProcessByName(ctx, toFind)
			require.NoError(t, err)
			assert.NotEmpty(t, ps)
			require.NoError(t, cmd.Wait())
			t.Run("cancelled context", func(t *testing.T) {
				cancelCtx, cancel := context.WithCancel(ctx)
				cancel()
				_, err = FindProcessByName(cancelCtx, toFind)
				errortest.AssertError(t, err, commonerrors.ErrCancelled)
			})
		})
	}
}
