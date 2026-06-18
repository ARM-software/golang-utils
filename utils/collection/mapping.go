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
//
// Mapping functions are useful for transforming collections from one shape into
// another, such as extracting a field, converting one type to another, or
// preparing values for serialisation or display.
type MapFunc[T1, T2 any] func(T1) T2

// MapRefFunc defines a mapping function that accepts a pointer to T1 and
// returns a pointer to T2.
//
// Reference-based mappers are useful when the mapping logic naturally works on
// pointers, when a value may be optional, or when larger values should not be
// copied unnecessarily.
type MapRefFunc[T1, T2 any] func(*T1) *T2

// MapWithErrorFunc defines a mapping function that may return an error.
type MapWithErrorFunc[T1, T2 any] func(T1) (T2, error)

// MapRefWithErrorFunc defines a reference-based mapping function that may return
// an error.
type MapRefWithErrorFunc[T1, T2 any] func(*T1) (*T2, error)

// IdentityMapFunc returns a mapping function that returns its input unchanged.
func IdentityMapFunc[T any]() MapFunc[T, T] {
	return func(i T) T { return i }
}

// Combine composes mapping functions of the same input/output type.
//
// Functions are applied from right to left, so
// `Combine(f1, f2, f3)(value)` is equivalent to `f1(f2(f3(value)))`.
// This corresponds to the composition `f1 o f2 o f3`.
//
// If no functions are supplied, the identity mapping is returned.
//
// Reference documentation:
//   - https://en.wikipedia.org/wiki/Function_composition
func Combine[T any](functions ...MapFunc[T, T]) MapFunc[T, T] {
	return func(value T) T {
		result := value
		ForEach(Reverse(functions), func(f MapFunc[T, T]) {
			if f != nil {
				result = f(result)
			}
		})
		return result
	}
}

// CombineRef composes reference-based mapping functions of the same input/output type.
//
// Functions are applied from right to left, so
// `CombineRef(f1, f2, f3)(value)` is equivalent to `f1(f2(f3(value)))`.
// This corresponds to the composition `f1 o f2 o f3`.
//
// If no functions are supplied, the identity mapping is returned.
//
// Reference documentation:
//   - https://en.wikipedia.org/wiki/Function_composition
func CombineRef[T any](functions ...MapRefFunc[T, T]) MapRefFunc[T, T] {
	return func(value *T) *T {
		result := value
		ForEach(Reverse(functions), func(f MapRefFunc[T, T]) {
			if f != nil {
				result = f(result)
			}
		})
		return result
	}
}

// CombineWithError composes error-returning mapping functions of the same input/output type.
//
// Functions are applied from right to left, so
// `CombineWithError(f1, f2, f3)(value)` is equivalent to `f1(f2(f3(value)))`.
// This corresponds to the composition `f1 o f2 o f3`.
//
// If any function returns an error, composition stops immediately and the error
// is returned.
//
// If no functions are supplied, the identity mapping is returned.
//
// Reference documentation:
//   - https://en.wikipedia.org/wiki/Function_composition
func CombineWithError[T any](functions ...MapWithErrorFunc[T, T]) MapWithErrorFunc[T, T] {
	return func(value T) (result T, err error) {
		result = value
		err = Each(slices.Values(Reverse(functions)), func(f MapWithErrorFunc[T, T]) error {
			if f == nil {
				return nil
			}
			result, err = f(result)
			return err
		})
		if err != nil {
			return result, err
		}
		return result, nil
	}
}

// CombineRefWithError composes reference-based error-returning mapping functions
// of the same input/output type.
//
// Functions are applied from right to left, so
// `CombineRefWithError(f1, f2, f3)(value)` is equivalent to `f1(f2(f3(value)))`.
// This corresponds to the composition `f1 o f2 o f3`.
//
// If any function returns an error, composition stops immediately and the error
// is returned.
//
// If no functions are supplied, the identity mapping is returned.
//
// Reference documentation:
//   - https://en.wikipedia.org/wiki/Function_composition
func CombineRefWithError[T any](functions ...MapRefWithErrorFunc[T, T]) MapRefWithErrorFunc[T, T] {
	return func(value *T) (result *T, err error) {
		result = value
		err = Each(slices.Values(Reverse(functions)), func(f MapRefWithErrorFunc[T, T]) error {
			if f == nil {
				return nil
			}
			result, err = f(result)
			return err
		})
		if err != nil {
			return result, err
		}
		return result, nil
	}
}

// MapSequence maps each element of s using f and returns a sequence of mapped
// values.
//
// This is useful when data should be transformed lazily, for example while
// streaming over iterators or when chaining other sequence operations.
//
// Reference documentation:
//   - https://en.wikipedia.org/wiki/Map_(higher-order_function)
//   - https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/map
func MapSequence[T1 any, T2 any](s iter.Seq[T1], f MapFunc[T1, T2]) iter.Seq[T2] {
	return MapSequenceWithError(s, func(t1 T1) (T2, error) {
		return f(t1), nil
	})
}

// MapSequenceWithError maps each element of s using f, which may return an
// error.
// Mapping stops if f returns an error or if the consumer declines the yielded
// value.
func MapSequenceWithError[T1 any, T2 any](s iter.Seq[T1], f MapWithErrorFunc[T1, T2]) iter.Seq[T2] {
	return func(yield func(T2) bool) {
		for v := range SequenceOrEmpty(s) {
			mapped, err := f(v)
			if err != nil || !yield(mapped) {
				return
			}
		}
	}
}

// MapSequenceRef maps a sequence by applying a reference-based mapper to each
// element.
//
// This is useful when values are consumed lazily and the mapping logic is
// already written against pointers.
func MapSequenceRef[T1 any, T2 any](s iter.Seq[T1], f MapRefFunc[T1, T2]) iter.Seq[T2] {
	return MapSequenceRefWithError(s, func(t1 *T1) (*T2, error) {
		return f(t1), nil
	})
}

// MapSequenceRefWithError maps a sequence using a reference-based mapper which
// may return an error.
// Mapping stops if the mapper returns an error, returns nil, or if the consumer
// declines the yielded value.
//
// This is useful when lazy iteration and pointer-oriented mapping need to be
// combined with validation or other error-producing transformations.
func MapSequenceRefWithError[T1 any, T2 any](s iter.Seq[T1], f MapRefWithErrorFunc[T1, T2]) iter.Seq[T2] {
	return func(yield func(T2) bool) {
		for v := range SequenceOrEmpty(s) {
			mapped, err := f(field.ToOptionalOrNilIfEmpty(v))
			if err != nil || mapped == nil || !yield(*mapped) {
				return
			}
		}
	}
}

// Map applies f to each element of s and returns a slice with the results.
//
// This is useful for eagerly transforming one slice into another, for example
// converting IDs to strings, extracting names from structs, or normalising
// values before further processing.
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

// MapRef applies a reference-based mapping function to s and returns the mapped
// slice.
//
// This is useful when eager slice transformation should reuse logic written for
// pointer-oriented mappers.
func MapRef[T1 any, T2 any](s []T1, f MapRefFunc[T1, T2]) []T2 {
	return slices.Collect[T2](MapSequenceRef(slices.Values(s), f))
}

// MapRefWithError applies a reference-based mapper that may return an error.
// If the mapper returns nil or an error, processing stops and an error is
// returned.
//
// This is useful when eager slice transformation should stop as soon as an
// invalid or missing mapped value is encountered.
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
