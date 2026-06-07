/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReverse(t *testing.T) {
	original := []int{1, 2, 3, 4}
	reversed := Reverse(original)

	assert.Equal(t, []int{4, 3, 2, 1}, reversed)
	assert.Equal(t, []int{1, 2, 3, 4}, original)
}

func TestReverseNil(t *testing.T) {
	assert.Nil(t, Reverse[[]int](nil))
}
