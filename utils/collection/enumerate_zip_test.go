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

func TestEnumerate(t *testing.T) {
	values := make([]string, 0)
	indices := make([]int, 0)
	for index, value := range Enumerate([]string{"a", "b"}) {
		indices = append(indices, index)
		values = append(values, value)
	}
	assert.Equal(t, []int{0, 1}, indices)
	assert.Equal(t, []string{"a", "b"}, values)
}

func TestEnumerateSequenceNil(t *testing.T) {
	var seq iter.Seq[int]
	called := false
	for range EnumerateSequence(seq) {
		called = true
	}
	assert.False(t, called)
}

func TestReverseSequence(t *testing.T) {
	values := make([]int, 0)
	for value := range ReverseSequence(slices.Values([]int{1, 2, 3})) {
		values = append(values, value)
	}
	assert.Equal(t, []int{3, 2, 1}, values)
}

func TestZip(t *testing.T) {
	leftValues := make([]int, 0)
	rightValues := make([]string, 0)
	for left, right := range Zip([]int{1, 2, 3}, []string{"a", "b"}) {
		leftValues = append(leftValues, left)
		rightValues = append(rightValues, right)
	}
	assert.Equal(t, []int{1, 2}, leftValues)
	assert.Equal(t, []string{"a", "b"}, rightValues)
}

func TestZipSequenceNil(t *testing.T) {
	var left iter.Seq[int]
	var right iter.Seq[string]
	called := false
	for range ZipSequence(left, right) {
		called = true
	}
	assert.False(t, called)
}
