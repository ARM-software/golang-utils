package pipeline

import (
	flowpipeline "github.com/ARM-software/golang-utils/utils/flow/pipeline"
	flowstep "github.com/ARM-software/golang-utils/utils/flow/step"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/retry"
)

type TransformFunc[I any, O any] = flowstep.TransformFunc[I, O]
type CompensationFunc[T any] = flowstep.CompensationFunc[T]
type FallbackFunc[I any, O any] = flowstep.FallbackFunc[I, O]
type ErrorAction = flowstep.ErrorAction
type StepOptions[I, O any] = flowstep.StepOptions[I, O]
type StepOption[I, O any] = flowstep.StepOption[I, O]
type Options[T any] = flowstep.Options[T]
type Option[T any] = flowstep.Option[T]
type GenericFallbackFunc[I any, O any] = flowstep.GenericFallbackFunc[I, O]
type GenericStepOptions[I, O any] = flowstep.GenericStepOptions[I, O]
type GenericStepOption[I, O any] = flowstep.GenericStepOption[I, O]
type IStep[I, O any] = flowstep.IStep[I, O]
type Step[T any] = flowstep.Step[T]
type GenericStep[I, O any] = flowstep.GenericStep[I, O]
type ISimplePipeline[T any] = flowpipeline.ISimplePipeline[T]
type IPipeline[I, O any] = flowpipeline.IPipeline[I, O]
type SimplePipeline[T any] = flowpipeline.SimplePipeline[T]
type Pipeline[I, O any] = flowpipeline.Pipeline[I, O]

const (
	ErrorActionStop     = flowstep.ErrorActionStop
	ErrorActionContinue = flowstep.ErrorActionContinue
	ErrorActionSkip     = flowstep.ErrorActionSkip
	ErrorActionFallback = flowstep.ErrorActionFallback
)

func DefaultStepOptions[I, O any]() *StepOptions[I, O] { return flowstep.DefaultStepOptions[I, O]() }
func DefaultOptions[T any]() *Options[T]               { return flowstep.DefaultOptions[T]() }
func WithOptions[T any](options ...Option[T]) *Options[T] {
	return flowstep.WithOptions[T](options...)
}
func DefaultGenericStepOptions[I, O any]() *GenericStepOptions[I, O] {
	return flowstep.DefaultGenericStepOptions[I, O]()
}
func WithGenericStepOptions[I, O any](options ...GenericStepOption[I, O]) *GenericStepOptions[I, O] {
	return flowstep.WithGenericStepOptions[I, O](options...)
}
func BreakOnFirstError[T any]() Option[T] { return flowstep.BreakOnFirstError[T]() }
func ExecuteAll[T any]() Option[T]        { return flowstep.ExecuteAll[T]() }
func ContinueOnError[T any]() Option[T]   { return flowstep.ContinueOnError[T]() }
func SkipStepOnError[T any]() Option[T]   { return flowstep.SkipStepOnError[T]() }
func FallbackOnError[T any](fallback FallbackFunc[T, T]) Option[T] {
	return flowstep.FallbackOnError[T](fallback)
}
func RetryOnError[T any](policy *retry.RetryPolicyConfiguration) Option[T] {
	return flowstep.RetryOnError[T](policy)
}
func WithCompensation[T any](compensation CompensationFunc[T]) Option[T] {
	return flowstep.WithCompensation[T](compensation)
}
func CompensatePreviousSteps[T any]() Option[T] { return flowstep.CompensatePreviousSteps[T]() }
func JoinErrorsAndContinue[T any]() Option[T]   { return flowstep.JoinErrorsAndContinue[T]() }
func WithTransformGroupOptions[T any](options ...parallelisation.StoreOption) Option[T] {
	return flowstep.WithTransformGroupOptions[T](options...)
}
func WithGenericFallbackOnError[I, O any](fallback GenericFallbackFunc[I, O]) GenericStepOption[I, O] {
	return flowstep.WithGenericFallbackOnError[I, O](fallback)
}
func WithGenericRetryOnError[I, O any](policy *retry.RetryPolicyConfiguration) GenericStepOption[I, O] {
	return flowstep.WithGenericRetryOnError[I, O](policy)
}
func WithGenericCompensation[I, O any](compensation CompensationFunc[O]) GenericStepOption[I, O] {
	return flowstep.WithGenericCompensation[I, O](compensation)
}
func WithGenericCompensatePreviousSteps[I, O any]() GenericStepOption[I, O] {
	return flowstep.WithGenericCompensatePreviousSteps[I, O]()
}
func WithGenericTransformGroupOptions[I, O any](options ...parallelisation.StoreOption) GenericStepOption[I, O] {
	return flowstep.WithGenericTransformGroupOptions[I, O](options...)
}
func NewStep[T any](transform parallelisation.TransformFunc[T, T], options ...Option[T]) *Step[T] {
	return flowstep.NewStep[T](transform, options...)
}
func NewNamedStep[T any](name string, transform parallelisation.TransformFunc[T, T], options ...Option[T]) *Step[T] {
	return flowstep.NewNamedStep[T](name, transform, options...)
}
func NewTransformGroupStep[T any](transform parallelisation.TransformFunc[T, T], transformGroupOptions []parallelisation.StoreOption, options ...Option[T]) *Step[T] {
	return flowstep.NewTransformGroupStep[T](transform, transformGroupOptions, options...)
}
func NewNamedTransformGroupStep[T any](name string, transform parallelisation.TransformFunc[T, T], transformGroupOptions []parallelisation.StoreOption, options ...Option[T]) *Step[T] {
	return flowstep.NewNamedTransformGroupStep[T](name, transform, transformGroupOptions, options...)
}
func NewGenericStep[I, O any](transform parallelisation.TransformFunc[I, O], options ...GenericStepOption[I, O]) *GenericStep[I, O] {
	return flowstep.NewGenericStep[I, O](transform, options...)
}
func NewNamedGenericStep[I, O any](name string, transform parallelisation.TransformFunc[I, O], options ...GenericStepOption[I, O]) *GenericStep[I, O] {
	return flowstep.NewNamedGenericStep[I, O](name, transform, options...)
}
func NewGenericTransformGroupStep[I, O any](transform parallelisation.TransformFunc[I, O], transformGroupOptions []parallelisation.StoreOption, options ...GenericStepOption[I, O]) *GenericStep[I, O] {
	return flowstep.NewGenericTransformGroupStep[I, O](transform, transformGroupOptions, options...)
}
func NewNamedGenericTransformGroupStep[I, O any](name string, transform parallelisation.TransformFunc[I, O], transformGroupOptions []parallelisation.StoreOption, options ...GenericStepOption[I, O]) *GenericStep[I, O] {
	return flowstep.NewNamedGenericTransformGroupStep[I, O](name, transform, transformGroupOptions, options...)
}
func NewSimplePipeline[T any](options ...Option[T]) *SimplePipeline[T] {
	return flowpipeline.NewSimplePipeline[T](options...)
}
func NewPipeline[I, O any](step IStep[I, O]) *Pipeline[I, O] {
	return flowpipeline.NewPipeline[I, O](step)
}
func Chain[I, M, O any](pipeline *Pipeline[I, M], step IStep[M, O]) *Pipeline[I, O] {
	return flowpipeline.Chain[I, M, O](pipeline, step)
}
