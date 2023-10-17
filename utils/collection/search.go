/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package collection

import "strings"

// Find looks for an element in a slice. If found it will
// return its index and true; otherwise it will return -1 and false.
func Find(slice *[]string, val string) (int, bool) {
	if slice == nil {
		return -1, false
	}
	return FindInSlice(true, *slice, val)
}

// FindInSlice finds if any values val are present in the slice and if so returns the first index.
// if strict, it check of an exact match; otherwise discard whitespaces and case.
func FindInSlice(strict bool, slice []string, val ...string) (int, bool) {
	if len(val) == 0 || len(slice) == 0 {
		return -1, false
	}
	if len(slice) > len(val) {
		for i := range slice {
			item := slice[i]
			if !strict {
				item = strings.TrimSpace(item)
			}
			for j := range val {
				if !strict && strings.EqualFold(item, val[j]) {
					return i, true
				} else if strict && item == val[j] {
					return i, true
				}
			}
		}
	} else {
		for j := range val {
			for i := range slice {
				item := slice[i]
				if !strict {
					item = strings.TrimSpace(item)
				}
				if !strict && strings.EqualFold(item, val[j]) {
					return i, true
				} else if strict && item == val[j] {
					return i, true
				}
			}
		}
	}

	return -1, false
}

// Any returns true if there is at least one element of the slice which is true.
func Any(slice []bool) bool {
	for i := range slice {
		if slice[i] {
			return true
		}
	}
	return false
}

// AnyEmpty returns whether there is one entry in the slice which is empty.
// If strict, then whitespaces are considered as empty strings
func AnyEmpty(strict bool, slice []string) bool {
	_, found := FindInSlice(!strict, slice, "")
	return found
}

// All returns true if all items of the slice are true.
func All(slice []bool) bool {
	for i := range slice {
		if !slice[i] {
			return false
		}
	}
	return true
}

// AllNotEmpty returns whether all elements of the slice are not empty.
// If strict, then whitespaces are considered as empty strings
func AllNotEmpty(strict bool, slice []string) bool {
	_, found := FindInSlice(!strict, slice, "")
	return !found
}
