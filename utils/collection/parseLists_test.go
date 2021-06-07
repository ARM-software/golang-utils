package collection

import (
	"math/rand"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/require"
)

func TestParseCommaSeparatedListWordsOnly(t *testing.T) {
	stringList := ""
	stringArray := []string{}
	// we don't need cryptographically secure random numbers for generating a number of elements in a list
	lengthOfList := rand.Intn(10) //nolint:gosec
	for i := 0; i < lengthOfList; i++ {
		word := faker.Word()
		stringList += word
		stringArray = append(stringArray, word)
		numSpacesToAdd := rand.Intn(5) //nolint:gosec
		for j := 0; j < numSpacesToAdd; j++ {
			stringList += " "
		}
		stringList += ","
	}
	finalList := ParseCommaSeparatedList(stringList)
	require.Equal(t, stringArray, finalList)
}

// Test to makje sure that spaces that show up within the words aren't removed
func TestParseCommaSeparatedListWithSpacesBetweenWords(t *testing.T) {
	stringList := ""
	stringArray := []string{}
	// we don't need cryptographically secure random numbers for generating a number of elements in a list
	lengthOfList := rand.Intn(10) //nolint:gosec
	for i := 0; i < lengthOfList; i++ {
		word := faker.Sentence()
		stringList += word
		stringArray = append(stringArray, word)
		numSpacesToAdd := rand.Intn(5) //nolint:gosec
		for j := 0; j < numSpacesToAdd; j++ {
			stringList += " "
		}
		stringList += ","
	}
	finalList := ParseCommaSeparatedList(stringList)
	require.Equal(t, stringArray, finalList)
}
