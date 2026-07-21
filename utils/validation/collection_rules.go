package validation

// collection_rules.go contains validation helpers focused on generic Go
// collections such as arrays and slices, without using JSON Schema
// terminology as the primary organising principle.

import (
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
		return collection.EachSlice(items, func(item any) error {
			return rule.Validate(item)
		})
	})
}
