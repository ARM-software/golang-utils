package parallelisation

import (
	"context"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/collection"
)

func TestNewTransformGroup(t *testing.T) {
	defer goleak.VerifyNone(t)
	tr := func(ctx context.Context, i string) (o int, success bool, err error) {
		err = DetermineContextError(ctx)
		if err != nil {
			return
		}
		o, err = strconv.Atoi(i)
		if err == nil {
			success = true
		}
		return
	}
	g := NewTransformGroup[string, int](tr, RetainAfterExecution, Parallel)
	assert.Zero(t, g.Len())
	o, err := g.Outputs(context.Background())
	require.NoError(t, err)
	assert.Empty(t, o)
	numberOfInput := 50
	in := collection.Range(0, numberOfInput, nil)
	in2 := collection.Map[int, string](in, strconv.Itoa)
	err = g.Inputs(context.Background(), in2...)
	require.NoError(t, err)
	assert.Equal(t, numberOfInput, g.Len())
	o, err = g.Outputs(context.Background())
	require.NoError(t, err)
	assert.Empty(t, o)
	err = g.Transform(context.Background())
	require.NoError(t, err)
	o, err = g.Outputs(context.Background())
	require.NoError(t, err)
	assert.ElementsMatch(t, in, o)
	o, err = Transform[string, int](context.Background(), slices.Values(in2), tr, RetainAfterExecution, Parallel)
	require.NoError(t, err)
	assert.ElementsMatch(t, in, o)
	o, err = g.OrderedOutputs(context.Background())
	require.NoError(t, err)
	assert.Empty(t, o)
	err = g.Transform(context.Background())
	require.NoError(t, err)
	o, err = g.OrderedOutputs(context.Background())
	require.NoError(t, err)
	assert.Equal(t, in, o)
	o, err = TransformInOrder[string, int](context.Background(), slices.Values(in2), tr, RetainAfterExecution, Parallel)
	require.NoError(t, err)
	assert.Equal(t, in, o)
	err = g.Inputs(context.Background(), in2...)
	require.NoError(t, err)
	assert.Equal(t, 2*numberOfInput, g.Len())
	o, err = g.Outputs(context.Background())
	require.NoError(t, err)
	assert.Empty(t, o)
	err = g.Transform(context.Background())
	require.NoError(t, err)
	o, err = g.Outputs(context.Background())
	require.NoError(t, err)
	assert.ElementsMatch(t, append(in, in...), o)
	err = g.Transform(context.Background())
	require.NoError(t, err)
	o, err = g.OrderedOutputs(context.Background())
	require.NoError(t, err)
	assert.Equal(t, append(in, in...), o)
}

func TestNewTransformGroup_Clone(t *testing.T) {
	defer goleak.VerifyNone(t)
	tr := func(ctx context.Context, i string) (o int, success bool, err error) {
		err = DetermineContextError(ctx)
		if err != nil {
			return
		}
		o, err = strconv.Atoi(i)
		if err == nil {
			success = true
		}
		return
	}
	g := NewTransformGroup[string, int](tr, RetainAfterExecution, Parallel)
	assert.Zero(t, g.Len())
	o, err := g.Outputs(context.Background())
	require.NoError(t, err)
	assert.Empty(t, o)
	numberOfInput := 50
	in := collection.RangeSequence(0, numberOfInput, nil)
	err = g.InputSequence(context.Background(), collection.MapSequence[int, string](in, strconv.Itoa))
	require.NoError(t, err)
	assert.Equal(t, numberOfInput, g.Len())
	clone := g.Clone()
	require.NotNil(t, clone)
	gClone, ok := clone.(*TransformGroup[string, int])
	require.True(t, ok)
	assert.Equal(t, numberOfInput, gClone.Len())
	err = gClone.Transform(context.Background())
	require.NoError(t, err)
	o, err = gClone.Outputs(context.Background())
	require.NoError(t, err)
	assert.ElementsMatch(t, slices.Collect[int](in), o)
}
