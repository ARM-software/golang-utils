package validation

import (
	"iter"
	"regexp"
	"strings"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
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
		errortest.AssertErrorDescription(t, validation.Validate("abc", PrefixItems(Type("string"))), "must be an array or slice")

		var stringPtr *string
		errortest.AssertErrorDescription(t, validation.Validate(stringPtr, PrefixItems(Type("string"))), "must be an array or slice")

		var nilFunc func() = nil
		errortest.AssertErrorDescription(t, validation.Validate(nilFunc, PrefixItems(Type("string"))), "must be an array or slice")

		var nilMap map[int]string
		errortest.AssertErrorDescription(t, validation.Validate(nilMap, UniqueItems[string](func(item string) string { return item })), "must be an array or slice")
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

		seq := iter.Seq2[string, any](func(yield func(string, any) bool) {
			_ = yield("a", 1)
			_ = yield("b", 2)
		})
		assert.NoError(t, validation.Validate(seq, RequiredProperties("a")))
		assert.NoError(t, validation.Validate(seq, DependentRequired(map[string][]string{"a": {"b"}})))
		assert.NoError(t, validation.Validate(seq, PropertyNames(Pattern(re))))
		assert.NoError(t, validation.Validate(seq, AdditionalProperties("a", "b")))

		var stringPtr *string
		errortest.AssertErrorDescription(t, validation.Validate(stringPtr, RequiredProperties("a")), "must be a map")

		var nilFunc func() = nil
		errortest.AssertErrorDescription(t, validation.Validate(nilFunc, AdditionalProperties("a", "b")), "must be a map")
	})

	t.Run("contains", func(t *testing.T) {
		assert.NoError(t, validation.Validate([]string{"a", "b"}, Contains(Const("a"))))
		assert.Error(t, validation.Validate([]string{"b", "c"}, Contains(Const("a"))))
		assert.NoError(t, validation.Validate([]string{"a", "b", "a"}, MinContains(2, Const("a"))))
		assert.Error(t, validation.Validate([]string{"a", "b"}, MinContains(2, Const("a"))))
		assert.NoError(t, validation.Validate([]string{"a", "b"}, MaxContains(1, Const("a"))))
		assert.Error(t, validation.Validate([]string{"a", "a"}, MaxContains(1, Const("a"))))
		errortest.AssertErrorDescription(t, validation.Validate("abc", Contains(Const("a"))), "must be an array or slice")

		var stringPtr *string
		errortest.AssertErrorDescription(t, validation.Validate(stringPtr, MinContains(1, Const("a"))), "must be an array or slice")

		var nilFunc func() = nil
		errortest.AssertErrorDescription(t, validation.Validate(nilFunc, MaxContains(1, Const("a"))), "must be an array or slice")

		var nilMap map[int]string
		errortest.AssertErrorDescription(t, validation.Validate(nilMap, Contains(Const("a"))), "must be an array or slice")
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

		var stringPtr *string
		errortest.AssertErrorDescription(t, validation.Validate(stringPtr, MutuallyExclusiveWith("A", "B")), "must be a map")

		var nilFunc func() = nil
		errortest.AssertErrorDescription(t, validation.Validate(nilFunc, MutuallyExclusiveWith("A", "B")), "must be a map")
	})

	t.Run("item keys", func(t *testing.T) {
		keyFunc := func(value string) string { return value }
		assert.NoError(t, validation.Validate([]string{"a", "b"}, RequiredItems(keyFunc, "a", "b")))
		assert.Error(t, validation.Validate([]string{"a"}, RequiredItems(keyFunc, "a", "b")))
		assert.NoError(t, validation.Validate([]string{"a", "b"}, DependentRequiredItems(keyFunc, map[string][]string{"a": {"b"}})))
		assert.Error(t, validation.Validate([]string{"a"}, DependentRequiredItems(keyFunc, map[string][]string{"a": {"b"}})))
		assert.NoError(t, validation.Validate([]string{"a"}, AdditionalItems(keyFunc, "a", "b")))
		assert.Error(t, validation.Validate([]string{"c"}, AdditionalItems(keyFunc, "a", "b")))
		assert.NoError(t, validation.Validate([]string{"a"}, MutuallyExclusiveItems(keyFunc, "a", "b")))
		assert.Error(t, validation.Validate([]string{"a", "b"}, MutuallyExclusiveItems(keyFunc, "a", "b")))
		assert.NoError(t, validation.Validate([]string{"a"}, AtMostOneItem(keyFunc, "a", "b")))
		assert.Error(t, validation.Validate([]string{"a", "b"}, AtMostOneItem(keyFunc, "a", "b")))
		assert.NoError(t, validation.Validate([]string{"a"}, OneOfItems(keyFunc, "a", "b")))
		assert.Error(t, validation.Validate([]string{}, OneOfItems(keyFunc, "a", "b")))
		assert.Error(t, validation.Validate([]string{"a", "b"}, OneOfItems(keyFunc, "a", "b")))
		assert.NoError(t, validation.Validate([]string{"a"}, AtLeastOneItem(keyFunc, "a", "b")))
		assert.Error(t, validation.Validate([]string{}, AtLeastOneItem(keyFunc, "a", "b")))
		assert.NoError(t, validation.Validate([]string{"a"}, ForbiddenItems(keyFunc, "b")))
		assert.Error(t, validation.Validate([]string{"a"}, ForbiddenItems(keyFunc, "a")))

		seq := iter.Seq[string](func(yield func(string) bool) {
			_ = yield("a")
			_ = yield("b")
		})
		assert.NoError(t, validation.Validate(seq, RequiredItems(keyFunc, "a", "b")))

		errortest.AssertErrorDescription(t, validation.Validate("abc", MutuallyExclusiveItems(keyFunc, "a", "b")), "must be an array or slice")
	})

	t.Run("item key strings", func(t *testing.T) {
		type valueT struct {
			i int
			j string
		}
		keyFunc := func(value valueT) string { return value.j }

		assert.NoError(t, validation.Validate([]valueT{{i: 0, j: "a"}}, AtMostOneItemKey(keyFunc, "a", "b")))
		assert.Error(t, validation.Validate([]valueT{{i: 0, j: "a"}, {i: 1, j: "b"}}, AtMostOneItemKey(keyFunc, "a", "b")))
		assert.NoError(t, validation.Validate([]valueT{{i: 0, j: "a"}, {i: 1, j: "b"}}, RequiredItemKeys(keyFunc, "a", "b")))
		assert.Error(t, validation.Validate([]valueT{{i: 0, j: "a"}}, RequiredItemKeys(keyFunc, "a", "b")))
		assert.NoError(t, validation.Validate([]valueT{{i: 0, j: "a"}, {i: 1, j: "b"}}, DependentRequiredItemKeys(keyFunc, map[string][]string{"a": {"b"}})))
		assert.Error(t, validation.Validate([]valueT{{i: 0, j: "a"}}, DependentRequiredItemKeys(keyFunc, map[string][]string{"a": {"b"}})))
		assert.NoError(t, validation.Validate([]valueT{{i: 0, j: "a"}}, AdditionalItemKeys(keyFunc, "a", "b")))
		assert.Error(t, validation.Validate([]valueT{{i: 0, j: "c"}}, AdditionalItemKeys(keyFunc, "a", "b")))
		assert.NoError(t, validation.Validate([]valueT{{i: 0, j: "a"}}, MutuallyExclusiveItemKeys(keyFunc, "a", "b")))
		assert.Error(t, validation.Validate([]valueT{{i: 0, j: "a"}, {i: 1, j: "b"}}, MutuallyExclusiveItemKeys(keyFunc, "a", "b")))
		assert.NoError(t, validation.Validate([]valueT{{i: 0, j: "a"}}, OneOfItemKeys(keyFunc, "a", "b")))
		assert.Error(t, validation.Validate([]valueT{}, OneOfItemKeys(keyFunc, "a", "b")))
		assert.Error(t, validation.Validate([]valueT{{i: 0, j: "a"}, {i: 1, j: "b"}}, OneOfItemKeys(keyFunc, "a", "b")))
		assert.NoError(t, validation.Validate([]valueT{{i: 0, j: "a"}}, AtLeastOneItemKey(keyFunc, "a", "b")))
		assert.Error(t, validation.Validate([]valueT{}, AtLeastOneItemKey(keyFunc, "a", "b")))
		assert.NoError(t, validation.Validate([]valueT{{i: 0, j: "a"}}, ForbiddenItemKeys(keyFunc, "b")))
		assert.Error(t, validation.Validate([]valueT{{i: 0, j: "a"}}, ForbiddenItemKeys(keyFunc, "a")))
	})

	t.Run("schema terminology aliases", func(t *testing.T) {
		assert.NoError(t, validation.Validate("blue", Enum("red", "blue")))
		assert.Error(t, validation.Validate("green", Enum("red", "blue")))

		assert.NoError(t, validation.Validate("v1", Const("v1")))
		assert.Error(t, validation.Validate("v2", Const("v1")))

		re := regexp.MustCompile(`^[a-z]+$`)
		assert.NoError(t, validation.Validate("abc", Pattern(re)))
		assert.Error(t, validation.Validate("123", Pattern(re)))
		assert.NoError(t, validation.Validate(3, XIntOrString()))
		assert.NoError(t, validation.Validate("3", XIntOrString()))
		assert.NoError(t, validation.Validate(float64(3), XIntOrString()))
		assert.Error(t, validation.Validate(float64(3.5), XIntOrString()))

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

func TestJSONSchemaInspiredRulesFieldReferences(t *testing.T) {
	type fields struct {
		A int
		B int
		C int
	}

	t.Run("required and additional properties", func(t *testing.T) {
		value := &fields{A: 1, B: 2}
		assert.NoError(t, validation.Validate(value, RequiredPropertiesBy(&value.A, &value.B)))
		assert.NoError(t, validation.Validate(value, RequiredPropertiesBy([]string{"A", "B"})))
		assert.Error(t, validation.Validate(value, RequiredPropertiesBy(&value.A, &value.C)))
		assert.NoError(t, validation.Validate(value, AdditionalPropertiesBy(&value.A, "B", &value.C)))
		assert.NoError(t, validation.Validate(value, AdditionalPropertiesBy([]string{"A", "B", "C"})))
		assert.Error(t, validation.Validate(value, AdditionalPropertiesBy(&value.A)))
	})

	t.Run("dependent properties and schemas", func(t *testing.T) {
		value := &fields{A: 1, B: 2, C: 3}
		assert.NoError(t, validation.Validate(value, DependentRequiredBy(map[any]any{&value.A: []any{&value.B, &value.C}})))
		assert.Error(t, validation.Validate(&fields{A: 1, B: 2}, DependentRequiredBy(map[any]any{&value.A: []any{&value.B, &value.C}})))
		assert.NoError(t, validation.Validate(value, DependentSchemasBy(map[any]validation.Rule{&value.A: RequiredPropertiesBy(&value.B)})))
		assert.Error(t, validation.Validate(&fields{A: 1}, DependentSchemasBy(map[any]validation.Rule{&value.A: RequiredPropertiesBy(&value.B)})))
	})

	t.Run("exclusive variants", func(t *testing.T) {
		value := &fields{A: 1}
		assert.NoError(t, validation.Validate(value, MutuallyExclusiveWithBy(&value.A, &value.B)))
		assert.NoError(t, validation.Validate(value, AtMostOnePropertyBy(&value.A, &value.B)))
		assert.NoError(t, validation.Validate(value, OneOfPropertiesBy(&value.A, &value.B)))
		assert.NoError(t, validation.Validate(value, AtLeastOnePropertyBy(&value.A, &value.B)))
		assert.NoError(t, validation.Validate(value, ForbiddenPropertiesBy(&value.B, &value.C)))

		value.B = 2
		assert.Error(t, validation.Validate(value, MutuallyExclusiveWithBy(&value.A, &value.B)))
		assert.Error(t, validation.Validate(value, AtMostOnePropertyBy(&value.A, &value.B)))
		assert.Error(t, validation.Validate(value, OneOfPropertiesBy(&value.A, &value.B)))
		assert.Error(t, validation.Validate(value, ForbiddenPropertiesBy(&value.A, &value.B)))

		assert.Error(t, validation.Validate(&fields{}, OneOfPropertiesBy(&value.A, &value.B)))
		assert.Error(t, validation.Validate(&fields{}, AtLeastOnePropertyBy(&value.A, &value.B)))
	})

	t.Run("unmatched reference", func(t *testing.T) {
		value := &fields{A: 1}
		other := &fields{}
		err := validation.Validate(value, RequiredPropertiesBy(&other.A))
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
	})

	t.Run("embedded field reference", func(t *testing.T) {
		type embedded struct {
			B int
		}
		type withEmbedded struct {
			embedded
			A int
		}
		value := &withEmbedded{embedded: embedded{B: 2}, A: 1}
		assert.NoError(t, validation.Validate(value, RequiredPropertiesBy(&value.B, &value.A)))
	})
}
