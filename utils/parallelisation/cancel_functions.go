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

func newFunctionStore[T any](clearOnExecution, stopOnFirstError bool, executeFunc func(context.Context, T) error) *store[T] {
	return &store[T]{
		mu:               deadlock.RWMutex{},
		functions:        make([]T, 0),
		executeFunc:      executeFunc,
		clearOnExecution: clearOnExecution,
		stopOnFirstError: stopOnFirstError,
	}
}

type store[T any] struct {
	mu               deadlock.RWMutex
	functions        []T
	executeFunc      func(ctx context.Context, element T) error
	clearOnExecution bool
	stopOnFirstError bool
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

func (s *store[T]) Execute(ctx context.Context) error {
	defer s.mu.Unlock()
	s.mu.Lock()
	if reflection.IsEmpty(s.executeFunc) {
		return commonerrors.New(commonerrors.ErrUndefined, "the cancel store was not initialised correctly")
	}
	g, gCtx := errgroup.WithContext(ctx)
	if !s.stopOnFirstError {
		gCtx = ctx
	}
	g.SetLimit(len(s.functions))
	for i := range s.functions {
		g.Go(func() error {
			err := DetermineContextError(gCtx)
			if err != nil {
				return err
			}
			return s.executeFunc(gCtx, s.functions[i])
		})
	}

	err := g.Wait()
	if err == nil && s.clearOnExecution {
		s.functions = make([]T, 0, len(s.functions))
	}
	return err
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

// NewCancelFunctionsStore creates a store for cancel functions.
func NewCancelFunctionsStore() *CancelFunctionStore {
	return &CancelFunctionStore{
		store: *newFunctionStore[context.CancelFunc](true, false, func(_ context.Context, cancelFunc context.CancelFunc) error {
			cancelFunc()
			return nil
		}),
	}
}
