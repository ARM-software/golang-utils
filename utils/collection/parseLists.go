/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package collection

import "strings"

// ParseListWithCleanup splits a string into a list like strings.Split but also removes any whitespace surrounding  the different items
// for example,
// ParseListWithCleanup("a, b ,  c", ",") returns []{"a","b","c"}
func ParseListWithCleanup(input string, sep string) (newS []string) {
	if len(input) == 0 {
		newS = []string{} // initialisation of empty arrays in function returns []string(nil) instead of []string{}
		return
	}
	split := strings.Split(input, sep)
	for _, s := range split {
		tempString := strings.TrimSpace(s)
		if tempString != "" {
			newS = append(newS, tempString)
		}
	}
	return
}

// ParseCommaSeparatedList returns the list of string separated by a comma
func ParseCommaSeparatedList(input string) []string {
	return ParseListWithCleanup(input, ",")
}
