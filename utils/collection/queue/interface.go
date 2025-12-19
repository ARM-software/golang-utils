package queue

import "iter"

// IQueue specifies the behaviour of a first-in, first-out (FIFO) collection.
// It is inspired by the work of https://github.com/hayageek/threadsafe/ and
// https://github.com/golang-collections/collections.
type IQueue[T any] interface {
	// Enqueue adds an element to the queue.
	Enqueue(value ...T)
	// EnqueueSequence adds an element to the queue.
	EnqueueSequence(value iter.Seq[T])
	// Dequeue removes and returns an element from the queue. It returns ok true if the queue is not empty.
	Dequeue() (element T, ok bool)
	// Peek returns the element at the front of the queue without removing it. It returns ok true if the queue is not empty.
	Peek() (element T, ok bool)
	// IsEmpty states whether the queue is empty.
	IsEmpty() bool
	// Clear removes all elements from the queue.
	Clear()
	// Values returns all the elements in the queue. The queue will be empty as a result.
	Values() iter.Seq[T]
	// Len returns the number of elements in the queue.
	Len() int
}
