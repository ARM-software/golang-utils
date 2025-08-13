/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parallelisation

import (
	"context"

	"github.com/sasha-s/go-deadlock"
	"golang.org/x/sync/errgroup"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

type StoreOptions struct {
	clearOnExecution bool
	stopOnFirstError bool
	sequential       bool
	reverse          bool
	joinErrors       bool
}
type StoreOption func(*StoreOptions) *StoreOptions

// StopOnFirstError stops store execution on first error.
var StopOnFirstError StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		return o
	}
	o.stopOnFirstError = true
	o.joinErrors = false
	return o
}

// JoinErrors will collate any errors which happened when executing functions in store.
// This option should not be used in combination to StopOnFirstError.
var JoinErrors StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		return o
	}
	o.stopOnFirstError = false
	o.joinErrors = true
	return o
}

// ExecuteAll executes all functions in the store even if an error is raised. the first error raised is then returned.
var ExecuteAll StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		return o
	}
	o.stopOnFirstError = false
	return o
}

// ClearAfterExecution clears the store after execution.
var ClearAfterExecution StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		return o
	}
	o.clearOnExecution = true
	return o
}

// RetainAfterExecution keep the store intact after execution (no reset).
var RetainAfterExecution StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		return o
	}
	o.clearOnExecution = false
	return o
}

// Parallel ensures every function registered in the store is executed concurrently in the order they were registered.
var Parallel StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		return o
	}
	o.sequential = false
	return o
}

// Sequential ensures every function registered in the store is executed sequentially in the order they were registered.
var Sequential StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		return o
	}
	o.sequential = true
	return o
}

// SequentialInReverse ensures every function registered in the store is executed sequentially but in the reverse order they were registered.
var SequentialInReverse StoreOption = func(o *StoreOptions) *StoreOptions {
	if o == nil {
		return o
	}
	o.sequential = true
	o.reverse = true
	return o
}

func newFunctionStore[T any](executeFunc func(context.Context, T) error, options ...StoreOption) *store[T] {

	opts := &StoreOptions{}

	for i := range options {
		opts = options[i](opts)
	}
	return &store[T]{
		mu:          deadlock.RWMutex{},
		functions:   make([]T, 0),
		executeFunc: executeFunc,
		options:     *opts,
	}
}

type store[T any] struct {
	mu          deadlock.RWMutex
	functions   []T
	executeFunc func(ctx context.Context, element T) error
	options     StoreOptions
}

func (s *store[T]) RegisterFunction(function ...T) {
	defer s.mu.Unlock()
	s.mu.Lock()
	s.functions = append(s.functions, function...)
}

func (s *store[T]) Len() int {
	defer s.mu.RUnlock()
	s.mu.RLock()
	return len(s.functions)
}

func (s *store[T]) Execute(ctx context.Context) (err error) {
	defer s.mu.Unlock()
	s.mu.Lock()
	if reflection.IsEmpty(s.executeFunc) {
		return commonerrors.New(commonerrors.ErrUndefined, "the store was not initialised correctly")
	}

	if s.options.sequential {
		err = s.executeSequentially(ctx, s.options.stopOnFirstError, s.options.reverse, s.options.joinErrors)
	} else {
		err = s.executeConcurrently(ctx, s.options.stopOnFirstError, s.options.joinErrors)
	}

	if err == nil && s.options.clearOnExecution {
		s.functions = make([]T, 0, len(s.functions))
	}
	return
}

func (s *store[T]) executeConcurrently(ctx context.Context, stopOnFirstError bool, collateErrors bool) error {
	g, gCtx := errgroup.WithContext(ctx)
	if !stopOnFirstError {
		gCtx = ctx
	}
	funcNum := len(s.functions)
	errCh := make(chan error, funcNum)
	g.SetLimit(funcNum)
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

func (s *store[T]) executeSequentially(ctx context.Context, stopOnFirstError, reverse, collateErrors bool) (err error) {
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

func (s *store[T]) executeFunction(ctx context.Context, element T) (mustBreak bool, err error) {
	err = DetermineContextError(ctx)
	if err != nil {
		mustBreak = true
		return
	}
	err = s.executeFunc(ctx, element)
	return
}

type CancelFunctionStore struct {
	store[context.CancelFunc]
}

func (s *CancelFunctionStore) RegisterCancelFunction(cancel ...context.CancelFunc) {
	s.store.RegisterFunction(cancel...)
}

// Cancel will execute the cancel functions in the store. Any errors will be ignored and Execute() is recommended if you need to know if a cancellation failed
func (s *CancelFunctionStore) Cancel() {
	_ = s.Execute(context.Background())
}

func (s *CancelFunctionStore) Len() int {
	return s.store.Len()
}

// NewCancelFunctionsStore creates a store for cancel functions. Whatever the options passed, all cancel functions will be executed and cleared. In other words, options `RetainAfterExecution` and `StopOnFirstError` would be discarded if selected to create the Cancel store
func NewCancelFunctionsStore(options ...StoreOption) *CancelFunctionStore {
	return &CancelFunctionStore{
		store: *newFunctionStore[context.CancelFunc](func(_ context.Context, cancelFunc context.CancelFunc) error {
			cancelFunc()
			return nil
		}, append(options, ClearAfterExecution, ExecuteAll)...),
	}
}
