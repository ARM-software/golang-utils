package collection

import (
	"iter"
	"slices"
)

// Slice returns a subsection of slice using Python-style indexing semantics.
//
// Negative indices are interpreted relative to the end of the slice and the
// stop index is exclusive. Out-of-range bounds are clamped to the valid range.
// If the resulting bounds are empty, an empty slice is returned.
//
// References:
//   - https://docs.python.org/3/tutorial/introduction.html#lists
//   - https://www.geeksforgeeks.org/python/what-is-negative-indexing-in-python/
func Slice[S ~[]E, E any](slice S, start, stop int) S {
	start, stop = normaliseSliceBounds(len(slice), start, stop)
	if start >= stop {
		return S{}
	}
	return slice[start:stop]
}

// SliceSequence returns a subsection of seq using Python-style indexing semantics.
func SliceSequence[E any](seq iter.Seq[E], start, stop int) iter.Seq[E] {
	collected := slices.Collect(SequenceOrEmpty(seq))
	return slices.Values(Slice(collected, start, stop))
}

// Take returns the first count elements of slice.
func Take[S ~[]E, E any](slice S, count int) S {
	if count <= 0 {
		return S{}
	}
	if count >= len(slice) {
		return slice
	}
	return slice[:count]
}

// TakeSequence yields the first count elements of seq.
func TakeSequence[E any](seq iter.Seq[E], count int) iter.Seq[E] {
	if count <= 0 {
		return EmptySequence[E]()
	}
	return func(yield func(E) bool) {
		for v := range SequenceOrEmpty(seq) {
			if count == 0 || !yield(v) {
				return
			}
			count--
		}
	}
}

// Drop returns slice without its first count elements.
func Drop[S ~[]E, E any](slice S, count int) S {
	if count <= 0 {
		return slice
	}
	if count >= len(slice) {
		return S{}
	}
	return slice[count:]
}

// DropSequence skips the first count elements of seq.
func DropSequence[E any](seq iter.Seq[E], count int) iter.Seq[E] {
	if count <= 0 {
		return SequenceOrEmpty(seq)
	}
	return func(yield func(E) bool) {
		for v := range SequenceOrEmpty(seq) {
			if count > 0 {
				count--
				continue
			}
			if !yield(v) {
				return
			}
		}
	}
}

// TakeWhile returns the leading elements of slice that satisfy predicate.
func TakeWhile[S ~[]E, E any](slice S, predicate Predicate[E]) S {
	return slices.Collect(TakeWhileSequence(slices.Values(slice), predicate))
}

// TakeWhileSequence yields the leading elements of seq that satisfy predicate.
func TakeWhileSequence[E any](seq iter.Seq[E], predicate Predicate[E]) iter.Seq[E] {
	return func(yield func(E) bool) {
		for v := range SequenceOrEmpty(seq) {
			if !predicate(v) || !yield(v) {
				return
			}
		}
	}
}

// DropWhile returns slice without its leading elements that satisfy predicate.
func DropWhile[S ~[]E, E any](slice S, predicate Predicate[E]) S {
	return slices.Collect(DropWhileSequence(slices.Values(slice), predicate))
}

// DropWhileSequence skips the leading elements of seq that satisfy predicate.
func DropWhileSequence[E any](seq iter.Seq[E], predicate Predicate[E]) iter.Seq[E] {
	return func(yield func(E) bool) {
		dropping := true
		for v := range SequenceOrEmpty(seq) {
			if dropping && predicate(v) {
				continue
			}
			dropping = false
			if !yield(v) {
				return
			}
		}
	}
}

// PopAt removes and returns the element at index from slice.
//
// Negative indices follow the same Python-style semantics as [At].
func PopAt[S ~[]E, E any](slice S, index int) (element E, remaining S, ok bool) {
	index = normaliseIndex(len(slice), index)
	if index < 0 || index >= len(slice) {
		remaining = slice
		return
	}
	element = slice[index]
	remaining = append(slices.Clone(slice[:index]), slice[index+1:]...)
	ok = true
	return
}

func normaliseSliceBounds(length, start, stop int) (int, int) {
	start = normaliseSliceBound(length, start)
	stop = normaliseSliceBound(length, stop)
	if start < 0 {
		start = 0
	}
	if stop < 0 {
		stop = 0
	}
	if start > length {
		start = length
	}
	if stop > length {
		stop = length
	}
	return start, stop
}

func normaliseSliceBound(length, index int) int {
	if index < 0 {
		return length + index
	}
	return index
}
