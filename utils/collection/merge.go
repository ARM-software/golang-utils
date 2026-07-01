package collection

import (
	"maps"
	"slices"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/ptr"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// MergeFunc combines two values of the same type into one.
type MergeFunc[T any] func(e1, e2 T) (merged T, compatible bool)

// MergeRefFunc combines two references to values of the same type into one.
type MergeRefFunc[T any] func(e1, e2 *T) (merged *T, compatible bool)

// MergeBy groups values by key and merges each group into one or more values.
//
// Within each group, values are coalesced from left to right using merge.
//
// Because grouping is map-based, the order of the returned groups is not
// guaranteed.
//
// If items is nil, nil is returned. If items is empty, an empty slice is
// returned. If keyFunc or merge is nil, a clone of items is returned unchanged.
func MergeBy[T any, K comparable](items []T, keyFunc KeyFunc[T, K], merge MergeFunc[T]) (result []T) {
	if reflection.IsEmpty(items) {
		return []T{}
	}
	if keyFunc == nil || merge == nil {
		return slices.Clone(items)
	}

	groups := GroupBy(items, keyFunc)
	result = ReducesSequence(maps.Values(groups), make([]T, 0, len(groups)), func(acc []T, group []T) []T {
		return append(acc, Coalesce(group, merge)...)
	})
	return result
}

// MergeByRef groups values by a reference-based key function and merges each
// group using a reference-based merge function.
func MergeByRef[T any, K comparable](items []T, keyFunc KeyRefFunc[T, K], merge MergeRefFunc[T]) (result []T) {
	if keyFunc == nil || merge == nil {
		if items == nil {
			return nil
		}
		return slices.Clone(items)
	}
	return MergeBy(items,
		func(item T) K {
			return keyFunc(field.ToOptional(item))
		},
		func(left, right T) (merged T, compatible bool) {
			mergedRef, compatible := merge(field.ToOptional(left), field.ToOptional(right))
			if !compatible || mergedRef == nil {
				return merged, false
			}
			return ptr.From(mergedRef), true
		},
	)
}
