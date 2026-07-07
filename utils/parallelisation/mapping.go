package parallelisation

import (
	"context"
	"iter"
	"slices"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/field"
)

// MapConcurrent is similar to [collection.Map] but uses store options instead of
// a dedicated worker-count parameter.
//
// Results are returned in input order. Use [Workers] to limit parallelism.
func MapConcurrent[I any, O any](ctx context.Context, items []I, fn collection.MapFunc[I, O], options ...StoreOption) ([]O, error) {
	return MapConcurrentSequence(ctx, slices.Values(items), fn, options...)

}

// MapConcurrentRef is similar to [collection.MapRef] but uses store options
// instead of a dedicated worker-count parameter.
//
// Results are returned in input order. Nil mapped values are skipped.
func MapConcurrentRef[I any, O any](ctx context.Context, items []I, fn collection.MapRefFunc[I, O], options ...StoreOption) ([]O, error) {
	return TransformInOrder(ctx, slices.Values(items), func(fCtx context.Context, item I) (output O, success bool, err error) {
		err = DetermineContextError(fCtx)
		if err != nil {
			return
		}
		mapped := fn(field.ToOptionalOrNilIfEmpty(item))
		if mapped == nil {
			return
		}
		output = *mapped
		success = true
		return
	}, options...)
}

// MapConcurrentSequence is similar to [collection.MapSequence] but evaluates the
// mapping work concurrently and returns the eagerly collected result slice.
//
// Results are returned in input order. Use [Workers] to limit parallelism.
func MapConcurrentSequence[I any, O any](ctx context.Context, items iter.Seq[I], fn collection.MapFunc[I, O], options ...StoreOption) ([]O, error) {
	return TransformInOrder(ctx, items, func(fCtx context.Context, item I) (output O, success bool, err error) {
		err = DetermineContextError(fCtx)
		if err != nil {
			return
		}
		output = fn(item)
		success = true
		return
	}, options...)
}
