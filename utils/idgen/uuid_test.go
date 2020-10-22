package idgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUuidUniqueness(t *testing.T) {
	uuid1, err := GenerateUuid4()
	require.Nil(t, err)

	uuid2, err := GenerateUuid4()
	require.Nil(t, err)

	assert.NotEqual(t, uuid1, uuid2)
}

func TestUuidLength(t *testing.T) {
	uuid, err := GenerateUuid4()
	require.Nil(t, err)

	assert.Equal(t, 36, len(uuid))
}
