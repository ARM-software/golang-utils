package collection

import (
	"iter"

	"github.com/ARM-software/golang-utils/utils/field"
)

func sign(x int) int {
	if x < 0 {
		return -1
	}
	return 1
}

// Range returns a slice of integers similar to Python's built-in range().
// https://docs.python.org/2/library/functions.html#range
//
//	Note: The stop value is always exclusive.
func Range(start, stop int, step *int) (result []int) {
	it, length := rangeSequence(start, stop, step)
	result = make([]int, length)
	i := 0
	for v := range it {
		result[i] = v
		i++
	}
	return result
}

// RangeSequence returns an iterator over a range
func RangeSequence(start, stop int, step *int) iter.Seq[int] {
	it, _ := rangeSequence(start, stop, step)
	return it
}

func rangeSequence(start, stop int, step *int) (it iter.Seq[int], length int) {
	s := field.OptionalInt(step, 1)
	length = 0
	if s == 0 {
		it = func(yield func(int) bool) {}
		return
	}
	// Compute length
	if (s > 0 && start < stop) || (s < 0 && start > stop) {
		length = (stop - start + s - sign(s)) / s
	}
	it = func(yield func(int) (end bool)) {
		for i, v := 0, start; i < length; i, v = i+1, v+s {
			if !yield(v) {
				return
			}
		}
	}
	return
}
