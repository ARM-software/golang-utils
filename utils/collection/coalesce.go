package collection

import (
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/ptr"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// Coalesce compacts items by merging adjacent compatible elements.
//
// Items are processed from left to right. Each new item is only considered for
// merging with the last item already emitted in the output. If merge returns
// ok=true, the last output item is replaced with the merged value; otherwise the
// item is appended as a new output element.
//
// This is similar to Rust itertools::coalesce and is useful when neighbouring
// values may collapse into a more compact representation.
//
// References:
//   - https://docs.rs/itertools/latest/itertools/trait.Itertools.html#method.coalesce
func Coalesce[T any](items []T, merge MergeFunc[T]) []T {
	if reflection.IsEmpty(items) {
		return []T{}
	}

	result := make([]T, 0, len(items))
	ForEach(items, func(item T) {
		last, found := Last(result)
		if !found {
			result = append(result, item)
			return
		}

		merged, ok := merge(last, item)
		if !ok {
			result = append(result, item)
			return
		}

		result[len(result)-1] = merged
	})
	return result
}

// CoalesceRef compacts items by merging adjacent compatible elements using a
// reference-based merge function.
func CoalesceRef[T any](items []T, merge MergeRefFunc[T]) []T {
	return Coalesce(items, func(left, right T) (merged T, compatible bool) {
		if merge == nil {
			return merged, false
		}
		mergedRef, compatible := merge(field.ToOptional(left), field.ToOptional(right))
		if !compatible || mergedRef == nil {
			return merged, false
		}
		return ptr.From(mergedRef), true
	})
}
