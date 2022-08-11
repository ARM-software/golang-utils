package strings

import (
	"math"
	"strings"

	"github.com/ARM-software/golang-utils/utils/reflection"
)

// CalculateStringShannonEntropy measures the Shannon entropy of a string
// the returned value is a bits/byte 'entropy' value between 0.0 and 8.0,
// 0 being no, and 8 being maximal entropy.
// See http://bearcave.com/misl/misl_tech/wavelets/compression/shannon.html for the algorithmic explanation.
// Code comes from https://rosettacode.org/wiki/Entropy#Go:_Slice_version
func CalculateStringShannonEntropy(str string) (entropy float64) {
	if reflection.IsEmpty(str) {
		return 0
	}
	for i := 0; i < 256; i++ {
		px := float64(strings.Count(str, string(byte(i)))) / float64(len(str))
		if px > 0 {
			entropy += -px * math.Log2(px)
		}
	}
	return entropy
}
