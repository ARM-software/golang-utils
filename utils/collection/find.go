package collection

import (
	"iter"
	"slices"
	"strings"

	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/safecast"
)

//
// Find utilities
//

// Find searches for val in the slice pointed to by slice.
// It returns the index of the first match and true when found.
// If slice is nil or val is not present, it returns -1 and false.
func Find(slice *[]string, val string) (int, bool) {
	if slice == nil {
		return -1, false
	}
	return FindInSlice(true, *slice, val)
}

// FindInSequence searches elements (a sequence) for the first item that
// satisfies predicate. It returns the zero-based index of the matching
// element and true when a match is found. If elements is nil or no
// match exists, it returns -1 and false.
func FindInSequence[E any](elements iter.Seq[E], predicate Predicate[E]) (int, bool) {
	if elements == nil {
		return -1, false
	}
	idx := atomic.NewUint64(0)
	for e := range elements {
		if predicate(e) {
			return safecast.ToInt(idx.Load()), true
		}
		idx.Inc()
	}
	return -1, false
}

// FindInSequenceRef behaves like FindInSequence but accepts a predicate
// that operates on element references.
func FindInSequenceRef[E any](elements iter.Seq[E], predicate PredicateRef[E]) (int, bool) {
	return FindInSequence(elements, toPredicateFunc(predicate))
}

// FindInSlice checks whether any of the provided val arguments exist in
// slice. It returns the index of the first match and true if found.
// When strict is true, matching uses exact string equality. When strict
// is false, matching is case-insensitive and ignores surrounding
// whitespace.
//
// If no values are provided or the slice is empty, the function returns
// -1 and false.
func FindInSlice(strict bool, slice []string, val ...string) (int, bool) {
	if len(val) == 0 || len(slice) == 0 {
		return -1, false
	}
	if strict && len(val) == 1 {
		idx := slices.Index(slice, val[0])
		return idx, idx >= 0
	}

	inSlice := make(map[string]int, len(slice))
	for i := range slice {
		item := slice[i]
		if !strict {
			item = strings.ToLower(strings.TrimSpace(item))
		}
		if _, ok := inSlice[item]; !ok {
			inSlice[item] = i
		}
	}

	for i := range val {
		item := val[i]
		if !strict {
			item = strings.ToLower(strings.TrimSpace(item))
		}
		if idx, ok := inSlice[item]; ok {
			return idx, true
		}
	}
	return -1, false
}
