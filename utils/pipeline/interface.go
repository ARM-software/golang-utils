package pipeline

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

//go:generate go tool mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE IStep,ISimplePipeline,IPipeline

// IStep defines one executable stage in a pipeline.
//
// `I` is the step input type and `O` is the step output type.
//
// Use this interface when you want to model a transformation as a reusable unit
// that can be executed directly, inserted into a [Pipeline], or inserted into a
// [SimplePipeline] when `I` and `O` are the same type.
//
// Step names are optional. Implementations may return nil from
// [IStep.GetName] when no human-readable label is needed. Names are mainly
// useful for diagnostics, logs, and test readability.
//
// Example:
//
//	step := NewNamedGenericStep("parse", func(ctx context.Context, input string) (int, bool, error) {
//		value, err := strconv.Atoi(input)
//		if err != nil {
//			return 0, false, err
//		}
//		return value, true, nil
//	})
type IStep[I, O any] interface {
	// GetName returns the human-readable name of the step.
	//
	// The name may be nil for unnamed steps.
	GetName() *string
	// GetOptions returns the step-specific execution options.
	//
	// These options are merged with any pipeline-level options before execution.
	GetOptions() *StepOptions[I, O]
	// Execute runs the step against the current pipeline state.
	Execute(context.Context, I) (output O, success bool, err error)
	// ExecuteAsNew runs the step against the provided input using the supplied
	// execution options instead of the step's stored options.
	ExecuteAsNew(context.Context, *StepOptions[I, O], I) (output O, success bool, err error)
}

// ISimplePipeline defines a context-controlled pipeline over a single state
// value.
//
// `T` is both the input and output type of every registered step. This variant
// is useful when the pipeline represents progressive enrichment or mutation of a
// single state object.
//
// Typical uses include request processing, configuration normalisation, or any
// workflow where each step receives the current state and returns the next state
// of the same type.
//
// Example:
//
//	p := NewSimplePipeline[int]()
//	p.RegisterStep(NewStep(func(ctx context.Context, value int) (int, bool, error) {
//		return value + 1, true, nil
//	}))
//	result, err := p.Execute(context.Background(), 1)
type ISimplePipeline[T any] interface {
	// RegisterStep appends steps to the pipeline.
	RegisterStep(...IStep[T, T])
	// RegisterTransform appends a new named step built from transform.
	RegisterTransform(name string, transform TransformFunc[T, T], options ...Option[T])
	// RegisterTransformGroup appends a step backed by a transform group.
	RegisterTransformGroup(name *string, group *parallelisation.TransformGroup[T, T])
	// Execute runs the pipeline over input until completion, cancellation, or a
	// configured stop condition.
	Execute(context.Context, T) (T, error)
}

// IPipeline defines a typed pipeline whose step output types may differ from
// earlier step input types while still ensuring that each step output matches
// the next step input at compile time.
//
// `I` is the pipeline input type and `O` is the final pipeline output type.
//
// This variant is useful when a workflow naturally moves through multiple data
// shapes, for example: raw input -> parsed model -> validated model -> rendered
// output.
//
// Example:
//
//	parse := NewNamedGenericStep("parse", func(ctx context.Context, input string) (int, bool, error) {
//		value, err := strconv.Atoi(input)
//		if err != nil {
//			return 0, false, err
//		}
//		return value, true, nil
//	})
//	double := NewGenericStep(func(ctx context.Context, value int) (int, bool, error) {
//		return value * 2, true, nil
//	})
//	p := Chain(NewPipeline(parse), double)
//	result, err := p.Execute(context.Background(), "21")
type IPipeline[I, O any] interface {
	Execute(context.Context, I) (O, error)
}
