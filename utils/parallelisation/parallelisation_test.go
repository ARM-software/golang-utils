package parallelisation

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"github.com/ARMmbed/golang-utils/utils/commonerrors"
)

func TestParallelisationWithResults(t *testing.T) {
	var values []int
	length := 100
	for i := 0; i < length; i++ {
		values = append(values, i)
	}
	action := func(arg interface{}) (result interface{}, err error) {
		result = int64(arg.(int))
		return
	}
	var temp []int64
	results, err := Parallelise(values, action, reflect.TypeOf(temp))
	require.Nil(t, err)

	temp = results.([]int64)
	assert.Equal(t, length, len(temp))
}
func TestParallelisationWithoutResults(t *testing.T) {
	var values []int
	length := 30
	for i := 0; i < length; i++ {
		values = append(values, i)
	}
	action := func(arg interface{}) (result interface{}, err error) {
		return
	}
	results, err := Parallelise(values, action, nil)
	assert.Nil(t, err)
	assert.Nil(t, results)
}

func TestParallelisationWithErrors(t *testing.T) {
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
	assert.Equal(t, anError, err)
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
			testSleep(t, test.sleep)
		})
	}
}

func TestSchedule(t *testing.T) {
	var ticks []int
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	Schedule(ctx, 10*time.Millisecond, 15*time.Millisecond, func(time time.Time) {
		ticks = append(ticks, time.Nanosecond())
	})

	time.Sleep(500 * time.Millisecond)
	// Expected number should be 49 but there is some timing variance depending on the state of the environment this is run on.
	// Therefore, we accept that the number of ticks achieved is not always accurate but close to what is expected.
	assert.GreaterOrEqual(t, len(ticks), 20)
	assert.LessOrEqual(t, len(ticks), 80)
	require.Nil(t, ctx.Err())
}

func TestRunBlockingActionWithTimeout(t *testing.T) {
	for i := 0; i < 200; i++ {
		testTimeout(t)
	}
}

func TestRunBlockingActionWithTimeoutAndContex(t *testing.T) {
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
	require.NotNil(t, err)
	assert.True(t, errors.Is(err, commonerrors.ErrTimeout))
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
	require.NotNil(t, err)
	assert.True(t, errors.Is(err, commonerrors.ErrTimeout))
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
	require.NotNil(t, err)
	assert.True(t, errors.Is(err, commonerrors.ErrTimeout))
	assert.False(t, isrunning.Load())

	isrunning.Store(true)
	nonblockingAction := func(stop chan bool) error {
		isrunning.Store(true)
		isrunning.Store(false)
		return nil
	}
	assert.True(t, isrunning.Load())
	err = RunActionWithTimeout(nonblockingAction, 10*time.Millisecond)
	require.Nil(t, err)
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
	require.NotNil(t, err)
	assert.True(t, errors.Is(err, anError))
	assert.False(t, isrunning.Load())
}

func testTimeoutWithContext(t *testing.T, ctx context.Context) {
	isrunning := atomic.NewBool(true)
	blockingAction := func(ctx context.Context) error {
		isrunning.Store(true)
		<-ctx.Done()
		isrunning.Store(false)
		return nil
	}
	assert.True(t, isrunning.Load())
	err := RunActionWithTimeoutAndContext(ctx, 10*time.Millisecond, blockingAction)
	require.Nil(t, DetermineContextError(ctx))
	require.NotNil(t, err)
	assert.True(t, errors.Is(err, commonerrors.ErrTimeout))
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
	require.Nil(t, DetermineContextError(ctx))
	require.NotNil(t, err)
	assert.True(t, errors.Is(err, commonerrors.ErrTimeout))
	assert.False(t, isrunning.Load())

	isrunning.Store(true)
	nonblockingAction := func(ctx context.Context) error {
		isrunning.Store(true)
		isrunning.Store(false)
		return nil
	}
	assert.True(t, isrunning.Load())
	err = RunActionWithTimeoutAndContext(ctx, 10*time.Millisecond, nonblockingAction)
	require.Nil(t, DetermineContextError(ctx))
	require.Nil(t, err)
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
	require.Nil(t, DetermineContextError(ctx))
	require.NotNil(t, err)
	assert.True(t, errors.Is(err, anError))
	assert.False(t, isrunning.Load())
}

func TestRunActionWithParallelCheckHappy(t *testing.T) {
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("test #%v", i), func(t *testing.T) {
			runActionWithParallelCheckHappy(t, ctx)
		})
	}
}

func TestRunActionWithParallelCheckFail(t *testing.T) {
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("test #%v", i), func(t *testing.T) {
			runActionWithParallelCheckFail(t, ctx)
		})
	}
}

func TestRunActionWithParallelCheckFailAtRandom(t *testing.T) {
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("test #%v", i), func(t *testing.T) {
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
	require.Nil(t, err)
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
	require.NotNil(t, err)
	assert.Error(t, commonerrors.ErrCancelled, err)
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
	require.NotNil(t, err)
	assert.Error(t, commonerrors.ErrCancelled, err)
}
