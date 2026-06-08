package pipeline

import (
	"context"
	"slices"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	flowstep "github.com/ARM-software/golang-utils/utils/flow/step"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

type compensationRecord[T any] struct {
	compensate flowstep.CompensationFunc[T]
	state      T
}

var _ ISimplePipeline[any] = (*SimplePipeline[any])(nil)

// SimplePipeline executes ordered steps over a single state value.
type SimplePipeline[T any] struct {
	steps   []flowstep.IStep[T, T]
	options *flowstep.Options[T]
}

func NewSimplePipeline[T any](options ...flowstep.Option[T]) *SimplePipeline[T] {
	return &SimplePipeline[T]{steps: make([]flowstep.IStep[T, T], 0), options: flowstep.WithOptions(options...)}
}

func (p *SimplePipeline[T]) RegisterStep(steps ...flowstep.IStep[T, T]) {
	p.steps = append(p.steps, steps...)
}

func (p *SimplePipeline[T]) RegisterTransform(name string, transform flowstep.TransformFunc[T, T], options ...flowstep.Option[T]) {
	p.RegisterStep(flowstep.NewNamedStep(name, transform, options...))
}

func (p *SimplePipeline[T]) RegisterTransformGroup(name *string, group *parallelisation.TransformGroup[T, T]) {
	p.RegisterStep(flowstep.NewStepFromTransformGroup(name, group))
}

func (p *SimplePipeline[T]) Execute(ctx context.Context, input T) (current T, executeErr error) {
	current = input
	compensations := make([]compensationRecord[T], 0)

	err := collection.EachRef(slices.Values(p.steps), func(step *flowstep.IStep[T, T]) error {
		if err := parallelisation.DetermineContextError(ctx); err != nil {
			return err
		}
		if step == nil || *step == nil {
			return commonerrors.UndefinedVariable("pipeline step")
		}
		currentStep := *step

		effective := flowstep.WithOptions[T]()
		effective.Merge(p.options)
		effective.Merge(currentStep.GetOptions())

		next, success, err := currentStep.ExecuteAsNew(ctx, effective, current)
		if err == nil {
			if success {
				current = next
			}
			if compensation := effective.GetCompensation(); compensation != nil {
				compensations = append(compensations, compensationRecord[T]{compensate: compensation, state: current})
			}
			return nil
		}

		if effective.GetCompensatePrevious() {
			err = commonerrors.Join(err, compensateAll(ctx, compensations))
			compensations = nil
		}

		switch effective.GetAction() {
		case flowstep.ErrorActionContinue, flowstep.ErrorActionSkip:
			executeErr = commonerrors.Join(executeErr, err)
			return nil
		case flowstep.ErrorActionFallback:
			fallback := effective.GetFallback()
			if fallback == nil {
				executeErr = commonerrors.Join(executeErr, err)
				return commonerrors.ErrEOF
			}
			fallbackValue, fallbackSuccess, fallbackErr := fallback(ctx, current, err)
			if fallbackErr != nil {
				executeErr = commonerrors.Join(executeErr, err, fallbackErr)
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

func compensateAll[T any](ctx context.Context, compensations []compensationRecord[T]) (err error) {
	for _, record := range collection.Reverse(compensations) {
		if record.compensate == nil {
			continue
		}
		err = commonerrors.Join(err, record.compensate(ctx, record.state))
	}
	return
}
