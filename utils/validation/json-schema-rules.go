package validation

// json-schema-rules.go contains validation helpers whose naming and behaviour
// are intentionally aligned with JSON Schema vocabulary where that can be
// expressed cleanly on top of ozzo-validation.
//
// Some helpers are thin aliases over existing ozzo rules, while others provide
// small convenience layers for applying JSON-Schema-like constraints to generic
// Go values such as decoded maps, slices, and structs.
//
// References:
//   - JSON Schema reference:
//     https://json-schema.org/understanding-json-schema/
//   - OpenAPI Schema Object:
//     https://spec.openapis.org/oas/latest.html#schema-object

import (
	"math"
	"reflect"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	utilreflection "github.com/ARM-software/golang-utils/utils/reflection"
)

var (
	errArrayOrSliceRequired = validation.NewError("validation_array_or_slice_required", "must be an array or slice")
	errMapRequired          = validation.NewError("validation_map_required", "must be a map")
	errUniqueItems          = validation.NewError("validation_unique_items", "must contain unique items")
	errMutuallyExclusive    = validation.NewError("validation_mutually_exclusive", "must not define more than one mutually exclusive field")
	errContains             = validation.NewError("validation_contains", "must contain a matching item")
	errRequiredProperties   = validation.NewError("validation_required_properties", "must define all required properties")
	errDependentRequired    = validation.NewError("validation_dependent_required", "must define dependent properties")
	errDependentSchemas     = validation.NewError("validation_dependent_schemas", "must satisfy dependent schemas")
	errPropertyNames        = validation.NewError("validation_property_names", "contains an invalid property name")
	errPatternProperties    = validation.NewError("validation_pattern_properties", "contains an invalid property value for a matching property pattern")
	errAdditionalProperties = validation.NewError("validation_additional_properties", "contains unsupported additional properties")
	errType                 = validation.NewError("validation_type", "must be of an allowed type")
	errIntOrString          = validation.NewError("validation_int_or_string", "must be an integer or a string")
)

// PatternProperty couples a property-name pattern with the rule that should be
// applied to matching properties.
type PatternProperty struct {
	Pattern *regexp.Regexp
	Rule    validation.Rule
}

// Type validates the JSON Schema `type` constraint for decoded Go values.
//
// Supported schema type names are: `string`, `number`, `integer`, `object`,
// `array`, `boolean`, and `null`.
//
// Example: `Type("string", "null")` accepts a string value or nil.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/type
func Type(types ...string) validation.Rule {
	normalised := collection.Map(types, strings.ToLower)
	return validation.By(func(value any) error {
		v, isNil := validation.Indirect(value)
		if isNil {
			if collection.In(normalised, "null", collection.StringCaseInsensitiveMatch) {
				return nil
			}
			return errType
		}

		kind := reflect.ValueOf(v).Kind()
		matched := collection.AnyFunc(normalised, func(expected string) bool {
			switch expected {
			case "string":
				return kind == reflect.String
			case "number":
				return collection.In([]reflect.Kind{reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64}, kind, func(a, b reflect.Kind) (bool, error) { return a == b, nil })
			case "integer":
				return collection.In([]reflect.Kind{reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr}, kind, func(a, b reflect.Kind) (bool, error) { return a == b, nil })
			case "object":
				return kind == reflect.Map || kind == reflect.Struct
			case "array":
				return kind == reflect.Array || kind == reflect.Slice
			case "boolean":
				return kind == reflect.Bool
			case "null":
				return false
			default:
				return false
			}
		})
		if !matched {
			return errType
		}
		return nil
	})
}

// MultipleOf validates the JSON Schema `multiple_of` constraint.
//
// Example: `MultipleOf(5)` accepts `10` and rejects `11`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/numeric#multiples
func MultipleOf(base any) validation.Rule {
	return validation.MultipleOf(base)
}

// Maximum validates the JSON Schema `maximum` constraint.
//
// Example: `Maximum(10)` accepts `10` and rejects `11`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/numeric#range
func Maximum(max any) validation.Rule {
	return validation.Max(max)
}

// ExclusiveMaximum validates the JSON Schema `exclusive_maximum` constraint.
//
// Example: `ExclusiveMaximum(10)` accepts `9` and rejects `10`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/numeric#range
func ExclusiveMaximum(max any) validation.Rule {
	return validation.Max(max).Exclusive()
}

// Minimum validates the JSON Schema `minimum` constraint.
//
// Example: `Minimum(10)` accepts `10` and rejects `9`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/numeric#range
func Minimum(min any) validation.Rule {
	return validation.Min(min)
}

// ExclusiveMinimum validates the JSON Schema `exclusive_minimum` constraint.
//
// Example: `ExclusiveMinimum(10)` accepts `11` and rejects `10`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/numeric#range
func ExclusiveMinimum(min any) validation.Rule {
	return validation.Min(min).Exclusive()
}

// MaxLength validates the JSON Schema `max_length` constraint.
//
// This counts Unicode code points rather than bytes.
// Example: `MaxLength(5)` accepts `hello` and rejects `hello!`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/string#length
func MaxLength(max int) validation.Rule {
	return RuneLengthRule(nil, field.ToOptionalInt(max))
}

// MinLength validates the JSON Schema `min_length` constraint.
//
// This counts Unicode code points rather than bytes.
// Example: `MinLength(5)` accepts `hello` and rejects `hell`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/string#length
func MinLength(min int) validation.Rule {
	return RuneLengthRule(field.ToOptionalInt(min), nil)
}

// MaxItems validates the JSON Schema `max_items` constraint.
//
// Example: `MaxItems(2)` accepts `[1,2]` and rejects `[1,2,3]`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/array#length
func MaxItems(max int) validation.Rule {
	return LengthRule(nil, field.ToOptionalInt(max))
}

// MinItems validates the JSON Schema `min_items` constraint.
//
// Example: `MinItems(2)` accepts `[1,2]` and rejects `[1]`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/array#length
func MinItems(min int) validation.Rule {
	return LengthRule(field.ToOptionalInt(min), nil)
}

// PrefixItems validates the JSON Schema `prefixItems` constraint.
//
// Each rule is applied to the item at the same index. Extra items are ignored.
//
// Example: `PrefixItems(Type("string"), Type("integer"))` accepts
// `[]any{"a", 1}`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/array#tuple-validation
func PrefixItems(rules ...validation.Rule) validation.Rule {
	return validation.By(func(value any) error {
		items, err := typedSequence[any](value)
		if err != nil || items == nil {
			return err
		}
		for i := 0; i < len(items) && i < len(rules); i++ {
			if rules[i] == nil {
				continue
			}
			if subErr := rules[i].Validate(items[i]); subErr != nil {
				return subErr
			}
		}
		return nil
	})
}

// Contains validates the JSON Schema `contains` constraint.
//
// The rule succeeds when at least one item in the array or slice satisfies rule.
// Example: `Contains(Const("a"))` accepts `[]string{"a", "b"}`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/array#contains
func Contains(rule validation.Rule) validation.Rule {
	return MinContains(1, rule)
}

// MinContains validates the JSON Schema `minContains` constraint.
//
// The rule succeeds when at least min items in the array or slice satisfy rule.
// Example: `MinContains(2, Const("a"))` accepts `[]string{"a", "b", "a"}`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/array#contains
func MinContains(min int, rule validation.Rule) validation.Rule {
	return validation.By(func(value any) error {
		items, err := typedSequence[any](value)
		if err != nil || items == nil {
			return err
		}
		matches := collection.CountBy(items, func(item any) bool {
			return rule.Validate(item) == nil
		})
		if matches < min {
			return errContains
		}
		return nil
	})
}

// MaxContains validates the JSON Schema `maxContains` constraint.
//
// The rule succeeds when at most max items in the array or slice satisfy rule.
// Example: `MaxContains(1, Const("a"))` rejects `[]string{"a", "a"}`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/array#contains
func MaxContains(max int, rule validation.Rule) validation.Rule {
	return validation.By(func(value any) error {
		items, err := typedSequence[any](value)
		if err != nil || items == nil {
			return err
		}
		matches := collection.CountBy(items, func(item any) bool {
			return rule.Validate(item) == nil
		})
		if matches > max {
			return errContains
		}
		return nil
	})
}

// UniqueItems validates the JSON Schema `unique_items` constraint using a key
// function to decide whether two items should be considered the same.
//
// The provided key function should return a comparable identity for each item.
// If the number of items changes after applying [collection.UniqueBy], the input
// contains duplicates.
//
// Example: `UniqueItems[string](strings.ToLower)` rejects `[]string{"a", "A"}`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/array#uniqueness
func UniqueItems[T any, K comparable](keyFunc collection.KeyFunc[T, K]) validation.Rule {
	return validation.By(func(value any) error {
		items, err := typedSequence[T](value)
		if err != nil || items == nil {
			return err
		}
		if len(collection.UniqueBy(items, keyFunc)) != len(items) {
			return errUniqueItems
		}
		return nil
	})
}

// MaxProperties validates the JSON Schema `max_properties` constraint.
//
// Example: `MaxProperties(2)` accepts a map with two keys and rejects one with three.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/object#size
func MaxProperties(max int) validation.Rule {
	return LengthRule(nil, field.ToOptionalInt(max))
}

// MinProperties validates the JSON Schema `min_properties` constraint.
//
// Example: `MinProperties(2)` accepts a map with two keys and rejects one with one.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/object#size
func MinProperties(min int) validation.Rule {
	return LengthRule(field.ToOptionalInt(min), nil)
}

// RequiredProperties validates the JSON Schema `required` constraint for object
// properties.
//
// For map values, a property is considered present if the key exists. For
// structs, a property is considered present if a field of that name exists and
// its value is not empty according to the repository's reflection helpers.
//
// Example: `RequiredProperties("a", "b")` rejects `map[string]any{"a": 1}`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/object#required
func RequiredProperties(keys ...string) validation.Rule {
	normalizedKeys := collection.UniqueEntries(keys)
	return validation.By(func(value any) error {
		rv, isNil, err := objectValue(value)
		if err != nil || isNil {
			return err
		}
		missing := collection.CountBy(normalizedKeys, func(key string) bool {
			return !hasObjectProperty(rv, key)
		})
		if missing > 0 {
			return errRequiredProperties
		}
		return nil
	})
}

// DependentRequired validates the JSON Schema `dependentRequired` constraint.
//
// For each trigger property key in dependencies, if that property is present,
// all listed dependent properties must also be present.
//
// Example: `DependentRequired(map[string][]string{"a": {"b"}})` rejects
// `map[string]any{"a": 1}`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/conditionals#dependentrequired
func DependentRequired(dependencies map[string][]string) validation.Rule {
	normalised := make(map[string][]string, len(dependencies))
	for key, dependents := range dependencies {
		normalised[key] = collection.UniqueEntries(dependents)
	}
	return validation.By(func(value any) error {
		rv, isNil, err := objectValue(value)
		if err != nil || isNil {
			return err
		}
		for key, dependents := range normalised {
			if !hasObjectProperty(rv, key) {
				continue
			}
			missing := collection.CountBy(dependents, func(dependent string) bool {
				return !hasObjectProperty(rv, dependent)
			})
			if missing > 0 {
				return errDependentRequired
			}
		}
		return nil
	})
}

// DependentSchemas validates the JSON Schema `dependentSchemas` constraint.
//
// For each trigger property key in dependencies, if that property is present,
// the corresponding rule is applied to the whole object.
//
// Example: `DependentSchemas(map[string]validation.Rule{"a": RequiredProperties("b")})`
// rejects `map[string]any{"a": 1}`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/conditionals#dependentschemas
func DependentSchemas(dependencies map[string]validation.Rule) validation.Rule {
	return validation.By(func(value any) error {
		rv, isNil, err := objectValue(value)
		if err != nil || isNil {
			return err
		}
		for key, rule := range dependencies {
			if !hasObjectProperty(rv, key) {
				continue
			}
			if rule.Validate(value) != nil {
				return errDependentSchemas
			}
		}
		return nil
	})
}

// PropertyNames validates the JSON Schema `propertyNames` constraint.
//
// The supplied rule is applied to every property name in a map or every field
// name in a struct.
//
// Example: `PropertyNames(Pattern(regexp.MustCompile("^[a-z]+$")))` rejects a
// property named `Alpha`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/object#property-names
func PropertyNames(rule validation.Rule) validation.Rule {
	return validation.By(func(value any) error {
		rv, isNil, err := objectValue(value)
		if err != nil || isNil {
			return err
		}
		for _, key := range objectPropertyNames(rv) {
			if rule.Validate(key) != nil {
				return errPropertyNames
			}
		}
		return nil
	})
}

// PatternProperties validates the JSON Schema `patternProperties` constraint.
//
// For each pattern/rule pair, every matching property name in a map or struct
// must have a value that satisfies the associated rule.
//
// Example: a pattern `^s_` with rule `Type("string")` ensures every property
// whose name starts with `s_` contains a string value.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/object#pattern-properties
func PatternProperties(patterns ...PatternProperty) validation.Rule {
	return validation.By(func(value any) error {
		rv, isNil, err := objectValue(value)
		if err != nil || isNil {
			return err
		}
		for _, property := range patterns {
			if property.Pattern == nil || property.Rule == nil {
				continue
			}
			for _, key := range objectPropertyNames(rv) {
				if !property.Pattern.MatchString(key) {
					continue
				}
				fieldValue, found := objectPropertyValue(rv, key)
				if !found {
					continue
				}
				if property.Rule.Validate(fieldValue) != nil {
					return errPatternProperties
				}
			}
		}
		return nil
	})
}

// AdditionalProperties validates that a map or struct contains no property name
// outside the supplied known set.
//
// This is a simplified helper corresponding to the JSON Schema
// `additionalProperties: false` pattern.
//
// Example: `AdditionalProperties("a", "b")` rejects
// `map[string]any{"c": 1}`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/object#additional-properties
func AdditionalProperties(keys ...string) validation.Rule {
	normalizedKeys := collection.UniqueEntries(keys)
	return validation.By(func(value any) error {
		rv, isNil, err := objectValue(value)
		if err != nil || isNil {
			return err
		}
		invalid := collection.CountBy(objectPropertyNames(rv), func(key string) bool {
			return !collection.In(normalizedKeys, key, collection.StringMatch)
		})
		if invalid > 0 {
			return errAdditionalProperties
		}
		return nil
	})
}

// MutuallyExclusiveWith validates that at most one of the named keys or fields
// is set in a map or struct value.
//
// This helper is inspired by non-standard schema-style object constraints and is
// useful when several alternative fields are allowed but must not appear
// together. It is not a direct JSON Schema keyword, but it is often useful for
// OpenAPI-style request and configuration validation.
//
// OpenAPI reference: https://spec.openapis.org/oas/latest.html#schema-object
//
// Example: `MutuallyExclusiveWith("A", "B")` rejects a value where both `A`
// and `B` are non-empty.
func MutuallyExclusiveWith(keys ...string) validation.Rule {
	normalizedKeys := collection.UniqueEntries(keys)
	return validation.By(func(value any) error {
		v, isNil := validation.Indirect(value)
		if isNil {
			return nil
		}
		rv := reflect.ValueOf(v)
		count := 0
		switch rv.Kind() {
		case reflect.Map:
			count = collection.CountBy(normalizedKeys, func(key string) bool {
				mapValue := rv.MapIndex(reflect.ValueOf(key))
				if !mapValue.IsValid() {
					return false
				}
				return !utilreflection.IsEmpty(mapValue.Interface())
			})
		case reflect.Struct:
			count = collection.CountBy(normalizedKeys, func(key string) bool {
				fieldValue := rv.FieldByName(key)
				if !fieldValue.IsValid() {
					return false
				}
				return !utilreflection.IsEmpty(fieldValue.Interface())
			})
		default:
			return errMapRequired
		}
		if count > 1 {
			return errMutuallyExclusive
		}
		return nil
	})
}

// AtMostOneProperty validates that no more than one of the named keys or fields
// is set in a map or struct value.
//
// Example: `AtMostOneProperty("A", "B")` rejects a value where both `A` and
// `B` are non-empty.
//
// This is not a direct JSON Schema keyword, but is useful for OpenAPI-like
// object validation.
//
// OpenAPI reference: https://spec.openapis.org/oas/latest.html#schema-object
func AtMostOneProperty(keys ...string) validation.Rule {
	return MutuallyExclusiveWith(keys...)
}

// OneOfProperties validates that exactly one of the named keys or fields is set
// in a map or struct value.
//
// Example: `OneOfProperties("A", "B")` accepts `{A: 1}` and rejects both
// `{}` and `{A: 1, B: 2}`.
//
// This is not a direct JSON Schema keyword, but is useful for OpenAPI-like
// object validation.
//
// OpenAPI reference: https://spec.openapis.org/oas/latest.html#schema-object
func OneOfProperties(keys ...string) validation.Rule {
	normalizedKeys := collection.UniqueEntries(keys)
	return validation.By(func(value any) error {
		rv, isNil, err := objectValue(value)
		if err != nil || isNil {
			return err
		}
		if countPresentObjectProperties(rv, normalizedKeys) != 1 {
			return errMutuallyExclusive
		}
		return nil
	})
}

// AtLeastOneProperty validates that at least one of the named keys or fields
// is set in a map or struct value.
//
// Example: `AtLeastOneProperty("A", "B")` rejects a value where neither `A`
// nor `B` is set.
//
// This is not a direct JSON Schema keyword, but is useful for OpenAPI-like
// object validation.
//
// OpenAPI reference: https://spec.openapis.org/oas/latest.html#schema-object
func AtLeastOneProperty(keys ...string) validation.Rule {
	normalizedKeys := collection.UniqueEntries(keys)
	return validation.By(func(value any) error {
		rv, isNil, err := objectValue(value)
		if err != nil || isNil {
			return err
		}
		if countPresentObjectProperties(rv, normalizedKeys) == 0 {
			return errRequiredProperties
		}
		return nil
	})
}

// ForbiddenProperties validates that none of the named keys or fields is set in
// a map or struct value.
//
// Example: `ForbiddenProperties("debug")` rejects `{"debug": true}`.
//
// This is not a direct JSON Schema keyword, but is useful for OpenAPI-like
// object validation.
//
// OpenAPI reference: https://spec.openapis.org/oas/latest.html#schema-object
func ForbiddenProperties(keys ...string) validation.Rule {
	normalizedKeys := collection.UniqueEntries(keys)
	return validation.By(func(value any) error {
		rv, isNil, err := objectValue(value)
		if err != nil || isNil {
			return err
		}
		if countPresentObjectProperties(rv, normalizedKeys) > 0 {
			return errAdditionalProperties
		}
		return nil
	})
}

// XIntOrString validates the Kubernetes/OpenAPI `x-kubernetes-int-or-string`
// style constraint.
//
// Example: `XIntOrString()` accepts `3`, `"3"`, and JSON-decoded integer
// numbers represented as `float64(3)`.
func XIntOrString() validation.Rule {
	return validation.By(func(value any) error {
		_, isNil := validation.Indirect(value)
		if isNil {
			return nil
		}
		if isString, _, _, _ := validation.StringOrBytes(value); isString {
			return nil
		}
		if _, err := validation.ToInt(value); err == nil {
			return nil
		}
		if _, err := validation.ToUint(value); err == nil {
			return nil
		}
		if f, err := validation.ToFloat(value); err == nil && math.Trunc(f) == f {
			return nil
		}
		return errIntOrString
	})
}

// Enum validates that a value is one of a fixed set of allowed values.
//
// This is a schema-oriented alias for ozzo's `validation.In(...)` helper.
// Example: `Enum("red", "blue")` accepts `"blue"` and rejects `"green"`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/enum
func Enum(values ...any) validation.Rule {
	return validation.In(values...)
}

// Const validates that a value is exactly equal to expected.
//
// This is useful for schema-style validations where one field must have a fixed
// discriminator or version value.
// Example: `Const("v1")` accepts `"v1"` and rejects `"v2"`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/const
func Const(expected any) validation.Rule {
	return validation.In(expected)
}

// Pattern validates that a string or byte slice matches re.
//
// This is a schema-oriented alias for ozzo's regexp rule.
// Example: `Pattern(regexp.MustCompile("^[a-z]+$"))` accepts `"abc"`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/string#regular-expressions
func Pattern(re *regexp.Regexp) validation.Rule {
	return validation.Match(re)
}

// Not returns a rule that succeeds only if rule fails.
//
// This is a schema-oriented helper corresponding to JSON Schema `not`.
// Example: `Not(is.Email)` rejects `"user@example.com"`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/combining#not
func Not(rule validation.Rule) validation.Rule {
	return NewNoneRule(rule)
}

// Nullable validates that a value is nil or satisfies rule.
//
// Example: `Nullable(Type("string"))` accepts `nil` and `"hello"`.
//
// This is a convenience helper corresponding conceptually to JSON Schema
// patterns such as `type: ["string", "null"]` and to OpenAPI 3.0-style
// `nullable` handling. It is not a direct JSON Schema keyword.
//
// References:
//   - JSON Schema type reference:
//     https://json-schema.org/understanding-json-schema/reference/type
//   - OpenAPI 3.0 Schema Object:
//     https://spec.openapis.org/oas/v3.0.3.html#schema-object
func Nullable(rule validation.Rule) validation.Rule {
	return validation.By(func(value any) error {
		_, isNil := validation.Indirect(value)
		if isNil {
			return nil
		}
		if rule == nil {
			return nil
		}
		return rule.Validate(value)
	})
}

// AnyOf returns a rule that succeeds if at least one nested rule succeeds.
//
// This is a schema-oriented alias for [NewAnyRule].
// Example: `AnyOf(is.Email, is.UUID)` accepts `"user@example.com"`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/combining#anyof
func AnyOf(rules ...validation.Rule) validation.Rule {
	return AtLeast(1, rules...)
}

// AllOf returns a rule that succeeds only if all nested rules succeed.
//
// This is a schema-oriented alias for [NewAllRule].
// Example: `AllOf(validation.Required, is.Email)` requires a non-empty email.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/combining#allof
func AllOf(rules ...validation.Rule) validation.Rule {
	return NewAllRule(rules...)
}

// NoneOf returns a rule that succeeds only if none of the nested rules succeed.
//
// This is a schema-oriented alias for [NewNoneRule].
// Example: `NoneOf(is.Email, is.UUID)` accepts `"plain-text"`.
func NoneOf(rules ...validation.Rule) validation.Rule {
	return NewNoneRule(rules...)
}

// OneOf returns a rule that succeeds only if exactly one nested rule succeeds.
//
// Example: `OneOf(is.Email, is.UUID)` accepts a valid email or a valid UUID,
// but rejects values that satisfy both or neither rule.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/combining#oneof
func OneOf(rules ...validation.Rule) validation.Rule {
	return NewOneOfRule(rules...)
}

// NotEmpty validates that a value is not empty according to the repository's
// reflection-based emptiness semantics.
//
// Example: `NotEmpty()` rejects `"   "`.
func NotEmpty() validation.Rule {
	return validation.By(func(value any) error {
		if utilreflection.IsNotEmpty(value) {
			return nil
		}
		return commonerrors.New(commonerrors.ErrInvalid, "cannot be empty")
	})
}

// LengthRule returns an ozzo length rule with optional minimum and maximum
// bounds.
//
// A nil minimum means zero. A nil maximum means unbounded.
//
// Strings are treated as a special case and validated using rune length rather
// than byte length so the behaviour is closer to JSON Schema string length
// semantics. Other length-aware values use ozzo's standard Length rule.
//
// Example: `LengthRule(field.ToOptionalInt(1), field.ToOptionalInt(3))`
// accepts strings, slices, arrays, and maps whose length is between one and
// three inclusive.
func LengthRule(min, max *int) validation.Rule {
	lengthRule := validation.Length(field.OptionalInt(min, 0), field.OptionalInt(max, math.MaxInt))
	runeLengthRule := validation.RuneLength(field.OptionalInt(min, 0), field.OptionalInt(max, math.MaxInt))
	return validation.By(func(value any) error {
		v, isNil := validation.Indirect(value)
		if !isNil {
			if isString, _, _, _ := validation.StringOrBytes(v); isString {
				return runeLengthRule.Validate(v)
			}
		}
		return lengthRule.Validate(value)
	})
}

// RuneLengthRule returns an ozzo rune-length rule with optional minimum and
// maximum bounds.
//
// A nil minimum means zero. A nil maximum means unbounded.
//
// Example: `RuneLengthRule(nil, field.ToOptionalInt(2))` accepts `éé` and
// rejects `ééé`.
func RuneLengthRule(min, max *int) validation.Rule {
	return validation.RuneLength(field.OptionalInt(min, 0), field.OptionalInt(max, math.MaxInt))
}
