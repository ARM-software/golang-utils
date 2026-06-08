package step

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/retry"
)

var _ IStep[any, any] = (*Step[any])(nil)

// Step represents one stage in a same-type flow.
type Step[T any] struct {
	StepName  string
	Transform parallelisation.TransformFunc[T, T]
	options   *Options[T]
}

func NewStep[T any](transform parallelisation.TransformFunc[T, T], options ...Option[T]) *Step[T] {
	return NewNamedStep("", transform, options...)
}

func NewNamedStep[T any](name string, transform parallelisation.TransformFunc[T, T], options ...Option[T]) *Step[T] {
	return &Step[T]{StepName: name, Transform: transform, options: WithOptions(options...)}
}

func NewTransformGroupStep[T any](transform parallelisation.TransformFunc[T, T], transformGroupOptions []parallelisation.StoreOption, options ...Option[T]) *Step[T] {
	return NewNamedTransformGroupStep("", transform, transformGroupOptions, options...)
}

func NewNamedTransformGroupStep[T any](name string, transform parallelisation.TransformFunc[T, T], transformGroupOptions []parallelisation.StoreOption, options ...Option[T]) *Step[T] {
	resolved := WithOptions[T](options...).Apply(WithTransformGroupOptions[T](transformGroupOptions...))
	return &Step[T]{StepName: name, Transform: transform, options: resolved}
}

func (s *Step[T]) GetName() *string { return field.ToOptionalOrNilIfEmpty(s.StepName) }

func (s *Step[T]) GetOptions() *Options[T] {
	if s == nil {
		return nil
	}
	return s.options
}

func (s *Step[T]) Execute(ctx context.Context, current T) (output T, success bool, err error) {
	return s.ExecuteAsNew(ctx, s.GetOptions(), current)
}

func (s *Step[T]) ExecuteAsNew(ctx context.Context, options *StepOptions[T, T], current T) (output T, success bool, err error) {
	if s == nil {
		err = commonerrors.UndefinedVariable("flow step")
		return
	}
	if s.Transform == nil {
		err = commonerrors.Newf(commonerrors.ErrUndefined, "flow step [%v] has no transform", field.OptionalString(s.GetName(), unnamedStep))
		return
	}

	transform := s.Transform
	var retryPolicyValue *retry.RetryPolicyConfiguration
	if options != nil {
		retryPolicyValue = options.retryPolicy
	}
	if options != nil && len(options.transformGroupOptions) > 0 {
		transform = transformWithGroup(s.Transform, options.transformGroupOptions)
	}

	return executeWithRetry(ctx, field.OptionalString(s.GetName(), unnamedStep), current, transform, retryPolicyValue)
}
