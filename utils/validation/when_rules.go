package validation

// when_rules.go contains conditional validation helpers that extend
// ozzo-validation's built-in `validation.When(...)` support with property- and
// field-oriented conditions.

import (
	"reflect"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/collection"
	utilreflection "github.com/ARM-software/golang-utils/utils/reflection"
)

type propertyConditionFunc func(actual any, found bool) (bool, error)

func filteredWhenRules(rules []validation.Rule) []validation.Rule {
	return collection.Filter(rules, func(rule validation.Rule) bool {
		return rule != nil
	})
}

func evaluateConditionalRules(value any, matched bool, rules []validation.Rule) error {
	if !matched || len(rules) == 0 {
		return nil
	}
	return NewAllRule(rules...).Validate(value)
}

func whenProperty(key string, condition propertyConditionFunc, rules ...validation.Rule) validation.Rule {
	filteredRules := filteredWhenRules(rules)
	return validation.By(func(value any) error {
		replayableValue, err := replayableValidationValue(value)
		if err != nil {
			return err
		}
		props, isNil, err := objectProperties(replayableValue)
		if err != nil || isNil {
			return err
		}
		actual, found := props.value(key)
		matched, err := condition(actual, found)
		if err != nil {
			return err
		}
		return evaluateConditionalRules(replayableValue, matched, filteredRules)
	})
}

func whenField(field any, propertyRule func(fieldName string) validation.Rule) validation.Rule {
	return validation.By(func(value any) error {
		replayableValue, err := replayableValidationValue(value)
		if err != nil {
			return err
		}
		fieldName, err := propertyNameForValue(replayableValue, field)
		if err != nil {
			return err
		}
		return propertyRule(fieldName).Validate(replayableValue)
	})
}

// WhenPropertyEquals applies rules when the value stored under key equals expected.
//
// Equality is evaluated with `reflect.DeepEqual`.
//
// This helper is primarily intended for property-oriented validation of maps or
// decoded object content where conditions are expressed in terms of named
// properties. For struct-oriented cross-field validation, prefer
// [WhenFieldEqualsValue].
//
// If the structure in question has its own `Validate()` method, prefer the
// field-oriented helpers for nested composition. They are designed to avoid the
// recursive re-entry pitfalls that older conditional-validation patterns can
// trigger.
//
// Example: `WhenPropertyEquals("mode", "strict", RequiredProperties("name"))`
// validates `RequiredProperties("name")` only when `mode == "strict"`.
func WhenPropertyEquals(key string, expected any, rules ...validation.Rule) validation.Rule {
	return WhenPropertyMatches(key, expected, func(left, right any) (bool, error) {
		return reflect.DeepEqual(left, right), nil
	}, rules...)
}

// WhenPropertyNotEquals applies rules when the value stored under key does not equal expected.
//
// Equality is evaluated with `reflect.DeepEqual`.
//
// This helper is primarily intended for property-oriented validation of maps or
// decoded object content. For struct-oriented cross-field validation, prefer
// [WhenFieldNotEqualsValue].
//
// If the structure in question has its own `Validate()` method, prefer the
// field-oriented helpers for nested composition. They are designed to avoid the
// recursive re-entry pitfalls that older conditional-validation patterns can
// trigger.
//
// Example: `WhenPropertyNotEquals("mode", "strict", ForbiddenProperties("name"))`
// rejects a payload that defines `name` unless `mode == "strict"`.
func WhenPropertyNotEquals(key string, expected any, rules ...validation.Rule) validation.Rule {
	return WhenPropertyMatches(key, expected, func(left, right any) (bool, error) {
		return !reflect.DeepEqual(left, right), nil
	}, rules...)
}

// WhenPropertyAbsent applies rules when the value stored under key is absent.
//
// A property is considered absent when it is missing or when its value is empty
// according to `reflection.IsEmpty`.
//
// This helper is primarily intended for property-oriented validation of maps or
// decoded object content. For struct-oriented cross-field validation, prefer
// [WhenFieldAbsent].
//
// If the structure in question has its own `Validate()` method, prefer the
// field-oriented helpers for nested composition. They are designed to avoid the
// recursive re-entry pitfalls that older conditional-validation patterns can
// trigger.
//
// Example: `WhenPropertyAbsent("token", RequiredProperties("username"))`
// requires `username` whenever `token` is missing or empty.
func WhenPropertyAbsent(key string, rules ...validation.Rule) validation.Rule {
	return whenProperty(key, func(actual any, found bool) (bool, error) {
		return !found || utilreflection.IsEmpty(actual), nil
	}, rules...)
}

// WhenPropertyPresent applies rules when the value stored under key is present.
//
// A property is considered present when it exists and its value is not empty
// according to `reflection.IsNotEmpty`.
//
// This helper is primarily intended for property-oriented validation of maps or
// decoded object content. For struct-oriented cross-field validation, prefer
// [WhenFieldPresent].
//
// If the structure in question has its own `Validate()` method, prefer the
// field-oriented helpers for nested composition. They are designed to avoid the
// recursive re-entry pitfalls that older conditional-validation patterns can
// trigger.
//
// Example: `WhenPropertyPresent("token", ForbiddenProperties("username"))`
// rejects a payload that defines `username` whenever `token` is already set.
func WhenPropertyPresent(key string, rules ...validation.Rule) validation.Rule {
	return whenProperty(key, func(actual any, found bool) (bool, error) {
		return found && utilreflection.IsNotEmpty(actual), nil
	}, rules...)
}

// WhenPropertyEmpty applies rules when the value stored under key is empty.
//
// Deprecated: prefer [WhenPropertyAbsent] for schema-style presence checks.
func WhenPropertyEmpty(key string, rules ...validation.Rule) validation.Rule {
	return WhenPropertyAbsent(key, rules...)
}

// WhenPropertyNotEmpty applies rules when the value stored under key is not empty.
//
// Deprecated: prefer [WhenPropertyPresent] for schema-style presence checks.
func WhenPropertyNotEmpty(key string, rules ...validation.Rule) validation.Rule {
	return WhenPropertyPresent(key, rules...)
}

// WhenPropertyMatches applies rules when the value stored under key matches expected.
//
// The comparison is delegated to match so callers can define case-insensitive or
// other domain-specific matching behaviour.
//
// This helper is primarily intended for property-oriented validation of maps or
// decoded object content. For struct-oriented cross-field validation, prefer the
// corresponding `WhenField...` helpers.
//
// If the structure in question has its own `Validate()` method, prefer the
// field-oriented helpers for nested composition. They are designed to avoid the
// recursive re-entry pitfalls that older conditional-validation patterns can
// trigger.
func WhenPropertyMatches[T any](key string, expected T, match collection.MatchFunc[T], rules ...validation.Rule) validation.Rule {
	return whenProperty(key, func(actual any, found bool) (bool, error) {
		if !found {
			return false, nil
		}
		cast, ok := actual.(T)
		if !ok {
			return false, nil
		}
		return match(cast, expected)
	}, rules...)
}

// WhenFieldEquals applies rules when the resolved field value equals expected.
//
// Example:
//
//	cfg := &Config{}
//	err := validation.Validate(cfg, WhenFieldEquals(&cfg.Mode, "strict", RequiredPropertiesBy(&cfg.Name)))
func WhenFieldEquals(field any, expected any, rules ...validation.Rule) validation.Rule {
	return WhenFieldMatches(field, expected, func(left, right any) (bool, error) {
		return reflect.DeepEqual(left, right), nil
	}, rules...)
}

// WhenFieldEqualsValue applies rules when the resolved field value equals expected.
//
// It is the preferred value-oriented entrypoint for composing cross-field rules
// such as `RequiredFieldsBy(...)` against a containing struct.
//
// The current implementation is equivalent to [WhenFieldEquals], but this name
// makes the intended usage clearer when the nested rules validate field values
// rather than mere property presence.
//
// Unlike older patterns that rely on `validation.When(...).Validate(value)`, the
// `WhenField...` helper family now executes nested rules directly against the
// containing object so it can be used safely from a struct's own `Validate()`
// implementation without recursive re-entry into that method.
//
// Example:
//
//	cfg := &Config{}
//	err := validation.Validate(cfg, NewAllRule(
//		WhenFieldEqualsValue(&cfg.Mode, "strict", RequiredFieldsBy(&cfg.Name)),
//		WhenFieldPresent(&cfg.Name, RequiredFieldsBy(&cfg.Mode)),
//	))
func WhenFieldEqualsValue(field any, expected any, rules ...validation.Rule) validation.Rule {
	return WhenFieldEquals(field, expected, rules...)
}

// WhenFieldNotEqualsValue applies rules when the resolved field value does not equal expected.
//
// It is the value-oriented counterpart to [WhenFieldNotEquals].
//
// Example:
//
// 	cfg := &Config{}
// 	err := validation.Validate(cfg, WhenFieldNotEqualsValue(&cfg.Mode, "strict", RequiredFieldsBy(&cfg.Profile)))
//
// This requires `cfg.Profile` whenever `cfg.Mode != "strict"`.
func WhenFieldNotEqualsValue(field any, expected any, rules ...validation.Rule) validation.Rule {
	return WhenFieldNotEquals(field, expected, rules...)
}

// WhenFieldNotEquals applies rules when the resolved field value does not equal expected.
func WhenFieldNotEquals(field any, expected any, rules ...validation.Rule) validation.Rule {
	return WhenFieldMatches(field, expected, func(left, right any) (bool, error) {
		return !reflect.DeepEqual(left, right), nil
	}, rules...)
}

// WhenFieldInValues applies rules when the resolved field value equals any of
// the supplied expected values.
//
// Example:
//
// 	cfg := &Config{}
// 	err := validation.Validate(cfg, WhenFieldInValues(&cfg.Mode, []any{"strict", "relaxed"}, RequiredFieldsBy(&cfg.Profile)))
//
// This requires `cfg.Profile` whenever `cfg.Mode` is either `strict` or
// `relaxed`.
func WhenFieldInValues(field any, expected []any, rules ...validation.Rule) validation.Rule {
	return whenField(field, func(fieldName string) validation.Rule {
		return whenProperty(fieldName, func(actual any, found bool) (bool, error) {
			if !found {
				return false, nil
			}
			return collection.In(expected, actual, func(left any, right any) (bool, error) {
				return reflect.DeepEqual(left, right), nil
			}), nil
		}, rules...)
	})
}

// WhenFieldNotInValues applies rules when the resolved field value equals none
// of the supplied expected values.
//
// Example:
//
// 	cfg := &Config{}
// 	err := validation.Validate(cfg, WhenFieldNotInValues(&cfg.Mode, []any{"strict", "relaxed"}, RequiredFieldsBy(&cfg.FallbackProfile)))
//
// This requires `cfg.FallbackProfile` whenever `cfg.Mode` is neither `strict`
// nor `relaxed`.
func WhenFieldNotInValues(field any, expected []any, rules ...validation.Rule) validation.Rule {
	return whenField(field, func(fieldName string) validation.Rule {
		return whenProperty(fieldName, func(actual any, found bool) (bool, error) {
			if !found {
				return false, nil
			}
			inValues := collection.In(expected, actual, func(left any, right any) (bool, error) {
				return reflect.DeepEqual(left, right), nil
			})
			return !inValues, nil
		}, rules...)
	})
}

// WhenFieldAbsent applies rules when the resolved field value is absent.
//
// A field is considered absent when its value is empty according to
// `reflection.IsEmpty`.
//
// Example:
//
// 	cfg := &Config{}
// 	err := validation.Validate(cfg, WhenFieldAbsent(&cfg.Token, RequiredFieldsBy(&cfg.Username)))
//
// This requires `cfg.Username` whenever `cfg.Token` is empty.
func WhenFieldAbsent(field any, rules ...validation.Rule) validation.Rule {
	return whenField(field, func(fieldName string) validation.Rule {
		return WhenPropertyAbsent(fieldName, rules...)
	})
}

// WhenFieldPresent applies rules when the resolved field value is present.
//
// A field is considered present when its value is not empty according to
// `reflection.IsNotEmpty`.
//
// Example:
//
// 	cfg := &Config{}
// 	err := validation.Validate(cfg, WhenFieldPresent(&cfg.Token, ForbiddenFieldsBy(&cfg.Username)))
//
// This rejects `cfg.Username` whenever `cfg.Token` is non-empty.
func WhenFieldPresent(field any, rules ...validation.Rule) validation.Rule {
	return whenField(field, func(fieldName string) validation.Rule {
		return WhenPropertyPresent(fieldName, rules...)
	})
}

// WhenFieldEmpty applies rules when the resolved field value is empty.
//
// Deprecated: prefer [WhenFieldAbsent] for schema-style presence checks.
func WhenFieldEmpty(field any, rules ...validation.Rule) validation.Rule {
	return WhenFieldAbsent(field, rules...)
}

// WhenFieldNotEmpty applies rules when the resolved field value is not empty.
//
// Deprecated: prefer [WhenFieldPresent] for schema-style presence checks.
func WhenFieldNotEmpty(field any, rules ...validation.Rule) validation.Rule {
	return WhenFieldPresent(field, rules...)
}

// WhenFieldMatches applies rules when the resolved field value matches expected.
func WhenFieldMatches[T any](field any, expected T, match collection.MatchFunc[T], rules ...validation.Rule) validation.Rule {
	return whenField(field, func(fieldName string) validation.Rule {
		return WhenPropertyMatches(fieldName, expected, match, rules...)
	})
}
