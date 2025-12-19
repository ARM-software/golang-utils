package queue

import (
	"iter"
	"slices"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

// ChanQueue is a channel-backed implementation of IQueue.
//
// All methods are safe for concurrent use.
// Enqueue blocks when the queue is full, while Dequeue and Peek are non-blocking
// and return the zero value when the queue is empty.
// The queue has a fixed capacity defined at creation time.
// If performance is not enough, worth looking at https://github.com/alphadose/ZenQ
type ChanQueue[T any] struct {
	ch chan T
}

// NewChannelQueue creates a new channel-backed queue with the given capacity.
// Capacity must be > 0.
func NewChannelQueue[T any](capacity int) (IQueue[T], error) {
	if capacity <= 0 {
		return nil, commonerrors.Newf(commonerrors.ErrInvalid, "invalid capacity value [%d]", capacity)
	}
	return &ChanQueue[T]{
		ch: make(chan T, capacity),
	}, nil
}

// Enqueue adds one or more elements to the queue.
// This blocks if the queue is full.
func (q *ChanQueue[T]) Enqueue(values ...T) {
	q.EnqueueSequence(slices.Values(values))
}

// EnqueueSequence adds a sequence to the queue.
// This blocks if the queue is full.
func (q *ChanQueue[T]) EnqueueSequence(seq iter.Seq[T]) {
	for v := range seq {
		q.ch <- v
	}
}

// Dequeue removes and returns the next element.
// If the queue is empty, it returns the zero value.
func (q *ChanQueue[T]) Dequeue() (element T, ok bool) {
	select {
	case v := <-q.ch:
		element = v
		ok = true
		return
	default:
		return
	}
}

// Peek returns the next element without removing it.
// If the queue is empty, it returns the zero value.
func (q *ChanQueue[T]) Peek() (element T, ok bool) {
	select {
	case v := <-q.ch:
		// put it back
		q.ch <- v
		element = v
		ok = true
		return
	default:
		return
	}
}

// IsEmpty reports whether the queue is empty.
func (q *ChanQueue[T]) IsEmpty() bool {
	return q.Len() == 0
}

// Len returns the number of elements currently in the queue.
func (q *ChanQueue[T]) Len() int {
	return len(q.ch)
}

// Clear removes all elements from the queue.
func (q *ChanQueue[T]) Clear() {
	for {
		select {
		case <-q.ch:
			// drain
		default:
			return
		}
	}
}

// Values returns all elements in FIFO order and drains the queue.
func (q *ChanQueue[T]) Values() iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			select {
			case v := <-q.ch:
				if !yield(v) {
					return
				}
			default:
				return
			}
		}
	}
}
