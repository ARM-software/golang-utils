package collection

import (
	"iter"
	"slices"
)

//
// Reduce utilities
//

// ReduceFunc defines a reducer that combines an accumulator and an element to
// produce a new accumulator.
//
// Reducers are useful for aggregating collections into a single result, such as
// totals, counts, grouped values, flattened structures, or derived summary
// objects.
type ReduceFunc[T1, T2 any] func(T2, T1) T2

// ReduceRefFunc defines a reducer that operates on references.
//
// Reference-based reducers are useful when the accumulator or elements are more
// naturally handled as pointers, for example when building up mutable state or
// sharing reducer logic with other pointer-oriented helpers.
type ReduceRefFunc[T1, T2 any] func(*T2, *T1) *T2

// Reduce folds over the slice s using f, starting with accumulator.
//
// This is useful when a collection needs to be collapsed into one value, for
// example summing values, building a lookup map, or computing a derived result
// from many elements.
//
// Reference documentation:
//   - https://en.wikipedia.org/wiki/Fold_(higher-order_function)
//   - https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/reduce
func Reduce[T1, T2 any](s []T1, accumulator T2, f ReduceFunc[T1, T2]) T2 {
	return ReducesSequence(slices.Values(s), accumulator, f)
}

// ReducesSequence folds over a sequence using f, starting with accumulator.
//
// This is the sequence-oriented counterpart to Reduce and is useful when values
// are consumed lazily rather than from a prebuilt slice, such as values flowing
// through iterators or other on-demand pipelines.
//
// Reference documentation:
//   - https://pkg.go.dev/iter
func ReducesSequence[T1, T2 any](s iter.Seq[T1], accumulator T2, f ReduceFunc[T1, T2]) T2 {
	result := accumulator
	for e := range sequenceOrEmpty(s) {
		result = f(result, e)
	}
	return result
}
