/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package parallelisation

import (
	"context"
	"errors"
	"fmt"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
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
}

func runActionWithParallelCheckFailAtRandom(t *testing.T, ctx context.Context) {
	counter := atomic.NewInt32(0)
	checkAction := func(ctx context.Context) bool {
		counter.Add(1)
		fmt.Println("Check #", counter.String())
		return rand.Intn(2) != 0 && counter.Load() < 10 //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
	}
	action := func(ctx context.Context) error {
		time.Sleep(150 * time.Millisecond)
		return nil
	}
	err := RunActionWithParallelCheck(ctx, action, checkAction, 10*time.Millisecond)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
}
