package parallelisation

import (
	"context"

	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
)

type TransformFunc[I any, O any] func(context.Context, I) (output O, success bool, err error)

type results[O any] struct {
	terminated *atomic.Bool
	r          chan O
}

func (r *results[O]) Append(o O) {
	if !r.terminated.Load() {
		r.r <- o
	}
}

func (r *results[O]) Results(ctx context.Context) (slice []O, err error) {
	if !r.terminated.Swap(true) {
		close(r.r)
	}
	err = DetermineContextError(ctx)
	if err != nil {
		return
	}
	slice = make([]O, 0, len(r.r))
	for output := range r.r {
		err = DetermineContextError(ctx)
		if err != nil {
			return
		}
		slice = append(slice, output)
	}
	return
}

func newResults[O any](numberOfInput *int) *results[O] {
	i := field.OptionalInt(numberOfInput, 0)
	var channel chan O
	if i <= 0 {
		channel = make(chan O)
	} else {
		channel = make(chan O, i)
	}

	return &results[O]{
		terminated: atomic.NewBool(false),
		r:          channel,
	}
}

type TransformGroup[I any, O any] struct {
	ExecutionGroup[I]
	results *atomic.Pointer[results[O]]
}

func (g *TransformGroup[I, O]) appendResult(o O) {
	r := g.results.Load()
	if r != nil {
		r.Append(o)
	}
}

// Inputs registers inputs to transform.
func (g *TransformGroup[I, O]) Inputs(ctx context.Context, i ...I) error {
	for j := range i {
		err := DetermineContextError(ctx)
		if err != nil {
			return err
		}
		g.RegisterFunction(i[j])
	}
	return nil
}

// Outputs returns any input which have been transformed when the Transform function was called.
func (g *TransformGroup[I, O]) Outputs(ctx context.Context) ([]O, error) {
	r := g.results.Load()
	if r == nil {
		return nil, commonerrors.UndefinedVariable("results")
	}
	return r.Results(ctx)
}

// Transform actually performs the transformation
func (g *TransformGroup[I, O]) Transform(ctx context.Context) error {
	g.results.Store(newResults[O](field.ToOptionalInt(g.Len())))
	return g.ExecutionGroup.Execute(ctx)
}

// NewTransformGroup returns a group transforming inputs into outputs.
// To register inputs, call the Input function
// To perform the transformation of inputs, then call Transform
// To retrieve the output, then call Output
func NewTransformGroup[I any, O any](transform TransformFunc[I, O], options ...StoreOption) *TransformGroup[I, O] {
	g := &TransformGroup[I, O]{
		results: atomic.NewPointer[results[O]](newResults[O](nil)),
	}
	g.ExecutionGroup = *NewExecutionGroup[I](func(fCtx context.Context, i I) error {
		err := DetermineContextError(fCtx)
		if err != nil {
			return err
		}
		o, success, err := transform(fCtx, i)
		if err != nil {
			return commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "an error occurred whilst handling an input [%+v]", i)
		}
		if success {
			g.appendResult(o)
		}
		return nil
	}, options...)
	return g
}
