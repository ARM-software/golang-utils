package git

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	require.NoError(t, DefaultLimits().Validate()) //nolint:typecheck
}
