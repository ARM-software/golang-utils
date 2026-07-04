package validation

import (
	"iter"
	"regexp"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
)

func TestCollectionRules(t *testing.T) {
	t.Run("array items", func(t *testing.T) {
		assert.NoError(t, validation.Validate([]any{"a", "b"}, ArrayItems(Type("string"))))
		assert.Error(t, validation.Validate([]any{"a", 1}, ArrayItems(Type("string"))))

		seq := iter.Seq[any](func(yield func(any) bool) {
			_ = yield("a")
			_ = yield("b")
		})
		assert.NoError(t, validation.Validate(seq, ArrayItems(Type("string"))))
	})

	t.Run("map keys", func(t *testing.T) {
		re := regexp.MustCompile(`^[a-z]+$`)
		assert.NoError(t, validation.Validate(map[string]any{"alpha": 1}, MapKeys(Pattern(re))))
		assert.Error(t, validation.Validate(map[string]any{"Alpha": 1}, MapKeys(Pattern(re))))

		seq := iter.Seq2[string, any](func(yield func(string, any) bool) {
			_ = yield("alpha", 1)
		})
		assert.NoError(t, validation.Validate(seq, MapKeys(Pattern(re))))
	})

	t.Run("map values", func(t *testing.T) {
		assert.NoError(t, validation.Validate(map[string]any{"a": "x"}, MapValues(Type("string"))))
		assert.Error(t, validation.Validate(map[string]any{"a": 1}, MapValues(Type("string"))))

		seq := iter.Seq2[string, any](func(yield func(string, any) bool) {
			_ = yield("a", "x")
		})
		assert.NoError(t, validation.Validate(seq, MapValues(Type("string"))))
	})
}
