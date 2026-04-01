package collection

import (
	"iter"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func TestInSequence(t *testing.T) {
	values := []string{"alpha", "Beta", "gamma"}

	assert.True(t, InSequence(slices.Values(values), " beta ", StringCleanCaseInsensitiveMatch))
	assert.False(t, InSequence(slices.Values(values), "beta", StringCleanMatch))
	assert.False(t, InSequence(slices.Values(values), "delta", StringCleanCaseInsensitiveMatch))

	var nilSeq iter.Seq[string]
	assert.False(t, InSequence(nilSeq, "alpha", StringMatch))

	errMatch := func(pattern, value string) (bool, error) {
		return true, commonerrors.ErrUnexpected
	}
	assert.False(t, InSequence(slices.Values(values), "alpha", errMatch))
}

func TestInSequenceRef(t *testing.T) {
	values := []int{1, 2, 3}
	target := 2

	assert.True(t, InSequenceRef(slices.Values(values), &target, StrictRefMatch))

	missing := 4
	assert.False(t, InSequenceRef(slices.Values(values), &missing, StrictRefMatch))
	assert.False(t, InSequenceRef(slices.Values(values), &target, func(a, b *int) (bool, error) {
		return true, commonerrors.ErrUnexpected
	}))

	var nilSeq iter.Seq[int]
	assert.False(t, InSequenceRef(nilSeq, &target, StrictRefMatch))
}

func TestIn(t *testing.T) {
	values := []string{"alpha", "beta", "gamma"}

	assert.True(t, In(values, "^be", StringRegexMatch))
	assert.False(t, In(values, "^de", StringRegexMatch))
	assert.True(t, In(values, "BETA", StringCaseInsensitiveMatch))
	assert.True(t, In(values, ".+ta.*", StringRegexMatch))
}

func TestInRef(t *testing.T) {
	values := []string{"alpha", "beta", "gamma"}
	target := "beta"

	assert.True(t, InRef(values, &target, StrictRefMatch))

	missing := "delta"
	assert.False(t, InRef(values, &missing, StrictRefMatch))
	assert.False(t, InRef(values, &target, func(a, b *string) (bool, error) {
		return true, commonerrors.ErrUnexpected
	}))
}
