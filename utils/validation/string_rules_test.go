package validation

import (
	"regexp"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
)

func TestStringRules(t *testing.T) {
	t.Run("string rules", func(t *testing.T) {
		assert.NoError(t, validation.Validate("prefix-value", Prefix("pre")))
		assert.Error(t, validation.Validate("value", Prefix("pre")))

		assert.NoError(t, validation.Validate("file.txt", Suffix(".txt")))
		assert.Error(t, validation.Validate("file.log", Suffix(".txt")))

		assert.NoError(t, validation.Validate("hello world", ContainsString("world")))
		assert.Error(t, validation.Validate("hello", ContainsString("world")))

		re := regexp.MustCompile(`^[a-z]+$`)
		assert.NoError(t, validation.Validate("hello", Like(re)))
		assert.Error(t, validation.Validate("123", Like(re)))

		assert.NoError(t, validation.Validate("hello", NotContains("world")))
		assert.Error(t, validation.Validate("hello world", NotContains("world")))

		assert.NoError(t, validation.Validate("helloworld", NotContainsWhitespaces()))
		assert.Error(t, validation.Validate("hello world", NotContainsWhitespaces()))
	})

	t.Run("length and occurs", func(t *testing.T) {
		assert.NoError(t, validation.Validate("ab", LengthExact(2)))
		assert.Error(t, validation.Validate("abc", LengthExact(2)))

		assert.NoError(t, validation.Validate("banana", MinOccurs("a", 3)))
		assert.Error(t, validation.Validate("banana", MinOccurs("a", 4)))

		assert.NoError(t, validation.Validate("banana", MaxOccurs("a", 3)))
		assert.Error(t, validation.Validate("banana", MaxOccurs("a", 2)))

		assert.NoError(t, validation.Validate("banana", OccursExactly("a", 3)))
		assert.Error(t, validation.Validate("banana", OccursExactly("a", 2)))
	})
}
