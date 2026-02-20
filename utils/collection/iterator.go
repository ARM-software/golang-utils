package collection

import (
	"iter"
	"slices"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
)

//
// Iteration utilities
//

// OperationFunc defines an operation on a value that may return an error.
type OperationFunc[E any] func(E) error

// OperationRefFunc defines an operation on a pointer to a value that may return an error.
type OperationRefFunc[E any] func(*E) error

// OperationWithoutErrorFunc defines an operation on a value that does not return an error.
type OperationWithoutErrorFunc[E any] func(E)

// OperationWithoutErrorRefFunc defines an operation on a pointer that does not return an error.
type OperationWithoutErrorRefFunc[E any] func(*E)

// toOperationFunc adapts an OperationRefFunc to an OperationFunc by
// converting the value to an optional reference.
func toOperationFunc[E any](f OperationRefFunc[E]) OperationFunc[E] {
	return func(e E) error {
		return f(field.ToOptional(e))
	}
}

// toOperationWithoutErrorFunc adapts an OperationWithoutErrorRefFunc to
// an OperationWithoutErrorFunc by converting the value to an optional reference.
func toOperationWithoutErrorFunc[E any](f OperationWithoutErrorRefFunc[E]) OperationWithoutErrorFunc[E] {
	return func(e E) {
		f(field.ToOptional(e))
	}
}

// convertOperationWithoutError wraps a non-error operation so it conforms
// to OperationFunc by always returning nil.
func convertOperationWithoutError[E any](f OperationWithoutErrorFunc[E]) OperationFunc[E] {
	return func(e E) error {
		f(e)
		return nil
	}
}

// Each iterates over a sequence and invokes f for each element. If f
// returns a non-EOF error, iteration stops and that error is returned.
// If f returns EOF, the EOF is ignored and iteration ends without error.
func Each[T any](s iter.Seq[T], f OperationFunc[T]) error {
	for e := range s {
		err := f(e)
		if err != nil {
			err = commonerrors.Ignore(err, commonerrors.ErrEOF)
			return err
		}
	}
	return nil
}

// EachRef behaves like Each but invokes f with a reference to each element.
func EachRef[T any](s iter.Seq[T], f OperationRefFunc[T]) error {
	return Each(s, toOperationFunc(f))
}

// ForEach invokes f on every element of the provided slice. Any error
// returned by f is ignored.
func ForEach[S ~[]E, E any](s S, f OperationWithoutErrorFunc[E]) {
	_ = Each[E](slices.Values(s), convertOperationWithoutError(f))
}

// ForEachValues invokes f for each value passed in values.
func ForEachValues[E any](f func(E), values ...E) {
	ForEach(values, f)
}

// ForEachRef invokes f on every element of the provided slice, passing a reference.
func ForEachRef[S ~[]E, E any](s S, f OperationWithoutErrorRefFunc[E]) {
	ForEach(s, toOperationWithoutErrorFunc(f))
}

// ForAll invokes f on every element of the provided slice. Any non-EOF
// errors returned by f are collected and aggregated into a single error
// returned to the caller. If f returns EOF, iteration stops immediately.
func ForAll[S ~[]E, E any](s S, f OperationFunc[E]) error {
	return ForAllSequence[E](slices.Values(s), f)
}

// ForAllSequence invokes f for every element of s (a sequence). Non-EOF
// errors are wrapped and aggregated; EOF causes immediate termination.
func ForAllSequence[T any](s iter.Seq[T], f OperationFunc[T]) error {
	var err error
	err = commonerrors.Join(err, Each[T](s, func(e T) error {
		subErr := f(e)
		if commonerrors.Any(subErr, commonerrors.ErrEOF) {
			return subErr
		}
		if subErr != nil {
			err = commonerrors.Join(err,
				commonerrors.Newf(subErr, "error during iteration over value [%v]", e))
		}
		return nil
	}))
	return err
}

// ForAllRef behaves like ForAll but applies f to references of elements.
func ForAllRef[S ~[]E, E any](s S, f OperationRefFunc[E]) error {
	return ForAll(s, toOperationFunc(f))
}

// ForAllSequenceRef behaves like ForAllSequence but adapts a reference-based operation.
func ForAllSequenceRef[T any](s iter.Seq[T], f OperationRefFunc[T]) error {
	return ForAllSequence(s, toOperationFunc(f))
}
