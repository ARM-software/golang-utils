package collection

import "iter"

// At returns the element at index from slice.
//
// If index is out of range, it returns the zero value and false.
func At[S ~[]E, E any](slice S, index int) (element E, ok bool) {
	if index < 0 || index >= len(slice) {
		return
	}
	element = slice[index]
	ok = true
	return
}

// AtSequence returns the element at index from seq.
//
// If index is out of range, it returns the zero value and false.
func AtSequence[E any](seq iter.Seq[E], index int) (element E, ok bool) {
	if index < 0 {
		return
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
