/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package parallelisation

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/field"
)

var (
	random = rand.New(rand.NewSource(time.Now().Unix())) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec as this is just for
)

func TestParallelisationWithResults(t *testing.T) {
	defer goleak.VerifyNone(t)
	var values []int
	length := 100
	for i := 0; i < length; i++ {
		values = append(values, i)
	}
	action := func(arg interface{}) (result interface{}, err error) {
		result = int64(arg.(int))
		return
	}
	var results []int64
	rawResults, err := Parallelise(values, action, reflect.TypeOf(results))
	require.NoError(t, err)

	results = rawResults.([]int64)
	assert.Equal(t, length, len(results))
}
func TestParallelisationWithoutResults(t *testing.T) {
	defer goleak.VerifyNone(t)
	var values []int
	length := 30
	for i := 0; i < length; i++ {
		values = append(values, i)
	}
	action := func(arg interface{}) (result interface{}, err error) {
		return
	}
	results, err := Parallelise(values, action, nil)
	assert.NoError(t, err)
	assert.Nil(t, results)
}

func TestParallelisationWithErrors(t *testing.T) {
	defer goleak.VerifyNone(t)
	var values []int
	length := 30
	for i := 0; i < length; i++ {
		values = append(values, i)
	}
	anError := errors.New("a failure")
	action := func(arg interface{}) (result interface{}, err error) {
		modulo := (arg.(int)) % 10
		if modulo == 0 {
			err = anError
		}
		return
	}
	results, err := Parallelise(values, action, nil)
	assert.Nil(t, results)
	errortest.AssertError(t, err, anError)
}

func TestSleepWithInterruption(t *testing.T) {
	tests := []struct {
		name  string
		sleep func(context.Context, time.Duration, chan time.Duration)
	}{
		{
			name: "Sleep with interruption",
			sleep: func(ctx context.Context, duration time.Duration, wait chan time.Duration) {
				start := time.Now()
				stop := make(chan bool, 1)
				go func(ctx context.Context, stop chan bool) {
					<-ctx.Done()
					stop <- true
				}(ctx, stop)
				SleepWithInterruption(stop, duration)
				wait <- time.Since(start)
			},
		},
		{
			name: "Sleep with context",
			sleep: func(ctx context.Context, duration time.Duration, wait chan time.Duration) {
				start := time.Now()
				SleepWithContext(ctx, duration)
				wait <- time.Since(start)
			},
		},
	}

	testSleep := func(t *testing.T, sleep func(context.Context, time.Duration, chan time.Duration)) {
		times := make(chan time.Duration)
		ctx, cancel := context.WithCancel(context.Background())
		timeToSleep := 100 * time.Millisecond
		go sleep(ctx, timeToSleep, times)
		timeSlept := <-times
		assert.GreaterOrEqual(t, timeSlept.Milliseconds(), timeToSleep.Milliseconds())
		timeToSleep = time.Hour
		go sleep(ctx, timeToSleep, times)
		time.Sleep(time.Millisecond)
		cancel()
		timeSlept = <-times
		assert.Less(t, timeSlept.Milliseconds(), timeToSleep.Milliseconds())
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			testSleep(t, test.sleep)
		})
	}
}

func TestSchedule(t *testing.T) {
	defer goleak.VerifyNone(t)
	var ticks atomic.Uint64
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	Schedule(ctx, 10*time.Millisecond, 15*time.Millisecond, func(time.Time) {
		ticks.Inc()
	})
	time.Sleep(500 * time.Millisecond)
	// Expected number should be 49 but there is some timing variance depending on the state of the environment this is run on.
	// Therefore, we accept that the number of ticks achieved is not always accurate but close to what is expected.
	tickNumbers := ticks.Load()
	require.NoError(t, ctx.Err())
	cancel()
	assert.GreaterOrEqual(t, tickNumbers, uint64(20))
	assert.LessOrEqual(t, tickNumbers, uint64(80))
}

func TestScheduleAfter(t *testing.T) {
	defer goleak.VerifyNone(t)
	var timeS atomic.Value
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	time1 := time.Now()
	expectedOffset := 10 * time.Millisecond
	ScheduleAfter(ctx, expectedOffset, func(time.Time) {
		timeS.Store(time.Now())
	})
	time.Sleep(50 * time.Millisecond)

	duration := timeS.Load().(time.Time).Sub(time1)
	require.NoError(t, ctx.Err())
	cancel()
	assert.GreaterOrEqual(t, duration, expectedOffset)
}

func TestRunBlockingActionWithTimeout(t *testing.T) {
	defer goleak.VerifyNone(t)
	for i := 0; i < 200; i++ {
		testTimeout(t)
	}
}

func TestRunBlockingActionWithTimeoutAndContext(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx := context.Background()
	for i := 0; i < 200; i++ {
		testTimeoutWithContext(t, ctx)
	}
}

func testTimeout(t *testing.T) {
	isrunning := atomic.NewBool(true)
	blockingAction := func(stop chan bool) error {
		isrunning.Store(true)
		<-stop
		isrunning.Store(false)
		return nil
	}
	assert.True(t, isrunning.Load())
	err := RunActionWithTimeout(blockingAction, 10*time.Millisecond)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrTimeout)
	assert.False(t, isrunning.Load())

	isrunning.Store(true)
	blockingAction2 := func(stop chan bool) error {
		isrunning.Store(true)
		<-stop
		isrunning.Store(false)
		time.Sleep(5 * time.Millisecond)
		return nil
	}
	assert.True(t, isrunning.Load())
	err = RunActionWithTimeout(blockingAction2, 10*time.Millisecond)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrTimeout)
	assert.False(t, isrunning.Load())

	isrunning.Store(true)
	blockingAction3 := func(stop chan bool) error {
		for {
			isrunning.Store(true)
			select {
			case <-stop:
				isrunning.Store(false)
				time.Sleep(5 * time.Millisecond)
				return nil
			default:
				time.Sleep(time.Millisecond)
			}
		}
	}
	assert.True(t, isrunning.Load())
	err = RunActionWithTimeout(blockingAction3, 10*time.Millisecond)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrTimeout)
	assert.False(t, isrunning.Load())

	isrunning.Store(true)
	nonblockingAction := func(stop chan bool) error {
		isrunning.Store(true)
		isrunning.Store(false)
		return nil
	}
	assert.True(t, isrunning.Load())
	err = RunActionWithTimeout(nonblockingAction, 10*time.Millisecond)
	require.NoError(t, err)
	assert.False(t, isrunning.Load())

	isrunning.Store(true)
	anError := errors.New("action error")
	failingnonblockingAction := func(stop chan bool) error {
		isrunning.Store(true)
		isrunning.Store(false)
		return anError
	}
	assert.True(t, isrunning.Load())
	err = RunActionWithTimeout(failingnonblockingAction, 10*time.Millisecond)
	require.Error(t, err)
	errortest.AssertError(t, err, anError)
	assert.False(t, isrunning.Load())
}

func testTimeoutWithContext(t *testing.T, ctx context.Context) {
	isrunning := atomic.NewBool(true)
	blockingAction := func(actionCtx context.Context) error {
		isrunning.Store(true)
		<-actionCtx.Done()
		isrunning.Store(false)
		return nil
	}
	assert.True(t, isrunning.Load())
	err := RunActionWithTimeoutAndContext(ctx, 10*time.Millisecond, blockingAction)
	require.NoError(t, DetermineContextError(ctx))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrTimeout)
	assert.False(t, isrunning.Load())

	isrunning.Store(true)
	blockingAction2 := func(ctx context.Context) error {
		isrunning.Store(true)
		<-ctx.Done()
		isrunning.Store(false)
		time.Sleep(5 * time.Millisecond)
		return nil
	}
	assert.True(t, isrunning.Load())
	err = RunActionWithTimeoutAndContext(ctx, 10*time.Millisecond, blockingAction2)
	require.NoError(t, DetermineContextError(ctx))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrTimeout)
	assert.False(t, isrunning.Load())

	isrunning.Store(true)
	nonblockingAction := func(ctx context.Context) error {
		isrunning.Store(true)
		isrunning.Store(false)
		return nil
	}
	assert.True(t, isrunning.Load())
	err = RunActionWithTimeoutAndContext(ctx, 10*time.Millisecond, nonblockingAction)
	require.NoError(t, DetermineContextError(ctx))
	require.NoError(t, err)
	assert.False(t, isrunning.Load())

	isrunning.Store(true)
	anError := errors.New("action error")
	failingnonblockingAction := func(ctx context.Context) error {
		isrunning.Store(true)
		isrunning.Store(false)
		return anError
	}
	assert.True(t, isrunning.Load())
	err = RunActionWithTimeoutAndContext(ctx, 10*time.Millisecond, failingnonblockingAction)
	require.NoError(t, DetermineContextError(ctx))
	require.Error(t, err)
	errortest.AssertError(t, err, anError)
	assert.False(t, isrunning.Load())

	isrunning.Store(true)
	var funcCtxatomic atomic.Value
	nonblockingAction2 := func(funcCtx context.Context) error {
		isrunning.Store(true)
		isrunning.Store(false)
		funcCtxatomic.Store(funcCtx)
		return nil
	}
	assert.True(t, isrunning.Load())
	err = RunActionWithTimeoutAndContext(ctx, 100*time.Millisecond, nonblockingAction2)
	require.NoError(t, DetermineContextError(ctx))
	require.NoError(t, err)
	errortest.AssertError(t, DetermineContextError(funcCtxatomic.Load().(context.Context)), commonerrors.ErrCancelled)
	assert.False(t, isrunning.Load())

	isrunning.Store(true)
	assert.True(t, isrunning.Load())
	err = RunActionWithTimeoutAndCancelStore(ctx, 100*time.Millisecond, NewCancelFunctionsStore(), nonblockingAction2)
	require.NoError(t, DetermineContextError(ctx))
	require.NoError(t, err)
	require.NoError(t, DetermineContextError(funcCtxatomic.Load().(context.Context)))
	assert.False(t, isrunning.Load())
}

func TestRunActionWithParallelCheckHappy(t *testing.T) {
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("test #%v", i), func(t *testing.T) {
			defer goleak.VerifyNone(t)
			runActionWithParallelCheckHappy(t, ctx)
		})
	}
}

func TestRunActionWithParallelCheckFail(t *testing.T) {
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("test #%v", i), func(t *testing.T) {
			defer goleak.VerifyNone(t)
			runActionWithParallelCheckFail(t, ctx)
		})
	}
}

func TestRunActionWithParallelCheckFailAtRandom(t *testing.T) {
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("test #%v", i), func(t *testing.T) {
			defer goleak.VerifyNone(t)
			runActionWithParallelCheckFailAtRandom(t, ctx)
		})
	}
}

func runActionWithParallelCheckHappy(t *testing.T, ctx context.Context) {
	counter := atomic.NewInt32(0)
	checkAction := func(ctx context.Context) bool {
		counter.Inc()
		fmt.Println("Check #", counter.String())
		return true
	}
	action := func(ctx context.Context) error {
		time.Sleep(150 * time.Millisecond)
		return nil
	}
	err := RunActionWithParallelCheck(ctx, action, checkAction, 10*time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, int32(15), counter.Load())
}

func runActionWithParallelCheckFail(t *testing.T, ctx context.Context) {
	counter := atomic.NewInt32(0)
	checkAction := func(ctx context.Context) bool {
		counter.Inc()
		fmt.Println("Check #", counter.String())
		return false
	}
	action := func(ctx context.Context) error {
		time.Sleep(150 * time.Millisecond)
		return nil
	}
	err := RunActionWithParallelCheck(ctx, action, checkAction, 10*time.Millisecond)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
	assert.Equal(t, int32(1), counter.Load())
}

func runActionWithParallelCheckFailAtRandom(t *testing.T, ctx context.Context) {
	counter := atomic.NewInt32(0)
	checkAction := func(ctx context.Context) bool {
		counter.Add(1)
		fmt.Println("Check #", counter.String())
		return random.Intn(2) != 0 && counter.Load() < 10 //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
	}
	action := func(ctx context.Context) error {
		time.Sleep(150 * time.Millisecond)
		return nil
	}
	err := RunActionWithParallelCheck(ctx, action, checkAction, 10*time.Millisecond)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
	assert.GreaterOrEqual(t, counter.Load(), int32(1))
}

func TestRunActionWithParallelCheckAndResult(t *testing.T) {
	type parallelisationCheckResult struct {
		checks int32
		status string
	}

	t.Run("Happy", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		checkCounter := atomic.NewInt32(0)
		checkResultCounter := atomic.NewInt32(0)

		res, ok, err := RunActionWithParallelCheckAndResult(
			context.Background(),
			func(ctx context.Context) (err error) {
				time.Sleep(120 * time.Millisecond)
				return
			},
			func(ctx context.Context) (res parallelisationCheckResult, ok bool) {
				return parallelisationCheckResult{
					checks: checkCounter.Inc(),
					status: "healthy",
				}, true
			},
			func(_ parallelisationCheckResult) error {
				checkResultCounter.Inc()
				return nil
			},
			10*time.Millisecond,
		)

		require.NoError(t, err)
		require.True(t, ok)

		assert.GreaterOrEqual(t, res.checks, int32(10))
		assert.Equal(t, res.checks, checkCounter.Load())
		assert.Equal(t, "healthy", res.status)
		assert.Equal(t, checkCounter.Load(), checkResultCounter.Load())
	})

	t.Run("Check Fails With Reason", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		checkCounter := atomic.NewInt32(0)
		checkResultCounter := atomic.NewInt32(0)
		actionStarted := atomic.NewBool(false)

		status := "adrien"

		res, ok, err := RunActionWithParallelCheckAndResult(
			context.Background(),
			func(ctx context.Context) error {
				actionStarted.Store(true)
				<-ctx.Done()
				return DetermineContextError(ctx)
			},
			func(ctx context.Context) (res parallelisationCheckResult, ok bool) {
				if n := checkCounter.Inc(); n >= 5 {
					return parallelisationCheckResult{
						checks: n,
						status: status,
					}, false
				} else {
					return parallelisationCheckResult{
						checks: n,
						status: "ok",
					}, true
				}
			},
			func(_ parallelisationCheckResult) error {
				checkResultCounter.Inc()
				return nil
			},
			5*time.Millisecond,
		)

		require.True(t, actionStarted.Load())
		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)

		require.False(t, ok)
		assert.Equal(t, status, res.status)
		assert.Equal(t, int32(5), res.checks)
		assert.Equal(t, int32(5), checkCounter.Load())
		assert.Equal(t, checkCounter.Load(), checkResultCounter.Load())
	})
	t.Run("Action Error (no context cancel)", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		checkCounter := atomic.NewInt32(0)
		checkResultCounter := atomic.NewInt32(0)
		status := "abdel"

		res, ok, err := RunActionWithParallelCheckAndResult(
			context.Background(),
			func(ctx context.Context) error {
				time.Sleep(30 * time.Millisecond)
				return commonerrors.New(commonerrors.ErrForbidden, faker.Sentence())
			},
			func(ctx context.Context) (parallelisationCheckResult, bool) {
				return parallelisationCheckResult{
					checks: checkCounter.Inc(),
					status: status,
				}, true
			},
			func(_ parallelisationCheckResult) error {
				checkResultCounter.Inc()
				return nil
			},
			5*time.Millisecond,
		)

		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrForbidden)
		require.True(t, ok)

		assert.Equal(t, status, res.status)
		assert.GreaterOrEqual(t, res.checks, int32(1))
		assert.Equal(t, res.checks, checkCounter.Load())
		assert.Equal(t, checkCounter.Load(), checkResultCounter.Load())
	})

	t.Run("Context cancel", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		checkCounter := atomic.NewInt32(0)
		checkResultCounter := atomic.NewInt32(0)
		status := "kem"

		res, ok, err := RunActionWithParallelCheckAndResult(
			ctx,
			func(ctx context.Context) error {
				<-ctx.Done()
				return DetermineContextError(ctx)
			},
			func(ctx context.Context) (parallelisationCheckResult, bool) {
				return parallelisationCheckResult{
					checks: checkCounter.Inc(),
					status: status,
				}, true
			},
			func(_ parallelisationCheckResult) error {
				checkResultCounter.Inc()
				return nil
			},
			5*time.Millisecond,
		)

		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrTimeout)
		assert.True(t, ok)
		assert.GreaterOrEqual(t, res.checks, int32(1))
		assert.Equal(t, res.checks, checkCounter.Load())
		assert.Equal(t, checkCounter.Load(), checkResultCounter.Load())
	})

	t.Run("Check result error", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		checkCounter := atomic.NewInt32(0)
		checkResultCounter := atomic.NewInt32(0)
		status := "kem"

		res, ok, err := RunActionWithParallelCheckAndResult(
			context.Background(),
			func(ctx context.Context) error {
				<-ctx.Done()
				return DetermineContextError(ctx)
			},
			func(ctx context.Context) (parallelisationCheckResult, bool) {
				return parallelisationCheckResult{
					checks: checkCounter.Inc(),
					status: status,
				}, true
			},
			func(_ parallelisationCheckResult) error {
				checkResultCounter.Inc()
				return commonerrors.ErrUnexpected
			},
			5*time.Millisecond,
		)

		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrUnexpected)
		assert.True(t, ok)
		assert.GreaterOrEqual(t, res.checks, int32(1))
		assert.Equal(t, res.checks, checkCounter.Load())
		assert.Equal(t, checkCounter.Load(), checkResultCounter.Load())
	})
}

func TestRunActionWithCancelStore(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("action no error", func(t *testing.T) {
		actionRan := atomic.NewBool(false)

		action := func(actionCtx context.Context) error {
			actionRan.Store(true)
			return nil
		}

		err := RunActionWithCancelStore(t.Context(), NewCancelFunctionsStore(), action)

		require.NoError(t, err)
		assert.True(t, actionRan.Load())
	})

	t.Run("action error", func(t *testing.T) {
		actionRan := atomic.NewBool(false)
		actionErr := commonerrors.ErrUnexpected
		action := func(actionCtx context.Context) error {
			actionRan.Store(true)
			return actionErr
		}

		err := RunActionWithCancelStore(t.Context(), NewCancelFunctionsStore(), action)

		require.Error(t, err)
		errortest.AssertError(t, err, actionErr)
		assert.True(t, actionRan.Load())
	})

	t.Run("action canceled via parent context", func(t *testing.T) {
		actionStarted := atomic.NewBool(false)

		action := func(actionCtx context.Context) error {
			actionStarted.Store(true)
			<-actionCtx.Done()
			return nil
		}

		parentCtx, cancel := context.WithCancel(t.Context())
		cancel()

		err := RunActionWithCancelStore(parentCtx, NewCancelFunctionsStore(), action)

		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
		assert.False(t, actionStarted.Load())
	})

	t.Run("action canceled via store", func(t *testing.T) {
		actionStarted := atomic.NewBool(false)
		store := NewCancelFunctionsStore()

		action := func(actionCtx context.Context) error {
			actionStarted.Store(true)
			<-actionCtx.Done()
			return DetermineContextError(actionCtx)
		}

		errCh := make(chan error, 1)
		go func() {
			errCh <- RunActionWithCancelStore(t.Context(), store, action)
		}()

		time.Sleep(5 * time.Millisecond)

		store.Cancel()

		err := <-errCh
		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
		assert.True(t, actionStarted.Load())
	})
}

func TestRunPeriodicCheckWithAction(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Run("action executed when check returns true", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		checkCount := atomic.NewInt32(0)
		actionCount := atomic.NewInt32(0)
		stopErr := commonerrors.ErrEOF
		err := SchedulePeriodicCheckWithAction(ctx,
			func(ctx context.Context) (bool, error) {
				val := checkCount.Inc()
				if val > 6 {
					return false, stopErr
				}
				return val%2 == 0, nil
			},
			func(ctx context.Context) error {
				actionCount.Inc()
				return nil
			},
			time.Millisecond,
		)
		require.Error(t, err)
		errortest.AssertError(t, err, stopErr)
		assert.Equal(t, int32(7), checkCount.Load())
		assert.Equal(t, int32(3), actionCount.Load())
	})

	t.Run("action error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		actionErr := commonerrors.ErrFailed
		err := SchedulePeriodicCheckWithAction(ctx,
			func(ctx context.Context) (bool, error) {
				return true, nil
			},
			func(ctx context.Context) error {
				return actionErr
			},
			time.Millisecond,
		)
		require.Error(t, err)
		errortest.AssertError(t, err, actionErr)
	})

	t.Run("check error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		checkErr := commonerrors.ErrFailed
		err := SchedulePeriodicCheckWithAction(ctx,
			func(ctx context.Context) (bool, error) {
				return false, checkErr
			},
			func(ctx context.Context) error {
				return nil
			},
			time.Millisecond,
		)
		require.Error(t, err)
		errortest.AssertError(t, err, checkErr)
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		actionCount := atomic.NewInt32(0)
		err := SchedulePeriodicCheckWithAction(ctx,
			func(ctx context.Context) (bool, error) {
				return true, nil
			},
			func(ctx context.Context) error {
				if actionCount.Inc() >= 2 {
					cancel()
				}
				return nil
			},
			time.Millisecond,
		)
		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
		assert.GreaterOrEqual(t, actionCount.Load(), int32(2))
	})
}

func TestWaitUntil(t *testing.T) {
	defer goleak.VerifyNone(t)
	verifiedCondition := func(ctx context.Context) (bool, error) {
		SleepWithContext(ctx, 50*time.Millisecond)
		return true, nil
	}

	t.Run("cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := WaitUntil(ctx, verifiedCondition, 10*time.Millisecond)
		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
	})
	t.Run("verified", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err := WaitUntil(ctx, verifiedCondition, 10*time.Millisecond)
		require.NoError(t, err)
	})
	t.Run("verified after multiple attempts", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		counter := atomic.NewInt32(0)
		verifiedConditionAfterAttempts := func(ctx context.Context) (bool, error) {
			SleepWithContext(ctx, time.Millisecond)
			if counter.Load() > 10 {
				return true, nil
			}
			counter.Inc()
			return false, nil
		}
		err := WaitUntil(ctx, verifiedConditionAfterAttempts, 10*time.Millisecond)
		require.NoError(t, err)
	})
	t.Run("verified with condition evaluation failure", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		counter := atomic.NewInt32(0)
		verifiedConditionAfterAttempts := func(ctx context.Context) (bool, error) {
			SleepWithContext(ctx, time.Millisecond)
			if counter.Load() > 10 {
				return false, commonerrors.ErrUnexpected
			}
			counter.Inc()
			return false, nil
		}
		err := WaitUntil(ctx, verifiedConditionAfterAttempts, 10*time.Millisecond)
		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrUnexpected)
	})
}

func TestWorkerPool(t *testing.T) {
	defer goleak.VerifyNone(t)
	for _, test := range []struct {
		name       string
		numWorkers int
		jobs       []int
		results    []int
		workerFunc func(context.Context, int) (int, bool, error)
		err        error
	}{
		{
			name:       "Success",
			numWorkers: 3,
			jobs:       []int{1, 2, 3, 4, 5},
			results:    []int{2, 4, 6, 8, 10},
			workerFunc: func(ctx context.Context, job int) (int, bool, error) {
				return job * 2, true, nil
			},
			err: nil,
		},
		{
			name:       "Invalid Num Workers",
			numWorkers: 0,
			jobs:       []int{1, 2, 3},
			results:    nil,
			workerFunc: func(ctx context.Context, job int) (int, bool, error) {
				return 0, true, nil
			},
			err: commonerrors.ErrInvalid,
		},
		{
			name:       "Worker Returns Error",
			numWorkers: 2,
			jobs:       []int{1, 2, 3},
			results:    nil,
			workerFunc: func(ctx context.Context, job int) (int, bool, error) {
				if job == 2 {
					return 0, false, errors.New("fail")
				}
				return job, true, nil
			},
			err: commonerrors.ErrUnexpected,
		},
		{
			name:       "Some ok False",
			numWorkers: 1,
			jobs:       []int{1, 2, 3},
			results:    []int{1, 3},
			workerFunc: func(ctx context.Context, job int) (int, bool, error) {
				return job, job != 2, nil
			},
			err: nil,
		},
		{
			name:       "All ok False",
			numWorkers: 1,
			jobs:       []int{1, 2, 3},
			results:    []int{},
			workerFunc: func(ctx context.Context, job int) (int, bool, error) {
				return job, false, nil
			},
			err: nil,
		},
		{
			name:       "Empty Jobs",
			numWorkers: 2,
			jobs:       []int{},
			results:    []int{},
			workerFunc: func(ctx context.Context, job int) (int, bool, error) {
				return job, true, nil
			},
			err: nil,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			results, err := WorkerPool(ctx, test.numWorkers, test.jobs, test.workerFunc)

			if test.err != nil {
				errortest.AssertError(t, err, test.err)
			} else {
				require.NoError(t, err)
				assert.ElementsMatch(t, test.results, results)
			}
		})
	}

	t.Run("Context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := WorkerPool(ctx, 100, []int{1, 2, 3}, func(ctx context.Context, job int) (int, bool, error) {
			return job, true, nil
		})

		errortest.AssertError(t, err, commonerrors.ErrCancelled)
	})
}

func TestFilterReject(t *testing.T) {
	defer goleak.VerifyNone(t)
	nums := []int{1, 2, 3, 4, 5}
	ctx := context.Background()
	results, err := Filter(ctx, 3, nums, func(n int) bool {
		return n%2 == 0
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []int{2, 4}, results)
	results, err = Reject(ctx, 3, nums, func(n int) bool {
		return n%2 == 0
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []int{1, 3, 5}, results)
	results, err = Filter(ctx, 3, nums, func(n int) bool {
		return n > 3
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []int{4, 5}, results)
	results, err = Reject(ctx, 3, nums, func(n int) bool {
		return n > 3
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []int{1, 2, 3}, results)
	results2, err := Filter(ctx, 3, []string{"", "foo", "", "bar", ""}, func(x string) bool {
		return len(x) > 0
	})

	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"foo", "bar"}, results2)
	results3, err := Reject(ctx, 3, []string{"", "foo", "", "bar", ""}, func(x string) bool {
		return len(x) > 0
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"", "", ""}, results3)
	t.Run("cancelled context", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := Filter(cancelledCtx, 3, nums, func(n int) bool {
			return n%2 == 0
		})
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
	})
}

func TestMap(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx := context.Background()
	mapped, err := Map(ctx, 3, []int{1, 2}, func(i int) string {
		return fmt.Sprintf("Hello world %v", i)
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"Hello world 1", "Hello world 2"}, mapped)
	mapped, err = Map(ctx, 3, []int64{1, 2, 3, 4}, func(x int64) string {
		return strconv.FormatInt(x, 10)
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"1", "2", "3", "4"}, mapped)
	t.Run("cancelled context", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := Map(cancelledCtx, 3, []int{1, 2}, func(i int) string {
			return fmt.Sprintf("Hello world %v", i)
		})
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
	})
}

func TestMapAndOrderedMap(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx := context.Background()
	mapped, err := OrderedMap(ctx, 3, []int{1, 2}, func(i int) string {
		return fmt.Sprintf("Hello world %v", i)
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"Hello world 1", "Hello world 2"}, mapped)
	mapped, err = OrderedMap(ctx, 3, []int64{1, 2, 3, 4}, func(x int64) string {
		return strconv.FormatInt(x, 10)
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2", "3", "4"}, mapped)
	t.Run("cancelled context", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := Map(cancelledCtx, 3, []int{1, 2}, func(i int) string {
			return fmt.Sprintf("Hello world %v", i)
		})
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
	})

	in := collection.Range(0, 1000, field.ToOptionalInt(5))
	mappedInt, err := OrderedMap(ctx, 3, in, collection.IdentityMapFunc[int]())
	require.NoError(t, err)
	assert.Equal(t, in, mappedInt)
	mappedInt, err = Map(ctx, 3, in, collection.IdentityMapFunc[int]())
	require.NoError(t, err)
	assert.NotEqual(t, in, mappedInt)
	assert.ElementsMatch(t, in, mappedInt)
}
