package parallelisation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Given a CancelFunctionsStore
// Functions can be registered
// and all functions will be called
func TestCancelFunctionStore(t *testing.T) {
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
}
