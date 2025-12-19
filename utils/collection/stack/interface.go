package stack

import "iter"

//go:generate go tool mockgen -destination=../../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/collection/$GOPACKAGE IStack

// IStack specifies the behaviour of a last-in, first-out (LIFO) collection.
// it is inspired by the work of https://github.com/hayageek/threadsafe/ and https://github.com/golang-collections/collections
type IStack[T any] interface {
	// Push adds elements to the stack.
	Push(value ...T)
	// PushSequence adds elements to the stack.
	PushSequence(seq iter.Seq[T])
	//Pop removes and returns an element from the stack.
	Pop() T
	//Peek returns the element at the top of the stack without removing it.
	Peek() T
	// IsEmpty states whether the stack is empty.
	IsEmpty() bool
	// Clear clears all elements from the stack.
	Clear()
	// Values returns all the elements in the stack. The stack will be empty as a result.
	Values() iter.Seq[T]
	// Len returns the number of elements in the stack.
	Len() int
}
