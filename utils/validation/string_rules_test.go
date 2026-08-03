package validation

import (
	"regexp"
	"testing"

	"github.com/go-faker/faker/v4"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
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

	t.Run("string rule type errors", func(t *testing.T) {
		invalidInputs := []struct {
			name  string
			value any
		}{
			{name: "int", value: 123},
			{name: "bool", value: true},
			{name: "float", value: 1.5},
			{name: "struct", value: struct{ Name string }{Name: faker.Word()}},
			{name: "map", value: map[string]int{faker.Word(): 1}},
		}

		tests := []struct {
			name        string
			rule        validation.Rule
			expectedErr string
		}{
			{name: "prefix", rule: Prefix("pre"), expectedErr: "must be either a string or byte slice"},
			{name: "suffix", rule: Suffix(".txt"), expectedErr: "must be either a string or byte slice"},
			{name: "contains", rule: ContainsString("world"), expectedErr: "must be either a string or byte slice"},
			{name: "like", rule: Like(regexp.MustCompile(`^[a-z]+$`)), expectedErr: "must be in a valid format"},
			{name: "not contains", rule: NotContains("world"), expectedErr: "must be either a string or byte slice"},
			{name: "not whitespace", rule: NotContainsWhitespaces(), expectedErr: "must be either a string or byte slice"},
			{name: "min occurs", rule: MinOccurs("a", 1), expectedErr: "must be either a string or byte slice"},
			{name: "max occurs", rule: MaxOccurs("a", 1), expectedErr: "must be either a string or byte slice"},
			{name: "occurs exactly", rule: OccursExactly("a", 1), expectedErr: "must be either a string or byte slice"},
		}

		for i := range tests {
			test := tests[i]
			t.Run(test.name, func(t *testing.T) {
				for j := range invalidInputs {
					invalid := invalidInputs[j]
					t.Run(invalid.name, func(t *testing.T) {
						errortest.AssertErrorDescription(t, validation.Validate(invalid.value, test.rule), test.expectedErr)
					})
				}
			})
		}
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
