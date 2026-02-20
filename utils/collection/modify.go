/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package collection

import "slices"

// Remove looks for elements in a slice. If they're found, it will
// remove them.
func Remove(slice []string, val ...string) []string {
	return GenericRemove(func(val1, val2 string) bool { return val1 == val2 }, slice, val...)
}

// GenericRemove looks for elements in a slice using the equal function. If they're found, it will
// remove them from the slice.
func GenericRemove(equal func(string, string) bool, slice []string, val ...string) []string {
	if len(val) == 0 {
		return slice
	}
	if len(val) == 1 {
		return slices.DeleteFunc(slice, func(v string) bool { return equal(v, val[0]) })
	}
	list := make([]string, 0, len(slice))
	found := make([]bool, len(val))

	for i := range slice {
		e := slice[i]
		for j := range val {
			found[j] = equal(e, val[j])
		}
		if !Any(found) {
			list = append(list, e)
		}
	}
	return list
}
