package parallelisation

import (
	"context"
	"math"

	"github.com/sasha-s/go-deadlock"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/safecast"
)

type StoreOptions struct {
	clearOnExecution bool
	stopOnFirstError bool
	sequential       bool
	reverse          bool
	joinErrors       bool
	onlyOnce         bool
	workers          int
}

func (o *StoreOptions) Default() *StoreOptions {
	o.clearOnExecution = false
	o.stopOnFirstError = false
	o.sequential = false
	o.reverse = false
	o.joinErrors = false
	o.onlyOnce = false
	o.workers = 0
	return o
}

func (o *StoreOptions) Merge(opts *StoreOptions) *StoreOptions {
	if opts == nil {
		return o
	}
	o.clearOnExecution = opts.clearOnExecution || o.clearOnExecution
	o.stopOnFirstError = opts.stopOnFirstError || o.stopOnFirstError
	o.sequential = opts.sequential || o.sequential
	o.reverse = opts.reverse || o.reverse
	o.joinErrors = opts.joinErrors || o.joinErrors
	o.onlyOnce = opts.onlyOnce || o.onlyOnce
	o.workers = safecast.ToInt(math.Max(float64(opts.workers), float64(o.workers)))
	return o
}

func (o *StoreOptions) MergeWithOptions(opt ...StoreOption) *StoreOptions {
	return o.Merge(WithOptions(opt...))
}

func (o *StoreOptions) Overwrite(opts *StoreOptions) *StoreOptions {
	return o.Default().Merge(opts)
}

func (o *StoreOptions) WithOptions(opts ...StoreOption) *StoreOptions {
	return o.Overwrite(WithOptions(opts...))
}

func (o *StoreOptions) Options() []StoreOption {
	return []StoreOption{
		func(opts *StoreOptions) *StoreOptions {
			op := o
			if op == nil {
				op = DefaultOptions()
			}
			return op.Merge(opts)
		},
	}
}

type StoreOption func(*StoreOptions) *StoreOptions

// StopOnFirstError stops ExecutionGroup execution on first error.
var StopOnFirstError StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		o = DefaultOptions()
	}
	o.stopOnFirstError = true
	o.joinErrors = false
	return o
}

// JoinErrors will collate any errors which happened when executing functions in ExecutionGroup.
// This option should not be used in combination to StopOnFirstError.
var JoinErrors StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		o = DefaultOptions()
	}
	o.stopOnFirstError = false
	o.joinErrors = true
	return o
}

// OnlyOnce will ensure the function are executed only once if they do.
var OnlyOnce StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		o = DefaultOptions()
	}
	o.onlyOnce = true
	return o
}

// AnyTimes will allow the functions to be executed as often that they might be.
var AnyTimes StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		o = DefaultOptions()
	}
	o.onlyOnce = false
	return o
}

// ExecuteAll executes all functions in the ExecutionGroup even if an error is raised. the first error raised is then returned.
var ExecuteAll StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		o = DefaultOptions()
	}
	o.stopOnFirstError = false
	return o
}

// ClearAfterExecution clears the ExecutionGroup after execution.
var ClearAfterExecution StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		o = DefaultOptions()
	}
	o.clearOnExecution = true
	return o
}

// RetainAfterExecution keep the ExecutionGroup intact after execution (no reset).
var RetainAfterExecution StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		o = DefaultOptions()
	}
	o.clearOnExecution = false
	return o
}

// Parallel ensures every function registered in the ExecutionGroup is executed concurrently in the order they were registered.
var Parallel StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		o = DefaultOptions()
	}
	o.sequential = false
	return o
}

// Workers defines a limit number of workers for executing the function registered in the ExecutionGroup.
func Workers(workers int) StoreOption {
	return func(o *StoreOptions) *StoreOptions {
		if o == nil {
			o = DefaultOptions()
		}
		o.workers = workers
		o.sequential = false
		return o
	}
}

// Sequential ensures every function registered in the ExecutionGroup is executed sequentially in the order they were registered.
var Sequential StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		o = DefaultOptions()
	}
	o.sequential = true
	return o
}

// SequentialInReverse ensures every function registered in the ExecutionGroup is executed sequentially but in the reverse order they were registered.
var SequentialInReverse StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		o = DefaultOptions()
	}
	o.sequential = true
	o.reverse = true
	return o
}

// WithOptions defines a store configuration.
func WithOptions(option ...StoreOption) (opts *StoreOptions) {
	for i := range option {
		opts = option[i](opts)
	}
	if opts == nil {
		opts = DefaultOptions()
	}
	return
}

// DefaultOptions returns the default store configuration
func DefaultOptions() *StoreOptions {
	opts := &StoreOptions{}
	return opts.Default()
}

type IExecutor interface {
	// Execute executes all the functions in the group.
	Execute(ctx context.Context) error
}

type IExecutionGroup[T any] interface {
	IExecutor
	RegisterFunction(function ...T)
	Len() int
}

type ICompoundExecutionGroup[T any] interface {
	IExecutionGroup[T]
	// RegisterExecutor registers executors of any kind to the group: they could be functions or sub-groups.
	RegisterExecutor(executor ...IExecutor)
}

// NewExecutionGroup returns an execution group which executes functions according to store options.
func NewExecutionGroup[T any](executeFunc ExecuteFunc[T], options ...StoreOption) *ExecutionGroup[T] {

	opts := WithOptions(options...)
	return &ExecutionGroup[T]{
		mu:          deadlock.RWMutex{},
		functions:   make([]wrappedElement[T], 0),
		executeFunc: executeFunc,
		options:     *opts,
	}
}

type ExecuteFunc[T any] func(ctx context.Context, element T) error

type ExecutionGroup[T any] struct {
	mu          deadlock.RWMutex
	functions   []wrappedElement[T]
	executeFunc ExecuteFunc[T]
	options     StoreOptions
}

// RegisterFunction registers functions to the group.
func (s *ExecutionGroup[T]) RegisterFunction(function ...T) {
	defer s.mu.Unlock()
	s.mu.Lock()
	wrapped := make([]wrappedElement[T], len(function))
	for i := range function {
		wrapped[i] = newWrapped(function[i], s.options.onlyOnce)
	}
	s.functions = append(s.functions, wrapped...)
}

func (s *ExecutionGroup[T]) Len() int {
	defer s.mu.RUnlock()
	s.mu.RLock()
	return len(s.functions)
}

// Execute executes all the function in the group according to store options.
func (s *ExecutionGroup[T]) Execute(ctx context.Context) (err error) {
	defer s.mu.Unlock()
	s.mu.Lock()
	if reflection.IsEmpty(s.executeFunc) {
		return commonerrors.New(commonerrors.ErrUndefined, "the group was not initialised correctly")
	}

	if s.options.sequential {
		err = s.executeSequentially(ctx, s.options.stopOnFirstError, s.options.reverse, s.options.joinErrors)
	} else {
		err = s.executeConcurrently(ctx, s.options.stopOnFirstError, s.options.joinErrors)
	}

	if err == nil && s.options.clearOnExecution {
		s.functions = make([]wrappedElement[T], 0, len(s.functions))
	}
	return
}

func (s *ExecutionGroup[T]) executeConcurrently(ctx context.Context, stopOnFirstError bool, collateErrors bool) error {
	g, gCtx := errgroup.WithContext(ctx)
	if !stopOnFirstError {
		gCtx = ctx
	}
	funcNum := len(s.functions)
	workers := s.options.workers
	if workers <= 0 {
		workers = funcNum
	}
	errCh := make(chan error, funcNum)

	g.SetLimit(workers)
	for i := range s.functions {
		g.Go(func() error {
			_, subErr := s.executeFunction(gCtx, s.functions[i])
			errCh <- subErr
			return subErr
		})
	}
	err := g.Wait()
	close(errCh)
	if collateErrors {
		collateErr := make([]error, funcNum)
		i := 0
		for subErr := range errCh {
			collateErr[i] = subErr
			i++
		}
		err = commonerrors.Join(collateErr...)
	}

	return err
}

func (s *ExecutionGroup[T]) executeSequentially(ctx context.Context, stopOnFirstError, reverse, collateErrors bool) (err error) {
	err = DetermineContextError(ctx)
	if err != nil {
		return
	}
	funcNum := len(s.functions)
	collateErr := make([]error, funcNum)
	if reverse {
		for i := funcNum - 1; i >= 0; i-- {
			shouldBreak, subErr := s.executeFunction(ctx, s.functions[i])
			collateErr[funcNum-i-1] = subErr
			if shouldBreak {
				err = subErr
				return
			}
			if subErr != nil && err == nil {
				err = subErr
				if stopOnFirstError {
					return
				}
			}
		}
	} else {
		for i := range s.functions {
			shouldBreak, subErr := s.executeFunction(ctx, s.functions[i])
			collateErr[i] = subErr
			if shouldBreak {
				err = subErr
				return
			}
			if subErr != nil && err == nil {
				err = subErr
				if stopOnFirstError {
					return
				}
			}
		}
	}

	if collateErrors {
		err = commonerrors.Join(collateErr...)
	}
	return
}

func (s *ExecutionGroup[T]) executeFunction(ctx context.Context, w wrappedElement[T]) (mustBreak bool, err error) {
	err = DetermineContextError(ctx)
	if err != nil {
		mustBreak = true
		return
	}
	if w == nil {
		err = commonerrors.UndefinedVariable("function element")
		mustBreak = true
		return
	}
	err = w.Execute(ctx, s.executeFunc)

	return
}

type wrappedElement[T any] interface {
	Execute(ctx context.Context, f ExecuteFunc[T]) error
}
type basicWrap[T any] struct {
	value T
}

func (w *basicWrap[T]) Execute(ctx context.Context, f ExecuteFunc[T]) error {
	return f(ctx, w.value)
}

func newBasicWrap[T any](e T) wrappedElement[T] {
	return &basicWrap[T]{
		value: e,
	}
}

func newOnce[T any](e T) wrappedElement[T] {
	return &once[T]{
		wrappedElement: newBasicWrap[T](e),
		once:           atomic.NewBool(false),
	}
}

type once[T any] struct {
	wrappedElement[T]
	once *atomic.Bool
}

func (w *once[T]) Execute(ctx context.Context, f ExecuteFunc[T]) error {
	if !w.once.Swap(true) {
		return w.wrappedElement.Execute(ctx, f)
	}
	return nil
}

func newWrapped[T any](e T, once bool) wrappedElement[T] {
	if once {
		return newOnce[T](e)
	} else {
		return newBasicWrap[T](e)
	}
}

var _ ICompoundExecutionGroup[ContextualFunc] = &CompoundExecutionGroup{}

// NewCompoundExecutionGroup returns an execution group made of executors
func NewCompoundExecutionGroup(options ...StoreOption) *CompoundExecutionGroup {
	return &CompoundExecutionGroup{
		ContextualFunctionGroup: *NewContextualGroup(options...),
	}
}

type CompoundExecutionGroup struct {
	ContextualFunctionGroup
}

// RegisterExecutor registers executors
func (g *CompoundExecutionGroup) RegisterExecutor(group ...IExecutor) {
	for i := range group {
		g.RegisterFunction(func(ctx context.Context) error {
			return group[i].Execute(ctx)
		})
	}
}
