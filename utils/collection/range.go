package collection

import "github.com/ARM-software/golang-utils/utils/field"

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
func Range(start, stop int, step *int) []int {
	s := field.OptionalInt(step, 1)
	if s == 0 {
		return []int{}
	}

	// Compute length
	length := 0
	if (s > 0 && start < stop) || (s < 0 && start > stop) {
		length = (stop - start + s - sign(s)) / s
	}

	result := make([]int, length)
	for i, v := 0, start; i < length; i, v = i+1, v+s {
		result[i] = v
	}
	return result
}
