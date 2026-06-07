package pipeline

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

func transformWithGroup[I, O any](transform parallelisation.TransformFunc[I, O], options []parallelisation.StoreOption) parallelisation.TransformFunc[I, O] {
	return func(ctx context.Context, input I) (output O, success bool, err error) {
		return executeTransformGroup(ctx, parallelisation.NewTransformGroup[I, O](transform, options...), input)
	}
}
