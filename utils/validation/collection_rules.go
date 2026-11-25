/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package validation

// collection_rules.go contains validation helpers focused on generic Go
// collections such as arrays and slices, without using JSON Schema
// terminology as the primary organising principle.

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/collection"
)

var errSequenceRequired = validation.NewError("validation_sequence_required", "must be a sequence")

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

// SequenceItems validates that every item in a function-backed sequence
// satisfies rule.
//
// Example: `SequenceItems(Type("string"))` accepts
// `iter.Seq[any](func(yield func(any) bool) { ... })`.
func SequenceItems(rule validation.Rule) validation.Rule {
	return validation.By(func(value any) error {
		items, ok := sequenceToSlice(value)
		if !ok {
			return errSequenceRequired
		}
		if items == nil {
			return nil
		}
		return collection.EachSlice(items, func(item any) error {
			return rule.Validate(item)
		})
	})
}
