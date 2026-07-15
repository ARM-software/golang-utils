package casing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInterfaceConventionDetection(t *testing.T) {
	assert.True(t, isInterfacePrefixedAcronym("IHTTP"))
	assert.True(t, isInterfacePrefixedAcronym("IHTTP2"))
	assert.False(t, isInterfacePrefixedAcronym("iHTTP"))
	assert.False(t, isInterfacePrefixedAcronym("HTTP"))
	assert.False(t, isInterfacePrefixedAcronym("IHttp"))

	assert.True(t, hasInterfacePrefixAcronymBoundary([]rune("IHTTP"), 1))
	assert.False(t, hasInterfacePrefixAcronymBoundary([]rune("HTTP"), 1))
	assert.False(t, hasInterfacePrefixAcronymBoundary([]rune("IHttp"), 1))
}

func TestInterfaceConventionNormalisation(t *testing.T) {
	r, err := NewReplacer(
		Rule{Token: "Http", Replacement: "HTTP"},
	)
	require.NoError(t, err)

	value, ok := normaliseInterfacePrefixedAcronym("IHTTP", r, false)
	require.True(t, ok)
	assert.Equal(t, "IHTTP", value)

	value, ok = normaliseInterfacePrefixedAcronym("ihttp", r, true)
	require.True(t, ok)
	assert.Equal(t, "iHTTP", value)

	value, ok = normaliseInterfacePrefixedAcronym("iHttp", r, false)
	require.True(t, ok)
	assert.Equal(t, "IHTTP", value)

	_, ok = normaliseInterfacePrefixedAcronym("HTTP", r, false)
	assert.False(t, ok)
}
