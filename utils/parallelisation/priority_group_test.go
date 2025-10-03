package parallelisation

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

type testExecutorFunc func(context.Context) error

func (f testExecutorFunc) Execute(ctx context.Context) error { return f(ctx) }

var _ IExecutor = (testExecutorFunc)(nil)

// testRecordingExecutor will emit supplied values into a shared slice for order comparison
type testRecordingExecutor struct {
	valueToEmit    uint
	executionOrder *atomicSlice
	duration       time.Duration
}

func newRecordingExec(valueToEmit uint, executionOrder *atomicSlice) *testRecordingExecutor {
	return newRecordingExecOfDuration(valueToEmit, executionOrder, 10*time.Millisecond)
}

func newRecordingExecOfDuration(valueToEmit uint, executionOrder *atomicSlice, duration time.Duration) *testRecordingExecutor {
	return &testRecordingExecutor{
		valueToEmit:    valueToEmit,
		executionOrder: executionOrder,
		duration:       duration,
	}
}

func (r *testRecordingExecutor) Execute(ctx context.Context) (err error) {
	r.executionOrder.Append(r.valueToEmit)
	SleepWithContext(ctx, r.duration)
	return
}

type atomicSlice struct {
	mu   sync.RWMutex
	data []uint
}

func newAtomicSlice(t *testing.T) *atomicSlice {
	t.Helper()
	return &atomicSlice{
		mu:   sync.RWMutex{},
		data: make([]uint, 0),
	}
}

func (a *atomicSlice) Append(v ...uint) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.data = append(a.data, v...)
}

func (a *atomicSlice) IsSorted() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return slices.IsSorted(a.data)
}

func (a *atomicSlice) Len() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.data)
}

func (a *atomicSlice) Data() []uint {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.data
}

func TestPriority(t *testing.T) {
	t.Run("single executor group", func(t *testing.T) {
		t.Run("all sequential", func(t *testing.T) {
			defer goleak.VerifyNone(t)

			executionOrder := newAtomicSlice(t)
			priorities := []uint{3, 1, 2, 2, 0}
			require.False(t, slices.IsSorted(priorities))

			priorityGroup := NewPriorityExecutionGroup(Sequential, RetainAfterExecution)

			priorityGroup.RegisterFunctionWithPriority(priorities[0], newRecordingExec(priorities[0], executionOrder))
			priorityGroup.RegisterFunctionWithPriority(priorities[1], newRecordingExec(priorities[1], executionOrder))
			priorityGroup.RegisterFunctionWithPriority(priorities[2], newRecordingExec(priorities[2], executionOrder))
			priorityGroup.RegisterFunctionWithPriority(priorities[3], newRecordingExec(priorities[3], executionOrder))
			priorityGroup.RegisterFunctionWithPriority(priorities[4], newRecordingExec(priorities[4], executionOrder))

			require.NoError(t, priorityGroup.Execute(context.Background()))

			assert.True(t, executionOrder.IsSorted())
			assert.EqualValues(t, executionOrder.Len(), len(priorities))
		})

		t.Run("all parallel", func(t *testing.T) {
			defer goleak.VerifyNone(t)

			executionOrder := newAtomicSlice(t)
			priorities := []uint{2, 3, 2}
			require.False(t, slices.IsSorted(priorities))

			priorityGroup := NewPriorityExecutionGroup(Parallel, Workers(4), RetainAfterExecution)

			eachRunDuration := 100 * time.Millisecond

			priorityGroup.RegisterFunctionWithPriority(priorities[0], newRecordingExecOfDuration(priorities[0], executionOrder, eachRunDuration))
			priorityGroup.RegisterFunctionWithPriority(priorities[1], newRecordingExecOfDuration(priorities[1], executionOrder, eachRunDuration))
			priorityGroup.RegisterFunctionWithPriority(priorities[2], newRecordingExecOfDuration(priorities[2], executionOrder, eachRunDuration))

			start := time.Now()
			require.NoError(t, priorityGroup.Execute(context.Background()))
			actualDuration := time.Since(start)

			assert.True(t, executionOrder.IsSorted())

			// total duration for parallel executions should be pretty much the same as the number of priorities * eachRunDuration. This will indicate that they ran concurrently
			prioritiesSorted := slices.Clone(priorities)
			slices.Sort(prioritiesSorted)
			expectedTotalDuration := eachRunDuration * time.Duration(len(slices.Compact(prioritiesSorted))) // account for different priorities
			diff := expectedTotalDuration - actualDuration
			assert.LessOrEqual(t, diff.Abs(), eachRunDuration/5)
		})

	})

	t.Run("newGroup not set", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		var priorityGroup PriorityExecutionGroup[IExecutor] // no constructor used

		var called atomic.Bool
		priorityGroup.RegisterFunction(testExecutorFunc(func(ctx context.Context) (err error) {
			called.Store(true)
			return
		}))

		err := priorityGroup.Execute(context.Background())
		assert.Error(t, err)
		errortest.AssertErrorDescription(t, err, "priority execution group has not been initialised correctly")
		assert.False(t, called.Load())
	})

	t.Run("multiple groups", func(t *testing.T) {
		t.Run("all sequential", func(t *testing.T) {
			defer goleak.VerifyNone(t)

			executionOrder := newAtomicSlice(t)

			priorities := []uint{2, 3, 2}
			require.False(t, slices.IsSorted(priorities))

			group1 := NewExecutionGroup[IExecutor](func(ctx context.Context, e IExecutor) error {
				return e.Execute(ctx)
			}, Sequential)
			group1.RegisterFunction(
				newRecordingExec(priorities[0], executionOrder),
				newRecordingExec(priorities[1], executionOrder),
			)

			group2 := NewExecutionGroup[IExecutor](func(ctx context.Context, e IExecutor) error {
				return e.Execute(ctx)
			}, Sequential)
			group2.RegisterFunction(
				newRecordingExec(priorities[2], executionOrder),
			)

			priorityGroup := NewPriorityExecutionGroup(Sequential)
			priorityGroup.RegisterFunctionWithPriority(5, group1)
			priorityGroup.RegisterFunctionWithPriority(1, group2)

			require.NoError(t, priorityGroup.Execute(context.Background()))

			expected := []uint{priorities[2], priorities[0], priorities[1]} // 2 then 2,3
			assert.Equal(t, expected, executionOrder.Data())
		})

		t.Run("two parallel groups (outer sequential)", func(t *testing.T) {
			defer goleak.VerifyNone(t)

			executionOrder := newAtomicSlice(t)
			priorities := []uint{20, 20, 10, 10}

			testDuration := 100 * time.Millisecond

			group1 := NewExecutionGroup[IExecutor](func(ctx context.Context, e IExecutor) error {
				return e.Execute(ctx)
			}, Parallel, Workers(4))
			group1.RegisterFunction(
				newRecordingExecOfDuration(priorities[0], executionOrder, testDuration),
				newRecordingExecOfDuration(priorities[1], executionOrder, testDuration),
			)

			group2 := NewExecutionGroup[IExecutor](func(ctx context.Context, e IExecutor) error {
				return e.Execute(ctx)
			}, Parallel, Workers(4))
			group2.RegisterFunction(
				newRecordingExecOfDuration(priorities[2], executionOrder, testDuration),
				newRecordingExecOfDuration(priorities[3], executionOrder, testDuration),
			)

			priorityGroup := NewPriorityExecutionGroup(Sequential)
			priorityGroup.RegisterFunctionWithPriority(5, group1)
			priorityGroup.RegisterFunctionWithPriority(1, group2)

			start := time.Now()
			require.NoError(t, priorityGroup.Execute(context.Background()))
			actualDuration := time.Since(start)

			require.EqualValues(t, executionOrder.Len(), 4)
			assert.IsNonDecreasing(t, executionOrder.Data())

			prioritiesSorted := slices.Clone(priorities)
			slices.Sort(prioritiesSorted)
			expectedTotalDuration := testDuration * 2 // two parallel tests in order so should take 2*testDuration
			diff := expectedTotalDuration - actualDuration
			assert.LessOrEqual(t, diff.Abs(), testDuration/5)
		})

		t.Run("mixed (group2 sequential and group1 parallel)", func(t *testing.T) {
			defer goleak.VerifyNone(t)

			executionOrder := newAtomicSlice(t)

			priorities := []uint{20, 21, 10, 11}

			group1 := NewExecutionGroup[IExecutor](func(ctx context.Context, e IExecutor) error {
				return e.Execute(ctx)
			}, Parallel, Workers(4))
			group1.RegisterFunction(
				newRecordingExec(priorities[0], executionOrder),
				newRecordingExec(priorities[1], executionOrder),
			)

			group2 := NewExecutionGroup[IExecutor](func(ctx context.Context, e IExecutor) error {
				return e.Execute(ctx)
			}, Sequential)
			group2.RegisterFunction(
				newRecordingExec(priorities[2], executionOrder),
				newRecordingExec(priorities[3], executionOrder),
			)

			priorityGroup := NewPriorityExecutionGroup(Sequential)
			priorityGroup.RegisterFunctionWithPriority(5, group1)
			priorityGroup.RegisterFunctionWithPriority(1, group2)

			require.NoError(t, priorityGroup.Execute(context.Background()))

			require.EqualValues(t, executionOrder.Len(), 4)
			assert.Equal(t, priorities[2:], executionOrder.Data()[:2])         // 10, 11 in order
			assert.ElementsMatch(t, priorities[:2], executionOrder.Data()[2:]) // 20 & 21 any order
		})

		t.Run("two parallel groups in outer parallel (outer parallel same priority)", func(t *testing.T) {
			defer goleak.VerifyNone(t)

			executionOrder := newAtomicSlice(t)

			priorities := []uint{20, 20, 10, 10} // each group will have all members run in parallel

			testDuration := 100 * time.Millisecond

			group1 := NewExecutionGroup[IExecutor](func(ctx context.Context, e IExecutor) error {
				return e.Execute(ctx)
			}, Parallel, Workers(4))
			group1.RegisterFunction(
				newRecordingExecOfDuration(priorities[0], executionOrder, testDuration),
				newRecordingExecOfDuration(priorities[1], executionOrder, testDuration),
			)

			group2 := NewExecutionGroup[IExecutor](func(ctx context.Context, e IExecutor) error {
				return e.Execute(ctx)
			}, Parallel)
			group2.RegisterFunction(
				newRecordingExecOfDuration(priorities[2], executionOrder, testDuration),
				newRecordingExecOfDuration(priorities[3], executionOrder, testDuration),
			)

			priorityGroup := NewPriorityExecutionGroup(Parallel, Workers(4))
			priorityGroup.RegisterFunctionWithPriority(1, group1)
			priorityGroup.RegisterFunctionWithPriority(1, group2)

			start := time.Now()
			require.NoError(t, priorityGroup.Execute(context.Background()))
			actualDuration := time.Since(start)

			prioritiesSorted := slices.Clone(priorities)
			slices.Sort(prioritiesSorted)
			expectedTotalDuration := testDuration // all should run at once since the different priorities are in different groups
			diff := expectedTotalDuration - actualDuration
			assert.LessOrEqual(t, diff.Abs(), testDuration/5)
		})

		t.Run("two parallel groups in outer parallel (outer parallel different priorities)", func(t *testing.T) {
			defer goleak.VerifyNone(t)

			executionOrder := newAtomicSlice(t)

			priorities := []uint{20, 20, 10, 10} // each group will have all members run in parallel

			testDuration := 100 * time.Millisecond

			group1 := NewExecutionGroup[IExecutor](func(ctx context.Context, e IExecutor) error {
				return e.Execute(ctx)
			}, Parallel, Workers(4))
			group1.RegisterFunction(
				newRecordingExecOfDuration(priorities[0], executionOrder, testDuration),
				newRecordingExecOfDuration(priorities[1], executionOrder, testDuration),
			)

			group2 := NewExecutionGroup[IExecutor](func(ctx context.Context, e IExecutor) error {
				return e.Execute(ctx)
			}, Parallel)
			group2.RegisterFunction(
				newRecordingExecOfDuration(priorities[2], executionOrder, testDuration),
				newRecordingExecOfDuration(priorities[3], executionOrder, testDuration),
			)

			priorityGroup := NewPriorityExecutionGroup(Parallel, Workers(4))
			priorityGroup.RegisterFunctionWithPriority(5, group1)
			priorityGroup.RegisterFunctionWithPriority(1, group2)

			start := time.Now()
			require.NoError(t, priorityGroup.Execute(context.Background()))
			actualDuration := time.Since(start)

			prioritiesSorted := slices.Clone(priorities)
			slices.Sort(prioritiesSorted)
			expectedTotalDuration := 2 * testDuration // parallel groups have different priorities so act in sequential manner
			diff := expectedTotalDuration - actualDuration
			assert.LessOrEqual(t, diff.Abs(), testDuration/5)
		})

	})

	t.Run("default priority (zero) is highest", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		executionOrder := newAtomicSlice(t)

		priorityGroup := NewPriorityExecutionGroup(Sequential)

		priorityGroup.RegisterFunctionWithPriority(1, newRecordingExec(1, executionOrder))

		priorityGroup.RegisterFunction(newRecordingExec(0, executionOrder))

		require.NoError(t, priorityGroup.Execute(context.Background()))
		assert.Equal(t, []uint{0, 1}, executionOrder.Data())
		assert.Equal(t, 2, priorityGroup.Len())
	})

	t.Run("cancel", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		priorityGroup := NewPriorityExecutionGroup(Parallel)

		priorityGroup.RegisterFunction(testExecutorFunc(func(ctx context.Context) error {
			return DetermineContextError(ctx)
		}))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := priorityGroup.Execute(ctx)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
	})

	t.Run("timeout", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		priorityGroup := NewPriorityExecutionGroup(Parallel)

		priorityGroup.RegisterFunction(testExecutorFunc(func(ctx context.Context) error {
			return DetermineContextError(ctx)
		}))

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		time.Sleep(100 * time.Millisecond)

		err := priorityGroup.Execute(ctx)
		errortest.AssertError(t, err, commonerrors.ErrTimeout)
	})
}
