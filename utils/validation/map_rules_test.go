package validation

import (
	"iter"
	"regexp"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestMapRules(t *testing.T) {
	t.Run("map keys", func(t *testing.T) {
		re := regexp.MustCompile(`^[a-z]+$`)
		assert.NoError(t, validation.Validate(map[string]any{"alpha": 1}, MapKeys(Pattern(re))))
		assert.Error(t, validation.Validate(map[string]any{"Alpha": 1}, MapKeys(Pattern(re))))

		var s *string
		errortest.AssertErrorDescription(t, validation.Validate(s, MapKeys(Pattern(re))), "must be a map")

		var f func() = nil
		errortest.AssertErrorDescription(t, validation.Validate(f, MapKeys(Pattern(re))), "must be a map")

		seq := iter.Seq2[string, any](func(yield func(string, any) bool) {
			_ = yield("alpha", 1)
		})
		assert.NoError(t, validation.Validate(seq, MapKeys(Pattern(re))))
	})

	t.Run("map values", func(t *testing.T) {
		assert.NoError(t, validation.Validate(map[string]any{"a": "x"}, MapValues(Type("string"))))
		assert.Error(t, validation.Validate(map[string]any{"a": 1}, MapValues(Type("string"))))

		var s *string
		errortest.AssertErrorDescription(t, validation.Validate(s, MapValues(Type("string"))), "must be a map")

		var f func() = nil
		errortest.AssertErrorDescription(t, validation.Validate(f, MapValues(Type("string"))), "must be a map")

		seq := iter.Seq2[string, any](func(yield func(string, any) bool) {
			_ = yield("a", "x")
		})
		assert.NoError(t, validation.Validate(seq, MapValues(Type("string"))))
	})
}
