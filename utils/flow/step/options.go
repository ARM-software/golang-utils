package step

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/logs/logrimp"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/retry"
)

const unnamedStep = "unnamed step"

// TransformFunc defines a flow step transformation.
type TransformFunc[I any, O any] = parallelisation.TransformFunc[I, O]

// CompensationFunc compensates a previously successful step.
type CompensationFunc[T any] func(context.Context, T) error

// FallbackFunc produces a replacement result when a step fails.
type FallbackFunc[I any, O any] func(context.Context, I, error) (output O, success bool, err error)

//go:generate go tool enumer -type=ErrorAction -text -json -yaml

// ErrorAction defines the error handling strategy applied to a step.
type ErrorAction int

const (
	// ErrorActionStop stops pipeline execution when a step returns an error.
	ErrorActionStop ErrorAction = iota
	// ErrorActionContinue records the error and continues with later steps.
	ErrorActionContinue
	// ErrorActionSkip ignores the failing step result and continues.
	ErrorActionSkip
	// ErrorActionFallback invokes the configured fallback function.
	ErrorActionFallback
)

// StepOptions configures step behaviour.
type StepOptions[I, O any] struct {
	action                ErrorAction
	actionDefined         bool
	compensatePrevious    bool
	retryPolicy           *retry.RetryPolicyConfiguration
	fallback              FallbackFunc[I, O]
	compensation          CompensationFunc[O]
	transformGroupOptions []parallelisation.StoreOption
}

// StepOption configures step behaviour.
type StepOption[I, O any] func(*StepOptions[I, O]) *StepOptions[I, O]

// Options configures homogeneous step behaviour.
type Options[T any] = StepOptions[T, T]

// Option configures homogeneous step behaviour.
type Option[T any] = StepOption[T, T]

// GenericFallbackFunc produces a replacement result when a generic step fails.
type GenericFallbackFunc[I any, O any] = FallbackFunc[I, O]

// GenericStepOptions configures a generic step.
type GenericStepOptions[I, O any] = StepOptions[I, O]

// GenericStepOption configures a generic step.
type GenericStepOption[I, O any] = StepOption[I, O]

func DefaultStepOptions[I, O any]() *StepOptions[I, O] {
	return (&StepOptions[I, O]{}).Default()
}

func (o *StepOptions[I, O]) Default() *StepOptions[I, O] {
	if o == nil {
		o = &StepOptions[I, O]{}
	}
	if !o.actionDefined {
		o.action = ErrorActionStop
	}
	return o
}

func (o *StepOptions[I, O]) Merge(opts *StepOptions[I, O]) *StepOptions[I, O] {
	if o == nil {
		o = DefaultStepOptions[I, O]()
	}
	if opts == nil {
		return o.Default()
	}
	if opts.actionDefined {
		o.action = opts.action
		o.actionDefined = true
	}
	o.compensatePrevious = o.compensatePrevious || opts.compensatePrevious
	if opts.retryPolicy != nil {
		o.retryPolicy = opts.retryPolicy
	}
	if opts.fallback != nil {
		o.fallback = opts.fallback
	}
	if opts.compensation != nil {
		o.compensation = opts.compensation
	}
	if len(opts.transformGroupOptions) > 0 {
		o.transformGroupOptions = append([]parallelisation.StoreOption{}, opts.transformGroupOptions...)
	}
	return o.Default()
}

func (o *StepOptions[I, O]) Apply(opt StepOption[I, O]) *StepOptions[I, O] {
	if opt == nil {
		return o.Default()
	}
	return opt(o).Default()
}

func (o *StepOptions[I, O]) MergeWithOptions(opts ...StepOption[I, O]) *StepOptions[I, O] {
	return o.Merge(WithStepOptions(opts...))
}

func (o *StepOptions[I, O]) Overwrite(opts *StepOptions[I, O]) *StepOptions[I, O] {
	return o.Default().Merge(opts)
}

func (o *StepOptions[I, O]) WithOptions(opts ...StepOption[I, O]) *StepOptions[I, O] {
	return o.Overwrite(WithStepOptions(opts...))
}

func (o *StepOptions[I, O]) Options() []StepOption[I, O] {
	return []StepOption[I, O]{
		func(opts *StepOptions[I, O]) *StepOptions[I, O] {
			op := o
			if op == nil {
				op = DefaultStepOptions[I, O]()
			}
			return op.Merge(opts)
		},
	}
}

func WithStepOptions[I, O any](options ...StepOption[I, O]) *StepOptions[I, O] {
	resolved := DefaultStepOptions[I, O]()
	for i := range options {
		resolved = resolved.Apply(options[i])
	}
	return resolved.Default()
}

func DefaultOptions[T any]() *Options[T] {
	return DefaultStepOptions[T, T]()
}

func WithOptions[T any](options ...Option[T]) *Options[T] {
	return WithStepOptions[T, T](options...)
}

func DefaultGenericStepOptions[I, O any]() *GenericStepOptions[I, O] {
	return DefaultStepOptions[I, O]()
}

func WithGenericStepOptions[I, O any](options ...GenericStepOption[I, O]) *GenericStepOptions[I, O] {
	return WithStepOptions[I, O](options...)
}

func BreakOnFirstError[T any]() Option[T] { return withBreakOnFirstError[T, T]() }

func withBreakOnFirstError[I, O any]() StepOption[I, O] {
	return func(o *StepOptions[I, O]) *StepOptions[I, O] {
		if o == nil {
			o = DefaultStepOptions[I, O]()
		}
		o.action = ErrorActionStop
		o.actionDefined = true
		return o
	}
}

func ExecuteAll[T any]() Option[T] { return ContinueOnError[T]() }

func ContinueOnError[T any]() Option[T] { return withContinueOnError[T, T]() }

func withContinueOnError[I, O any]() StepOption[I, O] {
	return func(o *StepOptions[I, O]) *StepOptions[I, O] {
		if o == nil {
			o = DefaultStepOptions[I, O]()
		}
		o.action = ErrorActionContinue
		o.actionDefined = true
		return o
	}
}

func SkipStepOnError[T any]() Option[T] { return withSkipStepOnError[T, T]() }

func withSkipStepOnError[I, O any]() StepOption[I, O] {
	return func(o *StepOptions[I, O]) *StepOptions[I, O] {
		if o == nil {
			o = DefaultStepOptions[I, O]()
		}
		o.action = ErrorActionSkip
		o.actionDefined = true
		return o
	}
}

func FallbackOnError[T any](fallback FallbackFunc[T, T]) Option[T] {
	return withFallbackOnError[T, T](fallback)
}

func withFallbackOnError[I, O any](fallback FallbackFunc[I, O]) StepOption[I, O] {
	return func(o *StepOptions[I, O]) *StepOptions[I, O] {
		if o == nil {
			o = DefaultStepOptions[I, O]()
		}
		o.action = ErrorActionFallback
		o.actionDefined = true
		o.fallback = fallback
		return o
	}
}

func RetryOnError[T any](policy *retry.RetryPolicyConfiguration) Option[T] {
	return withRetryOnError[T, T](policy)
}

func withRetryOnError[I, O any](policy *retry.RetryPolicyConfiguration) StepOption[I, O] {
	return func(o *StepOptions[I, O]) *StepOptions[I, O] {
		if o == nil {
			o = DefaultStepOptions[I, O]()
		}
		o.retryPolicy = policy
		return o
	}
}

func WithCompensation[T any](compensation CompensationFunc[T]) Option[T] {
	return withCompensation[T, T](compensation)
}

func withCompensation[I, O any](compensation CompensationFunc[O]) StepOption[I, O] {
	return func(o *StepOptions[I, O]) *StepOptions[I, O] {
		if o == nil {
			o = DefaultStepOptions[I, O]()
		}
		o.compensation = compensation
		return o
	}
}

func CompensatePreviousSteps[T any]() Option[T] {
	return withCompensatePreviousSteps[T, T]()
}

func withCompensatePreviousSteps[I, O any]() StepOption[I, O] {
	return func(o *StepOptions[I, O]) *StepOptions[I, O] {
		if o == nil {
			o = DefaultStepOptions[I, O]()
		}
		o.compensatePrevious = true
		return o
	}
}

func JoinErrorsAndContinue[T any]() Option[T] { return ContinueOnError[T]() }

func WithTransformGroupOptions[T any](options ...parallelisation.StoreOption) Option[T] {
	return withTransformGroupOptions[T, T](options...)
}

func withTransformGroupOptions[I, O any](options ...parallelisation.StoreOption) StepOption[I, O] {
	return func(o *StepOptions[I, O]) *StepOptions[I, O] {
		if o == nil {
			o = DefaultStepOptions[I, O]()
		}
		o.transformGroupOptions = append([]parallelisation.StoreOption{}, options...)
		return o
	}
}

func WithGenericFallbackOnError[I, O any](fallback GenericFallbackFunc[I, O]) GenericStepOption[I, O] {
	return withFallbackOnError[I, O](fallback)
}

func WithGenericRetryOnError[I, O any](policy *retry.RetryPolicyConfiguration) GenericStepOption[I, O] {
	return withRetryOnError[I, O](policy)
}

func WithGenericCompensation[I, O any](compensation CompensationFunc[O]) GenericStepOption[I, O] {
	return withCompensation[I, O](compensation)
}

func WithGenericCompensatePreviousSteps[I, O any]() GenericStepOption[I, O] {
	return withCompensatePreviousSteps[I, O]()
}

func WithGenericTransformGroupOptions[I, O any](options ...parallelisation.StoreOption) GenericStepOption[I, O] {
	return withTransformGroupOptions[I, O](options...)
}

func executeWithRetry[I any, O any](ctx context.Context, name string, current I, transform parallelisation.TransformFunc[I, O], policy *retry.RetryPolicyConfiguration) (O, bool, error) {
	if policy == nil || !policy.Enabled {
		return transform(ctx, current)
	}

	var output O
	var success bool
	err := retry.RetryIf(ctx, logrimp.NewNoopLogger(), policy, func() error {
		var runErr error
		output, success, runErr = transform(ctx, current)
		return runErr
	}, "retrying step "+name, func(err error) bool { return err != nil })
	return output, success, err
}
