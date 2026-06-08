package step

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/retry"
)

// GenericStep represents a step whose output type may differ from its input type.
type GenericStep[I, O any] struct {
	StepName  string
	Transform parallelisation.TransformFunc[I, O]
	options   *GenericStepOptions[I, O]
}

func NewGenericStep[I, O any](transform parallelisation.TransformFunc[I, O], options ...GenericStepOption[I, O]) *GenericStep[I, O] {
	return NewNamedGenericStep("", transform, options...)
}

func NewNamedGenericStep[I, O any](name string, transform parallelisation.TransformFunc[I, O], options ...GenericStepOption[I, O]) *GenericStep[I, O] {
	return &GenericStep[I, O]{StepName: name, Transform: transform, options: WithGenericStepOptions(options...)}
}

func NewGenericTransformGroupStep[I, O any](transform parallelisation.TransformFunc[I, O], transformGroupOptions []parallelisation.StoreOption, options ...GenericStepOption[I, O]) *GenericStep[I, O] {
	return NewNamedGenericTransformGroupStep("", transform, transformGroupOptions, options...)
}

func NewNamedGenericTransformGroupStep[I, O any](name string, transform parallelisation.TransformFunc[I, O], transformGroupOptions []parallelisation.StoreOption, options ...GenericStepOption[I, O]) *GenericStep[I, O] {
	resolved := WithGenericStepOptions(options...).Apply(WithGenericTransformGroupOptions[I, O](transformGroupOptions...))
	return &GenericStep[I, O]{StepName: name, Transform: transform, options: resolved}
}

func (s *GenericStep[I, O]) GetName() *string { return field.ToOptionalOrNilIfEmpty(s.StepName) }

func (s *GenericStep[I, O]) GetOptions() *StepOptions[I, O] {
	if s == nil {
		return nil
	}
	return s.options
}

func (s *GenericStep[I, O]) Execute(ctx context.Context, current I) (output O, success bool, err error) {
	return s.ExecuteAsNew(ctx, s.GetOptions(), current)
}

func (s *GenericStep[I, O]) ExecuteAsNew(ctx context.Context, options *StepOptions[I, O], current I) (output O, success bool, err error) {
	if s == nil {
		err = commonerrors.UndefinedVariable("flow step")
		return
	}
	if s.Transform == nil {
		err = commonerrors.Newf(commonerrors.ErrUndefined, "flow step [%v] has no transform", field.OptionalString(s.GetName(), unnamedStep))
		return
	}

	transform := s.Transform
	if options != nil && len(options.transformGroupOptions) > 0 {
		transform = transformWithGroup(s.Transform, options.transformGroupOptions)
	}

	var retryPolicyValue *retry.RetryPolicyConfiguration
	if options != nil {
		retryPolicyValue = options.retryPolicy
	}

	return executeWithRetry(ctx, field.OptionalString(s.GetName(), unnamedStep), current, transform, retryPolicyValue)
}
