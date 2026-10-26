/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package multiplication

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNumbers(t *testing.T) {
	assert.Equal(t, Kilo, float64(1000))
	assert.Equal(t, Mega, float64(1000000))

}
