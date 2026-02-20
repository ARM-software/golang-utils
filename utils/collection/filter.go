package collection

import (
	"iter"
	"slices"

	"github.com/ARM-software/golang-utils/utils/field"
)

//
// Predicate & filter types
//

// FilterFunc defines a function that evaluates a value and returns true
// when the value satisfies the condition.
type FilterFunc[E any] func(E) bool

// FilterRefFunc defines a function that evaluates a pointer to a value
// and returns true when the referenced value satisfies the condition.
type FilterRefFunc[E any] func(*E) bool

// Predicate is an alias for FilterFunc to express boolean tests.
type Predicate[E any] = FilterFunc[E]

// PredicateRef is an alias for FilterRefFunc to express boolean tests on references.
type PredicateRef[E any] = FilterRefFunc[E]

// toPredicateFunc adapts a PredicateRef (reference-based predicate) to
// a Predicate (value-based) by converting the value into an optional
// reference using field.ToOptional.
func toPredicateFunc[E any](f PredicateRef[E]) Predicate[E] {
	return func(e E) bool {
		return f(field.ToOptional(e))
	}
}

//
// Rejection / Filtering
//

// Filter returns a new slice containing elements from s for which f returns true.
func Filter[S ~[]E, E any](s S, f FilterFunc[E]) S {
	return slices.Collect[E](FilterSequence[E](slices.Values(s), f))
}

// FilterRef behaves like Filter but accepts a reference-based predicate.
func FilterRef[S ~[]E, E any](s S, f FilterRefFunc[E]) S {
	return Filter[S](s, toPredicateFunc(f))
}

// FilterSequence returns a sequence that yields only elements for which f returns true.
func FilterSequence[E any](s iter.Seq[E], f Predicate[E]) (result iter.Seq[E]) {
	return func(yield func(E) bool) {
		for v := range s {
			if f(v) && !yield(v) {
				return
			}
		}
	}
}

// FilterRefSequence behaves like FilterSequence but accepts a reference-based predicate.
func FilterRefSequence[E any](s iter.Seq[E], f PredicateRef[E]) (result iter.Seq[E]) {
	return FilterSequence(s, toPredicateFunc(f))
}

// OppositeFunc returns a predicate that negates the result of f.
func OppositeFunc[E any](f FilterFunc[E]) FilterFunc[E] { return func(e E) bool { return !f(e) } }

// Reject returns elements for which f returns false (the inverse of Filter).
// This returns a new slice rather than modifying the input.
func Reject[S ~[]E, E any](s S, f FilterFunc[E]) S {
	return Filter(s, OppositeFunc[E](f))
}

// RejectRef behaves like Reject but accepts a reference-based predicate.
func RejectRef[S ~[]E, E any](s S, f FilterRefFunc[E]) S {
	return Reject(s, toPredicateFunc(f))
}

// RejectSequence returns a sequence that yields elements for which f returns false.
func RejectSequence[E any](s iter.Seq[E], f FilterFunc[E]) iter.Seq[E] {
	return FilterSequence(s, OppositeFunc[E](f))
}
