package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	index, found := Find(&[]string{"A", "b", "c"}, "D")
	assert.False(t, found)
	assert.Equal(t, -1, index)

	index, found = Find(&[]string{"A", "B", "b", "c"}, "b")
	assert.True(t, found)
	assert.Equal(t, 2, index)
}

func TestAny(t *testing.T) {
	assert.True(t, Any([]bool{false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false}))
	assert.False(t, Any([]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}))
	assert.True(t, Any([]bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true}))
	assert.True(t, Any([]bool{true, true, true, true, true, true, true, true, true, false, true, true, true, true, true, true, true, true, true, true}))
}

func TestAll(t *testing.T) {
	assert.False(t, All([]bool{false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false}))
	assert.False(t, All([]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}))
	assert.True(t, All([]bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true}))
	assert.False(t, All([]bool{true, true, true, true, true, true, true, true, true, false, true, true, true, true, true, true, true, true, true, true}))
}
