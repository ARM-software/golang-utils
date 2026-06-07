package pipeline

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/retry"
)

func TestGenericPipelineChain(t *testing.T) {
	step1 := NewNamedGenericStep("int-to-string", func(_ context.Context, value int) (string, bool, error) {
		return strconv.Itoa(value + 1), true, nil
	})
	step2 := NewNamedGenericStep("string-to-int", func(_ context.Context, value string) (int, bool, error) {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return 0, false, err
		}
		return parsed * 2, true, nil
	})

	p := Chain(NewPipeline(step1), step2)
	result, err := p.Execute(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, 6, result)
}

func TestGenericPipelineFallback(t *testing.T) {
	step1 := NewNamedGenericStep("int-to-string", func(_ context.Context, value int) (string, bool, error) {
		return "", false, commonerrors.ErrInvalid
	}, WithGenericFallbackOnError[int, string](func(_ context.Context, value int, _ error) (string, bool, error) {
		return strconv.Itoa(value), true, nil
	}))
	step2 := NewNamedGenericStep("string-to-int", func(_ context.Context, value string) (int, bool, error) {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return 0, false, err
		}
		return parsed + 1, true, nil
	})

	p := Chain(NewPipeline(step1), step2)
	result, err := p.Execute(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, 3, result)
}

func TestGenericPipelineRetry(t *testing.T) {
	attempts := 0
	step := NewNamedGenericStep("retry", func(_ context.Context, value int) (string, bool, error) {
		attempts++
		if attempts < 3 {
			return "", false, commonerrors.ErrUnexpected
		}
		return strconv.Itoa(value), true, nil
	}, WithGenericRetryOnError[int, string](retry.DefaultBasicRetryPolicyConfiguration()))

	result, err := NewPipeline(step).Execute(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, "7", result)
	assert.Equal(t, 3, attempts)
}

func TestGenericPipelineCompensatePrevious(t *testing.T) {
	var compensated []string
	step1 := NewNamedGenericStep("first", func(_ context.Context, value int) (string, bool, error) {
		return strconv.Itoa(value), true, nil
	}, WithGenericCompensation[int, string](func(_ context.Context, value string) error {
		compensated = append(compensated, value)
		return nil
	}))
	step2 := NewNamedGenericStep("second", func(_ context.Context, value string) (int, bool, error) {
		return 0, false, commonerrors.ErrInvalid
	}, WithGenericCompensatePreviousSteps[string, int]())

	_, err := Chain(NewPipeline(step1), step2).Execute(context.Background(), 4)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	assert.Equal(t, []string{"4"}, compensated)
}

func TestGenericTransformGroupStep(t *testing.T) {
	step := NewNamedGenericTransformGroupStep("group", func(_ context.Context, value int) (string, bool, error) {
		return strconv.Itoa(value + 1), true, nil
	}, []parallelisation.StoreOption{parallelisation.Sequential})

	result, err := NewPipeline(step).Execute(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "2", result)
}

func TestGenericPipelineUnnamedStep(t *testing.T) {
	step := NewGenericStep(func(_ context.Context, value int) (string, bool, error) {
		return strconv.Itoa(value), true, nil
	})

	result, err := NewPipeline(step).Execute(context.Background(), 5)
	require.NoError(t, err)
	assert.Equal(t, "5", result)
	assert.Nil(t, step.GetName())
}
