package pipeline

import (
	"context"
	"slices"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/retry"
)

const unnamedStep = "unnamed step"

type compensationRecord[T any] struct {
	compensate CompensationFunc[T]
	state      T
}

var _ IStep[any, any] = (*Step[any])(nil)
var _ ISimplePipeline[any] = (*SimplePipeline[any])(nil)

// Step represents one stage in a pipeline.
type Step[T any] struct {
	StepName  string
	Transform parallelisation.TransformFunc[T, T]
	options   *Options[T]
}

// NewStep returns an unnamed pipeline step.
func NewStep[T any](transform parallelisation.TransformFunc[T, T], options ...Option[T]) *Step[T] {
	return NewNamedStep("", transform, options...)
}

// NewNamedStep returns a named pipeline step.
func NewNamedStep[T any](name string, transform parallelisation.TransformFunc[T, T], options ...Option[T]) *Step[T] {
	return &Step[T]{
		StepName:  name,
		Transform: transform,
		options:   WithOptions(options...),
	}
}

// NewTransformGroupStep returns an unnamed pipeline step that executes
// transform through a parallelisation transform group.
func NewTransformGroupStep[T any](transform parallelisation.TransformFunc[T, T], transformGroupOptions []parallelisation.StoreOption, options ...Option[T]) *Step[T] {
	return NewNamedTransformGroupStep("", transform, transformGroupOptions, options...)
}

// NewNamedTransformGroupStep returns a named pipeline step that executes
// transform through a parallelisation transform group.
func NewNamedTransformGroupStep[T any](name string, transform parallelisation.TransformFunc[T, T], transformGroupOptions []parallelisation.StoreOption, options ...Option[T]) *Step[T] {
	resolved := WithOptions[T](options...).Apply(WithTransformGroupOptions[T](transformGroupOptions...))
	return &Step[T]{
		StepName:  name,
		Transform: transform,
		options:   resolved,
	}
}

// GetName returns the human-readable name of the step.
func (s *Step[T]) GetName() *string {
	return field.ToOptionalOrNilIfEmpty(s.StepName)
}

// GetOptions returns the step-specific execution options.
func (s *Step[T]) GetOptions() *Options[T] {
	if s == nil {
		return nil
	}
	return s.options
}

// Execute runs the step against the current pipeline state.
func (s *Step[T]) Execute(ctx context.Context, current T) (output T, success bool, err error) {
	return s.ExecuteAsNew(ctx, s.GetOptions(), current)
}

// ExecuteAsNew runs the step against the current pipeline state using the
// supplied execution options.
func (s *Step[T]) ExecuteAsNew(ctx context.Context, options *StepOptions[T, T], current T) (output T, success bool, err error) {
	if s == nil {
		err = commonerrors.UndefinedVariable("pipeline step")
		return
	}
	if s.Transform == nil {
		err = commonerrors.Newf(commonerrors.ErrUndefined, "pipeline step [%v] has no transform", field.OptionalString(s.GetName(), unnamedStep))
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

// SimplePipeline executes ordered steps over a single state value.
type SimplePipeline[T any] struct {
	steps   []IStep[T, T]
	options *Options[T]
}

// NewSimplePipeline returns a same-type pipeline configured with the supplied options.
func NewSimplePipeline[T any](options ...Option[T]) *SimplePipeline[T] {
	return &SimplePipeline[T]{
		steps:   make([]IStep[T, T], 0),
		options: WithOptions(options...),
	}
}

// RegisterStep appends steps to the pipeline.
func (p *SimplePipeline[T]) RegisterStep(steps ...IStep[T, T]) {
	p.steps = append(p.steps, steps...)
}

// RegisterTransform appends a new step built from transform.
func (p *SimplePipeline[T]) RegisterTransform(name string, transform parallelisation.TransformFunc[T, T], options ...Option[T]) {
	p.RegisterStep(NewNamedStep(name, transform, options...))
}

// RegisterTransformGroup appends a step backed by a transform group.
func (p *SimplePipeline[T]) RegisterTransformGroup(name *string, group *parallelisation.TransformGroup[T, T]) {
	p.RegisterStep(newTransformGroupStep(name, group))
}

// Execute runs the pipeline over input until completion, cancellation, or a
// configured stop condition.
func (p *SimplePipeline[T]) Execute(ctx context.Context, input T) (current T, executeErr error) {
	current = input
	compensations := make([]compensationRecord[T], 0)

	err := collection.EachRef(slices.Values(p.steps), func(step *IStep[T, T]) error {
		if err := parallelisation.DetermineContextError(ctx); err != nil {
			return err
		}
		if step == nil || *step == nil {
			return commonerrors.UndefinedVariable("pipeline step")
		}
		currentStep := *step

		effective := WithOptions[T]()
		effective.Merge(p.options)
		effective.Merge(currentStep.GetOptions())

		next, success, err := p.executeStep(ctx, currentStep, current, effective)
		if err == nil {
			if success {
				current = next
			}
			if effective.compensation != nil {
				compensations = append(compensations, compensationRecord[T]{compensate: effective.compensation, state: current})
			}
			return nil
		}

		if effective.compensatePrevious {
			compErr := compensateAll(ctx, compensations)
			if compErr != nil {
				err = commonerrors.Join(err, compErr)
			}
			compensations = nil
		}

		switch effective.action {
		case ErrorActionContinue, ErrorActionSkip:
			executeErr = commonerrors.Join(executeErr, err)
			return nil
		case ErrorActionFallback:
			if effective.fallback == nil {
				executeErr = commonerrors.Join(executeErr, err)
				return commonerrors.ErrEOF
			}
			fallbackValue, fallbackSuccess, fallbackErr := effective.fallback(ctx, current, err)
			if fallbackErr != nil {
				executeErr = commonerrors.Join(executeErr, err)
				return commonerrors.ErrEOF
			}
			if fallbackSuccess {
				current = fallbackValue
			}
			return nil
		default:
			executeErr = commonerrors.Join(executeErr, err)
			return commonerrors.ErrEOF
		}
	})
	executeErr = commonerrors.Join(executeErr, commonerrors.Ignore(err, commonerrors.ErrEOF))

	return
}

func (p *SimplePipeline[T]) executeStep(ctx context.Context, step IStep[T, T], current T, options *Options[T]) (T, bool, error) {
	return step.ExecuteAsNew(ctx, options, current)
}

func compensateAll[T any](ctx context.Context, compensations []compensationRecord[T]) (err error) {
	for _, record := range collection.Reverse(compensations) {
		if record.compensate != nil {
			err = commonerrors.Join(err, record.compensate(ctx, record.state))
		}
	}
	return
}
