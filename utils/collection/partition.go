package collection

import (
	"iter"
	"slices"
)

// Partition splits slice into matching and non-matching values according to predicate.
func Partition[S ~[]E, E any](slice S, predicate Predicate[E]) (matched S, unmatched S) {
	return PartitionSequence[S](slices.Values(slice), predicate)
}

// PartitionRef behaves like Partition but accepts a reference-based predicate.
func PartitionRef[S ~[]E, E any](slice S, predicate PredicateRef[E]) (matched S, unmatched S) {
	return Partition(slice, toPredicateFunc(predicate))
}

// PartitionSequence splits seq into matching and non-matching values according to predicate.
func PartitionSequence[S ~[]E, E any](seq iter.Seq[E], predicate Predicate[E]) (matched S, unmatched S) {
	for value := range SequenceOrEmpty(seq) {
		if predicate(value) {
			matched = append(matched, value)
			continue
		}
		unmatched = append(unmatched, value)
	}
	return
}

// PartitionRefSequence behaves like PartitionSequence but accepts a reference-based predicate.
func PartitionRefSequence[S ~[]E, E any](seq iter.Seq[E], predicate PredicateRef[E]) (matched S, unmatched S) {
	return PartitionSequence[S](seq, toPredicateFunc(predicate))
}
