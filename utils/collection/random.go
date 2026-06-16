package collection

import (
	cryptorand "crypto/rand"
	"math/big"
	"slices"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/safecast"
)

// Random returns a random element from the slice.
//
// This is useful when one arbitrary element should be sampled from a slice,
// such as choosing a test case, backend, worker, or candidate value.
//
// The first return value is the selected element.
// The second return value reports whether an element could be selected.
// If the slice is empty, Random returns the zero value of E and false.
//
// Reference documentation:
//   - https://pkg.go.dev/crypto/rand
func Random[S ~[]E, E any](slice S) (E, bool) {
	if reflection.IsEmpty(slice) {
		var zero E
		return zero, false
	}

	idx, err := cryptoIntn(len(slice))
	if err != nil {
		var zero E
		return zero, false
	}

	return slice[idx], true
}

// Shuffle returns a shuffled copy of the slice.
//
// This is useful when the same elements should be processed in a random order
// without mutating the original slice.
//
// Reference documentation:
//   - https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
//   - https://pkg.go.dev/crypto/rand
func Shuffle[S ~[]E, E any](slice S) S {
	result := slices.Clone(slice)
	for i := len(result) - 1; i > 0; i-- {
		j, err := cryptoIntn(i + 1)
		if err != nil {
			return result
		}
		result[i], result[j] = result[j], result[i]
	}
	return result
}

func cryptoIntn(max int) (int, error) {
	if max <= 0 {
		return 0, commonerrors.UndefinedVariable("maximum random range")
	}
	value, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return safecast.ToInt(value.Int64()), nil
}
