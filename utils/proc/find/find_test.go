/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
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
	"github.com/ARM-software/golang-utils/utils/parallelisation"
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
			cmdWindows:         exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", fmt.Sprintf("Start-Sleep -Seconds %v", 5)), //nolint: gosec // G204 Subprocess launched with a potential tainted input or cmd arguments (gosec)
			cmdOther:           exec.Command("sh", "-c", fmt.Sprintf("sleep %v", 5)),                                                               //nolint: gosec // G204 Subprocess launched with a potential tainted input or cmd arguments (gosec)
			processNameWindows: "powershell",
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
			numOfProcesses := len(ps)
			require.NoError(t, cmd.Start())
			defer func() { _ = cmd.Process.Kill() }()
			timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			require.NoError(t, parallelisation.WaitUntil(timeoutCtx, func(waitCtx context.Context) (bool, error) {
				ps, err = FindProcessByName(waitCtx, toFind)
				if err != nil {
					return false, err
				}
				return len(ps) > 0, nil
			}, 100*time.Millisecond))
			assert.NotEmpty(t, ps)

			assert.GreaterOrEqual(t, len(ps), numOfProcesses)
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
