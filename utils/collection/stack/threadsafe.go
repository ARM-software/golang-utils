package stack

import (
	"iter"
	"sync"
)

// NewThreadSafeStack  returns a thread safe stack
// This is inspired from https://github.com/hayageek/threadsafe
func NewThreadSafeStack[T any]() IStack[T] {
	return &SafeStack[T]{
		s:  NewStack[T](),
		mu: sync.Mutex{},
	}
}

type SafeStack[T any] struct {
	s  IStack[T]
	mu sync.Mutex
}

func (s *SafeStack[T]) Push(value ...T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s.Push(value...)
}

func (s *SafeStack[T]) PushSequence(seq iter.Seq[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s.PushSequence(seq)
}

func (s *SafeStack[T]) Pop() T {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.s.Pop()
}

func (s *SafeStack[T]) Peek() T {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.s.Peek()
}

func (s *SafeStack[T]) Values() iter.Seq[T] {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.s.Values()
}

func (s *SafeStack[T]) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.s.Len()
}

func (s *SafeStack[T]) IsEmpty() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.s.IsEmpty()
}

func (s *SafeStack[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s.Clear()
}
