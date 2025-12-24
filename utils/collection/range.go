package collection

import (
	"iter"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/safecast"
)

func sign[T safecast.IConvertable](x T) int64 {
	if x < 0 {
		return -1
	}
	return 1
}

func determineRangeLength[T safecast.IConvertable](start, stop, step T) (length int) {
	if (step > 0 && start < stop) || (step < 0 && start > stop) {
		length = safecast.ToInt((safecast.ToInt64(stop-start+step) - sign(step)) / safecast.ToInt64(step))
	}
	return
}

// Range returns a slice of integers similar to Python's built-in range().
// https://docs.python.org/2/library/functions.html#range
//
//	Note: The stop value is always exclusive.
func Range[T safecast.IConvertable](start, stop T, step *T) (result []T) {
	it, length := rangeSequence(start, stop, step)
	result = make([]T, length)
	i := 0
	for v := range it {
		result[i] = v
		i++
	}
	return result
}

// RangeSequence returns an iterator over a range
func RangeSequence[T safecast.IConvertable](start, stop T, step *T) iter.Seq[T] {
	it, _ := rangeSequence(start, stop, step)
	return it
}

func rangeSequence[T safecast.IConvertable](start, stop T, step *T) (it iter.Seq[T], length int) {
	s := field.Optional[T](step, 1)
	if s == 0 {
		it = func(yield func(T) bool) {}
		return
	}
	length = determineRangeLength[T](start, stop, s)
	it = func(yield func(T) (end bool)) {
		for i, v := 0, start; i < length; i, v = i+1, v+s {
			if !yield(v) {
				return
			}
		}
	}
	return
}
