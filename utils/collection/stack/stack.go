package stack

import (
	"iter"
	"slices"
)

// NewStack returns a stack which is not thread safe
func NewStack[T any]() IStack[T] {
	return &Stack[T]{top: nil, length: 0}
}

type Stack[T any] struct {
	top    *node[T]
	length int
}

func (s *Stack[T]) IsEmpty() bool {
	return s.length == 0
}

func (s *Stack[T]) Clear() {
	s.top = nil
	s.length = 0
}

func (s *Stack[T]) Values() iter.Seq[T] {
	return func(yield func(T) bool) {
		length := s.Len()
		for i := 0; i < length; i++ {
			v, ok := s.Pop()
			if !yield(v) || !ok {
				return
			}
		}
	}

}

type node[T any] struct {
	value T
	prev  *node[T]
}

func (s *Stack[T]) Len() int {
	return s.length
}

func (s *Stack[T]) Peek() (element T, ok bool) {
	if s.length == 0 {
		return
	}
	ok = true
	element = s.top.value
	return
}

// Pop the top item of the stack and return it
func (s *Stack[T]) Pop() (element T, ok bool) {
	if s.length == 0 {
		return
	}
	ok = true
	n := s.top
	s.top = n.prev
	s.length--
	element = n.value
	return
}

func (s *Stack[T]) Push(value ...T) {
	s.PushSequence(slices.Values(value))
}

func (s *Stack[T]) PushSequence(seq iter.Seq[T]) {
	for v := range seq {
		s.push(v)
	}
}

func (s *Stack[T]) push(value T) {
	n := &node[T]{value, s.top}
	s.top = n
	s.length++
}
