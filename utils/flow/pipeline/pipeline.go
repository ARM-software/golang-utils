package pipeline

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	flowstep "github.com/ARM-software/golang-utils/utils/flow/step"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

type genericCompensationRecord struct {
	compensate func(context.Context) error
}

var _ IPipeline[any, any] = (*Pipeline[any, any])(nil)

// Pipeline executes typed steps where each step may transform the input into a different output type.
type Pipeline[I, O any] struct {
	run func(context.Context, I) (O, []genericCompensationRecord, error)
}

func NewPipeline[I, O any](step flowstep.IStep[I, O]) *Pipeline[I, O] {
	return &Pipeline[I, O]{run: func(ctx context.Context, input I) (O, []genericCompensationRecord, error) {
		return executeGenericPipelineStep(ctx, step, input, nil)
	}}
}

func Chain[I, M, O any](pipeline *Pipeline[I, M], step flowstep.IStep[M, O]) *Pipeline[I, O] {
	return &Pipeline[I, O]{run: func(ctx context.Context, input I) (output O, compensations []genericCompensationRecord, err error) {
		var intermediate M
		intermediate, compensations, err = pipeline.run(ctx, input)
		if err != nil {
			return output, compensations, err
		}
		return executeGenericPipelineStep(ctx, step, intermediate, compensations)
	}}
}

func (p *Pipeline[I, O]) Execute(ctx context.Context, input I) (output O, err error) {
	if p == nil || p.run == nil {
		err = commonerrors.UndefinedVariable("pipeline")
		return
	}
	output, _, err = p.run(ctx, input)
	return
}

func executeGenericPipelineStep[I, O any](ctx context.Context, step flowstep.IStep[I, O], input I, previous []genericCompensationRecord) (output O, compensations []genericCompensationRecord, err error) {
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
	output, success, err = step.ExecuteAsNew(ctx, options, input)
	if err != nil {
		if options.GetCompensatePrevious() {
			err = commonerrors.Join(err, runGenericCompensations(ctx, previous))
		}
		if fallback := options.GetFallback(); fallback != nil {
			fallbackValue, fallbackSuccess, fallbackErr := fallback(ctx, input, err)
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
	if compensation := options.GetCompensation(); compensation != nil {
		compensations = append(compensations, genericCompensationRecord{compensate: func(compCtx context.Context) error {
			return compensation(compCtx, output)
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
