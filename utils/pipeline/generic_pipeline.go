package pipeline

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/retry"
)

// GenericStep represents a pipeline step whose output type may differ from its
// input type.
type GenericStep[I, O any] struct {
	StepName  string
	Transform parallelisation.TransformFunc[I, O]
	options   *GenericStepOptions[I, O]
}

var _ IPipeline[any, any] = (*Pipeline[any, any])(nil)

// NewGenericStep returns a generic pipeline step.
func NewGenericStep[I, O any](transform parallelisation.TransformFunc[I, O], options ...GenericStepOption[I, O]) *GenericStep[I, O] {
	return NewNamedGenericStep("", transform, options...)
}

// NewNamedGenericStep returns a named generic pipeline step.
func NewNamedGenericStep[I, O any](name string, transform parallelisation.TransformFunc[I, O], options ...GenericStepOption[I, O]) *GenericStep[I, O] {
	return &GenericStep[I, O]{
		StepName:  name,
		Transform: transform,
		options:   WithGenericStepOptions(options...),
	}
}

// NewGenericTransformGroupStep returns an unnamed generic step that executes
// transform through a parallelisation transform group.
func NewGenericTransformGroupStep[I, O any](transform parallelisation.TransformFunc[I, O], transformGroupOptions []parallelisation.StoreOption, options ...GenericStepOption[I, O]) *GenericStep[I, O] {
	return NewNamedGenericTransformGroupStep("", transform, transformGroupOptions, options...)
}

// NewNamedGenericTransformGroupStep returns a named generic step that executes
// transform through a parallelisation transform group.
func NewNamedGenericTransformGroupStep[I, O any](name string, transform parallelisation.TransformFunc[I, O], transformGroupOptions []parallelisation.StoreOption, options ...GenericStepOption[I, O]) *GenericStep[I, O] {
	resolved := WithGenericStepOptions(options...).Apply(WithGenericTransformGroupOptions[I, O](transformGroupOptions...))
	return &GenericStep[I, O]{
		StepName:  name,
		Transform: transform,
		options:   resolved,
	}
}

// GetName returns the human-readable name of the step.
func (s *GenericStep[I, O]) GetName() *string {
	return field.ToOptionalOrNilIfEmpty(s.StepName)
}

// GetOptions returns the step-specific execution options.
func (s *GenericStep[I, O]) GetOptions() *StepOptions[I, O] {
	if s == nil {
		return nil
	}
	return s.options
}

// Execute runs the step against the current pipeline state.
func (s *GenericStep[I, O]) Execute(ctx context.Context, current I) (output O, success bool, err error) {
	return s.ExecuteAsNew(ctx, s.GetOptions(), current)
}

// ExecuteAsNew runs the step against the current pipeline state using the
// supplied execution options.
func (s *GenericStep[I, O]) ExecuteAsNew(ctx context.Context, options *StepOptions[I, O], current I) (output O, success bool, err error) {
	if s == nil {
		err = commonerrors.UndefinedVariable("pipeline step")
		return
	}
	if s.Transform == nil {
		err = commonerrors.Newf(commonerrors.ErrUndefined, "pipeline step [%v] has no transform", field.OptionalString(s.GetName(), unnamedStep))
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

type genericCompensationRecord struct {
	compensate func(context.Context) error
}

// Pipeline executes typed steps where each step may transform the input
// into a different output type.
type Pipeline[I, O any] struct {
	run func(context.Context, I) (O, []genericCompensationRecord, error)
}

// NewPipeline returns a pipeline starting with step.
func NewPipeline[I, O any](step IStep[I, O]) *Pipeline[I, O] {
	return &Pipeline[I, O]{
		run: func(ctx context.Context, input I) (O, []genericCompensationRecord, error) {
			return executeGenericPipelineStep(ctx, step, input, nil)
		},
	}
}

// Chain appends step to pipeline, ensuring at compile time that the previous
// output type matches the next step input type.
func Chain[I, M, O any](pipeline *Pipeline[I, M], step IStep[M, O]) *Pipeline[I, O] {
	return &Pipeline[I, O]{
		run: func(ctx context.Context, input I) (output O, compensations []genericCompensationRecord, err error) {
			var intermediate M
			intermediate, compensations, err = pipeline.run(ctx, input)
			if err != nil {
				return output, compensations, err
			}
			return executeGenericPipelineStep(ctx, step, intermediate, compensations)
		},
	}
}

// Execute runs the generic pipeline over input until completion or
// cancellation.
func (p *Pipeline[I, O]) Execute(ctx context.Context, input I) (output O, err error) {
	if p == nil || p.run == nil {
		err = commonerrors.UndefinedVariable("pipeline")
		return
	}
	output, _, err = p.run(ctx, input)
	return
}

func executeGenericPipelineStep[I, O any](ctx context.Context, step IStep[I, O], input I, previous []genericCompensationRecord) (output O, compensations []genericCompensationRecord, err error) {
	compensations = previous
	if err := parallelisation.DetermineContextError(ctx); err != nil {
		return output, compensations, err
	}
	if step == nil {
		err = commonerrors.UndefinedVariable("pipeline step")
		return
	}
	options := step.GetOptions()

	var success bool
	output, success, err = step.Execute(ctx, input)
	if err != nil {
		if options != nil && options.compensatePrevious {
			compErr := runGenericCompensations(ctx, previous)
			if compErr != nil {
				err = commonerrors.Join(err, compErr)
			}
		}
		if options != nil && options.fallback != nil {
			fallbackValue, fallbackSuccess, fallbackErr := options.fallback(ctx, input, err)
			if fallbackErr != nil {
				err = commonerrors.Join(err, fallbackErr)
				return
			}
			if fallbackSuccess {
				output = fallbackValue
				err = nil
				success = true
			}
		}
	}
	if err != nil {
		return
	}
	if !success {
		err = commonerrors.New(commonerrors.ErrUnexpected, "generic pipeline step did not produce an output")
		return
	}

	if options != nil && options.compensation != nil {
		compensations = append(compensations, genericCompensationRecord{compensate: func(compCtx context.Context) error {
			return options.compensation(compCtx, output)
		}})
	}

	return
}

func runGenericCompensations(ctx context.Context, compensations []genericCompensationRecord) (err error) {
	for _, compensation := range collection.Reverse(compensations) {
		if compensation.compensate == nil {
			continue
		}
		err = commonerrors.Join(err, compensation.compensate(ctx))
	}
	return
}
