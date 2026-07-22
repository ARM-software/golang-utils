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
		original := reflect.ValueOf(value)
		v, isNil := validation.Indirect(value)
		if isNil {
			if original.IsValid() && original.Kind() == reflect.Func {
				return errType
			}
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
				_, ok := jsonSchemaNumber(v)
				return ok
			case "integer":
				_, ok := jsonSchemaInteger(v)
				return ok
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
	return validation.By(func(value any) error {
		candidate, isNil := validation.Indirect(value)
		if isNil {
			return nil
		}
		baseValue, _ := validation.Indirect(base)
		if !utilreflection.IsNotEmpty(baseValue) {
			return commonerrors.New(commonerrors.ErrInvalid, "multipleOf base must be strictly positive")
		}
		// Ozzo's MultipleOf can be reused safely for positive integer bases, but a
		// custom path is still required for JSON Schema compatibility because ozzo
		// does not support decimal bases and does not reject invalid bases such as
		// zero or negative values before attempting the calculation.
		if baseInt, err := validation.ToInt(baseValue); err == nil {
			if baseInt <= 0 {
				return commonerrors.New(commonerrors.ErrInvalid, "multipleOf base must be strictly positive")
			}
			return validation.MultipleOf(baseInt).Validate(candidate)
		}
		if baseUint, err := validation.ToUint(baseValue); err == nil {
			if baseUint == 0 {
				return commonerrors.New(commonerrors.ErrInvalid, "multipleOf base must be strictly positive")
			}
			return validation.MultipleOf(baseUint).Validate(candidate)
		}
		divisor, ok := jsonSchemaNumber(baseValue)
		if !ok {
			return commonerrors.Newf(commonerrors.ErrInvalid, "type not supported: %T", baseValue)
		}
		if divisor <= 0 {
			return commonerrors.New(commonerrors.ErrInvalid, "multipleOf base must be strictly positive")
		}
		multiple, ok := jsonSchemaNumber(candidate)
		if !ok {
			return commonerrors.Newf(commonerrors.ErrInvalid, "cannot convert %T to number", candidate)
		}
		quotient := multiple / divisor
		if math.Abs(quotient-math.Round(quotient)) <= 1e-12 {
			return nil
		}
		return validation.ErrMultipleOfInvalid.SetParams(map[string]any{"base": base})
	})
}

// Maximum validates the JSON Schema `maximum` constraint.
//
// Example: `Maximum(10)` accepts `10` and rejects `11`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/numeric#range
func Maximum(max any) validation.Rule {
	return thresholdRule(max, false, false)
}

// ExclusiveMaximum validates the JSON Schema `exclusive_maximum` constraint.
//
// Example: `ExclusiveMaximum(10)` accepts `9` and rejects `10`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/numeric#range
func ExclusiveMaximum(max any) validation.Rule {
	return thresholdRule(max, false, true)
}

// Minimum validates the JSON Schema `minimum` constraint.
//
// Example: `Minimum(10)` accepts `10` and rejects `9`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/numeric#range
func Minimum(min any) validation.Rule {
	return thresholdRule(min, true, false)
}

// ExclusiveMinimum validates the JSON Schema `exclusive_minimum` constraint.
//
// Example: `ExclusiveMinimum(10)` accepts `11` and rejects `10`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/numeric#range
func ExclusiveMinimum(min any) validation.Rule {
	return thresholdRule(min, true, true)
}

// MaxLength validates the JSON Schema `max_length` constraint.
//
// This counts Unicode code points rather than bytes.
// Example: `MaxLength(5)` accepts `hello` and rejects `hello!`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/string#length
func MaxLength(max int) validation.Rule {
	if max < 0 {
		return fail(commonerrors.Newf(commonerrors.ErrInvalid, "maxLength must be non-negative, got %d", max))
	}
	return stringLengthKeywordRule(nil, field.ToOptionalInt(max))
}

// MinLength validates the JSON Schema `min_length` constraint.
//
// This counts Unicode code points rather than bytes.
// Example: `MinLength(5)` accepts `hello` and rejects `hell`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/string#length
func MinLength(min int) validation.Rule {
	if min < 0 {
		return fail(commonerrors.Newf(commonerrors.ErrInvalid, "minLength must be non-negative, got %d", min))
	}
	return stringLengthKeywordRule(field.ToOptionalInt(min), nil)
}

// MaxItems validates the JSON Schema `max_items` constraint.
//
// Example: `MaxItems(2)` accepts `[1,2]` and rejects `[1,2,3]`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/array#length
func MaxItems(max int) validation.Rule {
	if max < 0 {
		return fail(commonerrors.Newf(commonerrors.ErrInvalid, "maxItems must be non-negative, got %d", max))
	}
	return itemLengthKeywordRule(nil, field.ToOptionalInt(max))
}

// MinItems validates the JSON Schema `min_items` constraint.
//
// Example: `MinItems(2)` accepts `[1,2]` and rejects `[1]`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/array#length
func MinItems(min int) validation.Rule {
	if min < 0 {
		return fail(commonerrors.Newf(commonerrors.ErrInvalid, "minItems must be non-negative, got %d", min))
	}
	return itemLengthKeywordRule(field.ToOptionalInt(min), nil)
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
	if rule == nil {
		return fail(commonerrors.New(commonerrors.ErrInvalid, "contains rule must not be nil"))
	}
	return MinContains(1, rule)
}

// MinContains validates the JSON Schema `minContains` constraint.
//
// The rule succeeds when at least min items in the array or slice satisfy rule.
// Example: `MinContains(2, Const("a"))` accepts `[]string{"a", "b", "a"}`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/array#contains
func MinContains(min int, rule validation.Rule) validation.Rule {
	if min < 0 {
		return fail(commonerrors.Newf(commonerrors.ErrInvalid, "minContains must be non-negative, got %d", min))
	}
	if rule == nil {
		return fail(commonerrors.New(commonerrors.ErrInvalid, "contains rule must not be nil"))
	}
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
	if max < 0 {
		return fail(commonerrors.Newf(commonerrors.ErrInvalid, "maxContains must be non-negative, got %d", max))
	}
	if rule == nil {
		return fail(commonerrors.New(commonerrors.ErrInvalid, "contains rule must not be nil"))
	}
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
	if keyFunc == nil {
		return fail(commonerrors.New(commonerrors.ErrInvalid, "unique item key function must not be nil"))
	}
	return validation.By(func(value any) error {
		items, err := typedSequence[T](value)
		if err != nil || items == nil {
			return err
		}
		unique, err := uniqueByKeySafe(items, keyFunc)
		if err != nil {
			return err
		}
		if len(unique) != len(items) {
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
	if max < 0 {
		return fail(commonerrors.Newf(commonerrors.ErrInvalid, "maxProperties must be non-negative, got %d", max))
	}
	return propertyLengthKeywordRule(nil, field.ToOptionalInt(max))
}

// MinProperties validates the JSON Schema `min_properties` constraint.
//
// Example: `MinProperties(2)` accepts a map with two keys and rejects one with one.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/object#size
func MinProperties(min int) validation.Rule {
	if min < 0 {
		return fail(commonerrors.Newf(commonerrors.ErrInvalid, "minProperties must be non-negative, got %d", min))
	}
	return propertyLengthKeywordRule(field.ToOptionalInt(min), nil)
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
	normalisedKeys := collection.UniqueEntries(keys)
	return validation.By(func(value any) error {
		props, isNil, err := objectProperties(value)
		if err != nil || isNil {
			return err
		}
		missing := collection.CountBy(normalisedKeys, func(key string) bool {
			return !props.present(key)
		})
		if missing > 0 {
			return errRequiredProperties
		}
		return nil
	})
}

// RequiredItems validates that a collection contains items matching all of the
// supplied reference items.
//
// The comparison key for both the validated items and the reference items is
// derived with keyFunc.
//
// Example: `RequiredItems(func(value string) string { return value }, "a", "b")`
// rejects `[]string{"a"}`.
func RequiredItems[T any, K comparable](keyFunc collection.KeyFunc[T, K], items ...T) validation.Rule {
	normalisedKeys, err := uniqueKeysSafe(collection.Map(items, keyFunc))
	if err != nil {
		return fail(err)
	}
	return RequiredItemKeys(keyFunc, normalisedKeys...)
}

// RequiredItemKeys validates that a collection contains items matching all of
// the supplied keys derived by keyFunc.
//
// Example: `RequiredItemKeys(func(value user) string { return value.Role }, "admin", "editor")`
// rejects `[]user{{Role: "admin"}}`.
func RequiredItemKeys[T any, K comparable](keyFunc collection.KeyFunc[T, K], keys ...K) validation.Rule {
	normalisedKeys, err := uniqueKeysSafe(keys)
	if err != nil {
		return fail(err)
	}
	return validation.By(func(value any) error {
		itemsByKey, isNil, err := keyedItemsByKey(value, keyFunc)
		if err != nil || isNil {
			return err
		}
		if countPresentKeys(itemsByKey, normalisedKeys) != len(normalisedKeys) {
			return errRequiredProperties
		}
		return nil
	})
}

// RequiredPropertiesBy resolves strings, `[]string`, or field references such
// as `&cfg.Name` against the validated value and applies [RequiredProperties]
// using the resulting property names.
//
// String and `[]string` arguments are treated as literal keys. Field pointers
// are resolved back to their struct field names.
//
// Example:
//
//	cfg := &Config{}
//	err := validation.Validate(cfg, RequiredPropertiesBy(&cfg.Name, &cfg.Enabled, &cfg.Mode))
func RequiredPropertiesBy(keys ...any) validation.Rule {
	return validation.By(func(value any) error {
		replayableValue, err := replayableValidationValue(value)
		if err != nil {
			return err
		}
		normalisedKeys, err := propertyNamesForValue(replayableValue, keys...)
		if err != nil {
			return err
		}
		return RequiredProperties(normalisedKeys...).Validate(replayableValue)
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
		props, isNil, err := objectProperties(value)
		if err != nil || isNil {
			return err
		}
		for key, dependents := range normalised {
			if !props.present(key) {
				continue
			}
			missing := collection.CountBy(dependents, func(dependent string) bool {
				return !props.present(dependent)
			})
			if missing > 0 {
				return errDependentRequired
			}
		}
		return nil
	})
}

// DependentRequiredItems validates that if a collection contains an item
// matching a trigger item then it also contains items matching each dependent
// item.
//
// Example: `DependentRequiredItems(func(value string) string { return value }, map[string][]string{"a": {"b"}})`
// rejects `[]string{"a"}`.
func DependentRequiredItems[T comparable, K comparable](keyFunc collection.KeyFunc[T, K], dependencies map[T][]T) validation.Rule {
	normalised := make(map[K][]K, len(dependencies))
	for item, dependents := range dependencies {
		key := keyFunc(item)
		if err := ensureHashableDynamicKey(any(key)); err != nil {
			return fail(err)
		}
		merged, err := uniqueKeysSafe(append(normalised[key], collection.Map(dependents, keyFunc)...))
		if err != nil {
			return fail(err)
		}
		normalised[key] = merged
	}
	return DependentRequiredItemKeys(keyFunc, normalised)
}

// DependentRequiredItemKeys validates that if a collection contains an item for
// a trigger key then it also contains items for each dependent key.
//
// Example: `DependentRequiredItemKeys(func(value user) string { return value.Role }, map[string][]string{"admin": {"editor"}})`
// rejects `[]user{{Role: "admin"}}`.
func DependentRequiredItemKeys[T any, K comparable](keyFunc collection.KeyFunc[T, K], dependencies map[K][]K) validation.Rule {
	normalised := make(map[K][]K, len(dependencies))
	for key, dependents := range dependencies {
		if err := ensureHashableDynamicKey(any(key)); err != nil {
			return fail(err)
		}
		merged, err := uniqueKeysSafe(dependents)
		if err != nil {
			return fail(err)
		}
		normalised[key] = merged
	}
	return validation.By(func(value any) error {
		itemsByKey, isNil, err := keyedItemsByKey(value, keyFunc)
		if err != nil || isNil {
			return err
		}
		for key, dependents := range normalised {
			if _, found := itemsByKey[key]; !found {
				continue
			}
			if countPresentKeys(itemsByKey, dependents) != len(dependents) {
				return errDependentRequired
			}
		}
		return nil
	})
}

// DependentRequiredBy resolves dependency trigger keys from strings or field
// references and dependent properties from strings, `[]string`, or field
// references before applying [DependentRequired].
//
// Example:
//
//	cfg := &Config{}
//	err := validation.Validate(cfg, DependentRequiredBy(map[any]any{&cfg.Username: []any{&cfg.Password, &cfg.Scheme}}))
func DependentRequiredBy(dependencies map[any]any) validation.Rule {
	return validation.By(func(value any) error {
		replayableValue, err := replayableValidationValue(value)
		if err != nil {
			return err
		}
		normalised, err := propertyDependenciesForValue(replayableValue, dependencies)
		if err != nil {
			return err
		}
		return DependentRequired(normalised).Validate(replayableValue)
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
		replayableValue, err := replayableValidationValue(value)
		if err != nil {
			return err
		}
		props, isNil, err := objectProperties(replayableValue)
		if err != nil || isNil {
			return err
		}
		for key, rule := range dependencies {
			if !props.present(key) {
				continue
			}
			if rule.Validate(replayableValue) != nil {
				return errDependentSchemas
			}
		}
		return nil
	})
}

// DependentSchemasBy resolves dependency trigger properties from strings or
// field references before applying [DependentSchemas].
//
// Example:
//
//	cfg := &Config{}
//	err := validation.Validate(cfg, DependentSchemasBy(map[any]validation.Rule{&cfg.Username: RequiredPropertiesBy(&cfg.Password)}))
func DependentSchemasBy(dependencies map[any]validation.Rule) validation.Rule {
	return validation.By(func(value any) error {
		replayableValue, err := replayableValidationValue(value)
		if err != nil {
			return err
		}
		normalised, err := propertyRulesForValue(replayableValue, dependencies)
		if err != nil {
			return err
		}
		return DependentSchemas(normalised).Validate(replayableValue)
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
	if rule == nil {
		return fail(commonerrors.New(commonerrors.ErrInvalid, "propertyNames rule must not be nil"))
	}
	return validation.By(func(value any) error {
		props, isNil, err := objectProperties(value)
		if err != nil || isNil {
			return err
		}
		for _, key := range objectPropertyNamesFromAccessor(props) {
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
		props, isNil, err := objectProperties(value)
		if err != nil || isNil {
			return err
		}
		for _, property := range patterns {
			if property.Pattern == nil || property.Rule == nil {
				continue
			}
			for _, key := range objectPropertyNamesFromAccessor(props) {
				if !property.Pattern.MatchString(key) {
					continue
				}
				fieldValue, found := props.value(key)
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
	normalisedKeys := collection.UniqueEntries(keys)
	return validation.By(func(value any) error {
		props, isNil, err := objectProperties(value)
		if err != nil || isNil {
			return err
		}
		invalid := collection.CountBy(objectPropertyNamesFromAccessor(props), func(key string) bool {
			return !collection.In(normalisedKeys, key, collection.StringMatch)
		})
		if invalid > 0 {
			return errAdditionalProperties
		}
		return nil
	})
}

// AdditionalItems validates that a collection contains no item whose derived
// key lies outside the supplied reference item set.
//
// Example: `AdditionalItems(func(value string) string { return value }, "a", "b")`
// rejects `[]string{"c"}`.
func AdditionalItems[T any, K comparable](keyFunc collection.KeyFunc[T, K], items ...T) validation.Rule {
	normalisedKeys, err := uniqueKeysSafe(collection.Map(items, keyFunc))
	if err != nil {
		return fail(err)
	}
	return AdditionalItemKeys(keyFunc, normalisedKeys...)
}

// AdditionalItemKeys validates that a collection contains no item whose derived
// key lies outside the supplied key set.
//
// Example: `AdditionalItemKeys(func(value user) string { return value.Role }, "admin", "editor")`
// rejects `[]user{{Role: "viewer"}}`.
func AdditionalItemKeys[T any, K comparable](keyFunc collection.KeyFunc[T, K], keys ...K) validation.Rule {
	normalisedKeys, err := uniqueKeysSafe(keys)
	if err != nil {
		return fail(err)
	}
	return validation.By(func(value any) error {
		itemsByKey, isNil, err := keyedItemsByKey(value, keyFunc)
		if err != nil || isNil {
			return err
		}
		invalid := 0
		for key := range itemsByKey {
			if !collection.AnyFunc(normalisedKeys, func(expected K) bool { return expected == key }) {
				invalid++
			}
		}
		if invalid > 0 {
			return errAdditionalProperties
		}
		return nil
	})
}

// AdditionalPropertiesBy resolves strings, `[]string`, or field references
// against the validated value and applies [AdditionalProperties] using the
// resulting names.
//
// Example:
//
//	cfg := &Config{}
//	err := validation.Validate(cfg, AdditionalPropertiesBy(&cfg.Name, &cfg.Enabled, &cfg.Mode))
func AdditionalPropertiesBy(keys ...any) validation.Rule {
	return validation.By(func(value any) error {
		replayableValue, err := replayableValidationValue(value)
		if err != nil {
			return err
		}
		normalisedKeys, err := propertyNamesForValue(replayableValue, keys...)
		if err != nil {
			return err
		}
		return AdditionalProperties(normalisedKeys...).Validate(replayableValue)
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
	normalisedKeys := collection.UniqueEntries(keys)
	return validation.By(func(value any) error {
		rv, isNil, err := objectValue(value)
		if err != nil || isNil {
			return err
		}
		count := 0
		switch rv.Kind() {
		case reflect.Map:
			count = collection.CountBy(normalisedKeys, func(key string) bool {
				mapValue, found := utilreflection.MapPropertyValue(rv, key)
				if !found {
					return false
				}
				return !utilreflection.IsEmpty(mapValue.Interface())
			})
		case reflect.Struct:
			count = collection.CountBy(normalisedKeys, func(key string) bool {
				fieldValue := rv.FieldByName(key)
				if !fieldValue.IsValid() || !fieldValue.CanInterface() {
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

// MutuallyExclusiveItems validates that a collection contains items matching at
// most one of the supplied reference items.
//
// Example: `MutuallyExclusiveItems(func(value string) string { return value }, "a", "b")`
// rejects `[]string{"a", "b"}`.
func MutuallyExclusiveItems[T any, K comparable](keyFunc collection.KeyFunc[T, K], items ...T) validation.Rule {
	normalisedKeys, err := uniqueKeysSafe(collection.Map(items, keyFunc))
	if err != nil {
		return fail(err)
	}
	return MutuallyExclusiveItemKeys(keyFunc, normalisedKeys...)
}

// MutuallyExclusiveItemKeys validates that a collection contains items matching
// at most one of the supplied keys.
//
// Example: `MutuallyExclusiveItemKeys(func(value user) string { return value.Role }, "admin", "editor")`
// rejects `[]user{{Role: "admin"}, {Role: "editor"}}`.
func MutuallyExclusiveItemKeys[T any, K comparable](keyFunc collection.KeyFunc[T, K], keys ...K) validation.Rule {
	normalisedKeys, err := uniqueKeysSafe(keys)
	if err != nil {
		return fail(err)
	}
	return validation.By(func(value any) error {
		itemsByKey, isNil, err := keyedItemsByKey(value, keyFunc)
		if err != nil || isNil {
			return err
		}
		if countPresentKeys(itemsByKey, normalisedKeys) > 1 {
			return errMutuallyExclusive
		}
		return nil
	})
}

// MutuallyExclusiveWithBy resolves strings, `[]string`, or field references
// against the validated value and applies [MutuallyExclusiveWith] using the
// resulting property names.
//
// Example:
//
//	cfg := &Config{}
//	err := validation.Validate(cfg, MutuallyExclusiveWithBy(&cfg.Token, &cfg.Username, &cfg.APIKey))
func MutuallyExclusiveWithBy(keys ...any) validation.Rule {
	return validation.By(func(value any) error {
		replayableValue, err := replayableValidationValue(value)
		if err != nil {
			return err
		}
		normalisedKeys, err := propertyNamesForValue(replayableValue, keys...)
		if err != nil {
			return err
		}
		return MutuallyExclusiveWith(normalisedKeys...).Validate(replayableValue)
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

// AtMostOneItem validates that a collection contains items matching no more
// than one of the supplied reference items.
//
// Example: `AtMostOneItem(func(value string) string { return value }, "a", "b")`
// rejects `[]string{"a", "b"}`.
func AtMostOneItem[T any, K comparable](keyFunc collection.KeyFunc[T, K], items ...T) validation.Rule {
	return MutuallyExclusiveItems(keyFunc, items...)
}

// AtMostOneItemKey validates that a collection contains items matching no more
// than one of the supplied keys.
//
// Example: `AtMostOneItemKey(func(value user) string { return value.Role }, "admin", "editor")`
// rejects `[]user{{Role: "admin"}, {Role: "editor"}}`.
func AtMostOneItemKey[T any, K comparable](keyFunc collection.KeyFunc[T, K], keys ...K) validation.Rule {
	return MutuallyExclusiveItemKeys(keyFunc, keys...)
}

// AtMostOnePropertyBy resolves strings, `[]string`, or field references against
// the validated value and applies [AtMostOneProperty].
//
// Example:
//
//	cfg := &Config{}
//	err := validation.Validate(cfg, AtMostOnePropertyBy(&cfg.Token, &cfg.Username, &cfg.APIKey))
func AtMostOnePropertyBy(keys ...any) validation.Rule {
	return validation.By(func(value any) error {
		replayableValue, err := replayableValidationValue(value)
		if err != nil {
			return err
		}
		normalisedKeys, err := propertyNamesForValue(replayableValue, keys...)
		if err != nil {
			return err
		}
		return AtMostOneProperty(normalisedKeys...).Validate(replayableValue)
	})
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
	normalisedKeys := collection.UniqueEntries(keys)
	return validation.By(func(value any) error {
		props, isNil, err := objectProperties(value)
		if err != nil || isNil {
			return err
		}
		if countPresentProperties(props, normalisedKeys) != 1 {
			return errMutuallyExclusive
		}
		return nil
	})
}

// OneOfItems validates that a collection contains items matching exactly one of
// the supplied reference items.
//
// Example: `OneOfItems(func(value string) string { return value }, "a", "b")`
// accepts `[]string{"a"}` and rejects both `[]string{}` and `[]string{"a", "b"}`.
func OneOfItems[T any, K comparable](keyFunc collection.KeyFunc[T, K], items ...T) validation.Rule {
	normalisedKeys, err := uniqueKeysSafe(collection.Map(items, keyFunc))
	if err != nil {
		return fail(err)
	}
	return OneOfItemKeys(keyFunc, normalisedKeys...)
}

// OneOfItemKeys validates that a collection contains items matching exactly one
// of the supplied keys.
//
// Example: `OneOfItemKeys(func(value user) string { return value.Role }, "admin", "editor")`
// accepts `[]user{{Role: "admin"}}` and rejects both `[]user{}` and
// `[]user{{Role: "admin"}, {Role: "editor"}}`.
func OneOfItemKeys[T any, K comparable](keyFunc collection.KeyFunc[T, K], keys ...K) validation.Rule {
	normalisedKeys, err := uniqueKeysSafe(keys)
	if err != nil {
		return fail(err)
	}
	return validation.By(func(value any) error {
		itemsByKey, isNil, err := keyedItemsByKey(value, keyFunc)
		if err != nil || isNil {
			return err
		}
		if countPresentKeys(itemsByKey, normalisedKeys) != 1 {
			return errMutuallyExclusive
		}
		return nil
	})
}

// OneOfPropertiesBy resolves strings, `[]string`, or field references against
// the validated value and applies [OneOfProperties].
//
// Example:
//
//	cfg := &Config{}
//	err := validation.Validate(cfg, OneOfPropertiesBy(&cfg.Token, &cfg.Username, &cfg.APIKey))
func OneOfPropertiesBy(keys ...any) validation.Rule {
	return validation.By(func(value any) error {
		replayableValue, err := replayableValidationValue(value)
		if err != nil {
			return err
		}
		normalisedKeys, err := propertyNamesForValue(replayableValue, keys...)
		if err != nil {
			return err
		}
		return OneOfProperties(normalisedKeys...).Validate(replayableValue)
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
	normalisedKeys := collection.UniqueEntries(keys)
	return validation.By(func(value any) error {
		props, isNil, err := objectProperties(value)
		if err != nil || isNil {
			return err
		}
		if countPresentProperties(props, normalisedKeys) == 0 {
			return errRequiredProperties
		}
		return nil
	})
}

// AtLeastOneItem validates that a collection contains an item matching at least
// one of the supplied reference items.
//
// Example: `AtLeastOneItem(func(value string) string { return value }, "a", "b")`
// rejects `[]string{}`.
func AtLeastOneItem[T any, K comparable](keyFunc collection.KeyFunc[T, K], items ...T) validation.Rule {
	normalisedKeys, err := uniqueKeysSafe(collection.Map(items, keyFunc))
	if err != nil {
		return fail(err)
	}
	return AtLeastOneItemKey(keyFunc, normalisedKeys...)
}

// AtLeastOneItemKey validates that a collection contains an item matching at
// least one of the supplied keys.
//
// Example: `AtLeastOneItemKey(func(value user) string { return value.Role }, "admin", "editor")`
// rejects `[]user{}`.
func AtLeastOneItemKey[T any, K comparable](keyFunc collection.KeyFunc[T, K], keys ...K) validation.Rule {
	normalisedKeys, err := uniqueKeysSafe(keys)
	if err != nil {
		return fail(err)
	}
	return validation.By(func(value any) error {
		itemsByKey, isNil, err := keyedItemsByKey(value, keyFunc)
		if err != nil || isNil {
			return err
		}
		if countPresentKeys(itemsByKey, normalisedKeys) == 0 {
			return errRequiredProperties
		}
		return nil
	})
}

// AtLeastOnePropertyBy resolves strings, `[]string`, or field references
// against the validated value and applies [AtLeastOneProperty].
//
// Example:
//
//	cfg := &Config{}
//	err := validation.Validate(cfg, AtLeastOnePropertyBy(&cfg.Token, &cfg.Username, &cfg.APIKey))
func AtLeastOnePropertyBy(keys ...any) validation.Rule {
	return validation.By(func(value any) error {
		replayableValue, err := replayableValidationValue(value)
		if err != nil {
			return err
		}
		normalisedKeys, err := propertyNamesForValue(replayableValue, keys...)
		if err != nil {
			return err
		}
		return AtLeastOneProperty(normalisedKeys...).Validate(replayableValue)
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
	normalisedKeys := collection.UniqueEntries(keys)
	return validation.By(func(value any) error {
		props, isNil, err := objectProperties(value)
		if err != nil || isNil {
			return err
		}
		if countPresentProperties(props, normalisedKeys) > 0 {
			return errAdditionalProperties
		}
		return nil
	})
}

// ForbiddenItems validates that a collection contains no items matching any of
// the supplied reference items.
//
// Example: `ForbiddenItems(func(value string) string { return value }, "debug")`
// rejects `[]string{"debug"}`.
func ForbiddenItems[T any, K comparable](keyFunc collection.KeyFunc[T, K], items ...T) validation.Rule {
	normalisedKeys, err := uniqueKeysSafe(collection.Map(items, keyFunc))
	if err != nil {
		return fail(err)
	}
	return ForbiddenItemKeys(keyFunc, normalisedKeys...)
}

// ForbiddenItemKeys validates that a collection contains no items matching any
// of the supplied keys.
//
// Example: `ForbiddenItemKeys(func(value user) string { return value.Role }, "debug")`
// rejects `[]user{{Role: "debug"}}`.
func ForbiddenItemKeys[T any, K comparable](keyFunc collection.KeyFunc[T, K], keys ...K) validation.Rule {
	normalisedKeys, err := uniqueKeysSafe(keys)
	if err != nil {
		return fail(err)
	}
	return validation.By(func(value any) error {
		itemsByKey, isNil, err := keyedItemsByKey(value, keyFunc)
		if err != nil || isNil {
			return err
		}
		if countPresentKeys(itemsByKey, normalisedKeys) > 0 {
			return errAdditionalProperties
		}
		return nil
	})
}

// ForbiddenPropertiesBy resolves strings, `[]string`, or field references
// against the validated value and applies [ForbiddenProperties].
//
// Example:
//
//	cfg := &Config{}
//	err := validation.Validate(cfg, ForbiddenPropertiesBy(&cfg.Debug, &cfg.InternalOnly))
func ForbiddenPropertiesBy(keys ...any) validation.Rule {
	return validation.By(func(value any) error {
		replayableValue, err := replayableValidationValue(value)
		if err != nil {
			return err
		}
		normalisedKeys, err := propertyNamesForValue(replayableValue, keys...)
		if err != nil {
			return err
		}
		return ForbiddenProperties(normalisedKeys...).Validate(replayableValue)
	})
}

// XIntOrString validates the Kubernetes/OpenAPI `x-kubernetes-int-or-string`
// style constraint.
//
// Example: `XIntOrString()` accepts `3`, `"3"`, and JSON-decoded integer
// numbers represented as `float64(3)`.
func XIntOrString() validation.Rule {
	return validation.By(func(value any) error {
		if validation.NotNil.Validate(value) != nil {
			return errIntOrString
		}
		candidate, _ := validation.Indirect(value)
		if isString, _, _, _ := validation.StringOrBytes(candidate); isString {
			return nil
		}
		if _, err := validation.ToInt(candidate); err == nil {
			return nil
		}
		if _, err := validation.ToUint(candidate); err == nil {
			return nil
		}
		if f, err := validation.ToFloat(candidate); err == nil && !math.IsInf(f, 0) && !math.IsNaN(f) && math.Trunc(f) == f {
			return nil
		}
		return errIntOrString
	})
}

// Enum validates that a value is one of a fixed set of allowed values.
//
// Unlike ozzo's `validation.In(...)`, this helper does not treat empty values as
// automatically valid. It follows JSON Schema `enum` semantics instead.
// Example: `Enum("red", "blue")` accepts `"blue"` and rejects `"green"`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/enum
func Enum(values ...any) validation.Rule {
	return validation.By(func(value any) error {
		candidate, isNil := validation.Indirect(value)
		if isNil {
			candidate = nil
		}
		if collection.AnyFunc(values, func(expected any) bool {
			return jsonSchemaEqualValues(candidate, expected)
		}) {
			return nil
		}
		return commonerrors.WrapError(commonerrors.ErrInvalid, validation.NewError("validation_enum", "must be one of the allowed values"), "invalid value")
	})
}

// Const validates that a value is exactly equal to expected.
//
// This uses the same equality semantics as [Enum], including numeric equality
// across compatible JSON-style number representations.
// This is useful for schema-style validations where one field must have a fixed
// discriminator or version value.
// Example: `Const("v1")` accepts `"v1"` and rejects `"v2"`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/const
func Const(expected any) validation.Rule {
	return Enum(expected)
}

// Pattern validates the JSON Schema `pattern` constraint.
//
// JSON Schema applies `pattern` only to string instances. Non-string values are
// ignored rather than rejected. Unlike ozzo's `validation.Match(...)`, empty
// strings are still validated against the supplied regexp.
// Example: `Pattern(regexp.MustCompile("^[a-z]+$"))` accepts `"abc"`.
//
// Reference: https://json-schema.org/understanding-json-schema/reference/string#regular-expressions
func Pattern(re *regexp.Regexp) validation.Rule {
	if re == nil {
		return fail(commonerrors.New(commonerrors.ErrInvalid, "pattern regexp must not be nil"))
	}
	return validation.By(func(value any) error {
		candidate, isNil := validation.Indirect(value)
		if isNil {
			return nil
		}
		isString, str, _, _ := validation.StringOrBytes(candidate)
		if !isString {
			return nil
		}
		if str == "" {
			if re.MatchString(str) {
				return nil
			}
			return validation.ErrMatchInvalid
		}
		return validation.Match(re).Validate(str)
	})
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
		if validation.Nil.Validate(value) == nil {
			return nil
		}
		if rule == nil {
			return commonerrors.New(commonerrors.ErrInvalid, "nullable rule must not be nil")
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
//
// References:
//   - JSON Schema string length:
//     https://json-schema.org/understanding-json-schema/reference/string#length
//   - JSON Schema array length:
//     https://json-schema.org/understanding-json-schema/reference/array#length
//   - JSON Schema object size:
//     https://json-schema.org/understanding-json-schema/reference/object#size
func LengthRule(min, max *int) validation.Rule {
	minimum := field.OptionalInt(min, 0)
	maximum := field.OptionalInt(max, math.MaxInt)
	lengthRule := validation.Length(minimum, maximum)
	runeLengthRule := validation.RuneLength(minimum, maximum)
	return validation.By(func(value any) error {
		// Ozzo's length rules short-circuit empty values as valid before applying
		// the bounds. For JSON Schema-style minimum constraints we need empty
		// strings, slices, and maps to fail when the minimum is positive.
		// References:
		//   - https://pkg.go.dev/github.com/go-ozzo/ozzo-validation/v4#Length
		if minimum > 0 && validation.IsEmpty(value) {
			return commonerrors.New(commonerrors.ErrInvalid, "the length must be no less than the minimum")
		}
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
//
// Reference:
//   - JSON Schema string length:
//     https://json-schema.org/understanding-json-schema/reference/string#length
func RuneLengthRule(min, max *int) validation.Rule {
	minimum := field.OptionalInt(min, 0)
	maximum := field.OptionalInt(max, math.MaxInt)
	rule := validation.RuneLength(minimum, maximum)
	return validation.By(func(value any) error {
		// Like LengthRule above, this guards the JSON Schema expectation that a
		// positive minimum rejects empty values instead of treating them as
		// automatically valid.
		// References:
		//   - https://pkg.go.dev/github.com/go-ozzo/ozzo-validation/v4#RuneLength
		if minimum > 0 && validation.IsEmpty(value) {
			return commonerrors.New(commonerrors.ErrInvalid, "the length must be no less than the minimum")
		}
		return rule.Validate(value)
	})
}

func stringLengthKeywordRule(min, max *int) validation.Rule {
	rule := RuneLengthRule(min, max)
	return validation.By(func(value any) error {
		candidate, isNil := validation.Indirect(value)
		if isNil {
			return nil
		}
		isString, _, _, _ := validation.StringOrBytes(candidate)
		if !isString {
			return nil
		}
		return rule.Validate(candidate)
	})
}

func itemLengthKeywordRule(min, max *int) validation.Rule {
	rule := LengthRule(min, max)
	return validation.By(func(value any) error {
		items, err := typedSequence[any](value)
		if err != nil {
			if commonerrors.Any(err, errArrayOrSliceRequired) || commonerrors.CorrespondTo(err, errArrayOrSliceRequired.Error()) {
				return nil
			}
			return err
		}
		if items == nil {
			return nil
		}
		return rule.Validate(items)
	})
}

func propertyLengthKeywordRule(min, max *int) validation.Rule {
	rule := LengthRule(min, max)
	return validation.By(func(value any) error {
		props, isNil, err := objectProperties(value)
		if err != nil {
			if commonerrors.Any(err, errMapRequired) || commonerrors.CorrespondTo(err, errMapRequired.Error()) {
				return nil
			}
			return err
		}
		if isNil {
			return nil
		}
		return rule.Validate(objectPropertyNamesFromAccessor(props))
	})
}

// jsonSchemaEqualValues compares two values using JSON Schema-style equality.
// Numeric values are compared by numeric value rather than exact Go type.
func jsonSchemaEqualValues(left, right any) bool {
	if left == nil || right == nil {
		return left == right
	}
	if leftNumber, ok := jsonSchemaNumber(left); ok {
		if rightNumber, ok := jsonSchemaNumber(right); ok {
			return leftNumber == rightNumber
		}
	}
	return reflect.DeepEqual(left, right)
}

// jsonSchemaNumber converts a JSON-number-like value into a float64 for schema
// equality checks.
func jsonSchemaNumber(value any) (float64, bool) {
	if v, err := validation.ToInt(value); err == nil {
		return float64(v), true
	}
	if v, err := validation.ToUint(value); err == nil {
		return float64(v), true
	}
	if v, err := validation.ToFloat(value); err == nil {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, false
		}
		return v, true
	}
	return 0, false
}

func jsonSchemaInteger(value any) (float64, bool) {
	if number, ok := jsonSchemaNumber(value); ok && math.Trunc(number) == number {
		return number, true
	}
	return 0, false
}
