package collection

import (
	"iter"
	"slices"
)

//
// Reduce utilities
//

// ReduceFunc defines a reducer that combines an accumulator and an element to produce a new accumulator.
type ReduceFunc[T1, T2 any] func(T2, T1) T2

// ReduceRefFunc defines a reducer that operates on references.
type ReduceRefFunc[T1, T2 any] func(*T2, *T1) *T2

// Reduce folds over the slice s using f, starting with accumulator.
func Reduce[T1, T2 any](s []T1, accumulator T2, f ReduceFunc[T1, T2]) T2 {
	return ReducesSequence(slices.Values(s), accumulator, f)
}

// ReducesSequence folds over a sequence using f, starting with accumulator.
func ReducesSequence[T1, T2 any](s iter.Seq[T1], accumulator T2, f ReduceFunc[T1, T2]) T2 {
	result := accumulator
	for e := range s {
		result = f(result, e)
	}
	return result
}
