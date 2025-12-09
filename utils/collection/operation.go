/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package collection

import (
	"iter"
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/safecast"
)

// Find looks for an element in a slice. If found it will
// return its index and true; otherwise it will return -1 and false.
func Find(slice *[]string, val string) (int, bool) {
	if slice == nil {
		return -1, false
	}
	return FindInSlice(true, *slice, val)
}

// FindInSequence searches a collection for an element satisfying the predicate.
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

// FindInSlice finds if any values val are present in the slice and if so returns the first index.
// if strict, it checks for an exact match; otherwise it discards whitespaces and case.
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

// UniqueEntries returns all the unique values contained in a slice.
func UniqueEntries[T comparable](slice []T) []T {
	subSet := mapset.NewSet[T]()
	_ = subSet.Append(slice...)
	return subSet.ToSlice()
}

// Unique returns all the unique values contained in a sequence.
func Unique[T comparable](s iter.Seq[T]) []T {
	return UniqueEntries(slices.Collect(s))
}

// Union returns the union of two slices (only unique values are returned).
func Union[T comparable](slice1, slice2 []T) []T {
	subSet := mapset.NewSet[T]()
	_ = subSet.Append(slice1...)
	_ = subSet.Append(slice2...)
	return subSet.ToSlice()
}

// Intersection returns the intersection of two slices (only unique values are returned).
func Intersection[T comparable](slice1, slice2 []T) []T {
	subSet1 := mapset.NewSet[T]()
	subSet2 := mapset.NewSet[T]()
	_ = subSet1.Append(slice1...)
	_ = subSet2.Append(slice2...)
	return subSet1.Intersect(subSet2).ToSlice()
}

// Difference returns the Difference between slice1 and slice2 (only unique values are returned).
func Difference[T comparable](slice1, slice2 []T) []T {
	subSet1 := mapset.NewSet[T]()
	subSet2 := mapset.NewSet[T]()
	_ = subSet1.Append(slice1...)
	_ = subSet2.Append(slice2...)
	return subSet1.Difference(subSet2).ToSlice()
}

// SymmetricDifference returns the symmetric difference between slice1 and slice2 (only unique values are returned).
func SymmetricDifference[T comparable](slice1, slice2 []T) []T {
	subSet1 := mapset.NewSet[T]()
	subSet2 := mapset.NewSet[T]()
	_ = subSet1.Append(slice1...)
	_ = subSet2.Append(slice2...)
	return subSet1.SymmetricDifference(subSet2).ToSlice()
}

// AnyFunc returns whether there is at least one element of slice s for which f() returns true.
func AnyFunc[S ~[]E, E any](s S, f func(E) bool) bool {
	conditions := NewConditions(len(s))
	for i := range s {
		conditions.Add(f(s[i]))
	}
	return conditions.Any()
}

type FilterFunc[E any] func(E) bool

type Predicate[E any] = FilterFunc[E]

// Filter returns a new slice that contains elements from the input slice which return true when they’re passed as a parameter to the provided filtering function f.
func Filter[S ~[]E, E any](s S, f FilterFunc[E]) S {
	return slices.Collect[E](FilterSequence[E](slices.Values(s), f))
}

// FilterSequence returns a new sequence that contains elements from the input sequence which return true when they’re passed as a parameter to the provided filtering function f.
func FilterSequence[E any](s iter.Seq[E], f Predicate[E]) (result iter.Seq[E]) {
	return func(yield func(E) bool) {
		for v := range s {
			if f(v) {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// ForEachValues iterates over values and executes the passed function on each of them.
func ForEachValues[E any](f func(E), values ...E) {
	ForEach(values, f)
}

// ForEach iterates over elements and executes the passed function on each element.
func ForEach[S ~[]E, E any](s S, f func(E)) {
	_ = Each[E](slices.Values(s), func(e E) error {
		f(e)
		return nil
	})
}

// ForAll iterates over every element in the provided sequence and invokes f
// on each item in order. If f returns an error for one or more elements,
// ForAll continues processing the remaining elements and returns a single
// aggregated error containing all collected errors.  If no errors occur, the returned error is nil.
func ForAll[S ~[]E, E any](s S, f func(E) error) error {
	return ForAllSequence[E](slices.Values(s), f)
}

// ForAllSequence iterates over every element in the provided sequence and invokes f
// on each item in order. If f returns an error for one or more elements,
// ForAllSequence continues processing the remaining elements and returns a single
// aggregated error containing all collected errors.  If no errors occur, the returned error is nil.
func ForAllSequence[T any](s iter.Seq[T], f func(T) error) error {
	var err error
	err = commonerrors.Join(err, Each[T](s, func(e T) error {
		subErr := f(e)
		if commonerrors.Any(subErr, commonerrors.ErrEOF) {
			return subErr
		}
		if subErr != nil {
			err = commonerrors.Join(err, commonerrors.Newf(subErr, "error during iteration over value [%v]", e))
		}
		return nil
	}))
	return err
}

// Each iterates over a sequence and executes the passed function against each element.
// If passed func returns an error, the iteration stops and the error is returned, unless it is EOF.
func Each[T any](s iter.Seq[T], f func(T) error) error {
	for e := range s {
		err := f(e)
		if err != nil {
			err = commonerrors.Ignore(err, commonerrors.ErrEOF)
			return err
		}
	}
	return nil
}

type MapFunc[T1, T2 any] func(T1) T2
type MapWithErrorFunc[T1, T2 any] func(T1) (T2, error)

func IdentityMapFunc[T any]() MapFunc[T, T] {
	return func(i T) T {
		return i
	}
}

// MapSequence creates a new sequences and populates it with the results of calling the provided function on every element of the input sequence.
func MapSequence[T1 any, T2 any](s iter.Seq[T1], f MapFunc[T1, T2]) iter.Seq[T2] {
	return MapSequenceWithError[T1, T2](s, func(t1 T1) (T2, error) {
		return f(t1), nil
	})
}

// MapSequenceWithError creates a new sequences and populates it with the results of calling the provided function on every element of the input sequence. If an error happens, the mapping stops.
func MapSequenceWithError[T1 any, T2 any](s iter.Seq[T1], f MapWithErrorFunc[T1, T2]) iter.Seq[T2] {
	return func(yield func(T2) bool) {
		for v := range s {
			mapped, subErr := f(v)
			if subErr != nil || !yield(mapped) {
				return
			}
		}
	}
}

// Map creates a new slice and populates it with the results of calling the provided function on every element in input slice.
func Map[T1 any, T2 any](s []T1, f MapFunc[T1, T2]) []T2 {
	return slices.Collect[T2](MapSequence[T1, T2](slices.Values(s), f))
}

// MapWithError creates a new slice and populates it with the results of calling the provided function on every element in input slice. If an error happens, the mapping stops and the error returned.
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

// OppositeFunc returns the opposite of a FilterFunc.
func OppositeFunc[E any](f FilterFunc[E]) FilterFunc[E] { return func(e E) bool { return !f(e) } }

// Reject is the opposite of Filter and returns the elements of collection for which the filtering function f returns false.
// This is functionally equivalent to slices.DeleteFunc but it returns a new slice.
func Reject[S ~[]E, E any](s S, f FilterFunc[E]) S {
	return Filter(s, OppositeFunc[E](f))
}

// RejectSequence is the opposite of FilterSequence and returns the elements of collection for which the filtering function f returns false.
func RejectSequence[E any](s iter.Seq[E], f FilterFunc[E]) iter.Seq[E] {
	return FilterSequence(s, OppositeFunc[E](f))
}

// Reduce runs a reducer function f over all elements in the array, in ascending-index order, and accumulates them into a single value.
func Reduce[T1, T2 any](s []T1, accumulator T2, f ReduceFunc[T1, T2]) T2 {
	return ReducesSequence[T1, T2](slices.Values(s), accumulator, f)
}

// ReducesSequence runs a reducer function f over all elements of a sequence, in ascending-index order, and accumulates them into a single value.
func ReducesSequence[T1, T2 any](s iter.Seq[T1], accumulator T2, f ReduceFunc[T1, T2]) (result T2) {
	result = accumulator
	for e := range s {
		result = f(result, e)
	}
	return
}

func match[E any](e E, matches []FilterFunc[E]) *Conditions {
	conditions := NewConditions(len(matches))
	for i := range matches {
		conditions.Add(matches[i](e))
	}
	return &conditions
}

// Match checks whether an element e matches any of the matching functions.
func Match[E any](e E, matches ...FilterFunc[E]) bool {
	return match[E](e, matches).Any()
}

// MatchAll checks whether an element e matches all the matching functions.
func MatchAll[E any](e E, matches ...FilterFunc[E]) bool {
	return match[E](e, matches).All()
}

type ReduceFunc[T1, T2 any] func(T2, T1) T2

// AllFunc returns whether f returns true for all the elements of slice s.
func AllFunc[S ~[]E, E any](s S, f func(E) bool) bool {
	return AllTrueSequence(slices.Values(s), f)
}

// AllTrueSequence returns whether f returns true for all the elements in a sequence.
func AllTrueSequence[E any](s iter.Seq[E], f func(E) bool) bool {
	return AllSequence(MapSequence[E, bool](s, f))
}

// AnyEmpty returns whether there is one entry in the slice which is empty.
// If strict, then whitespaces are considered as empty strings
func AnyEmpty(strict bool, slice []string) bool {
	_, found := FindInSlice(!strict, slice, "")
	return found
}

// AllNotEmpty returns whether all elements of the slice are not empty.
// If strict, then whitespaces are considered as empty strings
func AllNotEmpty(strict bool, slice []string) bool {
	_, found := FindInSlice(!strict, slice, "")
	return !found
}
