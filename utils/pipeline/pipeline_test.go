package pipeline

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/retry"
)

func TestPipelineBreakOnFirstError(t *testing.T) {
	p := NewSimplePipeline[int](BreakOnFirstError[int]())
	p.RegisterTransform("step1", func(_ context.Context, v int) (int, bool, error) { return v + 1, true, nil })
	p.RegisterTransform("step2", func(_ context.Context, v int) (int, bool, error) { return 0, false, commonerrors.ErrInvalid })
	p.RegisterTransform("step3", func(_ context.Context, v int) (int, bool, error) { return v + 100, true, nil })

	result, err := p.Execute(context.Background(), 1)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	assert.Equal(t, 2, result)
}

func TestPipelineExecuteAll(t *testing.T) {
	p := NewSimplePipeline[int](ExecuteAll[int]())
	p.RegisterTransform("step1", func(_ context.Context, v int) (int, bool, error) { return v + 1, true, nil })
	p.RegisterTransform("step2", func(_ context.Context, v int) (int, bool, error) { return 0, false, commonerrors.ErrInvalid })
	p.RegisterTransform("step3", func(_ context.Context, v int) (int, bool, error) { return v + 1, true, nil })

	result, err := p.Execute(context.Background(), 1)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	assert.Equal(t, 3, result)
}

func TestPipelineRetryOnError(t *testing.T) {
	var attempts atomic.Int32
	p := NewSimplePipeline[int]()
	p.RegisterTransform("retry", func(_ context.Context, v int) (int, bool, error) {
		if attempts.Add(1) < 3 {
			return 0, false, commonerrors.ErrUnexpected
		}
		return v + 1, true, nil
	}, RetryOnError[int](retry.DefaultBasicRetryPolicyConfiguration()))

	result, err := p.Execute(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 2, result)
	assert.Equal(t, int32(3), attempts.Load())
}

func TestPipelineSkipStepOnError(t *testing.T) {
	p := NewSimplePipeline[int]()
	p.RegisterTransform("step1", func(_ context.Context, v int) (int, bool, error) { return v + 1, true, nil })
	p.RegisterTransform("step2", func(_ context.Context, v int) (int, bool, error) { return 0, false, commonerrors.ErrInvalid }, SkipStepOnError[int]())
	p.RegisterTransform("step3", func(_ context.Context, v int) (int, bool, error) { return v + 1, true, nil })

	result, err := p.Execute(context.Background(), 1)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	assert.Equal(t, 3, result)
}

func TestPipelineFallbackOnError(t *testing.T) {
	p := NewSimplePipeline[int]()
	p.RegisterTransform("step1", func(_ context.Context, v int) (int, bool, error) { return v + 1, true, nil })
	p.RegisterTransform("step2", func(_ context.Context, v int) (int, bool, error) { return 0, false, commonerrors.ErrInvalid }, FallbackOnError[int](func(_ context.Context, current int, _ error) (int, bool, error) {
		return current + 10, true, nil
	}))
	p.RegisterTransform("step3", func(_ context.Context, v int) (int, bool, error) { return v + 1, true, nil })

	result, err := p.Execute(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 13, result)
}

func TestPipelineCompensatePreviousSteps(t *testing.T) {
	var compensated []int
	p := NewSimplePipeline[int]()
	p.RegisterTransform("step1", func(_ context.Context, v int) (int, bool, error) { return v + 1, true, nil }, WithCompensation[int](func(_ context.Context, v int) error {
		compensated = append(compensated, v)
		return nil
	}))
	p.RegisterTransform("step2", func(_ context.Context, v int) (int, bool, error) { return v + 1, true, nil }, WithCompensation[int](func(_ context.Context, v int) error {
		compensated = append(compensated, v)
		return nil
	}))
	p.RegisterTransform("step3", func(_ context.Context, v int) (int, bool, error) { return 0, false, commonerrors.ErrInvalid }, CompensatePreviousSteps[int]())

	result, err := p.Execute(context.Background(), 1)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	assert.Equal(t, 3, result)
	assert.Equal(t, []int{3, 2}, compensated)
}

func TestPipelineJoinErrorsAndContinue(t *testing.T) {
	p := NewSimplePipeline[int](JoinErrorsAndContinue[int]())
	p.RegisterTransform("step1", func(_ context.Context, v int) (int, bool, error) { return 0, false, commonerrors.ErrInvalid })
	p.RegisterTransform("step2", func(_ context.Context, v int) (int, bool, error) { return 0, false, commonerrors.ErrUnexpected })
	p.RegisterTransform("step3", func(_ context.Context, v int) (int, bool, error) { return v + 1, true, nil })

	result, err := p.Execute(context.Background(), 1)
	require.Error(t, err)
	assert.True(t, errors.Is(err, commonerrors.ErrInvalid))
	assert.True(t, errors.Is(err, commonerrors.ErrUnexpected))
	assert.Equal(t, 2, result)
}

func TestPipelineStopsOnCancellation(t *testing.T) {
	p := NewSimplePipeline[int]()
	p.RegisterTransform("cancel", func(ctx context.Context, v int) (int, bool, error) {
		<-ctx.Done()
		return 0, false, ctx.Err()
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.Execute(ctx, 1)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
}

func TestPipelineStepUsesTransformGroup(t *testing.T) {
	p := NewSimplePipeline[int]()
	p.RegisterTransform("group", func(_ context.Context, v int) (int, bool, error) {
		return v + 1, true, nil
	}, WithTransformGroupOptions[int](parallelisation.Sequential))

	result, err := p.Execute(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 2, result)
}

func TestPipelineRegisterTransformGroup(t *testing.T) {
	p := NewSimplePipeline[int]()
	name := "group"
	p.RegisterTransformGroup(&name, parallelisation.NewTransformGroup(func(_ context.Context, v int) (int, bool, error) {
		return v + 1, true, nil
	}, parallelisation.Sequential))

	result, err := p.Execute(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 2, result)
}

func TestPipelineRegisterTransformGroupUsesPipelineRetry(t *testing.T) {
	var attempts atomic.Int32
	p := NewSimplePipeline[int](RetryOnError[int](retry.DefaultBasicRetryPolicyConfiguration()))
	p.RegisterTransformGroup(nil, parallelisation.NewTransformGroup(func(_ context.Context, v int) (int, bool, error) {
		if attempts.Add(1) < 3 {
			return 0, false, commonerrors.ErrUnexpected
		}
		return v + 1, true, nil
	}, parallelisation.Sequential))

	result, err := p.Execute(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 2, result)
	assert.Equal(t, int32(3), attempts.Load())
}

func TestPipelineUnnamedStep(t *testing.T) {
	p := NewSimplePipeline[int]()
	step := NewStep(func(_ context.Context, v int) (int, bool, error) {
		return v + 1, true, nil
	})
	p.RegisterStep(step)

	result, err := p.Execute(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 2, result)
	assert.Nil(t, step.GetName())
}
