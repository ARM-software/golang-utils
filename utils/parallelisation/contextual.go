package parallelisation

import (
	"context"
	"io"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

// DetermineContextError determines what the context error is if any.
func DetermineContextError(ctx context.Context) error {
	err := commonerrors.ErrFromContext(ctx)
	if commonerrors.Any(err, nil) {
		return err
	}
	return commonerrors.WrapError(err, context.Cause(ctx), "")
}

type ContextualFunc func(ctx context.Context) error

type ContextualFunctionGroup struct {
	ExecutionGroup[ContextualFunc]
}

func (s *ContextualFunctionGroup) Clone() IExecutionGroup[ContextualFunc] {
	g := NewContextualGroup(s.options.Options()...)
	s.CopyFunctions(g)
	return g
}

// NewContextualGroup returns a group executing contextual functions.
func NewContextualGroup(options ...StoreOption) *ContextualFunctionGroup {
	return &ContextualFunctionGroup{
		ExecutionGroup: *NewExecutionGroup[ContextualFunc](func(ctx context.Context, contextualF ContextualFunc) error {
			return contextualF(ctx)
		}, options...),
	}
}

// NewContextualGroupWithPriority returns a group executing contextual functions that will be run in priority order.
func NewPriorityContextualGroup(options ...StoreOption) *PriorityExecutionGroup[ContextualFunc] {
	return newPriorityExecutionGroup[ContextualFunc](
		func(options ...StoreOption) IExecutionGroup[ContextualFunc] {
			return NewExecutionGroup[ContextualFunc](func(ctx context.Context, f ContextualFunc) error {
				return f(ctx)
			}, options...)
		},
		options...,
	)
}

// ForEach executes all the contextual functions according to the store options and returns an error if one occurred.
func ForEach(ctx context.Context, executionOptions *StoreOptions, contextualFunc ...ContextualFunc) error {
	group := NewContextualGroup(ExecuteAll(executionOptions).Options()...)
	group.RegisterFunction(contextualFunc...)
	return group.Execute(ctx)
}

// BreakOnError executes each functions in the group until an error is found or the context gets cancelled.
func BreakOnError(ctx context.Context, executionOptions *StoreOptions, contextualFunc ...ContextualFunc) error {
	group := NewContextualGroup(StopOnFirstError(executionOptions).Options()...)
	group.RegisterFunction(contextualFunc...)
	return group.Execute(ctx)
}

// BreakOnErrorOrEOF is similar to BreakOnError but also stops on EOF. However, in this case, no error is returned
func BreakOnErrorOrEOF(ctx context.Context, executionOptions *StoreOptions, contextualFunc ...ContextualFunc) error {
	return commonerrors.Ignore(BreakOnError(ctx, executionOptions, contextualFunc...), commonerrors.ErrEOF, io.EOF)
}
