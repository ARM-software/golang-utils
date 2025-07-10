/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parallelisation

import (
	"context"

	"github.com/sasha-s/go-deadlock"
)

type CancelFunctionStore struct {
	mu              deadlock.RWMutex
	cancelFunctions []context.CancelFunc
}

func (s *CancelFunctionStore) RegisterCancelFunction(cancel ...context.CancelFunc) {
	defer s.mu.Unlock()
	s.mu.Lock()
	s.cancelFunctions = append(s.cancelFunctions, cancel...)
}

func (s *CancelFunctionStore) Cancel() {
	defer s.mu.Unlock()
	s.mu.Lock()
	for _, c := range s.cancelFunctions {
		c()
	}
	s.cancelFunctions = []context.CancelFunc{}
}

func (s *CancelFunctionStore) Len() int {
	defer s.mu.RUnlock()
	s.mu.RLock()
	return len(s.cancelFunctions)
}

func NewCancelFunctionsStore() *CancelFunctionStore {
	return &CancelFunctionStore{
		cancelFunctions: []context.CancelFunc{},
	}
}
