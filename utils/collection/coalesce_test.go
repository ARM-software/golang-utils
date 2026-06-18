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

func TestCoalesceEmptySlice(t *testing.T) {
	assert.Empty(t, Coalesce[int](nil, func(left, right int) (int, bool) { return 0, false }))
	assert.Empty(t, Coalesce[int]([]int{}, func(left, right int) (int, bool) { return 0, false }))
}

func TestCoalesceSingleElement(t *testing.T) {
	assert.Equal(t, []int{1}, Coalesce([]int{1}, func(left, right int) (int, bool) { return 0, false }))
}

func TestCoalesceNoMerges(t *testing.T) {
	items := []int{1, 2, 3}
	assert.Equal(t, items, Coalesce(items, func(left, right int) (int, bool) { return 0, false }))
}

func TestCoalesceRepeatedAdjacentMerges(t *testing.T) {
	items := []int{1, 1, 1, 2}
	result := Coalesce(items, func(left, right int) (int, bool) {
		if left == right {
			return left + right, true
		}
		return 0, false
	})
	assert.Equal(t, []int{2, 1, 2}, result)
}

func TestCoalesceNonAdjacentValuesAreNotMerged(t *testing.T) {
	items := []int{1, 2, 1}
	result := Coalesce(items, func(left, right int) (int, bool) {
		if left == right {
			return left + right, true
		}
		return 0, false
	})
	assert.Equal(t, []int{1, 2, 1}, result)
}

func TestCoalesceMergedResultCanMergeAgain(t *testing.T) {
	items := []int{1, 1, 2}
	result := Coalesce(items, func(left, right int) (int, bool) {
		if left+right <= 4 {
			return left + right, true
		}
		return 0, false
	})
	assert.Equal(t, []int{4}, result)
}

func TestCoalesceStructExample(t *testing.T) {
	type token struct {
		Kind string
		Text string
	}

	items := []token{
		{Kind: "id", Text: "foo"},
		{Kind: "id", Text: "bar"},
		{Kind: "op", Text: "+"},
	}

	result := Coalesce(items, func(left, right token) (token, bool) {
		if left.Kind != right.Kind {
			return token{}, false
		}
		return token{Kind: left.Kind, Text: left.Text + " " + right.Text}, true
	})

	assert.Equal(t, []token{{Kind: "id", Text: "foo bar"}, {Kind: "op", Text: "+"}}, result)
}

func TestCoalesceDoesNotModifyInput(t *testing.T) {
	items := []int{1, 1, 2}
	original := append([]int(nil), items...)
	_ = Coalesce(items, func(left, right int) (int, bool) {
		if left == right {
			return left + right, true
		}
		return 0, false
	})
	assert.Equal(t, original, items)
}

func TestCoalesceRef(t *testing.T) {
	items := []int{1, 1, 2}
	result := CoalesceRef(items, func(left, right *int) (*int, bool) {
		if field.OptionalInt(left, 0) != field.OptionalInt(right, 0) {
			return nil, false
		}
		return field.ToOptional(field.OptionalInt(left, 0) + field.OptionalInt(right, 0)), true
	})
	assert.Equal(t, []int{4}, result)
}
