/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package collection provides various utilities working on slices or sequences
package collection

import (
	"iter"
	"slices"
)

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
