package collection

import (
	"iter"
	"slices"
)

// First returns the first element of slice.
func First[S ~[]E, E any](slice S) (element E, ok bool) {
	return At(slice, 0)
}

// FirstRef behaves like FirstByRef.
func FirstRef[S ~[]E, E any](slice S, predicate PredicateRef[E]) (element E, ok bool) {
	return FirstByRef(slice, predicate)
}

// FirstBy returns the first element of slice that satisfies predicate.
func FirstBy[S ~[]E, E any](slice S, predicate Predicate[E]) (element E, ok bool) {
	return FirstSequence(FilterSequence(slices.Values(slice), predicate))
}

// FirstByRef returns the first element of slice that satisfies a reference-based predicate.
func FirstByRef[S ~[]E, E any](slice S, predicate PredicateRef[E]) (element E, ok bool) {
	return FirstBy(slice, toPredicateFunc(predicate))
}

// FirstSequence returns the first element yielded by seq.
func FirstSequence[E any](seq iter.Seq[E]) (element E, ok bool) {
	return AtSequence(seq, 0)
}

// FirstRefSequence behaves like FirstByRefSequence.
func FirstRefSequence[E any](seq iter.Seq[E], predicate PredicateRef[E]) (element E, ok bool) {
	return FirstByRefSequence(seq, predicate)
}

// FirstBySequence returns the first element yielded by seq that satisfies predicate.
func FirstBySequence[E any](seq iter.Seq[E], predicate Predicate[E]) (element E, ok bool) {
	return FirstSequence(FilterSequence(seq, predicate))
}

// FirstByRefSequence returns the first element yielded by seq that satisfies a reference-based predicate.
func FirstByRefSequence[E any](seq iter.Seq[E], predicate PredicateRef[E]) (element E, ok bool) {
	return FirstBySequence(seq, toPredicateFunc(predicate))
}

// Last returns the last element of slice.
func Last[S ~[]E, E any](slice S) (element E, ok bool) {
	return At(slice, -1)
}

// LastRef behaves like LastByRef.
func LastRef[S ~[]E, E any](slice S, predicate PredicateRef[E]) (element E, ok bool) {
	return LastByRef(slice, predicate)
}

// LastBy returns the last element of slice that satisfies predicate.
func LastBy[S ~[]E, E any](slice S, predicate Predicate[E]) (element E, ok bool) {
	return LastSequence(FilterSequence(slices.Values(slice), predicate))
}

// LastByRef returns the last element of slice that satisfies a reference-based predicate.
func LastByRef[S ~[]E, E any](slice S, predicate PredicateRef[E]) (element E, ok bool) {
	return LastBy(slice, toPredicateFunc(predicate))
}

// LastSequence returns the last element yielded by seq.
func LastSequence[E any](seq iter.Seq[E]) (element E, ok bool) {
	return AtSequence(seq, -1)
}

// LastRefSequence behaves like LastByRefSequence.
func LastRefSequence[E any](seq iter.Seq[E], predicate PredicateRef[E]) (element E, ok bool) {
	return LastByRefSequence(seq, predicate)
}

// LastBySequence returns the last element yielded by seq that satisfies predicate.
func LastBySequence[E any](seq iter.Seq[E], predicate Predicate[E]) (element E, ok bool) {
	return LastSequence(FilterSequence(seq, predicate))
}

// LastByRefSequence returns the last element yielded by seq that satisfies a reference-based predicate.
func LastByRefSequence[E any](seq iter.Seq[E], predicate PredicateRef[E]) (element E, ok bool) {
	return LastBySequence(seq, toPredicateFunc(predicate))
}
