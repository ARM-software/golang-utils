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
	"github.com/stretchr/testify/require"
)

func TestEmptySequence(t *testing.T) {
	assert.Empty(t, slices.Collect(EmptySequence[int]()))
}

func TestEmptySequence2(t *testing.T) {
	pairs := make(map[int]string)
	for key, value := range EmptySequence2[int, string]() {
		pairs[key] = value
	}
	assert.Empty(t, pairs)
}

func TestMapSequenceNil(t *testing.T) {
	var seq iter.Seq[int]
	assert.Empty(t, slices.Collect(MapSequence(seq, func(v int) int { return v + 1 })))
}

func TestFilterSequenceNil(t *testing.T) {
	var seq iter.Seq[int]
	assert.Empty(t, slices.Collect(FilterSequence(seq, func(v int) bool { return v%2 == 0 })))
}

func TestEachNilSequence(t *testing.T) {
	var seq iter.Seq[int]
	require.NoError(t, Each(seq, func(v int) error {
		t.Fatalf("unexpected callback for value %d", v)
		return nil
	}))
}

func TestUniqueNilSequence(t *testing.T) {
	var seq iter.Seq[int]
	assert.Empty(t, Unique(seq))
}

func TestSequence2OrEmptyNil(t *testing.T) {
	var seq iter.Seq2[int, string]
	pairs := make(map[int]string)
	for key, value := range Sequence2OrEmpty(seq) {
		pairs[key] = value
	}
	assert.Empty(t, pairs)
}
