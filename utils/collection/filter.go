package collection

import (
	"iter"
	"regexp"
	"slices"
	"strings"

	"github.com/ARM-software/golang-utils/utils/field"
)

//
// Predicate & filter types
//

// FilterFunc defines a function that evaluates a value and returns true when
// the value satisfies the condition.
//
// Filter predicates are useful for selecting only the elements that should be
// kept, such as enabled records, matching names, valid values, or items within
// a threshold.
type FilterFunc[E any] func(E) bool

// FilterRefFunc defines a function that evaluates a pointer to a value and
// returns true when the referenced value satisfies the condition.
//
// Reference-based predicates are useful when the predicate should be able to
// work with optional values, share logic with other pointer-oriented helpers,
// or avoid copying larger element values.
type FilterRefFunc[E any] func(*E) bool

// Predicate is an alias for FilterFunc to express boolean tests.
type Predicate[E any] = FilterFunc[E]

// PredicateRef is an alias for FilterRefFunc to express boolean tests on
// references.
type PredicateRef[E any] = FilterRefFunc[E]

// toPredicateFunc adapts a PredicateRef (reference-based predicate) to
// a Predicate (value-based) by converting the value into an optional
// reference using field.ToOptional.
func toPredicateFunc[E any](f PredicateRef[E]) Predicate[E] {
	return func(e E) bool {
		return f(field.ToOptional(e))
	}
}

// MatchFunc compares two values of type E and reports whether they match.
// It may return an error if the comparison requires additional processing,
// such as compiling or evaluating a regular expression.
type MatchFunc[E any] func(E, E) (bool, error)

// MatchRefFunc compares two references to values of type E and reports whether
// they match. It may return an error if the comparison requires additional
// processing before the match can be determined.
type MatchRefFunc[E any] func(*E, *E) (bool, error)

func matchToPredicateFunc[E any](v E, matchFunc MatchFunc[E]) Predicate[E] {
	return func(e E) bool {
		matched, err := matchFunc(v, e)
		return matched && err == nil
	}
}

func matchToPredicateRefFunc[E any](v *E, matchFunc MatchRefFunc[E]) PredicateRef[E] {
	return func(e *E) bool {
		matched, err := matchFunc(v, e)
		return matched && err == nil
	}
}

// StringMatch reports whether two strings are exactly equal.
var StringMatch MatchFunc[string] = func(a, b string) (bool, error) { return a == b, nil }

// StringCaseInsensitiveMatch reports whether two strings are equal but ignoring their case.
var StringCaseInsensitiveMatch MatchFunc[string] = func(a, b string) (bool, error) { return strings.EqualFold(a, b), nil }

// StringCleanCaseInsensitiveMatch reports whether two strings are equal after
// trimming surrounding whitespace and ignoring their case.
var StringCleanCaseInsensitiveMatch MatchFunc[string] = func(a, b string) (bool, error) {
	return StringCaseInsensitiveMatch(strings.TrimSpace(a), strings.TrimSpace(b))
}

// StringCleanMatch reports whether two strings are exactly equal after trimming
// surrounding whitespace.
var StringCleanMatch MatchFunc[string] = func(a, b string) (bool, error) { return StringMatch(strings.TrimSpace(a), strings.TrimSpace(b)) }

// StringRegexMatch reports whether a string matches the provided regular expression pattern.
// the pattern being the first argument.
var StringRegexMatch MatchFunc[string] = regexp.MatchString

// StrictRefMatch reports whether two references are equal using field.Equal.
func StrictRefMatch[E comparable](a, b *E) (bool, error) { return field.Equal(a, b), nil }

//
// Rejection / Filtering
//

// Filter returns a new slice containing elements from s for which f returns
// true.
//
// This is useful when a collection should be narrowed down to only the items of
// interest, for example active users, matching paths, or values that pass a
// validation predicate.
//
// Reference documentation:
//   - https://en.wikipedia.org/wiki/Filter_(higher-order_function)
//   - https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/filter
func Filter[S ~[]E, E any](s S, f FilterFunc[E]) S {
	return slices.Collect[E](FilterSequence[E](slices.Values(s), f))
}

// FilterRef behaves like Filter but accepts a reference-based predicate.
//
// This is useful when the predicate naturally works on pointers or when it is
// shared with other APIs that already use PredicateRef.
func FilterRef[S ~[]E, E any](s S, f FilterRefFunc[E]) S {
	return Filter[S](s, toPredicateFunc(f))
}

// FilterSequence returns a sequence that yields only elements for which f
// returns true.
//
// This is the lazy sequence-oriented counterpart to Filter and is useful when
// composing iterator pipelines, avoiding intermediate slices, or processing
// potentially large or streaming inputs incrementally.
//
// Reference documentation:
//   - https://pkg.go.dev/iter
func FilterSequence[E any](s iter.Seq[E], f Predicate[E]) (result iter.Seq[E]) {
	return func(yield func(E) bool) {
		for v := range s {
			if f(v) && !yield(v) {
				return
			}
		}
	}
}

// FilterRefSequence behaves like FilterSequence but accepts a reference-based
// predicate.
//
// This is useful when sequence processing should remain lazy while the
// predicate logic is written against PredicateRef.
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

// RejectSequence returns a sequence that yields elements for which f returns
// false.
//
// This is useful when the complement of a filter condition should be applied in
// a lazy pipeline.
func RejectSequence[E any](s iter.Seq[E], f FilterFunc[E]) iter.Seq[E] {
	return FilterSequence(s, OppositeFunc[E](f))
}
