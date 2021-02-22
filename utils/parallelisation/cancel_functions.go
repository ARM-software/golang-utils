package parallelisation

import (
	"context"
	"sync"
)

type CancelFunctionStore struct {
	mu              sync.RWMutex
	cancelFunctions []context.CancelFunc
}

func (s *CancelFunctionStore) RegisterCancelFunction(cancel ...context.CancelFunc) {
	defer s.mu.Unlock()
	s.mu.Lock()
	s.cancelFunctions = append(s.cancelFunctions, cancel...)
}

func (s *CancelFunctionStore) Cancel() {
	defer s.mu.RUnlock()
	s.mu.RLock()
	for _, c := range s.cancelFunctions {
		c()
	}
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
