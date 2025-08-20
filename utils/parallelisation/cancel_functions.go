/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parallelisation

import "context"

type CancelFunctionStore struct {
	ExecutionGroup[context.CancelFunc]
}

func (s *CancelFunctionStore) RegisterCancelFunction(cancel ...context.CancelFunc) {
	s.ExecutionGroup.RegisterFunction(cancel...)
}

func (s *CancelFunctionStore) RegisterCancelStore(store *CancelFunctionStore) {
	if store == nil {
		return
	}
	s.RegisterCancelFunction(func() {
		store.Cancel()
	})
}

// Cancel will execute the cancel functions in the store. Any errors will be ignored and Execute() is recommended if you need to know if a cancellation failed
func (s *CancelFunctionStore) Cancel() {
	_ = s.Execute(context.Background())
}

func (s *CancelFunctionStore) Len() int {
	return s.ExecutionGroup.Len()
}

// NewCancelFunctionsStore creates a store for cancel functions. Whatever the options passed, all cancel functions will be executed and cleared. In other words, options `RetainAfterExecution` and `StopOnFirstError` would be discarded if selected to create the Cancel store
func NewCancelFunctionsStore(options ...StoreOption) *CancelFunctionStore {
	return &CancelFunctionStore{
		ExecutionGroup: *NewExecutionGroup[context.CancelFunc](func(ctx context.Context, cancelFunc context.CancelFunc) error {
			return WrapCancelToContextualFunc(cancelFunc)(ctx)
		}, append(options, ClearAfterExecution, ExecuteAll)...),
	}
}
