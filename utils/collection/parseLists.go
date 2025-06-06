/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package collection

import (
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func lineIsOnlyWhitespace(line string) bool {
	for _, c := range line {
		if !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

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

// ParseListWithCleanupKeepBlankLines splits a string into a list like strings.Split but also removes any whitespace surrounding  the different items
// unless the entire item is whitespace in which case it is converted to an empty string. For example,
// ParseListWithCleanupKeepBlankLines("a, b ,  c", ",") returns []{"a","b","c"}
// ParseListWithCleanupKeepBlankLines("a, b ,    , c", ",") returns []{"a","b", "", "c"}
func ParseListWithCleanupKeepBlankLines(input string, sep string) (newS []string) {
	return parseListWithCleanup(input, sep, true)
}

// ParseCommaSeparatedList returns the list of string separated by a comma
func ParseCommaSeparatedList(input string) []string {
	return ParseListWithCleanup(input, ",")
}

// ParseCommaSeparatedListToMap returns a map of key value pairs from a string containing a comma separated list
func ParseCommaSeparatedListToMap(input string) (pairs map[string]string, err error) {
	pairs, err = ConvertSliceToMap[string](ParseCommaSeparatedList(input))
	return
}

// ConvertSliceToMap converts a slice of elements into a map e.g. [key1, value1, key2, values2] -> {key1: value1, key2: value2}
func ConvertSliceToMap[T comparable](input []T) (pairs map[T]T, err error) {
	if len(input) == 0 {
		return
	}
	numElements := len(input)

	if numElements%2 != 0 {
		err = commonerrors.Newf(commonerrors.ErrInvalid, "could not convert the list into a map as it does not have an even number of elements")
		return
	}

	pairs = make(map[T]T, numElements/2)
	for pair := range slices.Chunk(input, 2) {
		pairs[pair[0]] = pair[1]
	}
	return
}

// ParseCommaSeparatedListOfPairsToMap returns a map of key value pairs from a string containing a comma separated list of pairs using pairSeparator to separate between keys and values e.g. key1=value1,key2=value2
func ParseCommaSeparatedListOfPairsToMap(input, pairSeparator string) (pairs map[string]string, err error) {
	if pairSeparator == "," {
		pairs, err = ParseCommaSeparatedListToMap(input)
		return
	}
	pairs, err = ConvertListOfPairsToMap(ParseCommaSeparatedList(input), pairSeparator)
	return
}

func ConvertListOfPairsToMap(input []string, pairSeparator string) (pairs map[string]string, err error) {
	if len(input) == 0 {
		return
	}
	pairs = make(map[string]string, len(input))
	for i := range input {
		pair := ParseListWithCleanup(input[i], pairSeparator)
		switch len(pair) {
		case 0:
			continue
		case 2:
			pairs[pair[0]] = pair[1]
		default:
			err = commonerrors.Newf(commonerrors.ErrInvalid, "could not parse key value pair '%v'", input[i])
			return
		}
	}
	return
}

// ConvertSliceToCommaSeparatedList converts a slice into a string containing a coma separated list
func ConvertSliceToCommaSeparatedList[T any](slice []T) string {
	if len(slice) == 0 {
		return ""
	}
	sliceOfStrings := make([]string, 0, len(slice))
	for i := range slice {
		sliceOfStrings = append(sliceOfStrings, fmt.Sprintf("%v", slice[i]))
	}

	return strings.Join(sliceOfStrings, ",")
}

// ConvertMapToSlice converts a map to list of keys and values listed sequentially e.g. [key1, value1, key2, value2]
func ConvertMapToSlice[K comparable, V any](pairs map[K]V) []string {
	if len(pairs) == 0 {
		return nil
	}
	slice := make([]string, 0, len(pairs)*2)
	for key, value := range pairs {
		slice = append(slice, fmt.Sprintf("%v", key), fmt.Sprintf("%v", value))
	}
	return slice
}

// ConvertMapToPairSlice converts a map to list of key value pairs e.g. ["key1=value1", "key2=value2"]
func ConvertMapToPairSlice[K comparable, V any](pairs map[K]V, pairSeparator string) []string {
	if len(pairs) == 0 {
		return nil
	}
	slice := make([]string, 0, len(pairs)*2)
	for key, value := range pairs {
		slice = append(slice, fmt.Sprintf("%v%v%v", key, pairSeparator, value))
	}
	return slice
}

// ConvertMapToCommaSeparatedList converts a map to a string of comma separated list of keys and values defined sequentially
func ConvertMapToCommaSeparatedList[K comparable, V any](pairs map[K]V) string {
	return ConvertSliceToCommaSeparatedList[string](ConvertMapToSlice[K, V](pairs))
}

// ConvertMapToCommaSeparatedPairsList converts a map to a string of comma separated list of key, value pairs.
func ConvertMapToCommaSeparatedPairsList[K comparable, V any](pairs map[K]V, pairSeparator string) string {
	return ConvertSliceToCommaSeparatedList[string](ConvertMapToPairSlice[K, V](pairs, pairSeparator))
}
