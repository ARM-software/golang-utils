package validation

// when_rules.go contains conditional validation helpers that extend
// ozzo-validation's built-in `validation.When(...)` support with property- and
// field-oriented conditions.

import (
	"reflect"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/collection"
)

// WhenPropertyEquals applies rules when the value stored under key equals expected.
//
// Equality is evaluated with `reflect.DeepEqual`.
//
// Example: `WhenPropertyEquals("mode", "strict", RequiredProperties("name"))`
// validates `RequiredProperties("name")` only when `mode == "strict"`.
func WhenPropertyEquals(key string, expected any, rules ...validation.Rule) validation.Rule {
	return WhenPropertyMatches(key, expected, func(left, right any) (bool, error) {
		return reflect.DeepEqual(left, right), nil
	}, rules...)
}

// WhenPropertyMatches applies rules when the value stored under key matches expected.
//
// The comparison is delegated to match so callers can define case-insensitive or
// other domain-specific matching behaviour.
func WhenPropertyMatches[T any](key string, expected T, match collection.MatchFunc[T], rules ...validation.Rule) validation.Rule {
	filteredRules := collection.Filter(rules, func(rule validation.Rule) bool {
		return rule != nil
	})
	return validation.By(func(value any) error {
		props, isNil, err := objectProperties(value)
		if err != nil || isNil {
			return err
		}
		actual, found := props.value(key)
		if !found {
			return nil
		}
		cast, ok := actual.(T)
		if !ok {
			return nil
		}
		condition, err := match(cast, expected)
		if err != nil {
			return err
		}
		return validation.When(condition, filteredRules...).Validate(value)
	})
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

// WhenFieldMatches applies rules when the resolved field value matches expected.
func WhenFieldMatches[T any](field any, expected T, match collection.MatchFunc[T], rules ...validation.Rule) validation.Rule {
	filteredRules := collection.Filter(rules, func(rule validation.Rule) bool {
		return rule != nil
	})
	return validation.By(func(value any) error {
		fieldName, err := propertyNameForValue(value, field)
		if err != nil {
			return err
		}
		return WhenPropertyMatches(fieldName, expected, match, filteredRules...).Validate(value)
	})
}
