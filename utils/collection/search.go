/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package collection

import (
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

// Find looks for an element in a slice. If found it will
// return its index and true; otherwise it will return -1 and false.
func Find(slice *[]string, val string) (int, bool) {
	if slice == nil {
		return -1, false
	}
	return FindInSlice(true, *slice, val)
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

// AnyFunc returns whether there is at least one element of slice s for which f() returns true.
func AnyFunc[S ~[]E, E any](s S, f func(E) bool) bool {
	conditions := NewConditions(len(s))
	for i := range s {
		conditions.Add(f(s[i]))
	}
	return conditions.Any()
}

type FilterFunc[E any] func(E) bool

// Filter returns a new slice that contains elements from the input slice which return true when they’re passed as a parameter to the provided filtering function f.
func Filter[S ~[]E, E any](s S, f FilterFunc[E]) (result S) {
	result = make(S, 0, len(s))

	for i := range s {
		if f(s[i]) {
			result = append(result, s[i])
		}
	}

	return result
}

type MapFunc[T1, T2 any] func(T1) T2

func IdentityMapFunc[T any]() MapFunc[T, T] {
	return func(i T) T {
		return i
	}
}

// Map creates a new slice and populates it with the results of calling the provided function on every element in input slice.
func Map[T1 any, T2 any](s []T1, f MapFunc[T1, T2]) (result []T2) {
	result = make([]T2, len(s))

	for i := range s {
		result[i] = f(s[i])
	}

	return result
}

// Reject is the opposite of Filter and returns the elements of collection for which the filtering function f returns false.
// This is functionally equivalent to slices.DeleteFunc but it returns a new slice.
func Reject[S ~[]E, E any](s S, f FilterFunc[E]) S {
	return Filter(s, func(e E) bool { return !f(e) })
}

type ReduceFunc[T1, T2 any] func(T2, T1) T2

// Reduce runs a reducer function f over all elements in the array, in ascending-index order, and accumulates them into a single value.
func Reduce[T1, T2 any](s []T1, accumulator T2, f ReduceFunc[T1, T2]) (result T2) {
	result = accumulator
	for i := range s {
		result = f(result, s[i])
	}
	return
}

// AnyEmpty returns whether there is one entry in the slice which is empty.
// If strict, then whitespaces are considered as empty strings
func AnyEmpty(strict bool, slice []string) bool {
	_, found := FindInSlice(!strict, slice, "")
	return found
}

// AllFunc returns whether f returns true for all the elements of slice s.
func AllFunc[S ~[]E, E any](s S, f func(E) bool) bool {
	conditions := NewConditions(len(s))
	for i := range s {
		conditions.Add(f(s[i]))
	}
	return conditions.All()
}

// AllNotEmpty returns whether all elements of the slice are not empty.
// If strict, then whitespaces are considered as empty strings
func AllNotEmpty(strict bool, slice []string) bool {
	_, found := FindInSlice(!strict, slice, "")
	return !found
}
