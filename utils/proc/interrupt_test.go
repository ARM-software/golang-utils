/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
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
	type terminateGracefullyTest struct {
		name                                       string
		testFunc                                   func(ctx context.Context, pid int, gracePeriod time.Duration) error
		expectSingleProcessBeforeGracePeriod       bool
		singleProcessGracePeriodReason             string
		expectProcessWithChildrenBeforeGracePeriod bool
		processWithChildrenGracePeriodReason       string
	}

	tests := []terminateGracefullyTest{
		{
			name:                                 "TerminateGracefully",
			testFunc:                             TerminateGracefully,
			expectSingleProcessBeforeGracePeriod: true,
			singleProcessGracePeriodReason:       "TerminateGracefully sends both SigInt and SigTerm before the fallback kill, so the process should exit before the grace period.",
			expectProcessWithChildrenBeforeGracePeriod: true,
			processWithChildrenGracePeriodReason:       "TerminateGracefully sends both SigInt and SigTerm before the fallback kill, so the bash parent and its child should exit before the grace period.",
		},
		{
			name:                                 "TerminateGracefullyWithChildren",
			testFunc:                             TerminateGracefullyWithChildren,
			expectSingleProcessBeforeGracePeriod: true,
			singleProcessGracePeriodReason:       "TerminateGracefullyWithChildren ultimately applies the default graceful shutdown sequence, so the target process should exit before the grace period instead of waiting for the fallback kill.",
			expectProcessWithChildrenBeforeGracePeriod: true,
			processWithChildrenGracePeriodReason:       "TerminateGracefullyWithChildren applies the default graceful shutdown sequence to the child first and then the parent, so the process tree should exit before the grace period.",
		},
	}

	if runtime.GOOS == "windows" {
		tests = append(tests,
			terminateGracefullyTest{
				name:                                 "TerminateGracefullyWithSpecificInterrupts",
				expectSingleProcessBeforeGracePeriod: false,
				singleProcessGracePeriodReason:       "On Windows, this variant only sends SigInt and proc.Interrupt documents SigInt as a no-op, so the process should remain alive until the fallback kill after the grace period.",
				expectProcessWithChildrenBeforeGracePeriod: false,
				processWithChildrenGracePeriodReason:       "On Windows, this variant only sends SigInt and proc.Interrupt documents SigInt as a no-op, so the bash parent is expected to remain alive until the fallback kill after the grace period.",
				testFunc: func(ctx context.Context, pid int, gracePeriod time.Duration) error {
					return TerminateGracefullyWithSpecificInterrupts(ctx, pid, gracePeriod, SigInt)
				},
			},
			terminateGracefullyTest{
				name:                                 "TerminateGracefullyWithChildrenWithSpecificInterrupts",
				expectSingleProcessBeforeGracePeriod: false,
				singleProcessGracePeriodReason:       "On Windows, this variant only sends SigInt to each process and SigInt is a no-op, so the process should remain alive until the fallback kill after the grace period.",
				expectProcessWithChildrenBeforeGracePeriod: false,
				processWithChildrenGracePeriodReason:       "On Windows, this variant only sends SigInt to each process and SigInt is a no-op, so the process tree should remain alive until the fallback kill after the grace period.",
				testFunc: func(ctx context.Context, pid int, gracePeriod time.Duration) error {
					return TerminateGracefullyWithChildrenWithSpecificInterrupts(ctx, pid, gracePeriod, SigInt)
				},
			},
		)
	} else {
		tests = append(tests,
			terminateGracefullyTest{
				name:                                 "TerminateGracefullyWithSpecificInterrupts",
				expectSingleProcessBeforeGracePeriod: true,
				singleProcessGracePeriodReason:       "On Unix, this variant only sends SigInt and sleep exits on SigInt, so the process should finish before the grace period.",
				expectProcessWithChildrenBeforeGracePeriod: true,
				processWithChildrenGracePeriodReason:       "On Unix, this variant only sends SigInt to the bash parent and that causes this test command to exit before the grace period.",
				testFunc: func(ctx context.Context, pid int, gracePeriod time.Duration) error {
					return TerminateGracefullyWithSpecificInterrupts(ctx, pid, gracePeriod, SigInt)
				},
			},
			terminateGracefullyTest{
				name:                                 "TerminateGracefullyWithChildrenWithSpecificInterrupts",
				expectSingleProcessBeforeGracePeriod: true,
				singleProcessGracePeriodReason:       "On Unix, this variant only sends SigInt and sleep exits on SigInt, so the process should finish before the grace period.",
				expectProcessWithChildrenBeforeGracePeriod: false,
				processWithChildrenGracePeriodReason:       "On Unix, this test command runs `bash -c \"sleep 50 & wait\"`. Sending SigInt to the child sleep process does not make the non-interactive bash parent exit promptly, so the parent is expected to remain alive until the fallback kill after the grace period.",
				testFunc: func(ctx context.Context, pid int, gracePeriod time.Duration) error {
					return TerminateGracefullyWithChildrenWithSpecificInterrupts(ctx, pid, gracePeriod, SigInt)
				},
			},
		)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			t.Run("single process", func(t *testing.T) {
				cmd := exec.Command("sleep", "50")
				require.NoError(t, cmd.Start())
				waitErrCh := waitForCmdInBackground(cmd)
				defer func() {
					p, _ := FindProcess(context.Background(), cmd.Process.Pid)
					if p != nil && (p.IsRunning() || p.IsAZombie()) {
						_ = p.KillWithChildren(context.Background())
					}
					waitErr := <-waitErrCh
					if waitErr != nil {
						waitErr = commonerrors.WrapIfNotCommonError(commonerrors.ErrInterrupted, waitErr, "")
					}
					errortest.AssertError(t, waitErr, nil, commonerrors.ErrInterrupted)
				}()
				timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				var process IProcess
				var err error
				require.NoError(t, parallelisation.WaitUntil(timeoutCtx, func(waitCtx context.Context) (bool, error) {
					process, err = FindProcess(waitCtx, cmd.Process.Pid)
					if err != nil {
						return false, err
					}
					return process.IsRunning(), nil
				}, 100*time.Millisecond))

				now := time.Now()
				gracePeriod := 10 * time.Second
				require.NoError(t, test.testFunc(context.Background(), cmd.Process.Pid, gracePeriod))
				elapsed := time.Since(now)
				if test.expectSingleProcessBeforeGracePeriod {
					assert.Lessf(t, elapsed, gracePeriod, "expected completion before the grace period: %s", test.singleProcessGracePeriodReason)
				} else {
					assert.GreaterOrEqualf(t, elapsed, gracePeriod, "expected to wait for the grace period on this platform: %s", test.singleProcessGracePeriodReason)
				}

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
				cmd := newCommandWithChildProcess()
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
				if !p.IsAZombie() {
					assert.NotEmpty(t, children)
				}

				now := time.Now()
				gracePeriod := 10 * time.Second
				require.NoError(t, test.testFunc(context.Background(), cmd.Process.Pid, gracePeriod))
				elapsed := time.Since(now)
				if test.expectProcessWithChildrenBeforeGracePeriod {
					assert.Lessf(t, elapsed, gracePeriod, "expected completion before the grace period: %s", test.processWithChildrenGracePeriodReason)
				} else {
					assert.GreaterOrEqualf(t, elapsed, gracePeriod, "expected to wait for the grace period on this platform: %s", test.processWithChildrenGracePeriodReason)
				}

				p, err = FindProcess(context.Background(), cmd.Process.Pid)
				if err == nil {
					require.NotEmpty(t, p)
					assert.False(t, p.IsRunning())
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
