package url

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrl_IsParamSegment(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		assert.True(t, IsParamSegment("{abc}"))
	})

	t.Run("false", func(t *testing.T) {
		assert.False(t, IsParamSegment("abc"))
	})
}
