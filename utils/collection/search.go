/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package collection

import (
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

// Any returns true if there is at least one element of the slice which is true.
func Any(slice []bool) bool {
	if len(slice) == 0 {
		return false
	}
	for i := range slice {
		if slice[i] {
			return true
		}
	}
	return false
}

// AnyTrue returns whether there is a value set to true
func AnyTrue(values ...bool) bool {
	return Any(values)
}

// AnyFunc returns whether there is at least one element of slice s for which f() returns true.
func AnyFunc[S ~[]E, E any](s S, f func(E) bool) bool {
	all := make([]bool, 0, len(s))
	for i := range s {
		all = append(all, f(s[i]))

	}
	return Any(all)
}

// AnyEmpty returns whether there is one entry in the slice which is empty.
// If strict, then whitespaces are considered as empty strings
func AnyEmpty(strict bool, slice []string) bool {
	_, found := FindInSlice(!strict, slice, "")
	return found
}

// All returns true if all items of the slice are true.
func All(slice []bool) bool {
	if len(slice) == 0 {
		return false
	}
	for i := range slice {
		if !slice[i] {
			return false
		}
	}
	return true
}

// AllTrue returns whether all values are true.
func AllTrue(values ...bool) bool {
	return All(values)
}

// AllFunc returns whether f returns true for all the elements of slice s.
func AllFunc[S ~[]E, E any](s S, f func(E) bool) bool {
	all := make([]bool, 0, len(s))
	for i := range s {
		all = append(all, f(s[i]))

	}
	return All(all)
}

// AllNotEmpty returns whether all elements of the slice are not empty.
// If strict, then whitespaces are considered as empty strings
func AllNotEmpty(strict bool, slice []string) bool {
	_, found := FindInSlice(!strict, slice, "")
	return !found
}
