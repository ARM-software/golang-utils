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

func TestAt(t *testing.T) {
	value, ok := At([]string{"a", "b", "c"}, 1)
	assert.True(t, ok)
	assert.Equal(t, "b", value)

	value, ok = At([]string{"a", "b", "c"}, -1)
	assert.True(t, ok)
	assert.Equal(t, "c", value)

	value, ok = At([]string{"a", "b", "c"}, -3)
	assert.True(t, ok)
	assert.Equal(t, "a", value)

	_, ok = At([]string{"a", "b", "c"}, -4)
	assert.False(t, ok)

	_, ok = At([]string{"a", "b", "c"}, 3)
	assert.False(t, ok)
}

func TestAtSequence(t *testing.T) {
	value, ok := AtSequence(slices.Values([]int{10, 20, 30}), 2)
	assert.True(t, ok)
	assert.Equal(t, 30, value)

	value, ok = AtSequence(slices.Values([]int{10, 20, 30}), -1)
	assert.True(t, ok)
	assert.Equal(t, 30, value)

	value, ok = AtSequence(slices.Values([]int{10, 20, 30}), -3)
	assert.True(t, ok)
	assert.Equal(t, 10, value)

	_, ok = AtSequence(slices.Values([]int{10, 20, 30}), -4)
	assert.False(t, ok)

	var seq iter.Seq[int]
	_, ok = AtSequence(seq, 0)
	assert.False(t, ok)
}

func TestNth(t *testing.T) {
	value, ok := Nth([]int{1, 2, 3}, 0)
	assert.True(t, ok)
	assert.Equal(t, 1, value)

	value, ok = Nth([]int{1, 2, 3}, -1)
	assert.True(t, ok)
	assert.Equal(t, 3, value)
}

func TestNthSequence(t *testing.T) {
	value, ok := NthSequence(slices.Values([]int{1, 2, 3}), 1)
	assert.True(t, ok)
	assert.Equal(t, 2, value)

	value, ok = NthSequence(slices.Values([]int{1, 2, 3}), -2)
	assert.True(t, ok)
	assert.Equal(t, 2, value)
}
