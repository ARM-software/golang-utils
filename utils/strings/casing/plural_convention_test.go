package casing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPluralConventionHelpers(t *testing.T) {
	assert.True(t, hasPluralInitialismSuffix([]rune("URLs"), 2))
	assert.False(t, hasPluralInitialismSuffix([]rune("URL"), 2))
	assert.False(t, hasPluralInitialismSuffix([]rune("urls"), 2))

	assert.True(t, hasLowerPrefixAcronymBoundary([]rune("uRLs"), 1))
	assert.False(t, hasLowerPrefixAcronymBoundary([]rune("URLs"), 1))
	assert.False(t, hasLowerPrefixAcronymBoundary([]rune("userURLs"), 1))

	parts, ok := splitPluralInitialismSuffix("s")
	assert.True(t, ok)
	assert.ElementsMatch(t, []string{"s"}, parts)

	parts, ok = splitPluralInitialismSuffix("tail")
	assert.False(t, ok)
	assert.Nil(t, parts)
}
