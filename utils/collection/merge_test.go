/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/field"
)

func TestMergeBy(t *testing.T) {
	items := []string{"a", "bb", "c", "dd"}
	merged := MergeBy(items,
		func(item string) int { return len(item) },
		func(left, right string) (string, bool) { return left + "," + right, true },
	)
	assert.ElementsMatch(t, []string{"a,c", "bb,dd"}, merged)
}

func TestMergeByMergesGroups(t *testing.T) {
	items := []int{2, 1, 4, 3}
	merged := MergeBy(items,
		func(item int) int { return item % 2 },
		func(left, right int) (int, bool) { return left + right, true },
	)
	assert.ElementsMatch(t, []int{6, 4}, merged)
}

func TestMergeByEmpty(t *testing.T) {
	merged := MergeBy[int, int](nil,
		func(item int) int { return item },
		func(left, right int) (int, bool) { return left + right, true },
	)
	assert.Empty(t, merged)

	merged = MergeBy([]int{},
		func(item int) int { return item },
		func(left, right int) (int, bool) { return left + right, true },
	)
	assert.Empty(t, merged)
}

func TestMergeByNilFunctionsReturnClone(t *testing.T) {
	items := []int{1, 2}
	assert.Equal(t, items, MergeBy[int, int](items, nil, func(left, right int) (int, bool) { return left + right, true }))
	assert.Equal(t, items, MergeBy[int, int](items, func(item int) int { return item }, nil))
}

func TestMergeByCanDeclineWithinGroup(t *testing.T) {
	items := []string{"a", "c", "e"}
	merged := MergeBy(items,
		func(item string) int { return len(item) },
		func(left, right string) (string, bool) {
			if left+right == "ac" {
				return left + "," + right, true
			}
			return "", false
		},
	)
	assert.ElementsMatch(t, []string{"a,c", "e"}, merged)
}

func TestMergeByStruct(t *testing.T) {
	type token struct {
		Kind string
		Text string
	}

	items := []token{
		{Kind: "id", Text: "foo"},
		{Kind: "op", Text: "+"},
		{Kind: "id", Text: "bar"},
	}

	merged := MergeBy(items,
		func(item token) string { return item.Kind },
		func(left, right token) (token, bool) {
			return token{Kind: left.Kind, Text: left.Text + " " + right.Text}, true
		},
	)
	assert.ElementsMatch(t, []token{{Kind: "id", Text: "foo bar"}, {Kind: "op", Text: "+"}}, merged)
}

func TestMergeByRef(t *testing.T) {
	items := []string{"a", "bb", "c", "dd"}
	merged := MergeByRef(items,
		func(item *string) int {
			return len(field.OptionalString(item, ""))
		},
		func(left, right *string) (*string, bool) {
			merged := field.OptionalString(left, "") + "," + field.OptionalString(right, "")
			return field.ToOptional(merged), true
		},
	)
	assert.ElementsMatch(t, []string{"a,c", "bb,dd"}, merged)
}
