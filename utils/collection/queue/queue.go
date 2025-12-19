package queue

import (
	"iter"
	"slices"
)

// NewQueue returns a Queue which is not thread safe
func NewQueue[T any]() IQueue[T] {
	return &Queue[T]{nil, nil, 0}
}

type Queue[T any] struct {
	start, end *node[T]
	length     int
}

func (s *Queue[T]) IsEmpty() bool {
	return s.length == 0
}

func (s *Queue[T]) Clear() {
	s.start = nil
	s.end = nil
	s.length = 0
}

func (s *Queue[T]) Values() iter.Seq[T] {
	return func(yield func(T) bool) {
		length := s.Len()
		for i := 0; i < length; i++ {
			v := s.Dequeue()
			if !yield(v) {
				return
			}
		}
	}

}

type node[T any] struct {
	value T
	next  *node[T]
}

func (s *Queue[T]) Len() int {
	return s.length
}

func (s *Queue[T]) Peek() (element T) {
	if s.length == 0 {
		return
	}
	element = s.start.value
	return
}

func (s *Queue[T]) Dequeue() (element T) {
	if s.length == 0 {
		return
	}
	n := s.start
	if s.length == 1 {
		s.start = nil
		s.end = nil
	} else {
		s.start = s.start.next
	}
	s.length--
	element = n.value
	return
}

func (s *Queue[T]) Enqueue(value ...T) {
	s.EnqueueSequence(slices.Values(value))
}

func (s *Queue[T]) EnqueueSequence(seq iter.Seq[T]) {
	for v := range seq {
		s.enqueue(v)
	}
}

func (s *Queue[T]) enqueue(value T) {
	n := &node[T]{value, nil}
	if s.length == 0 {
		s.start = n
		s.end = n
	} else {
		s.end.next = n
		s.end = n
	}
	s.length++
}
