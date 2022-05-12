/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package collection

import (
	"strings"
	"unicode"
)

func lineIsOnlyWhitespace(line string) bool {
	for _, c := range line {
		if !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

// ParseListWithCleanup splits a string into a list like strings.Split but also removes any whitespace surrounding  the different items
// for example,
// ParseListWithCleanup("a, b ,  c", ",") returns []{"a","b","c"}
func parseListWithCleanup(input string, sep string, keepBlankLines bool) (newS []string) {
	if len(input) == 0 {
		newS = []string{} // initialisation of empty arrays in function returns []string(nil) instead of []string{}
		return
	}
	split := strings.Split(input, sep)
	for _, s := range split {
		tempString := strings.TrimSpace(s)
		if tempString != "" || (keepBlankLines && lineIsOnlyWhitespace(s)) {
			newS = append(newS, tempString)
		}
	}
	return
}

// ParseListWithCleanup splits a string into a list like strings.Split but also removes any whitespace surrounding  the different items
// for example,
// ParseListWithCleanup("a, b ,  c", ",") returns []{"a","b","c"}
func ParseListWithCleanup(input string, sep string) (newS []string) {
	return parseListWithCleanup(input, sep, false)
}

// ParseListWithCleanup splits a string into a list like strings.Split but also removes any whitespace surrounding  the different items
// unless the entire item is whitespace in which case it is converted to an empty string. For example,
// ParseListWithCleanup("a, b ,  c", ",") returns []{"a","b","c"}
// ParseListWithCleanup("a, b ,    , c", ",") returns []{"a","b", "", "c"}
func ParseListWithCleanupKeepBlankLines(input string, sep string) (newS []string) {
	return parseListWithCleanup(input, sep, true)
}

// ParseCommaSeparatedList returns the list of string separated by a comma
func ParseCommaSeparatedList(input string) []string {
	return ParseListWithCleanup(input, ",")
}
