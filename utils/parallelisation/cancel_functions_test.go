/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package parallelisation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

// Given a CancelFunctionsStore
// Functions can be registered
// and all functions will be called
func TestCancelFunctionStore(t *testing.T) {
	t.Run("valid cancel store", func(t *testing.T) {
		// Set up some fake CancelFuncs to make sure they are called
		called1 := false
		called2 := false
		cancelFunc1 := func() {
			called1 = true
		}
		cancelFunc2 := func() {
			called2 = true
		}

		store := NewCancelFunctionsStore()

		store.RegisterCancelFunction(cancelFunc1, cancelFunc2)

		assert.Equal(t, 2, store.Len())

		store.Cancel()

		assert.True(t, called1)
		assert.True(t, called2)
	})

	t.Run("incorrectly initialised cancel store", func(t *testing.T) {
		called1 := false
		called2 := false
		cancelFunc1 := func() {
			called1 = true
		}
		cancelFunc2 := func() {
			called2 = true
		}

		store := CancelFunctionStore{}

		store.RegisterCancelFunction(cancelFunc1, cancelFunc2)

		assert.Equal(t, 2, store.Len())

		err := store.Execute(context.Background())
		errortest.AssertError(t, err, commonerrors.ErrUndefined)

		assert.False(t, called1)
		assert.False(t, called2)
	})
}
