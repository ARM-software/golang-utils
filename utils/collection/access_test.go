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

	_, ok = At([]string{"a", "b", "c"}, -1)
	assert.False(t, ok)

	_, ok = At([]string{"a", "b", "c"}, 3)
	assert.False(t, ok)
}

func TestAtSequence(t *testing.T) {
	value, ok := AtSequence(slices.Values([]int{10, 20, 30}), 2)
	assert.True(t, ok)
	assert.Equal(t, 30, value)

	var seq iter.Seq[int]
	_, ok = AtSequence(seq, 0)
	assert.False(t, ok)
}

func TestNth(t *testing.T) {
	value, ok := Nth([]int{1, 2, 3}, 0)
	assert.True(t, ok)
	assert.Equal(t, 1, value)
}

func TestNthSequence(t *testing.T) {
	value, ok := NthSequence(slices.Values([]int{1, 2, 3}), 1)
	assert.True(t, ok)
	assert.Equal(t, 2, value)
}
