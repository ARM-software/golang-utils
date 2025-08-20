/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package parallelisation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func testCancelStore(t *testing.T, store *CancelFunctionStore) {
	t.Helper()
	require.NotNil(t, store)
	// Set up some fake CancelFuncs to make sure they are called
	called1 := atomic.NewBool(false)
	called2 := atomic.NewBool(false)
	called3 := atomic.NewBool(false)

	cancelFunc1 := func() {
		called1.Store(true)
	}
	cancelFunc2 := func() {
		called2.Store(true)
	}
	cancelFunc3 := func() {
		called3.Store(true)
	}
	subStore := NewCancelFunctionsStore()
	subStore.RegisterCancelFunction(cancelFunc3)

	store.RegisterCancelFunction(cancelFunc1, cancelFunc2)
	store.RegisterCancelStore(subStore)
	store.RegisterCancelStore(nil)

	assert.Equal(t, 3, store.Len())
	assert.False(t, called1.Load())
	assert.False(t, called2.Load())
	assert.False(t, called3.Load())
	store.Cancel()

	assert.True(t, called1.Load())
	assert.True(t, called2.Load())
	assert.True(t, called3.Load())
}

// Given a CancelFunctionsStore
// Functions can be registered
// and all functions will be called
func TestCancelFunctionStore(t *testing.T) {
	t.Run("valid cancel store", func(t *testing.T) {

		t.Run("parallel", func(t *testing.T) {
			testCancelStore(t, NewCancelFunctionsStore())
		})
		t.Run("sequential", func(t *testing.T) {
			testCancelStore(t, NewCancelFunctionsStore(Sequential))
		})
		t.Run("reverse", func(t *testing.T) {
			testCancelStore(t, NewCancelFunctionsStore(SequentialInReverse))
		})
		t.Run("execute all", func(t *testing.T) {
			testCancelStore(t, NewCancelFunctionsStore(StopOnFirstError))
		})
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
