package filesystem

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	validationRules "github.com/ARM-software/golang-utils/utils/validation"
)

func TestNewLimits(t *testing.T) {
	require.NoError(t, NoLimits().Validate())      //nolint:typecheck
	require.NoError(t, DefaultLimits().Validate()) //nolint:typecheck
	require.NoError(t, validation.Validate(NoLimits(), validationRules.Required))
	assert.True(t, DefaultLimits().Apply())
	assert.False(t, NoLimits().Apply())
}
