package collection

import (
	"iter"
	"slices"

	mapset "github.com/deckarep/golang-set/v2"
)

//
// Set operations
//

// UniqueEntries returns a slice containing the distinct values from the
// provided slice. The order of elements is not guaranteed.
func UniqueEntries[T comparable](slice []T) []T {
	subSet := mapset.NewSet[T]()
	_ = subSet.Append(slice...)
	return subSet.ToSlice()
}

// Unique returns the distinct values from the provided sequence.
// The order of elements is not guaranteed.
func Unique[T comparable](s iter.Seq[T]) []T {
	return UniqueEntries(slices.Collect(s))
}

// Union returns the union of slice1 and slice2, containing only unique
// values. The order of elements is not guaranteed.
func Union[T comparable](slice1, slice2 []T) []T {
	subSet := mapset.NewSet[T]()
	_ = subSet.Append(slice1...)
	_ = subSet.Append(slice2...)
	return subSet.ToSlice()
}

// Intersection returns the distinct values common to slice1 and slice2.
// The order of elements is not guaranteed.
func Intersection[T comparable](slice1, slice2 []T) []T {
	subSet1 := mapset.NewSet[T]()
	subSet2 := mapset.NewSet[T]()
	_ = subSet1.Append(slice1...)
	_ = subSet2.Append(slice2...)
	return subSet1.Intersect(subSet2).ToSlice()
}

// Difference returns distinct values present in slice1 but not in slice2.
func Difference[T comparable](slice1, slice2 []T) []T {
	subSet1 := mapset.NewSet[T]()
	subSet2 := mapset.NewSet[T]()
	_ = subSet1.Append(slice1...)
	_ = subSet2.Append(slice2...)
	return subSet1.Difference(subSet2).ToSlice()
}

// SymmetricDifference returns distinct values that are present in either
// slice1 or slice2 but not in both.
func SymmetricDifference[T comparable](slice1, slice2 []T) []T {
	subSet1 := mapset.NewSet[T]()
	subSet2 := mapset.NewSet[T]()
	_ = subSet1.Append(slice1...)
	_ = subSet2.Append(slice2...)
	return subSet1.SymmetricDifference(subSet2).ToSlice()
}
