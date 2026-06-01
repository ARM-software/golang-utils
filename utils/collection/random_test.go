package collection

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandom(t *testing.T) {
	value, found := Random([]int(nil))
	assert.False(t, found)
	assert.Zero(t, value)

	items := []string{faker.Word(), faker.Name(), faker.Word()}
	randomItem, foundString := Random(items)
	require.True(t, foundString)
	assert.Contains(t, items, randomItem)
}

func TestShuffle(t *testing.T) {
	items := []string{faker.Word(), faker.Name(), faker.Word(), faker.Name(), faker.Word()}
	shuffled := Shuffle(items)
	assert.ElementsMatch(t, items, shuffled)
	assert.Equal(t, items, items)

	empty := Shuffle([]int(nil))
	assert.Nil(t, empty)
}
