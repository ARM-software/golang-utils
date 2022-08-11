package strings

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStrings(t *testing.T) {
	// values given by https://rosettacode.org/wiki/Entropy#Groovy
	testCases := []struct {
		Input   string
		Entropy float64
	}{
		{
			"1223334444",
			1.846439344671,
		},
		{
			"1223334444555555555",
			1.969811065121,
		},
		{
			"122333",
			1.459147917061,
		},
		{
			"1227774444",
			1.846439344671,
		},
		{
			"aaBBcccDDDD",
			1.936260027482,
		},
		{
			"1234567890abcdefghijklmnopqrstuvwxyz",
			5.169925004424,
		},
		{
			"Rosetta Code",
			3.084962500407,
		},
	}
	for _, testCase := range testCases {
		entropy := CalculateStringShannonEntropy(testCase.Input)
		require.Equal(t, fmt.Sprintf("%.8f", testCase.Entropy), fmt.Sprintf("%.8f", entropy))
	}
}
