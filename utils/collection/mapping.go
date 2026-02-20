package collection

import (
	"iter"
	"slices"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
)

//
// Mapping utilities
//

// MapFunc defines a function that maps a value of type T1 to type T2.
type MapFunc[T1, T2 any] func(T1) T2

// MapRefFunc defines a mapping function that accepts a pointer to T1 and
// returns a pointer to T2.
type MapRefFunc[T1, T2 any] func(*T1) *T2

// MapWithErrorFunc defines a mapping function that may return an error.
type MapWithErrorFunc[T1, T2 any] func(T1) (T2, error)

// MapRefWithErrorFunc defines a reference-based mapping function that may return an error.
type MapRefWithErrorFunc[T1, T2 any] func(*T1) (*T2, error)

// IdentityMapFunc returns a mapping function that returns its input unchanged.
func IdentityMapFunc[T any]() MapFunc[T, T] {
	return func(i T) T { return i }
}

// MapSequence maps each element of s using f and returns a sequence of mapped values.
func MapSequence[T1 any, T2 any](s iter.Seq[T1], f MapFunc[T1, T2]) iter.Seq[T2] {
	return MapSequenceWithError(s, func(t1 T1) (T2, error) {
		return f(t1), nil
	})
}

// MapSequenceWithError maps each element of s using f, which may return an error.
// Mapping stops if f returns an error or if the consumer declines the yielded value.
func MapSequenceWithError[T1 any, T2 any](s iter.Seq[T1], f MapWithErrorFunc[T1, T2]) iter.Seq[T2] {
	return func(yield func(T2) bool) {
		for v := range s {
			mapped, err := f(v)
			if err != nil || !yield(mapped) {
				return
			}
		}
	}
}

// MapSequenceRef maps a sequence by applying a reference-based mapper to each element.
func MapSequenceRef[T1 any, T2 any](s iter.Seq[T1], f MapRefFunc[T1, T2]) iter.Seq[T2] {
	return MapSequenceRefWithError(s, func(t1 *T1) (*T2, error) {
		return f(t1), nil
	})
}

// MapSequenceRefWithError maps a sequence using a reference-based mapper which may return an error.
// Mapping stops if the mapper returns an error, returns nil, or if the consumer declines the yielded value.
func MapSequenceRefWithError[T1 any, T2 any](s iter.Seq[T1], f MapRefWithErrorFunc[T1, T2]) iter.Seq[T2] {
	return func(yield func(T2) bool) {
		for v := range s {
			mapped, err := f(field.ToOptionalOrNilIfEmpty(v))
			if err != nil || mapped == nil || !yield(*mapped) {
				return
			}
		}
	}
}

// Map applies f to each element of s and returns a slice with the results.
func Map[T1 any, T2 any](s []T1, f MapFunc[T1, T2]) []T2 {
	return slices.Collect[T2](MapSequence(slices.Values(s), f))
}

// MapWithError applies f to each element of s where f may return an error.
// If an error occurs, processing stops and the error is returned.
func MapWithError[T1 any, T2 any](s []T1, f MapWithErrorFunc[T1, T2]) (result []T2, err error) {
	result = make([]T2, len(s))

	for i := range s {
		var subErr error
		result[i], subErr = f(s[i])
		if subErr != nil {
			err = subErr
			return
		}
	}

	return
}

// MapRef applies a reference-based mapping function to s and returns the mapped slice.
func MapRef[T1 any, T2 any](s []T1, f MapRefFunc[T1, T2]) []T2 {
	return slices.Collect[T2](MapSequenceRef(slices.Values(s), f))
}

// MapRefWithError applies a reference-based mapper that may return an error.
// If the mapper returns nil or an error, processing stops and an error is returned.
func MapRefWithError[T1 any, T2 any](s []T1, f MapRefWithErrorFunc[T1, T2]) (result []T2, err error) {
	result = make([]T2, len(s))

	for i := range s {
		var subErr error
		r, subErr := f(field.ToOptionalOrNilIfEmpty(s[i]))
		if subErr != nil {
			err = subErr
			return
		}
		if r == nil {
			err = commonerrors.UndefinedParameterf("item #%v was nil", i)
			return
		}
		result[i] = *r
	}

	return
}
