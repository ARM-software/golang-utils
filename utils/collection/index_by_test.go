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

func TestIndexBy(t *testing.T) {
	indexed := IndexBy([]string{"a", "bb", "ccc"}, func(v string) int { return len(v) })
	assert.Equal(t, map[int]string{1: "a", 2: "bb", 3: "ccc"}, indexed)
}

func TestIndexByLastWins(t *testing.T) {
	indexed := IndexBy([]string{"a", "b"}, func(string) int { return 1 })
	assert.Equal(t, map[int]string{1: "b"}, indexed)
}

func TestIndexBySequenceNil(t *testing.T) {
	var seq iter.Seq[string]
	indexed := IndexBySequence(seq, func(v string) int { return len(v) })
	assert.Empty(t, indexed)
}

func TestIndexBySequence(t *testing.T) {
	indexed := IndexBySequence(slices.Values([]string{"a", "bb"}), func(v string) int { return len(v) })
	assert.Equal(t, map[int]string{1: "a", 2: "bb"}, indexed)
}
