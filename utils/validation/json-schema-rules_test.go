package validation

import (
	"encoding/json"
	"iter"
	"math"
	"regexp"
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/field"
)

func TestJSONSchemaInspiredRules(t *testing.T) {
	invalidCollectionInputs := []struct {
		name  string
		value any
	}{
		{name: "string", value: faker.Word()},
		{name: "bool", value: true},
		{name: "float", value: 1.5},
		{name: "struct", value: struct{ Name string }{Name: faker.Word()}},
		{name: "map", value: map[string]int{faker.Word(): 1}},
	}

	invalidObjectInputs := []struct {
		name  string
		value any
	}{
		{name: "string", value: faker.Word()},
		{name: "bool", value: true},
		{name: "float", value: 1.5},
		{name: "slice", value: []string{faker.Word()}},
		{name: "int", value: 123},
	}

	t.Run("numeric", func(t *testing.T) {
		var jsonInteger any
		require.NoError(t, json.Unmarshal([]byte(`1`), &jsonInteger))
		assert.NoError(t, validation.Validate(jsonInteger, Type("integer")))
		assert.NoError(t, validation.Validate(jsonInteger, Type("number")))
		// JSON numbers decoded into any are float64 which breaks the MultipleOf as it can't compare int and float
		require.NoError(t, json.Unmarshal([]byte(`10`), &jsonInteger))
		assert.NoError(t, validation.Validate(jsonInteger, MultipleOf(5)))

		assert.NoError(t, validation.Validate(10, MultipleOf(5)))
		assert.Error(t, validation.Validate(11, MultipleOf(5)))
		assert.NoError(t, validation.Validate(0, MultipleOf(5)))
		assert.NoError(t, validation.Validate(0.3, MultipleOf(0.1)))
		assert.Error(t, validation.Validate(0.31, MultipleOf(0.1)))
		assert.Error(t, validation.Validate(10, MultipleOf(0)))
		assert.Error(t, validation.Validate(10, MultipleOf(-5)))
		value := 10
		assert.NoError(t, validation.Validate(&value, MultipleOf(5)))

		assert.NoError(t, validation.Validate(10, Maximum(10)))
		assert.Error(t, validation.Validate(11, Maximum(10)))
		assert.NoError(t, validation.Validate(0, Maximum(10)))
		assert.NoError(t, validation.Validate(float64(9), Maximum(10)))
		assert.NoError(t, validation.Validate(9, Maximum(float64(10))))

		assert.NoError(t, validation.Validate(9, ExclusiveMaximum(10)))
		assert.Error(t, validation.Validate(10, ExclusiveMaximum(10)))
		assert.NoError(t, validation.Validate(0, ExclusiveMaximum(10)))

		assert.NoError(t, validation.Validate(10, Minimum(10)))
		assert.Error(t, validation.Validate(9, Minimum(10)))
		assert.Error(t, validation.Validate(0, Minimum(10)))

		assert.NoError(t, validation.Validate(11, ExclusiveMinimum(10)))
		assert.Error(t, validation.Validate(10, ExclusiveMinimum(10)))
		assert.Error(t, validation.Validate(0, ExclusiveMinimum(10)))

		errortest.AssertErrorDescription(t, validation.Validate("arbitrarily long", MaxLength(-1)), "maxLength must be non-negative")
		errortest.AssertErrorDescription(t, validation.Validate([]int{1, 2}, MaxItems(-1)), "maxItems must be non-negative")
		errortest.AssertErrorDescription(t, validation.Validate([]int{}, MinContains(-1, Const(1))), "minContains must be non-negative")
		errortest.AssertErrorDescription(t, validation.Validate([]int{}, MaxContains(-1, Const(1))), "maxContains must be non-negative")
	})

	t.Run("pattern", func(t *testing.T) {
		rule := Pattern(regexp.MustCompile(`^a`))
		assert.NoError(t, validation.Validate("abc", rule))
		assert.Error(t, validation.Validate("", rule))
		assert.Error(t, validation.Validate("zzz", rule))
		assert.NoError(t, validation.Validate(123, rule))
		assert.NoError(t, validation.Validate([]byte("abc"), rule))
		errortest.AssertErrorDescription(t, validation.Validate("x", Pattern(nil)), "pattern regexp must not be nil")
	})

	t.Run("string lengths", func(t *testing.T) {
		assert.NoError(t, validation.Validate("hello", MaxLength(5)))
		assert.Error(t, validation.Validate("hello!", MaxLength(5)))
		assert.NoError(t, validation.Validate("hello", MinLength(5)))
		assert.Error(t, validation.Validate("hell", MinLength(5)))
		errortest.AssertErrorDescription(t, validation.Validate("", MinLength(1)), "minimum")
		errortest.AssertErrorDescription(t, validation.Validate("", RuneLengthRule(field.ToOptionalInt(1), nil)), "minimum")
		assert.NoError(t, validation.Validate("éé", LengthRule(nil, field.ToOptionalInt(2))))
		assert.Error(t, validation.Validate("ééé", LengthRule(nil, field.ToOptionalInt(2))))
		assert.NoError(t, validation.Validate([]int{1}, MinLength(2)))
		assert.NoError(t, validation.Validate([]int{1}, MaxLength(2)))
	})

	t.Run("items", func(t *testing.T) {
		type decision bool
		type singlePassSequence func(func(int) decision)

		assert.NoError(t, validation.Validate([]any{"a", 1}, PrefixItems(Type("string"), Type("integer"))))
		assert.Error(t, validation.Validate([]any{1, "a"}, PrefixItems(Type("string"), Type("integer"))))

		var stringPtr *string
		errortest.AssertErrorDescription(t, validation.Validate(stringPtr, PrefixItems(Type("string"))), "must be an array or slice")

		var direct **[]int
		assert.NoError(t, validation.Validate(direct, PrefixItems(Type("integer"))))
		var inner *[]int
		wrapped := &inner
		assert.NoError(t, validation.Validate(wrapped, PrefixItems(Type("integer"))))

		var nilFunc func() = nil
		errortest.AssertErrorDescription(t, validation.Validate(nilFunc, PrefixItems(Type("string"))), "must be an array or slice")

		var nilMap map[int]string
		errortest.AssertErrorDescription(t, validation.Validate(nilMap, UniqueItems[string](func(item string) string { return item })), "must be an array or slice")
		for i := range invalidCollectionInputs {
			invalid := invalidCollectionInputs[i]
			t.Run("invalid input "+invalid.name, func(t *testing.T) {
				errortest.AssertErrorDescription(t, validation.Validate(invalid.value, PrefixItems(Type("string"))), "must be an array or slice")
				errortest.AssertErrorDescription(t, validation.Validate(invalid.value, UniqueItems[string](func(item string) string { return item })), "must be an array or slice")
			})
		}
		assert.NoError(t, validation.Validate([]int{1, 2}, MaxItems(2)))
		assert.Error(t, validation.Validate([]int{1, 2, 3}, MaxItems(2)))
		assert.NoError(t, validation.Validate([]int{1, 2}, MinItems(2)))
		assert.Error(t, validation.Validate([]int{1}, MinItems(2)))
		errortest.AssertErrorDescription(t, validation.Validate([]int{}, MinItems(1)), "minimum")
		assert.NoError(t, validation.Validate("a", MinItems(2)))
		assert.NoError(t, validation.Validate("a", MaxItems(2)))
		assert.NoError(t, validation.Validate([]int{1, 2, 3}, UniqueItems[int](collection.IdentityMapFunc[int]())))
		assert.Error(t, validation.Validate([]int{1, 2, 1}, UniqueItems[int](collection.IdentityMapFunc[int]())))
		assert.NoError(t, validation.Validate([]string{"a", "A"}, UniqueItems[string](func(item string) string { return item })))
		assert.Error(t, validation.Validate([]string{"a", "A"}, UniqueItems[string](strings.ToLower)))
		errortest.AssertErrorDescription(t, validation.Validate([]int{1}, UniqueItems[int, int](nil)), "unique item key function must not be nil")
		assert.NoError(t, validation.Validate([]int{1, 2}, Type("array")))
		assert.Error(t, validation.Validate([]int{1, 2}, Type("object")))
		assert.Error(t, validation.Validate(math.NaN(), Type("number")))
		assert.Error(t, validation.Validate(math.Inf(1), Type("number")))
		var fn func()
		assert.Error(t, validation.Validate(fn, Type("null")))

		value := singlePassSequence(func(yield func(int) decision) {
			yield(1)
		})
		require.NotPanics(t, func() {
			errortest.AssertErrorDescription(t, validation.Validate(value, PrefixItems(Type("integer"))), "must be an array or slice")
		})
	})

	t.Run("properties", func(t *testing.T) {
		nonStringKeyMap := map[int]string{1: faker.Word()}
		interfaceKeyMap := map[any]any{"a": 1, "": 2}
		type namedString string
		namedStringKeyMap := map[namedString]any{"a": 1, "": 2}
		assert.NoError(t, validation.Validate(map[string]any{"a": 1, "b": 2}, MaxProperties(2)))
		assert.Error(t, validation.Validate(map[string]any{"a": 1, "b": 2, "c": 3}, MaxProperties(2)))
		assert.NoError(t, validation.Validate(map[string]any{"a": 1, "b": 2}, MinProperties(2)))
		assert.Error(t, validation.Validate(map[string]any{"a": 1}, MinProperties(2)))
		errortest.AssertErrorDescription(t, validation.Validate(map[string]any{}, MinProperties(1)), "minimum")
		assert.NoError(t, validation.Validate([]int{1}, MinProperties(2)))
		assert.NoError(t, validation.Validate([]int{1}, MaxProperties(2)))
		type zeroStruct struct {
			Enabled bool
			Count   int
		}
		type zeroSizedFields struct {
			A struct{}
			B struct{}
		}
		assert.NoError(t, validation.Validate(struct{ A int }{A: 1}, MaxProperties(2)))
		assert.Error(t, validation.Validate(struct{ A int }{A: 1}, MinProperties(2)))
		assert.NoError(t, validation.Validate(zeroStruct{}, RequiredProperties("Enabled", "Count")))
		assert.NoError(t, validation.Validate(map[string]any{"a": 1}, RequiredProperties("a")))
		assert.Error(t, validation.Validate(map[string]any{"a": 1}, RequiredProperties("a", "b")))
		assert.NoError(t, validation.Validate(map[string]any{"a": 1, "b": 2}, DependentRequired(map[string][]string{"a": {"b"}})))
		assert.Error(t, validation.Validate(map[string]any{"a": 1}, DependentRequired(map[string][]string{"a": {"b"}})))
		assert.NoError(t, validation.Validate(map[string]any{"a": 1, "b": 2}, DependentSchemas(map[string]validation.Rule{"a": RequiredProperties("b")})))
		assert.Error(t, validation.Validate(map[string]any{"a": 1}, DependentSchemas(map[string]validation.Rule{"a": RequiredProperties("b")})))

		re := regexp.MustCompile(`^[a-z]+$`)
		assert.NoError(t, validation.Validate(map[string]any{"alpha": 1}, PropertyNames(Pattern(re))))
		assert.Error(t, validation.Validate(map[string]any{"Alpha": 1}, PropertyNames(Pattern(re))))
		errortest.AssertErrorDescription(t, validation.Validate(map[string]any{"alpha": 1}, PropertyNames(nil)), "propertyNames rule must not be nil")
		assert.NoError(t, validation.Validate(interfaceKeyMap, RequiredProperties("a")))
		assert.NoError(t, validation.Validate(interfaceKeyMap, RequiredProperties("")))
		assert.NoError(t, validation.Validate(namedStringKeyMap, RequiredProperties("a")))
		assert.NoError(t, validation.Validate(namedStringKeyMap, RequiredProperties("")))
		assert.Error(t, validation.Validate(nonStringKeyMap, RequiredProperties("a")))
		var directMap **map[string]any
		assert.NoError(t, validation.Validate(directMap, RequiredProperties("a")))
		var innerMap *map[string]any
		wrappedMap := &innerMap
		assert.NoError(t, validation.Validate(wrappedMap, RequiredProperties("a")))
		require.NotPanics(t, func() {
			assert.NoError(t, validation.Validate(nonStringKeyMap, MaxProperties(2)))
			assert.Error(t, validation.Validate(nonStringKeyMap, MinProperties(2)))
			assert.NoError(t, validation.Validate(nonStringKeyMap, DependentRequired(map[string][]string{"a": {"b"}})))
			assert.NoError(t, validation.Validate(nonStringKeyMap, DependentSchemas(map[string]validation.Rule{"a": RequiredProperties("b")})))
			assert.NoError(t, validation.Validate(nonStringKeyMap, PropertyNames(Pattern(re))))
			assert.NoError(t, validation.Validate(nonStringKeyMap, PatternProperties(
				PatternProperty{Pattern: regexp.MustCompile(`^s_`), Rule: Type("string")},
			)))
			assert.NoError(t, validation.Validate(nonStringKeyMap, AdditionalProperties("a", "b")))
			assert.NoError(t, validation.Validate(nonStringKeyMap, MutuallyExclusiveWith("a", "b")))
			assert.NoError(t, validation.Validate(nonStringKeyMap, AtMostOneProperty("a", "b")))
			assert.Error(t, validation.Validate(nonStringKeyMap, OneOfProperties("a", "b")))
			assert.Error(t, validation.Validate(nonStringKeyMap, AtLeastOneProperty("a", "b")))
			assert.NoError(t, validation.Validate(nonStringKeyMap, ForbiddenProperties("a")))
			assert.NoError(t, validation.Validate(interfaceKeyMap, PropertyNames(Pattern(regexp.MustCompile(`^$|^[a-z]+$`)))))
			assert.NoError(t, validation.Validate(namedStringKeyMap, PropertyNames(Pattern(regexp.MustCompile(`^$|^[a-z]+$`)))))
			assert.NoError(t, validation.Validate(interfaceKeyMap, AdditionalProperties("a", "")))
			assert.NoError(t, validation.Validate(namedStringKeyMap, AdditionalProperties("a", "")))
			assert.NoError(t, validation.Validate(interfaceKeyMap, MutuallyExclusiveWith("a", "b")))
			assert.NoError(t, validation.Validate(namedStringKeyMap, MutuallyExclusiveWith("a", "b")))
		})
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

		used := false
		singleUseSeq := iter.Seq2[string, any](func(yield func(string, any) bool) {
			if used {
				return
			}
			used = true
			_ = yield("a", 1)
			_ = yield("b", 2)
		})
		assert.NoError(t, validation.Validate(singleUseSeq, DependentSchemas(map[string]validation.Rule{"a": RequiredProperties("b")})))

		used = false
		singleUseSeq = iter.Seq2[string, any](func(yield func(string, any) bool) {
			if used {
				return
			}
			used = true
			_ = yield("a", 1)
			_ = yield("b", 2)
		})
		assert.NoError(t, validation.Validate(singleUseSeq, DependentSchemasBy(map[any]validation.Rule{"a": RequiredProperties("b")})))

		used = false
		singleUseSeq = iter.Seq2[string, any](func(yield func(string, any) bool) {
			if used {
				return
			}
			used = true
			_ = yield("mode", "strict")
			_ = yield("name", "alice")
		})
		assert.NoError(t, validation.Validate(singleUseSeq, WhenPropertyEquals("mode", "strict", RequiredProperties("name"))))

		zeroSized := &zeroSizedFields{}
		assert.Error(t, validation.Validate(zeroSized, AdditionalPropertiesBy(&zeroSized.A, &zeroSized.B)))

		type dependenciesStruct struct {
			A int
			B int
			C int
		}
		dependenciesValue := &dependenciesStruct{A: 1, B: 1}
		normalisedDependencies, err := propertyDependenciesForValue(dependenciesValue, map[any]any{
			"A":                  []string{"B"},
			&dependenciesValue.A: []string{"C"},
		})
		require.NoError(t, err)
		require.Contains(t, normalisedDependencies, "A")
		assert.ElementsMatch(t, []string{"B", "C"}, normalisedDependencies["A"])

		firstRuleCalled := false
		secondRuleCalled := false
		normalisedRules, err := propertyRulesForValue(dependenciesValue, map[any]validation.Rule{
			"A": validation.By(func(value any) error {
				firstRuleCalled = true
				return nil
			}),
			&dependenciesValue.A: validation.By(func(value any) error {
				secondRuleCalled = true
				return nil
			}),
		})
		require.NoError(t, err)
		require.Contains(t, normalisedRules, "A")
		require.NoError(t, normalisedRules["A"].Validate(dependenciesValue))
		assert.True(t, firstRuleCalled)
		assert.True(t, secondRuleCalled)

		type withUnexported struct {
			Exported   string
			unexported string
		}
		assert.NoError(t, validation.Validate(withUnexported{Exported: "a", unexported: "b"}, RequiredProperties("Exported")))
		assert.Error(t, validation.Validate(withUnexported{Exported: "a", unexported: "b"}, RequiredProperties("unexported")))
		require.NotPanics(t, func() {
			assert.NoError(t, validation.Validate(withUnexported{Exported: "a", unexported: "b"}, AdditionalProperties("Exported")))
			assert.NoError(t, validation.Validate(withUnexported{Exported: "a", unexported: "b"}, PatternProperties(
				PatternProperty{Pattern: regexp.MustCompile(`^u`), Rule: Type("integer")},
			)))
			assert.NoError(t, validation.Validate(withUnexported{Exported: "a", unexported: "b"}, MutuallyExclusiveWith("unexported", "other")))
		})

		type embedded struct{ Value string }
		type withNilEmbedded struct {
			*embedded
			Name string
		}
		type tagged struct {
			Name string `json:"name"`
		}
		assert.NoError(t, validation.Validate(tagged{Name: "alice"}, RequiredProperties("name")))
		assert.NoError(t, validation.Validate(tagged{Name: "alice"}, AdditionalProperties("name")))
		require.NotPanics(t, func() {
			assert.NoError(t, validation.Validate(withNilEmbedded{Name: "ok"}, RequiredProperties("Name")))
			assert.Error(t, validation.Validate(withNilEmbedded{Name: "ok"}, RequiredProperties("Value")))
			assert.NoError(t, validation.Validate(withNilEmbedded{Name: "ok"}, PatternProperties(
				PatternProperty{Pattern: regexp.MustCompile(`^V`), Rule: Type("string")},
			)))
		})

		var stringPtr *string
		errortest.AssertErrorDescription(t, validation.Validate(stringPtr, RequiredProperties("a")), "must be a map")

		var nilFunc func() = nil
		errortest.AssertErrorDescription(t, validation.Validate(nilFunc, AdditionalProperties("a", "b")), "must be a map")
		for i := range invalidObjectInputs {
			invalid := invalidObjectInputs[i]
			t.Run("invalid object input "+invalid.name, func(t *testing.T) {
				errortest.AssertErrorDescription(t, validation.Validate(invalid.value, RequiredProperties("a")), "must be a map")
				errortest.AssertErrorDescription(t, validation.Validate(invalid.value, AdditionalProperties("a", "b")), "must be a map")
			})
		}
	})

	t.Run("contains", func(t *testing.T) {
		assert.NoError(t, validation.Validate([]string{"a", "b"}, Contains(Const("a"))))
		assert.Error(t, validation.Validate([]string{"b", "c"}, Contains(Const("a"))))
		assert.NoError(t, validation.Validate([]string{"a", "b", "a"}, MinContains(2, Const("a"))))
		assert.Error(t, validation.Validate([]string{"a", "b"}, MinContains(2, Const("a"))))
		assert.NoError(t, validation.Validate([]string{"a", "b"}, MaxContains(1, Const("a"))))
		assert.Error(t, validation.Validate([]string{"a", "a"}, MaxContains(1, Const("a"))))
		errortest.AssertErrorDescription(t, validation.Validate([]int{1}, Contains(nil)), "contains rule must not be nil")

		var stringPtr *string
		errortest.AssertErrorDescription(t, validation.Validate(stringPtr, MinContains(1, Const("a"))), "must be an array or slice")

		var nilFunc func() = nil
		errortest.AssertErrorDescription(t, validation.Validate(nilFunc, MaxContains(1, Const("a"))), "must be an array or slice")

		var nilMap map[int]string
		errortest.AssertErrorDescription(t, validation.Validate(nilMap, Contains(Const("a"))), "must be an array or slice")
		for i := range invalidCollectionInputs {
			invalid := invalidCollectionInputs[i]
			t.Run("invalid contains input "+invalid.name, func(t *testing.T) {
				errortest.AssertErrorDescription(t, validation.Validate(invalid.value, Contains(Const("a"))), "must be an array or slice")
				errortest.AssertErrorDescription(t, validation.Validate(invalid.value, MinContains(1, Const("a"))), "must be an array or slice")
				errortest.AssertErrorDescription(t, validation.Validate(invalid.value, MaxContains(1, Const("a"))), "must be an array or slice")
			})
		}
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
		for i := range invalidObjectInputs {
			invalid := invalidObjectInputs[i]
			t.Run("invalid mutually exclusive input "+invalid.name, func(t *testing.T) {
				errortest.AssertErrorDescription(t, validation.Validate(invalid.value, MutuallyExclusiveWith("A", "B")), "must be a map")
			})
		}
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
		keyA := faker.Word()
		keyB := keyA + "-" + faker.Word()
		keyC := keyB + "-" + faker.Word()

		assert.NoError(t, validation.Validate([]valueT{{i: 0, j: keyA}}, AtMostOneItemKey(keyFunc, keyA, keyB)))
		assert.Error(t, validation.Validate([]valueT{{i: 0, j: keyA}, {i: 1, j: keyB}}, AtMostOneItemKey(keyFunc, keyA, keyB)))
		assert.NoError(t, validation.Validate([]valueT{{i: 0, j: keyA}, {i: 1, j: keyB}}, RequiredItemKeys(keyFunc, keyA, keyB)))
		assert.Error(t, validation.Validate([]valueT{{i: 0, j: keyA}}, RequiredItemKeys(keyFunc, keyA, keyB)))
		assert.NoError(t, validation.Validate([]valueT{{i: 0, j: keyA}, {i: 1, j: keyB}}, DependentRequiredItemKeys(keyFunc, map[string][]string{keyA: {keyB}})))
		assert.Error(t, validation.Validate([]valueT{{i: 0, j: keyA}}, DependentRequiredItemKeys(keyFunc, map[string][]string{keyA: {keyB}})))
		assert.NoError(t, validation.Validate([]valueT{{i: 0, j: keyA}}, AdditionalItemKeys(keyFunc, keyA, keyB)))
		assert.Error(t, validation.Validate([]valueT{{i: 0, j: keyC}}, AdditionalItemKeys(keyFunc, keyA, keyB)))
		assert.NoError(t, validation.Validate([]valueT{{i: 0, j: keyA}}, MutuallyExclusiveItemKeys(keyFunc, keyA, keyB)))
		assert.Error(t, validation.Validate([]valueT{{i: 0, j: keyA}, {i: 1, j: keyB}}, MutuallyExclusiveItemKeys(keyFunc, keyA, keyB)))
		assert.NoError(t, validation.Validate([]valueT{{i: 0, j: keyA}}, OneOfItemKeys(keyFunc, keyA, keyB)))
		assert.Error(t, validation.Validate([]valueT{}, OneOfItemKeys(keyFunc, keyA, keyB)))
		assert.Error(t, validation.Validate([]valueT{{i: 0, j: keyA}, {i: 1, j: keyB}}, OneOfItemKeys(keyFunc, keyA, keyB)))
		assert.NoError(t, validation.Validate([]valueT{{i: 0, j: keyA}}, AtLeastOneItemKey(keyFunc, keyA, keyB)))
		assert.Error(t, validation.Validate([]valueT{}, AtLeastOneItemKey(keyFunc, keyA, keyB)))
		assert.NoError(t, validation.Validate([]valueT{{i: 0, j: keyA}}, ForbiddenItemKeys(keyFunc, keyB)))
		assert.Error(t, validation.Validate([]valueT{{i: 0, j: keyA}}, ForbiddenItemKeys(keyFunc, keyA)))
		assert.Error(t, validation.Validate([]string{"A", "B"}, DependentRequiredItems(strings.ToLower, map[string][]string{
			"A": {"B"},
			"a": {"C"},
		})))
		assert.NoError(t, validation.Validate([]string{"A", "B", "C"}, DependentRequiredItems(strings.ToLower, map[string][]string{
			"A": {"B"},
			"a": {"C"},
		})))

		identityAny := func(value any) any { return value }
		require.NotPanics(t, func() {
			err := validation.Validate([]any{[]int{1}}, UniqueItems[any, any](identityAny))
			assert.Error(t, err)
			assert.True(t, commonerrors.Any(err, commonerrors.ErrInvalid))
		})
		require.NotPanics(t, func() {
			err := validation.Validate([]any{[]int{1}}, RequiredItemKeys[any, any](identityAny, []int{1}))
			assert.Error(t, err)
			assert.True(t, commonerrors.Any(err, commonerrors.ErrInvalid))
		})
		require.NotPanics(t, func() {
			err := validation.Validate([]any{[]int{1}}, ForbiddenItemKeys[any, any](identityAny, []int{1}))
			assert.Error(t, err)
			assert.True(t, commonerrors.Any(err, commonerrors.ErrInvalid))
		})
	})

	t.Run("schema terminology aliases", func(t *testing.T) {
		assert.NoError(t, validation.Validate("blue", Enum("red", "blue")))
		assert.Error(t, validation.Validate("green", Enum("red", "blue")))
		assert.Error(t, validation.Validate("", Enum("red", "blue")))
		assert.Error(t, validation.Validate(nil, Enum("value")))
		assert.Error(t, validation.Validate(0, Enum(1, 2)))
		assert.Error(t, validation.Validate(0, Enum()))
		assert.NoError(t, validation.Validate(float64(1), Enum(1, 2)))

		assert.NoError(t, validation.Validate("v1", Const("v1")))
		assert.Error(t, validation.Validate("v2", Const("v1")))
		assert.Error(t, validation.Validate("", Const("v1")))
		assert.Error(t, validation.Validate(false, Const(true)))
		assert.NoError(t, validation.Validate(float64(1), Const(1)))

		re := regexp.MustCompile(`^[a-z]+$`)
		assert.NoError(t, validation.Validate("abc", Pattern(re)))
		assert.Error(t, validation.Validate("123", Pattern(re)))
		assert.NoError(t, validation.Validate(3, XIntOrString()))
		assert.NoError(t, validation.Validate("3", XIntOrString()))
		assert.NoError(t, validation.Validate(float64(3), XIntOrString()))
		assert.Error(t, validation.Validate(float64(3.5), XIntOrString()))

		assert.NoError(t, validation.Validate("plain-text", Not(is.Email)))
		assert.Error(t, validation.Validate("user@example.com", Not(is.Email)))
		assert.NoError(t, validation.Validate("", Not(Const("v1"))))
		assert.NoError(t, validation.Validate(nil, Nullable(Type("string"))))
		assert.NoError(t, validation.Validate("hello", Nullable(Type("string"))))
		assert.Error(t, validation.Validate(1, Nullable(Type("string"))))
		errortest.AssertErrorDescription(t, validation.Validate("hello", Nullable(nil)), "nullable rule must not be nil")

		assert.NoError(t, validation.Validate("user@example.com", AnyOf(is.Email, is.UUID)))
		assert.Error(t, validation.Validate("not-valid", AnyOf(is.Email, is.UUID)))
		require.NotPanics(t, func() {
			assert.Error(t, validation.Validate("plain-text", AnyOf(nil)))
		})

		assert.NoError(t, validation.Validate("user@example.com", AllOf(validation.Required, is.Email)))
		assert.Error(t, validation.Validate("", AllOf(validation.Required, is.Email)))
		require.NotPanics(t, func() {
			assert.NoError(t, validation.Validate("plain-text", AllOf(nil)))
		})

		assert.NoError(t, validation.Validate("plain-text", NoneOf(is.Email, is.UUID)))
		assert.Error(t, validation.Validate("user@example.com", NoneOf(is.Email, is.UUID)))
		require.NotPanics(t, func() {
			assert.NoError(t, validation.Validate("plain-text", NoneOf(nil)))
		})

		assert.NoError(t, validation.Validate("user@example.com", OneOf(is.Email, is.UUID)))
		assert.Error(t, validation.Validate("user@example.com", OneOf(validation.Required, is.Email)))
		require.NotPanics(t, func() {
			assert.Error(t, validation.Validate("plain-text", OneOf(nil)))
		})

		assert.NoError(t, validation.Validate("hello", NotEmpty()))
		assert.Error(t, validation.Validate("   ", NotEmpty()))
	})
}

// Integers above 2^53 cannot all be represented exactly as float64 because IEEE-754 float64 has 53 bits of integer precision.
func TestMaximumPreservesLargeIntegerPrecision(t *testing.T) {
	maximum := uint64(math.Exp2(53))
	larger := maximum + 1

	assert.Error(t, validation.Validate(larger, Maximum(maximum)))
	assert.NoError(t, validation.Validate(maximum, ExclusiveMaximum(larger)))
	assert.Error(t, validation.Validate(maximum, Minimum(larger)))
	assert.NoError(t, validation.Validate(larger, ExclusiveMinimum(maximum)))
	assert.Error(t, validation.Validate(larger, Enum(maximum)))
	assert.Error(t, validation.Validate(larger, Const(maximum)))
	assert.Error(t, validation.Validate(larger, MultipleOf(float64(2))))
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
		assert.NoError(t, validation.Validate(value, RequiredPropertiesBy(&value.A, &value.C)))
		assert.NoError(t, validation.Validate(value, AdditionalPropertiesBy(&value.A, "B", &value.C)))
		assert.NoError(t, validation.Validate(value, AdditionalPropertiesBy([]string{"A", "B", "C"})))
		assert.Error(t, validation.Validate(value, AdditionalPropertiesBy(&value.A)))
	})

	t.Run("dependent properties and schemas", func(t *testing.T) {
		value := &fields{A: 1, B: 2, C: 3}
		assert.NoError(t, validation.Validate(value, DependentRequiredBy(map[any]any{&value.A: []any{&value.B, &value.C}})))
		zeroValueDependency := &fields{A: 1, B: 2}
		assert.NoError(t, validation.Validate(zeroValueDependency, DependentRequiredBy(map[any]any{&zeroValueDependency.A: []any{&zeroValueDependency.B, &zeroValueDependency.C}})))
		assert.NoError(t, validation.Validate(value, DependentSchemasBy(map[any]validation.Rule{&value.A: RequiredPropertiesBy(&value.B)})))
		zeroValueSchemaDependency := &fields{A: 1}
		assert.NoError(t, validation.Validate(zeroValueSchemaDependency, DependentSchemasBy(map[any]validation.Rule{&zeroValueSchemaDependency.A: RequiredPropertiesBy(&zeroValueSchemaDependency.B)})))
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

		emptyValue := &fields{}
		err := validation.Validate(emptyValue, OneOfPropertiesBy(&emptyValue.A, &emptyValue.B))
		errortest.AssertErrorDescription(t, err, "mutually exclusive")
		err = validation.Validate(emptyValue, AtLeastOnePropertyBy(&emptyValue.A, &emptyValue.B))
		errortest.AssertErrorDescription(t, err, "required properties")
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
