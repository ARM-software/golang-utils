package step

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/retry"
)

func transformWithGroup[I, O any](transform TransformFunc[I, O], options []parallelisation.StoreOption) TransformFunc[I, O] {
	return func(ctx context.Context, input I) (output O, success bool, err error) {
		return executeTransformGroup(ctx, parallelisation.NewTransformGroup[I, O](transform, options...), input)
	}
}

type transformGroupStep[T any] struct {
	stepName string
	group    *parallelisation.TransformGroup[T, T]
}

var _ IStep[any, any] = (*transformGroupStep[any])(nil)

// NewStepFromTransformGroup returns a step backed by a transform group.
func NewStepFromTransformGroup[T any](name *string, group *parallelisation.TransformGroup[T, T]) IStep[T, T] {
	return &transformGroupStep[T]{group: group, stepName: field.OptionalString(name, "")}
}

func (s *transformGroupStep[T]) GetName() *string { return field.ToOptionalOrNilIfEmpty(s.stepName) }

func (s *transformGroupStep[T]) GetOptions() *StepOptions[T, T] { return nil }

func (s *transformGroupStep[T]) Execute(ctx context.Context, current T) (output T, success bool, err error) {
	return s.ExecuteAsNew(ctx, nil, current)
}

func (s *transformGroupStep[T]) ExecuteAsNew(ctx context.Context, options *StepOptions[T, T], current T) (output T, success bool, err error) {
	if s == nil {
		err = commonerrors.UndefinedVariable("flow step")
		return
	}
	if s.group == nil {
		err = commonerrors.Newf(commonerrors.ErrUndefined, "flow step [%v] has no transform group", field.OptionalString(s.GetName(), unnamedStep))
		return
	}

	transform := func(stepCtx context.Context, value T) (output T, success bool, err error) {
		var clonedGroup *parallelisation.TransformGroup[T, T]
		clonedGroup, err = cloneTransformGroup(s.group)
		if err != nil {
			return
		}
		return executeTransformGroup(stepCtx, clonedGroup, value)
	}

	var retryPolicyValue *retry.RetryPolicyConfiguration
	if options != nil {
		retryPolicyValue = options.retryPolicy
	}

	return executeWithRetry(ctx, field.OptionalString(s.GetName(), unnamedStep), current, transform, retryPolicyValue)
}

func cloneTransformGroup[I, O any](group *parallelisation.TransformGroup[I, O]) (clonedGroup *parallelisation.TransformGroup[I, O], err error) {
	if group == nil {
		err = commonerrors.UndefinedVariable("transform group")
		return
	}
	clone, ok := group.Clone().(*parallelisation.TransformGroup[I, O])
	if !ok {
		err = commonerrors.Newf(commonerrors.ErrUnexpected, "unable to clone transform group [%T]", group)
		return
	}
	clonedGroup = clone
	return
}

func executeTransformGroup[I, O any](ctx context.Context, group *parallelisation.TransformGroup[I, O], input I) (output O, success bool, err error) {
	if group == nil {
		err = commonerrors.UndefinedVariable("transform group")
		return
	}
	if err = group.Inputs(ctx, input); err != nil {
		return
	}
	if err = group.Transform(ctx); err != nil {
		return
	}
	var outputs []O
	outputs, err = group.OrderedOutputs(ctx)
	if err != nil {
		return
	}
	if len(outputs) == 0 {
		return
	}
	output = outputs[0]
	success = true
	return
}
