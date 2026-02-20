/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package collection provides various utilities working on slices or sequences
package collection

import (
	"iter"
	"slices"
	"strings"

	"github.com/ARM-software/golang-utils/utils/field"
	mapset "github.com/deckarep/golang-set/v2"
	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/safecast"
)

//
// Find utilities
//

// Find searches for val in the slice pointed to by slice.
// It returns the index of the first match and true when found.
// If slice is nil or val is not present, it returns -1 and false.
func Find(slice *[]string, val string) (int, bool) {
	if slice == nil {
		return -1, false
	}
	return FindInSlice(true, *slice, val)
}

// FindInSequence searches elements (a sequence) for the first item that
// satisfies predicate. It returns the zero-based index of the matching
// element and true when a match is found. If elements is nil or no
// match exists, it returns -1 and false.
func FindInSequence[E any](elements iter.Seq[E], predicate Predicate[E]) (int, bool) {
	if elements == nil {
		return -1, false
	}
	idx := atomic.NewUint64(0)
	for e := range elements {
		if predicate(e) {
			return safecast.ToInt(idx.Load()), true
		}
		idx.Inc()
	}
	return -1, false
}

// FindInSequenceRef behaves like FindInSequence but accepts a predicate
// that operates on element references.
func FindInSequenceRef[E any](elements iter.Seq[E], predicate PredicateRef[E]) (int, bool) {
	return FindInSequence(elements, toPredicateFunc(predicate))
}

// FindInSlice checks whether any of the provided val arguments exist in
// slice. It returns the index of the first match and true if found.
// When strict is true, matching uses exact string equality. When strict
// is false, matching is case-insensitive and ignores surrounding
// whitespace.
//
// If no values are provided or the slice is empty, the function returns
// -1 and false.
func FindInSlice(strict bool, slice []string, val ...string) (int, bool) {
	if len(val) == 0 || len(slice) == 0 {
		return -1, false
	}
	if strict && len(val) == 1 {
		idx := slices.Index(slice, val[0])
		return idx, idx >= 0
	}

	inSlice := make(map[string]int, len(slice))
	for i := range slice {
		item := slice[i]
		if !strict {
			item = strings.ToLower(strings.TrimSpace(item))
		}
		if _, ok := inSlice[item]; !ok {
			inSlice[item] = i
		}
	}

	for i := range val {
		item := val[i]
		if !strict {
			item = strings.ToLower(strings.TrimSpace(item))
		}
		if idx, ok := inSlice[item]; ok {
			return idx, true
		}
	}
	return -1, false
}

//
// Set operations
//

// UniqueEntries returns a slice containing the distinct values from the
// provided slice. The order of elements is not guaranteed.
func UniqueEntries[T comparable](slice []T) []T {
	subSet := mapset.NewSet[T]()
	_ = subSet.Append(slice...)
	return subSet.ToSlice()
}

// Unique returns the distinct values from the provided sequence.
// The order of elements is not guaranteed.
func Unique[T comparable](s iter.Seq[T]) []T {
	return UniqueEntries(slices.Collect(s))
}

// Union returns the union of slice1 and slice2, containing only unique
// values. The order of elements is not guaranteed.
func Union[T comparable](slice1, slice2 []T) []T {
	subSet := mapset.NewSet[T]()
	_ = subSet.Append(slice1...)
	_ = subSet.Append(slice2...)
	return subSet.ToSlice()
}

// Intersection returns the distinct values common to slice1 and slice2.
// The order of elements is not guaranteed.
func Intersection[T comparable](slice1, slice2 []T) []T {
	subSet1 := mapset.NewSet[T]()
	subSet2 := mapset.NewSet[T]()
	_ = subSet1.Append(slice1...)
	_ = subSet2.Append(slice2...)
	return subSet1.Intersect(subSet2).ToSlice()
}

// Difference returns distinct values present in slice1 but not in slice2.
func Difference[T comparable](slice1, slice2 []T) []T {
	subSet1 := mapset.NewSet[T]()
	subSet2 := mapset.NewSet[T]()
	_ = subSet1.Append(slice1...)
	_ = subSet2.Append(slice2...)
	return subSet1.Difference(subSet2).ToSlice()
}

// SymmetricDifference returns distinct values that are present in either
// slice1 or slice2 but not in both.
func SymmetricDifference[T comparable](slice1, slice2 []T) []T {
	subSet1 := mapset.NewSet[T]()
	subSet2 := mapset.NewSet[T]()
	_ = subSet1.Append(slice1...)
	_ = subSet2.Append(slice2...)
	return subSet1.SymmetricDifference(subSet2).ToSlice()
}

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
// Iteration utilities
//

// OperationFunc defines an operation on a value that may return an error.
type OperationFunc[E any] func(E) error

// OperationRefFunc defines an operation on a pointer to a value that may return an error.
type OperationRefFunc[E any] func(*E) error

// OperationWithoutErrorFunc defines an operation on a value that does not return an error.
type OperationWithoutErrorFunc[E any] func(E)

// OperationWithoutErrorRefFunc defines an operation on a pointer that does not return an error.
type OperationWithoutErrorRefFunc[E any] func(*E)

// toOperationFunc adapts an OperationRefFunc to an OperationFunc by
// converting the value to an optional reference.
func toOperationFunc[E any](f OperationRefFunc[E]) OperationFunc[E] {
	return func(e E) error {
		return f(field.ToOptional(e))
	}
}

// toOperationWithoutErrorFunc adapts an OperationWithoutErrorRefFunc to
// an OperationWithoutErrorFunc by converting the value to an optional reference.
func toOperationWithoutErrorFunc[E any](f OperationWithoutErrorRefFunc[E]) OperationWithoutErrorFunc[E] {
	return func(e E) {
		f(field.ToOptional(e))
	}
}

// convertOperationWithoutError wraps a non-error operation so it conforms
// to OperationFunc by always returning nil.
func convertOperationWithoutError[E any](f OperationWithoutErrorFunc[E]) OperationFunc[E] {
	return func(e E) error {
		f(e)
		return nil
	}
}

// Each iterates over a sequence and invokes f for each element. If f
// returns a non-EOF error, iteration stops and that error is returned.
// If f returns EOF, the EOF is ignored and iteration ends without error.
func Each[T any](s iter.Seq[T], f OperationFunc[T]) error {
	for e := range s {
		err := f(e)
		if err != nil {
			err = commonerrors.Ignore(err, commonerrors.ErrEOF)
			return err
		}
	}
	return nil
}

// EachRef behaves like Each but invokes f with a reference to each element.
func EachRef[T any](s iter.Seq[T], f OperationRefFunc[T]) error {
	return Each(s, toOperationFunc(f))
}

// ForEach invokes f on every element of the provided slice. Any error
// returned by f is ignored.
func ForEach[S ~[]E, E any](s S, f OperationWithoutErrorFunc[E]) {
	_ = Each[E](slices.Values(s), convertOperationWithoutError(f))
}

// ForEachValues invokes f for each value passed in values.
func ForEachValues[E any](f func(E), values ...E) {
	ForEach(values, f)
}

// ForEachRef invokes f on every element of the provided slice, passing a reference.
func ForEachRef[S ~[]E, E any](s S, f OperationWithoutErrorRefFunc[E]) {
	ForEach(s, toOperationWithoutErrorFunc(f))
}

// ForAll invokes f on every element of the provided slice. Any non-EOF
// errors returned by f are collected and aggregated into a single error
// returned to the caller. If f returns EOF, iteration stops immediately.
func ForAll[S ~[]E, E any](s S, f OperationFunc[E]) error {
	return ForAllSequence[E](slices.Values(s), f)
}

// ForAllSequence invokes f for every element of s (a sequence). Non-EOF
// errors are wrapped and aggregated; EOF causes immediate termination.
func ForAllSequence[T any](s iter.Seq[T], f OperationFunc[T]) error {
	var err error
	err = commonerrors.Join(err, Each[T](s, func(e T) error {
		subErr := f(e)
		if commonerrors.Any(subErr, commonerrors.ErrEOF) {
			return subErr
		}
		if subErr != nil {
			err = commonerrors.Join(err,
				commonerrors.Newf(subErr, "error during iteration over value [%v]", e))
		}
		return nil
	}))
	return err
}

// ForAllRef behaves like ForAll but applies f to references of elements.
func ForAllRef[S ~[]E, E any](s S, f OperationRefFunc[E]) error {
	return ForAll(s, toOperationFunc(f))
}

// ForAllSequenceRef behaves like ForAllSequence but adapts a reference-based operation.
func ForAllSequenceRef[T any](s iter.Seq[T], f OperationRefFunc[T]) error {
	return ForAllSequence(s, toOperationFunc(f))
}

//
// Mapping utilities
//

// MapFunc defines a function that maps a value of type T1 to type T2.
type MapFunc[T1, T2 any] func(T1) T2

// MapRefFunc defines a mapping function that accepts a pointer to T1 and
// returns a pointer to T2.
type MapRefFunc[T1, T2 any] func(*T1) *T2

// MapWithErrorFunc defines a mapping function that may return an error.
type MapWithErrorFunc[T1, T2 any] func(T1) (T2, error)

// MapRefWithErrorFunc defines a reference-based mapping function that may return an error.
type MapRefWithErrorFunc[T1, T2 any] func(*T1) (*T2, error)

// IdentityMapFunc returns a mapping function that returns its input unchanged.
func IdentityMapFunc[T any]() MapFunc[T, T] {
	return func(i T) T { return i }
}

// MapSequence maps each element of s using f and returns a sequence of mapped values.
func MapSequence[T1 any, T2 any](s iter.Seq[T1], f MapFunc[T1, T2]) iter.Seq[T2] {
	return MapSequenceWithError(s, func(t1 T1) (T2, error) {
		return f(t1), nil
	})
}

// MapSequenceWithError maps each element of s using f, which may return an error.
// Mapping stops if f returns an error or if the consumer declines the yielded value.
func MapSequenceWithError[T1 any, T2 any](s iter.Seq[T1], f MapWithErrorFunc[T1, T2]) iter.Seq[T2] {
	return func(yield func(T2) bool) {
		for v := range s {
			mapped, err := f(v)
			if err != nil || !yield(mapped) {
				return
			}
		}
	}
}

// MapSequenceRef maps a sequence by applying a reference-based mapper to each element.
func MapSequenceRef[T1 any, T2 any](s iter.Seq[T1], f MapRefFunc[T1, T2]) iter.Seq[T2] {
	return MapSequenceRefWithError(s, func(t1 *T1) (*T2, error) {
		return f(t1), nil
	})
}

// MapSequenceRefWithError maps a sequence using a reference-based mapper which may return an error.
// Mapping stops if the mapper returns an error, returns nil, or if the consumer declines the yielded value.
func MapSequenceRefWithError[T1 any, T2 any](s iter.Seq[T1], f MapRefWithErrorFunc[T1, T2]) iter.Seq[T2] {
	return func(yield func(T2) bool) {
		for v := range s {
			mapped, err := f(field.ToOptionalOrNilIfEmpty(v))
			if err != nil || mapped == nil || !yield(*mapped) {
				return
			}
		}
	}
}

// Map applies f to each element of s and returns a slice with the results.
func Map[T1 any, T2 any](s []T1, f MapFunc[T1, T2]) []T2 {
	return slices.Collect[T2](MapSequence(slices.Values(s), f))
}

// MapWithError applies f to each element of s where f may return an error.
// If an error occurs, processing stops and the error is returned.
func MapWithError[T1 any, T2 any](s []T1, f MapWithErrorFunc[T1, T2]) (result []T2, err error) {
	result = make([]T2, len(s))

	for i := range s {
		var subErr error
		result[i], subErr = f(s[i])
		if subErr != nil {
			err = subErr
			return
		}
	}

	return
}

// MapRef applies a reference-based mapping function to s and returns the mapped slice.
func MapRef[T1 any, T2 any](s []T1, f MapRefFunc[T1, T2]) []T2 {
	return slices.Collect[T2](MapSequenceRef(slices.Values(s), f))
}

// MapRefWithError applies a reference-based mapper that may return an error.
// If the mapper returns nil or an error, processing stops and an error is returned.
func MapRefWithError[T1 any, T2 any](s []T1, f MapRefWithErrorFunc[T1, T2]) (result []T2, err error) {
	result = make([]T2, len(s))

	for i := range s {
		var subErr error
		r, subErr := f(field.ToOptionalOrNilIfEmpty(s[i]))
		if subErr != nil {
			err = subErr
			return
		}
		if r == nil {
			err = commonerrors.UndefinedParameterf("item #%v was nil", i)
			return
		}
		result[i] = *r
	}

	return
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

//
// Match utilities
//

// match applies each match function to e and returns a Conditions containing the outcomes.
func match[E any](e E, matches []FilterFunc[E]) *Conditions {
	conditions := NewConditions(len(matches))
	for i := range matches {
		conditions.Add(matches[i](e))
	}
	return &conditions
}

// Match returns true if any of the provided match predicates return true for e.
func Match[E any](e E, matches ...FilterFunc[E]) bool {
	return match[E](e, matches).Any()
}

// MatchAll returns true only if all the provided match predicates return true for e.
func MatchAll[E any](e E, matches ...FilterFunc[E]) bool {
	return match[E](e, matches).All()
}

//
// Any / All helpers
//

// AnyFunc returns true if at least one element in s satisfies f.
func AnyFunc[S ~[]E, E any](s S, f Predicate[E]) bool {
	conditions := NewConditions(len(s))
	for i := range s {
		conditions.Add(f(s[i]))
	}
	return conditions.Any()
}

// AnyRefFunc behaves like AnyFunc but accepts a reference-based predicate.
func AnyRefFunc[S ~[]E, E any](s S, f PredicateRef[E]) bool {
	return AnyFunc(s, toPredicateFunc(f))
}

// AllFunc returns true if f returns true for every element in s.
func AllFunc[S ~[]E, E any](s S, f Predicate[E]) bool {
	return AllTrueSequence(slices.Values(s), f)
}

// AllTrueSequence returns true if f returns true for every element in the sequence.
func AllTrueSequence[E any](s iter.Seq[E], f Predicate[E]) bool {
	return AllSequence(MapSequence[E, bool](s, MapFunc[E, bool](f)))
}

//
// Empty checks
//

// AnyEmpty returns true if any entry of slice is empty. If strict is true,
// strings containing only whitespace are considered empty.
func AnyEmpty(strict bool, slice []string) bool {
	_, found := FindInSlice(!strict, slice, "")
	return found
}

// AllNotEmpty returns true when every entry of slice is non-empty. If strict is true,
// strings containing only whitespace are considered empty.
func AllNotEmpty(strict bool, slice []string) bool {
	_, found := FindInSlice(!strict, slice, "")
	return !found
}
