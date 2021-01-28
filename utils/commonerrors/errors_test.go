package commonerrors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAny(t *testing.T) {
	assert.True(t, Any(ErrNotImplemented, ErrInvalid, ErrNotImplemented, ErrUnknown))
	assert.False(t, Any(ErrNotImplemented, ErrInvalid, ErrUnknown))
	assert.True(t, Any(fmt.Errorf("an error %w", ErrNotImplemented), ErrInvalid, ErrNotImplemented, ErrUnknown))
	assert.False(t, Any(fmt.Errorf("an error %w", ErrNotImplemented), ErrInvalid, ErrUnknown))
}

func TestNone(t *testing.T) {
	assert.False(t, None(ErrNotImplemented, ErrInvalid, ErrNotImplemented, ErrUnknown))
	assert.True(t, None(ErrNotImplemented, ErrInvalid, ErrUnknown))
	assert.False(t, None(fmt.Errorf("an error %w", ErrNotImplemented), ErrInvalid, ErrNotImplemented, ErrUnknown))
	assert.True(t, None(fmt.Errorf("an error %w", ErrNotImplemented), ErrInvalid, ErrUnknown))
}
