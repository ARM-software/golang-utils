package parallelisation

import (
	"context"
	"iter"
	"maps"
	"slices"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/sasha-s/go-deadlock"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

var _ IExecutionGroup[IExecutor] = &PriorityExecutionGroup[IExecutor]{}

const defaultPriority uint = 0

type PriorityExecutionGroup[T any] struct {
	mu           deadlock.RWMutex
	groups       map[uint]IExecutionGroup[T]
	options      []StoreOption
	newGroupFunc func(...StoreOption) IExecutionGroup[T]
}

func newPriorityExecutionGroup[T any](newGroup func(...StoreOption) IExecutionGroup[T], options ...StoreOption) *PriorityExecutionGroup[T] {
	return &PriorityExecutionGroup[T]{
		mu:           deadlock.RWMutex{},
		groups:       make(map[uint]IExecutionGroup[T]),
		options:      options,
		newGroupFunc: newGroup,
	}
}

// NewPriorityExecutionGroup returns an execution group that can execute functions in order according to priority rules.
// Parallel commands with differing priorities will be executed in groups according to their priority.
// Sequential commands will be executed in order of their priority, no guarantees are made about the order of when
// the priority is the same as another command.
func NewPriorityExecutionGroup(options ...StoreOption) *PriorityExecutionGroup[IExecutor] {
	return newPriorityExecutionGroup[IExecutor](
		func(options ...StoreOption) IExecutionGroup[IExecutor] {
			return NewExecutionGroup[IExecutor](func(ctx context.Context, e IExecutor) error {
				return e.Execute(ctx)
			}, options...)
		},
		options...,
	)
}

func (g *PriorityExecutionGroup[T]) check() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.groups == nil {
		g.groups = make(map[uint]IExecutionGroup[T])
	}
	if g.options == nil {
		g.options = DefaultOptions().Options()
	}
	if g.newGroupFunc == nil {
		g.newGroupFunc = func(options ...StoreOption) IExecutionGroup[T] {
			// since none of the methods return errors directly we inject executors that will force the error and reveal it to the consumer
			return NewExecutionGroup[T](func(context.Context, T) error {
				return commonerrors.UndefinedVariableWithMessage("g.newGroupFunc", "priority execution group has not been initialised correctly")
			})
		}
	}
}

// RegisterExecutor registers executors with a specific priority (lower values indidcate higher priority)
func (g *PriorityExecutionGroup[T]) RegisterFunctionWithPriority(priority uint, function ...T) {
	g.RegisterFunctionsWithPriority(priority, slices.Values(function))
}

func (g *PriorityExecutionGroup[T]) RegisterFunctionsWithPriority(priority uint, functionSeq iter.Seq[T]) {
	g.GetPriorityGroup(priority).RegisterFunctions(functionSeq)
}

func (g *PriorityExecutionGroup[T]) GetPriorityGroup(priority uint) IExecutionGroup[T] {
	g.check()

	g.mu.Lock()
	defer g.mu.Unlock()
	if g.groups[priority] == nil {
		g.groups[priority] = g.newGroupFunc(g.options...)
	}
	return g.groups[priority]
}

func (g *PriorityExecutionGroup[T]) CopyAllPriorityFunctions(d *PriorityExecutionGroup[T]) {
	if d == nil {
		return
	}
	_ = collection.Each(maps.Keys(g.groups), func(i uint) error {
		g.CopyPriorityFunctions(i, d.GetPriorityGroup(i))
		return nil
	})

}

func (g *PriorityExecutionGroup[T]) CopyPriorityFunctions(priority uint, d IExecutionGroup[T]) {
	g.GetPriorityGroup(priority).CopyFunctions(d)
}

func (g *PriorityExecutionGroup[T]) CopyFunctions(d IExecutionGroup[T]) {
	if d == nil {
		return
	}
	if p, ok := d.(*PriorityExecutionGroup[T]); ok {
		g.CopyAllPriorityFunctions(p)
	} else {
		g.CopyPriorityFunctions(defaultPriority, d)
	}
}

func (g *PriorityExecutionGroup[T]) RegisterFunctions(functionSeq iter.Seq[T]) {
	g.RegisterFunctionsWithPriority(defaultPriority, functionSeq)
}

// RegisterExecutor registers executors with a priority of zero (highest priority)
func (g *PriorityExecutionGroup[T]) RegisterFunction(function ...T) {
	g.RegisterFunctionWithPriority(defaultPriority, function...)
}

func (g *PriorityExecutionGroup[T]) Clone() IExecutionGroup[T] {
	c := newPriorityExecutionGroup[T](g.newGroupFunc, g.options...)
	g.CopyAllPriorityFunctions(c)
	return c
}

func (g *PriorityExecutionGroup[T]) Len() (n int) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for _, group := range g.groups {
		n += group.Len()
	}
	return
}

func (g *PriorityExecutionGroup[T]) executors() (executor *CompoundExecutionGroup) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	opts := DefaultOptions().MergeWithOptions(g.options...)
	opts.MergeWithOptions(Sequential)
	executor = NewCompoundExecutionGroup(opts.Options()...)
	for _, key := range slices.Sorted(maps.Keys(g.groups)) {
		executor.RegisterExecutor(g.groups[key])
	}
	return
}

// Execute will execute all the groups according to the priorities of the functions
func (g *PriorityExecutionGroup[T]) Execute(ctx context.Context) error {
	return g.executors().Execute(ctx)
}
