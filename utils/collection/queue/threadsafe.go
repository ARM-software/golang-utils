package queue

import (
	"iter"
	"sync"
)

// NewThreadSafeQueue returns a thread safe queue.
// This is inspired from https://github.com/hayageek/threadsafe.
func NewThreadSafeQueue[T any]() IQueue[T] {
	return &SafeQueue[T]{
		q:  NewQueue[T](),
		mu: sync.Mutex{},
	}
}

type SafeQueue[T any] struct {
	q  IQueue[T]
	mu sync.Mutex
}

func (q *SafeQueue[T]) Enqueue(value ...T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.q.Enqueue(value...)
}

func (q *SafeQueue[T]) EnqueueSequence(seq iter.Seq[T]) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.q.EnqueueSequence(seq)
}

func (q *SafeQueue[T]) Dequeue() T {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.q.Dequeue()
}

func (q *SafeQueue[T]) Peek() T {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.q.Peek()
}

func (q *SafeQueue[T]) Values() iter.Seq[T] {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.q.Values()
}

func (q *SafeQueue[T]) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.q.Len()
}

func (q *SafeQueue[T]) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.q.IsEmpty()
}

func (q *SafeQueue[T]) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.q.Clear()
}
