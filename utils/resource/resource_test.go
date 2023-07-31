package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CloseableResource(t *testing.T) {
	desc := "test nil resource"
	testResource := NewCloseableResource(nil, desc)
	assert.Contains(t, testResource.String(), desc)
	assert.False(t, testResource.IsClosed())
	require.NoError(t, testResource.Close())
	assert.True(t, testResource.IsClosed())
}

func Test_NonCloseableResource(t *testing.T) {
	testResource := NewNonCloseableResource()
	assert.NotEmpty(t, testResource.String())
	assert.False(t, testResource.IsClosed())
	require.NoError(t, testResource.Close())
	assert.False(t, testResource.IsClosed())
}
