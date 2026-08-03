package validation

// map_rules.go contains validation helpers focused specifically on maps and
// other key/value object-like inputs.

import (
	"reflect"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/collection"
)

// MapKeys validates that every key in a map satisfies rule.
//
// It also accepts `iter.Seq2[string, any]`-style inputs and validates each
// yielded key.
//
// Example: `MapKeys(Pattern(regexp.MustCompile("^[a-z]+$")))` accepts
// `map[string]any{"alpha": 1}`.
func MapKeys(rule validation.Rule) validation.Rule {
	return PropertyNames(rule)
}

// MapValues validates that every value in a map satisfies rule.
//
// It also accepts `iter.Seq2[string, any]`-style inputs and validates each
// yielded value.
//
// Example: `MapValues(Type("string"))` accepts `map[string]any{"a": "x"}`.
func MapValues(rule validation.Rule) validation.Rule {
	return validation.By(func(value any) error {
		if props, ok, isNil, err := objectSequence2ToAccessor(value); err != nil {
			return err
		} else if isNil {
			return nil
		} else if ok {
			return collection.EachSlice(objectPropertyNamesFromAccessor(props), func(key string) error {
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
		return collection.EachSlice(objectPropertyNames(rv), func(key string) error {
			fieldValue, found := objectPropertyValue(rv, key)
			if !found {
				return nil
			}
			return rule.Validate(fieldValue)
		})
	})
}
