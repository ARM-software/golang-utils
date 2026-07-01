package collection

import (
	"iter"
	"slices"
)

// Reverse returns a reversed copy of s.
func Reverse[S ~[]E, E any](s S) S {
	clone := slices.Clone(s)
	slices.Reverse(clone)
	return clone
}

// Enumerate yields pairs of index and value for slice.
func Enumerate[S ~[]E, E any](slice S) iter.Seq2[int, E] {
	return EnumerateSequence(slices.Values(slice))
}

// EnumerateSequence yields pairs of index and value for seq.
func EnumerateSequence[E any](seq iter.Seq[E]) iter.Seq2[int, E] {
	return func(yield func(int, E) bool) {
		index := 0
		for value := range SequenceOrEmpty(seq) {
			if !yield(index, value) {
				return
			}
			index++
		}
	}
}

// ReverseSequence yields the values of seq in reverse order.
func ReverseSequence[E any](seq iter.Seq[E]) iter.Seq[E] {
	return slices.Values(Reverse(slices.Collect(SequenceOrEmpty(seq))))
}

// Zip yields pairs of corresponding values from left and right.
//
// Iteration stops when either side runs out of values.
func Zip[S1 ~[]E1, S2 ~[]E2, E1, E2 any](left S1, right S2) iter.Seq2[E1, E2] {
	return ZipSequence(slices.Values(left), slices.Values(right))
}

// ZipSequence yields pairs of corresponding values from left and right.
//
// Iteration stops when either side runs out of values.
func ZipSequence[E1, E2 any](left iter.Seq[E1], right iter.Seq[E2]) iter.Seq2[E1, E2] {
	leftValues := slices.Collect(SequenceOrEmpty(left))
	rightValues := slices.Collect(SequenceOrEmpty(right))
	length := min(len(leftValues), len(rightValues))
	return func(yield func(E1, E2) bool) {
		for i := 0; i < length; i++ {
			if !yield(leftValues[i], rightValues[i]) {
				return
			}
		}
	}
}
