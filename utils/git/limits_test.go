package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLimits(t *testing.T) {
	require.NoError(t, NoLimits().Validate())
	require.NoError(t, DefaultLimits().Validate())
	assert.True(t, DefaultLimits().Apply())
	assert.False(t, NoLimits().Apply())
}
