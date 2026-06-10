package collection

import (
	"iter"
	"slices"
)

// At returns the element at index from slice.
//
// Negative indices are interpreted relative to the end of the slice, following
// the same pattern as Python indexing, so `-1` selects the last element, `-2`
// the penultimate one, and so on.
//
// If index is out of range, it returns the zero value and false.
//
// References:
//   - https://docs.python.org/3/tutorial/introduction.html#lists
//   - https://www.geeksforgeeks.org/python/what-is-negative-indexing-in-python/
func At[S ~[]E, E any](slice S, index int) (element E, ok bool) {
	index = normaliseIndex(len(slice), index)
	if index < 0 || index >= len(slice) {
		return
	}
	element = slice[index]
	ok = true
	return
}

// AtSequence returns the element at index from seq.
//
// Negative indices are interpreted relative to the end of the sequence,
// following the same pattern as Python indexing.
//
// If index is out of range, it returns the zero value and false.
//
// References:
//   - https://docs.python.org/3/tutorial/introduction.html#lists
//   - https://www.geeksforgeeks.org/python/what-is-negative-indexing-in-python/
func AtSequence[E any](seq iter.Seq[E], index int) (element E, ok bool) {
	if index < 0 {
		return At(slices.Collect(SequenceOrEmpty(seq)), index)
	}
	for v := range SequenceOrEmpty(seq) {
		if index == 0 {
			element = v
			ok = true
			return
		}
		index--
	}
	return
}

// Nth behaves like At.
func Nth[S ~[]E, E any](slice S, index int) (element E, ok bool) {
	return At(slice, index)
}

// NthSequence behaves like AtSequence.
func NthSequence[E any](seq iter.Seq[E], index int) (element E, ok bool) {
	return AtSequence(seq, index)
}

func normaliseIndex(length int, index int) int {
	if index < 0 {
		return length + index
	}
	return index
}
