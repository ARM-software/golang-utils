package validation

import (
	"regexp"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
)

func TestCollectionRules(t *testing.T) {
	t.Run("array items", func(t *testing.T) {
		assert.NoError(t, validation.Validate([]any{"a", "b"}, ArrayItems(Type("string"))))
		assert.Error(t, validation.Validate([]any{"a", 1}, ArrayItems(Type("string"))))
	})

	t.Run("map keys", func(t *testing.T) {
		re := regexp.MustCompile(`^[a-z]+$`)
		assert.NoError(t, validation.Validate(map[string]any{"alpha": 1}, MapKeys(Pattern(re))))
		assert.Error(t, validation.Validate(map[string]any{"Alpha": 1}, MapKeys(Pattern(re))))
	})

	t.Run("map values", func(t *testing.T) {
		assert.NoError(t, validation.Validate(map[string]any{"a": "x"}, MapValues(Type("string"))))
		assert.Error(t, validation.Validate(map[string]any{"a": 1}, MapValues(Type("string"))))
	})
}
