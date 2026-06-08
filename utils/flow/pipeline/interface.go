package pipeline

import (
	"context"

	flowstep "github.com/ARM-software/golang-utils/utils/flow/step"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

//go:generate go tool mockgen -destination=../../mocks/mock_flow_pipeline.go -package=mocks github.com/ARM-software/golang-utils/utils/flow/pipeline ISimplePipeline,IPipeline

type ISimplePipeline[T any] interface {
	RegisterStep(...flowstep.IStep[T, T])
	RegisterTransform(name string, transform flowstep.TransformFunc[T, T], options ...flowstep.Option[T])
	RegisterTransformGroup(name *string, group *parallelisation.TransformGroup[T, T])
	Execute(context.Context, T) (T, error)
}

type IPipeline[I, O any] interface {
	Execute(context.Context, I) (O, error)
}
