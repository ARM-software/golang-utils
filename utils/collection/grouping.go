package collection

import (
	"iter"
	"slices"

	"github.com/ARM-software/golang-utils/utils/field"
)

// KeyFunc defines a function that maps an element to a comparable key.
//
// Key functions are useful when values need to be grouped, counted, or
// deduplicated based on a derived identity such as an ID, category, or
// normalised form.
type KeyFunc[T1 any, T2 comparable] = MapFunc[T1, T2]

// KeyRefFunc defines a function that maps a pointer to a value to a comparable
// key.
//
// Reference-based key functions are useful when grouping or counting logic is
// naturally written against optional or pointer-based values.
type KeyRefFunc[T1 any, T2 comparable] func(*T1) T2

func toKeyFunc[E any, K comparable](f KeyRefFunc[E, K]) KeyFunc[E, K] {
	return func(e E) K {
		return f(field.ToOptionalOrNilIfEmpty(e))
	}
}

// CountBy counts how many elements in a slice satisfy the predicate.
//
// This is useful when a collection should be reduced to the number of matching
// elements, for example, counting even values, enabled items, or values that
// match a threshold.
//
// Reference documentation:
//   - https://underscorejs.org/#countBy
func CountBy[S ~[]E, E any](slice S, predicate Predicate[E]) int {
	return CountBySequence(slices.Values(slice), predicate)
}

// CountByRef behaves like CountBy but accepts a reference-based predicate.
func CountByRef[S ~[]E, E any](slice S, predicate PredicateRef[E]) int {
	return CountBy(slice, toPredicateFunc(predicate))
}

// CountBySequence counts how many elements in a sequence satisfy predicate.
//
// This is useful when values are processed lazily through iterators rather than
// from a prebuilt slice.
func CountBySequence[E any](sequence iter.Seq[E], predicate Predicate[E]) int {
	return ReducesSequence(sequence, 0, func(acc int, element E) int {
		if predicate(element) {
			return acc + 1
		}
		return acc
	})
}

// CountByRefSequence behaves like CountBySequence but accepts a
// reference-based predicate.
func CountByRefSequence[E any](sequence iter.Seq[E], predicate PredicateRef[E]) int {
	return CountBySequence(sequence, toPredicateFunc(predicate))
}

// GroupBy groups slice elements by the key returned by keyFunc.
//
// This is useful when a collection should be partitioned into buckets, such as
// grouping users by role, strings by length, or values by status.
//
// Reference documentation:
//   - https://underscorejs.org/#groupBy
func GroupBy[S ~[]E, E any, K comparable](slice S, keyFunc KeyFunc[E, K]) map[K][]E {
	return GroupBySequence(slices.Values(slice), keyFunc)
}

// GroupByRef behaves like GroupBy but accepts a reference-based key function.
func GroupByRef[S ~[]E, E any, K comparable](slice S, keyFunc KeyRefFunc[E, K]) map[K][]E {
	return GroupBy(slice, toKeyFunc(keyFunc))
}

// GroupBySequence groups sequence elements by the key returned by keyFunc.
//
// This is useful when grouping should happen during lazy iteration rather than
// after collecting values into a slice.
func GroupBySequence[E any, K comparable](sequence iter.Seq[E], keyFunc KeyFunc[E, K]) map[K][]E {
	return ReducesSequence(sequence, map[K][]E{}, func(acc map[K][]E, element E) map[K][]E {
		key := keyFunc(element)
		acc[key] = append(acc[key], element)
		return acc
	})
}

// GroupByRefSequence behaves like GroupBySequence but accepts a reference-based
// key function.
func GroupByRefSequence[E any, K comparable](sequence iter.Seq[E], keyFunc KeyRefFunc[E, K]) map[K][]E {
	return GroupBySequence(sequence, toKeyFunc(keyFunc))
}

// FrequenciesBy counts how often each derived key occurs in slice.
//
// This is useful for building histograms or frequency tables, for example
// counting values by lowercase normalisation, status, or category.
//
// Reference documentation:
//   - https://underscorejs.org/#countBy
func FrequenciesBy[S ~[]E, E any, K comparable](slice S, keyFunc KeyFunc[E, K]) map[K]int {
	return FrequenciesBySequence(slices.Values(slice), keyFunc)
}

// FrequenciesByRef behaves like FrequenciesBy but accepts a reference-based key
// function.
func FrequenciesByRef[S ~[]E, E any, K comparable](slice S, keyFunc KeyRefFunc[E, K]) map[K]int {
	return FrequenciesBy(slice, toKeyFunc(keyFunc))
}

// FrequenciesBySequence counts how often each derived key occurs in a sequence.
func FrequenciesBySequence[E any, K comparable](sequence iter.Seq[E], keyFunc KeyFunc[E, K]) map[K]int {
	return ReducesSequence(sequence, map[K]int{}, func(acc map[K]int, element E) map[K]int {
		key := keyFunc(element)
		acc[key]++
		return acc
	})
}

// FrequenciesByRefSequence behaves like FrequenciesBySequence but accepts a
// reference-based key function.
func FrequenciesByRefSequence[E any, K comparable](sequence iter.Seq[E], keyFunc KeyRefFunc[E, K]) map[K]int {
	return FrequenciesBySequence(sequence, toKeyFunc(keyFunc))
}

// FlatMap maps each element and concatenates the returned slices.
//
// This is useful when each input element expands into zero or more output
// elements, such as splitting strings, flattening nested values, or projecting
// structs into multiple derived values.
//
// Reference documentation:
//   - https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/flatMap
func FlatMap[S ~[]E, E any, T any](slice S, mapper MapFunc[E, []T]) []T {
	return FlatMapSequence(slices.Values(slice), mapper)
}

// FlatMapRef behaves like FlatMap but accepts a reference-based mapper.
func FlatMapRef[S ~[]E, E any, T any](slice S, mapper MapRefFunc[E, []T]) []T {
	return FlatMapRefSequence(slices.Values(slice), mapper)
}

// FlatMapSequence maps each sequence element and concatenates the returned
// slices.
func FlatMapSequence[E any, T any](sequence iter.Seq[E], mapper MapFunc[E, []T]) []T {
	return ReducesSequence(sequence, []T{}, func(acc []T, element E) []T {
		return append(acc, mapper(element)...)
	})
}

// FlatMapRefSequence behaves like FlatMapSequence but accepts a reference-based
// mapper.
func FlatMapRefSequence[E any, T any](sequence iter.Seq[E], mapper MapRefFunc[E, []T]) []T {
	return FlatMapSequence(sequence, func(element E) []T {
		mapped := mapper(field.ToOptionalOrNilIfEmpty(element))
		if mapped == nil {
			return nil
		}
		return *mapped
	})
}
