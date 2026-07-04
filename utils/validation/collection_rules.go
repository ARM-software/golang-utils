package validation

// collection_rules.go contains validation helpers focused on generic Go
// collections such as arrays, slices, and maps, without using JSON Schema
// terminology as the primary organising principle.

import (
	"reflect"
	"slices"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/collection"
)

// ArrayItems validates that every item in an array or slice satisfies rule.
//
// Example: `ArrayItems(Type("string"))` accepts `[]any{"a", "b"}`.
func ArrayItems(rule validation.Rule) validation.Rule {
	return validation.By(func(value any) error {
		items, err := typedSequence[any](value)
		if err != nil || items == nil {
			return err
		}
		return collection.Each(slices.Values(items), func(item any) error {
			return rule.Validate(item)
		})
	})
}

// MapKeys validates that every key in a map satisfies rule.
//
// Example: `MapKeys(Pattern(regexp.MustCompile("^[a-z]+$")))` accepts
// `map[string]any{"alpha": 1}`.
func MapKeys(rule validation.Rule) validation.Rule {
	return PropertyNames(rule)
}

// MapValues validates that every value in a map satisfies rule.
//
// Example: `MapValues(Type("string"))` accepts `map[string]any{"a": "x"}`.
func MapValues(rule validation.Rule) validation.Rule {
	return validation.By(func(value any) error {
		if props, ok, err := objectSequence2ToAccessor(value); err != nil {
			return err
		} else if ok {
			return collection.Each(slices.Values(objectPropertyNamesFromAccessor(props)), func(key string) error {
				fieldValue, found := props.value(key)
				if !found {
					return nil
				}
				return rule.Validate(fieldValue)
			})
		}

		rv, isNil, err := objectValue(value)
		if err != nil || isNil {
			return err
		}
		if rv.Kind() != reflect.Map {
			return errMapRequired
		}
		return collection.Each(slices.Values(objectPropertyNames(rv)), func(key string) error {
			fieldValue, found := objectPropertyValue(rv, key)
			if !found {
				return nil
			}
			return rule.Validate(fieldValue)
		})
	})
}
