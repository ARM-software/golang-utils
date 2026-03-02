// Package ptr provides utilities for working with pointers.
package ptr

import "github.com/ARM-software/golang-utils/utils/reflection"

func To[T any](v T) *T {
	return &v
}

// ToOrNilIfEmpty returns a pointer to v if v is not considered empty; otherwise it returns nil.
//
// Emptiness is determined via utils/reflection.IsEmpty (e.g. "", whitespace-only strings, 0, false, nil,
// empty slices/maps, etc.).
func ToOrNilIfEmpty[T any](v T) *T {
	if reflection.IsEmpty(v) {
		return nil
	}
	return To[T](v)
}

// From returns the value pointed to by v.
//
// If v is nil, it returns the zero value of T.
func From[T any](v *T) T {
	var zero T
	return FromOrDefault(v, zero)
}

// FromOrDefault returns the value pointed to by v.
//
// If v is nil, it returns defaultValue instead.
func FromOrDefault[T any](v *T, defaultValue T) T {
	if v == nil {
		return defaultValue
	}
	return *v
}

// Deref returns the value pointed to by v.
//
// It is an alias for From.
func Deref[T any](v *T) T {
	return From[T](v)
}
