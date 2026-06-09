/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package collection

import (
	"iter"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirst(t *testing.T) {
	value, ok := First([]int{1, 2, 3})
	assert.True(t, ok)
	assert.Equal(t, 1, value)

	_, ok = First([]int(nil))
	assert.False(t, ok)
}

func TestFirstBy(t *testing.T) {
	value, ok := FirstBy([]int{1, 2, 3, 4}, func(v int) bool { return v%2 == 0 })
	assert.True(t, ok)
	assert.Equal(t, 2, value)

	_, ok = FirstBy([]int{1, 3, 5}, func(v int) bool { return v%2 == 0 })
	assert.False(t, ok)
}

func TestFirstSequence(t *testing.T) {
	value, ok := FirstSequence(slices.Values([]string{"a", "b"}))
	assert.True(t, ok)
	assert.Equal(t, "a", value)

	var seq iter.Seq[string]
	_, ok = FirstSequence(seq)
	assert.False(t, ok)
}

func TestFirstBySequence(t *testing.T) {
	value, ok := FirstBySequence(slices.Values([]int{1, 2, 3, 4}), func(v int) bool { return v > 2 })
	assert.True(t, ok)
	assert.Equal(t, 3, value)

	var seq iter.Seq[int]
	_, ok = FirstBySequence(seq, func(v int) bool { return v > 0 })
	assert.False(t, ok)
}

func TestLast(t *testing.T) {
	value, ok := Last([]int{1, 2, 3})
	assert.True(t, ok)
	assert.Equal(t, 3, value)

	_, ok = Last([]int(nil))
	assert.False(t, ok)
}

func TestLastBy(t *testing.T) {
	value, ok := LastBy([]int{1, 2, 3, 4}, func(v int) bool { return v%2 == 0 })
	assert.True(t, ok)
	assert.Equal(t, 4, value)

	_, ok = LastBy([]int{1, 3, 5}, func(v int) bool { return v%2 == 0 })
	assert.False(t, ok)
}

func TestLastSequence(t *testing.T) {
	value, ok := LastSequence(slices.Values([]string{"a", "b"}))
	assert.True(t, ok)
	assert.Equal(t, "b", value)

	var seq iter.Seq[string]
	_, ok = LastSequence(seq)
	assert.False(t, ok)
}

func TestLastBySequence(t *testing.T) {
	value, ok := LastBySequence(slices.Values([]int{1, 2, 3, 4}), func(v int) bool { return v < 4 })
	assert.True(t, ok)
	assert.Equal(t, 3, value)

	var seq iter.Seq[int]
	_, ok = LastBySequence(seq, func(v int) bool { return v > 0 })
	assert.False(t, ok)
}
