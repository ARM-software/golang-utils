/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package collection

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlice(t *testing.T) {
	assert.Equal(t, []int{2, 3}, Slice([]int{1, 2, 3, 4}, 1, 3))
	assert.Equal(t, []int{3, 4}, Slice([]int{1, 2, 3, 4}, -2, 4))
	assert.Equal(t, []int{2, 3}, Slice([]int{1, 2, 3, 4}, 1, -1))
	assert.Equal(t, []int{}, Slice([]int{1, 2, 3, 4}, 5, 8))
	assert.Equal(t, []int{}, Slice([]int{1, 2, 3, 4}, 3, 1))
}

func TestSliceSequence(t *testing.T) {
	assert.Equal(t, []int{2, 3}, slices.Collect(SliceSequence(slices.Values([]int{1, 2, 3, 4}), 1, -1)))
}

func TestTakeDrop(t *testing.T) {
	assert.Equal(t, []int{1, 2}, Take([]int{1, 2, 3}, 2))
	assert.Equal(t, []int{}, Take([]int{1, 2, 3}, 0))
	assert.Equal(t, []int{3}, Drop([]int{1, 2, 3}, 2))
	assert.Equal(t, []int{}, Drop([]int{1, 2, 3}, 5))
}

func TestTakeDropSequence(t *testing.T) {
	assert.Equal(t, []int{1, 2}, slices.Collect(TakeSequence(slices.Values([]int{1, 2, 3}), 2)))
	assert.Equal(t, []int{3}, slices.Collect(DropSequence(slices.Values([]int{1, 2, 3}), 2)))
}

func TestTakeWhileDropWhile(t *testing.T) {
	assert.Equal(t, []int{1, 2}, TakeWhile([]int{1, 2, 3, 1}, func(v int) bool { return v < 3 }))
	assert.Equal(t, []int{3, 1}, DropWhile([]int{1, 2, 3, 1}, func(v int) bool { return v < 3 }))
}

func TestPopAt(t *testing.T) {
	value, remaining, ok := PopAt([]int{1, 2, 3}, -1)
	assert.True(t, ok)
	assert.Equal(t, 3, value)
	assert.Equal(t, []int{1, 2}, remaining)

	_, remaining, ok = PopAt([]int{1, 2, 3}, 5)
	assert.False(t, ok)
	assert.Equal(t, []int{1, 2, 3}, remaining)
}
