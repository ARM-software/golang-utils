package validation

import (
	"iter"
	"regexp"
	"testing"

	"github.com/go-faker/faker/v4"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestMapRules(t *testing.T) {
	t.Run("map keys", func(t *testing.T) {
		re := regexp.MustCompile(`^[a-z]+$`)
		value := faker.Word()
		type namedString string
		assert.NoError(t, validation.Validate(map[string]any{"alpha": 1}, MapKeys(Pattern(re))))
		assert.Error(t, validation.Validate(map[string]any{"Alpha": 1}, MapKeys(Pattern(re))))
		assert.NoError(t, validation.Validate(map[any]any{"alpha": 1}, MapKeys(Pattern(re))))
		assert.NoError(t, validation.Validate(map[any]any{"": 1}, MapKeys(Pattern(regexp.MustCompile(`^$`)))))
		assert.NoError(t, validation.Validate(map[namedString]any{"alpha": 1}, MapKeys(Pattern(re))))
		require.NotPanics(t, func() {
			assert.NoError(t, validation.Validate(map[int]string{1: value}, MapKeys(Pattern(re))))
		})

		var s *string
		errortest.AssertErrorDescription(t, validation.Validate(s, MapKeys(Pattern(re))), "must be a map")

		var f func() = nil
		errortest.AssertErrorDescription(t, validation.Validate(f, MapKeys(Pattern(re))), "must be a map")

		seq := iter.Seq2[string, any](func(yield func(string, any) bool) {
			_ = yield("alpha", 1)
		})
		assert.NoError(t, validation.Validate(seq, MapKeys(Pattern(re))))

		nonStringKeySeq := iter.Seq2[int, any](func(yield func(int, any) bool) {})
		errortest.AssertErrorDescription(t, validation.Validate(nonStringKeySeq, MapKeys(Pattern(re))), "must be a map")

		nonStringKeySeq = iter.Seq2[int, any](func(yield func(int, any) bool) {
			_ = yield(1, 1)
		})
		errortest.AssertErrorDescription(t, validation.Validate(nonStringKeySeq, MapKeys(Pattern(re))), "must be a map")

		var nilSeq2 iter.Seq2[string, any]
		assert.NoError(t, validation.Validate(nilSeq2, MapKeys(Pattern(re))))
	})

	t.Run("map values", func(t *testing.T) {
		value := faker.Word()
		type namedString string
		assert.NoError(t, validation.Validate(map[string]any{"a": "x"}, MapValues(Type("string"))))
		assert.Error(t, validation.Validate(map[string]any{"a": 1}, MapValues(Type("string"))))
		assert.NoError(t, validation.Validate(map[any]any{"a": value}, MapValues(Type("string"))))
		assert.NoError(t, validation.Validate(map[namedString]string{"a": value}, MapValues(Type("string"))))
		require.NotPanics(t, func() {
			assert.NoError(t, validation.Validate(map[int]string{1: value}, MapValues(Type("string"))))
		})

		var s *string
		errortest.AssertErrorDescription(t, validation.Validate(s, MapValues(Type("string"))), "must be a map")

		var f func() = nil
		errortest.AssertErrorDescription(t, validation.Validate(f, MapValues(Type("string"))), "must be a map")

		seq := iter.Seq2[string, any](func(yield func(string, any) bool) {
			_ = yield("a", "x")
		})
		assert.NoError(t, validation.Validate(seq, MapValues(Type("string"))))

		var nilSeq2 iter.Seq2[string, any]
		assert.NoError(t, validation.Validate(nilSeq2, MapValues(Type("string"))))
	})
}
