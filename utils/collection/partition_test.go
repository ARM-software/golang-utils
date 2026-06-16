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

func TestPartition(t *testing.T) {
	matched, unmatched := Partition([]int{1, 2, 3, 4}, func(v int) bool { return v%2 == 0 })
	assert.Equal(t, []int{2, 4}, matched)
	assert.Equal(t, []int{1, 3}, unmatched)
}

func TestPartitionSequence(t *testing.T) {
	matched, unmatched := PartitionSequence[[]int](slices.Values([]int{1, 2, 3, 4}), func(v int) bool { return v > 2 })
	assert.Equal(t, []int{3, 4}, matched)
	assert.Equal(t, []int{1, 2}, unmatched)
}

func TestPartitionNilSequence(t *testing.T) {
	var seq iter.Seq[int]
	matched, unmatched := PartitionSequence[[]int](seq, func(v int) bool { return v > 0 })
	assert.Empty(t, matched)
	assert.Empty(t, unmatched)
}
