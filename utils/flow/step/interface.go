package step

import "context"

//go:generate go tool mockgen -destination=../../mocks/mock_flow_step.go -package=mocks github.com/ARM-software/golang-utils/utils/flow/step IStep

// IStep defines one executable stage in a flow.
//
// `I` is the step input type and `O` is the step output type.
//
// Step names are optional. Implementations may return nil from [IStep.GetName]
// when no human-readable label is needed.
type IStep[I, O any] interface {
	GetName() *string
	GetOptions() *StepOptions[I, O]
	Execute(context.Context, I) (output O, success bool, err error)
	ExecuteAsNew(context.Context, *StepOptions[I, O], I) (output O, success bool, err error)
}
