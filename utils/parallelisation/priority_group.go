package parallelisation

import (
	"context"
	"maps"
	"slices"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/sasha-s/go-deadlock"
)

var _ IExecutionGroup[IExecutor] = &PriorityExecutionGroup[IExecutor]{}

type PriorityExecutionGroup[T any] struct {
	mu       deadlock.RWMutex
	groups   map[uint]IExecutionGroup[T]
	options  []StoreOption
	newGroup func(...StoreOption) IExecutionGroup[T]
}

func newPriorityExecutionGroup[T any](newGroup func(...StoreOption) IExecutionGroup[T], options ...StoreOption) *PriorityExecutionGroup[T] {
	return &PriorityExecutionGroup[T]{
		mu:       deadlock.RWMutex{},
		groups:   make(map[uint]IExecutionGroup[T]),
		options:  options,
		newGroup: newGroup,
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
	if g.newGroup == nil {
		g.newGroup = func(options ...StoreOption) IExecutionGroup[T] {
			// since none of the methods return errors directly we inject executors that will force the error and reveal it to the consumer
			return NewExecutionGroup[T](func(context.Context, T) error {
				return commonerrors.UndefinedVariableWithMessage("g.newGroup", "priority execution group has not been initialised correctly")
			})
		}
	}
}

// RegisterExecutor registers executors with a specific priority (lower values indidcate higher priority)
func (g *PriorityExecutionGroup[T]) RegisterFunctionWithPriority(priority uint, function ...T) {
	g.check()

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.groups[priority] == nil {
		g.groups[priority] = g.newGroup(g.options...)
	}
	g.groups[priority].RegisterFunction(function...)
}

// RegisterExecutor registers executors with a priority of zero (highest priority)
func (g *PriorityExecutionGroup[T]) RegisterFunction(function ...T) {
	g.RegisterFunctionWithPriority(0, function...)
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

	executor = NewCompoundExecutionGroup(DefaultOptions().MergeWithOptions(Sequential).Options()...)
	for _, key := range slices.Sorted(maps.Keys(g.groups)) {
		executor.RegisterExecutor(g.groups[key])
	}
	return
}

// Execute will execute all the groups according to the priorities of the functions
func (g *PriorityExecutionGroup[T]) Execute(ctx context.Context) error {
	return g.executors().Execute(ctx)
}
