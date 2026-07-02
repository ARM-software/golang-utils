package validation

import (
	"regexp"
	"strings"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/field"
)

func TestJSONSchemaInspiredRules(t *testing.T) {
	t.Run("numeric", func(t *testing.T) {
		assert.NoError(t, validation.Validate(10, MultipleOf(5)))
		assert.Error(t, validation.Validate(11, MultipleOf(5)))

		assert.NoError(t, validation.Validate(10, Maximum(10)))
		assert.Error(t, validation.Validate(11, Maximum(10)))

		assert.NoError(t, validation.Validate(9, ExclusiveMaximum(10)))
		assert.Error(t, validation.Validate(10, ExclusiveMaximum(10)))

		assert.NoError(t, validation.Validate(10, Minimum(10)))
		assert.Error(t, validation.Validate(9, Minimum(10)))

		assert.NoError(t, validation.Validate(11, ExclusiveMinimum(10)))
		assert.Error(t, validation.Validate(10, ExclusiveMinimum(10)))
	})

	t.Run("string lengths", func(t *testing.T) {
		assert.NoError(t, validation.Validate("hello", MaxLength(5)))
		assert.Error(t, validation.Validate("hello!", MaxLength(5)))
		assert.NoError(t, validation.Validate("hello", MinLength(5)))
		assert.Error(t, validation.Validate("hell", MinLength(5)))
		assert.NoError(t, validation.Validate("éé", LengthRule(nil, field.ToOptionalInt(2))))
		assert.Error(t, validation.Validate("ééé", LengthRule(nil, field.ToOptionalInt(2))))
	})

	t.Run("items", func(t *testing.T) {
		assert.NoError(t, validation.Validate([]any{"a", 1}, PrefixItems(Type("string"), Type("integer"))))
		assert.Error(t, validation.Validate([]any{1, "a"}, PrefixItems(Type("string"), Type("integer"))))
		assert.NoError(t, validation.Validate([]int{1, 2}, MaxItems(2)))
		assert.Error(t, validation.Validate([]int{1, 2, 3}, MaxItems(2)))
		assert.NoError(t, validation.Validate([]int{1, 2}, MinItems(2)))
		assert.Error(t, validation.Validate([]int{1}, MinItems(2)))
		assert.NoError(t, validation.Validate([]int{1, 2, 3}, UniqueItems[int](collection.IdentityMapFunc[int]())))
		assert.Error(t, validation.Validate([]int{1, 2, 1}, UniqueItems[int](collection.IdentityMapFunc[int]())))
		assert.NoError(t, validation.Validate([]string{"a", "A"}, UniqueItems[string](func(item string) string { return item })))
		assert.Error(t, validation.Validate([]string{"a", "A"}, UniqueItems[string](strings.ToLower)))
		assert.NoError(t, validation.Validate([]int{1, 2}, Type("array")))
		assert.Error(t, validation.Validate([]int{1, 2}, Type("object")))
	})

	t.Run("properties", func(t *testing.T) {
		assert.NoError(t, validation.Validate(map[string]any{"a": 1, "b": 2}, MaxProperties(2)))
		assert.Error(t, validation.Validate(map[string]any{"a": 1, "b": 2, "c": 3}, MaxProperties(2)))
		assert.NoError(t, validation.Validate(map[string]any{"a": 1, "b": 2}, MinProperties(2)))
		assert.Error(t, validation.Validate(map[string]any{"a": 1}, MinProperties(2)))
		assert.NoError(t, validation.Validate(map[string]any{"a": 1}, RequiredProperties("a")))
		assert.Error(t, validation.Validate(map[string]any{"a": 1}, RequiredProperties("a", "b")))
		assert.NoError(t, validation.Validate(map[string]any{"a": 1, "b": 2}, DependentRequired(map[string][]string{"a": {"b"}})))
		assert.Error(t, validation.Validate(map[string]any{"a": 1}, DependentRequired(map[string][]string{"a": {"b"}})))
		assert.NoError(t, validation.Validate(map[string]any{"a": 1, "b": 2}, DependentSchemas(map[string]validation.Rule{"a": RequiredProperties("b")})))
		assert.Error(t, validation.Validate(map[string]any{"a": 1}, DependentSchemas(map[string]validation.Rule{"a": RequiredProperties("b")})))

		re := regexp.MustCompile(`^[a-z]+$`)
		assert.NoError(t, validation.Validate(map[string]any{"alpha": 1}, PropertyNames(Pattern(re))))
		assert.Error(t, validation.Validate(map[string]any{"Alpha": 1}, PropertyNames(Pattern(re))))
		assert.NoError(t, validation.Validate(map[string]any{"s_name": "alice", "n_count": 2}, PatternProperties(
			PatternProperty{Pattern: regexp.MustCompile(`^s_`), Rule: Type("string")},
		)))
		assert.Error(t, validation.Validate(map[string]any{"s_name": 2}, PatternProperties(
			PatternProperty{Pattern: regexp.MustCompile(`^s_`), Rule: Type("string")},
		)))

		assert.NoError(t, validation.Validate(map[string]any{"a": 1}, AdditionalProperties("a", "b")))
		assert.Error(t, validation.Validate(map[string]any{"c": 1}, AdditionalProperties("a", "b")))
	})

	t.Run("contains", func(t *testing.T) {
		assert.NoError(t, validation.Validate([]string{"a", "b"}, Contains(Const("a"))))
		assert.Error(t, validation.Validate([]string{"b", "c"}, Contains(Const("a"))))
		assert.NoError(t, validation.Validate([]string{"a", "b", "a"}, MinContains(2, Const("a"))))
		assert.Error(t, validation.Validate([]string{"a", "b"}, MinContains(2, Const("a"))))
		assert.NoError(t, validation.Validate([]string{"a", "b"}, MaxContains(1, Const("a"))))
		assert.Error(t, validation.Validate([]string{"a", "a"}, MaxContains(1, Const("a"))))
	})

	t.Run("mutually exclusive", func(t *testing.T) {
		assert.NoError(t, validation.Validate(map[string]any{"a": 1}, MutuallyExclusiveWith("a", "b")))
		assert.NoError(t, validation.Validate(map[string]any{"a": 1, "b": nil}, MutuallyExclusiveWith("a", "b")))
		assert.Error(t, validation.Validate(map[string]any{"a": 1, "b": 2}, MutuallyExclusiveWith("a", "b")))
		assert.NoError(t, validation.Validate(map[string]any{"a": 1}, AtMostOneProperty("a", "b")))
		assert.Error(t, validation.Validate(map[string]any{"a": 1, "b": 2}, AtMostOneProperty("a", "b")))
		assert.NoError(t, validation.Validate(map[string]any{"a": 1}, OneOfProperties("a", "b")))
		assert.Error(t, validation.Validate(map[string]any{}, OneOfProperties("a", "b")))
		assert.Error(t, validation.Validate(map[string]any{"a": 1, "b": 2}, OneOfProperties("a", "b")))
		assert.NoError(t, validation.Validate(map[string]any{"a": 1}, AtLeastOneProperty("a", "b")))
		assert.Error(t, validation.Validate(map[string]any{}, AtLeastOneProperty("a", "b")))
		assert.NoError(t, validation.Validate(map[string]any{"a": 1}, ForbiddenProperties("b")))
		assert.Error(t, validation.Validate(map[string]any{"a": 1}, ForbiddenProperties("a")))

		type fields struct {
			A int
			B int
		}
		assert.NoError(t, validation.Validate(fields{A: 1}, MutuallyExclusiveWith("A", "B")))
		assert.Error(t, validation.Validate(fields{A: 1, B: 2}, MutuallyExclusiveWith("A", "B")))
		assert.NoError(t, validation.Validate(fields{A: 1}, OneOfProperties("A", "B")))
		assert.Error(t, validation.Validate(fields{}, OneOfProperties("A", "B")))
		assert.NoError(t, validation.Validate(fields{A: 1}, AtLeastOneProperty("A", "B")))
		assert.Error(t, validation.Validate(fields{}, AtLeastOneProperty("A", "B")))
		assert.NoError(t, validation.Validate(fields{A: 1}, ForbiddenProperties("B")))
		assert.Error(t, validation.Validate(fields{A: 1}, ForbiddenProperties("A")))
	})

	t.Run("schema terminology aliases", func(t *testing.T) {
		assert.NoError(t, validation.Validate("blue", Enum("red", "blue")))
		assert.Error(t, validation.Validate("green", Enum("red", "blue")))

		assert.NoError(t, validation.Validate("v1", Const("v1")))
		assert.Error(t, validation.Validate("v2", Const("v1")))

		re := regexp.MustCompile(`^[a-z]+$`)
		assert.NoError(t, validation.Validate("abc", Pattern(re)))
		assert.Error(t, validation.Validate("123", Pattern(re)))

		assert.NoError(t, validation.Validate("plain-text", Not(is.Email)))
		assert.Error(t, validation.Validate("user@example.com", Not(is.Email)))
		assert.NoError(t, validation.Validate(nil, Nullable(Type("string"))))
		assert.NoError(t, validation.Validate("hello", Nullable(Type("string"))))
		assert.Error(t, validation.Validate(1, Nullable(Type("string"))))

		assert.NoError(t, validation.Validate("user@example.com", AnyOf(is.Email, is.UUID)))
		assert.Error(t, validation.Validate("not-valid", AnyOf(is.Email, is.UUID)))

		assert.NoError(t, validation.Validate("user@example.com", AllOf(validation.Required, is.Email)))
		assert.Error(t, validation.Validate("", AllOf(validation.Required, is.Email)))

		assert.NoError(t, validation.Validate("plain-text", NoneOf(is.Email, is.UUID)))
		assert.Error(t, validation.Validate("user@example.com", NoneOf(is.Email, is.UUID)))

		assert.NoError(t, validation.Validate("user@example.com", OneOf(is.Email, is.UUID)))
		assert.Error(t, validation.Validate("user@example.com", OneOf(validation.Required, is.Email)))

		assert.NoError(t, validation.Validate("hello", NotEmpty()))
		assert.Error(t, validation.Validate("   ", NotEmpty()))
	})
}
