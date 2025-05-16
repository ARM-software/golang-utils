/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package collection

import (
	"maps"
	"math/rand"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

var (
	random = rand.New(rand.NewSource(time.Now().Unix())) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec as this is just for
)

func TestParseCommaSeparatedListWordsOnly(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		stringArray := []string{faker.Word(), faker.Word(), faker.Word(), faker.UUIDDigit(), faker.URL(), faker.Username(), faker.DomainName()}
		require.Equal(t, stringArray, ParseCommaSeparatedList(ConvertSliceToCommaSeparatedList[string](stringArray)))
	})
	t.Run("with whitespaces", func(t *testing.T) {
		stringList := ""
		var stringArray []string
		// we don't need cryptographically secure random numbers for generating a number of elements in a list
		lengthOfList := random.Intn(10) //nolint:gosec
		for i := 0; i < lengthOfList; i++ {
			word := faker.Word()
			stringList += word
			stringArray = append(stringArray, word)
			numSpacesToAdd := random.Intn(5) //nolint:gosec
			for j := 0; j < numSpacesToAdd; j++ {
				stringList += " "
			}
			stringList += ","
		}
		finalList := ParseCommaSeparatedList(stringList)
		require.Equal(t, stringArray, finalList)
	})
}

// Test to make sure that spaces that show up within the words aren't removed
func TestParseCommaSeparatedListWithSpacesBetweenWords(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		stringArray := []string{faker.Paragraph(), faker.Word(), faker.Sentence(), faker.UUIDDigit(), faker.Name(), faker.Username(), faker.DomainName()}
		require.Equal(t, stringArray, ParseCommaSeparatedList(ConvertSliceToCommaSeparatedList[string](stringArray)))
	})
	t.Run("with whitespaces", func(t *testing.T) {
		stringList := ""
		var stringArray []string
		// we don't need cryptographically secure random numbers for generating a number of elements in a list
		lengthOfList := random.Intn(10) //nolint:gosec
		for i := 0; i < lengthOfList; i++ {
			word := faker.Sentence()
			stringList += word
			stringArray = append(stringArray, word)
			numSpacesToAdd := random.Intn(5) //nolint:gosec
			for j := 0; j < numSpacesToAdd; j++ {
				stringList += " "
			}
			stringList += ","
		}
		finalList := ParseCommaSeparatedList(stringList)
		require.Equal(t, stringArray, finalList)
	})
}

func TestParseCommaSeparatedListWithSpacesBetweenWordsKeepBlanks(t *testing.T) {
	stringList := ""
	var stringArray []string
	// we don't need cryptographically secure random numbers for generating a number of elements in a list
	lengthOfList := random.Intn(10) + 8 //nolint:gosec
	for i := 0; i < lengthOfList; i++ {
		word := faker.Sentence()
		stringList += word
		stringArray = append(stringArray, word)
		numSpacesToAdd := random.Intn(5) //nolint:gosec
		for j := 0; j < numSpacesToAdd; j++ {
			stringList += " "
		}
		stringList += ","
		if i%3 == 2 {
			numSpacesToAdd := random.Intn(5) //nolint:gosec
			for j := 0; j < numSpacesToAdd; j++ {
				stringList += " "
			}
			stringArray = append(stringArray, "")
			stringList += ","
		}
	}
	stringArray = append(stringArray, "") // account for final ,

	finalList1 := ParseCommaSeparatedList(stringList)
	require.NotEqual(t, stringArray, finalList1)

	finalList2 := ParseListWithCleanupKeepBlankLines(stringList, ",")
	require.Equal(t, stringArray, finalList2)
}

func TestParseCommaSeparatedListToMap(t *testing.T) {
	randomMap := map[string]string{
		faker.Sentence(): faker.Sentence(),
		faker.Word():     faker.Paragraph(),
		faker.Name():     faker.Sentence(),
		faker.Sentence(): faker.Sentence(),
		faker.Word():     faker.Paragraph(),
		faker.Name():     faker.Sentence(),
		faker.Sentence(): faker.Sentence(),
		faker.Word():     faker.Paragraph(),
		faker.Name():     faker.Sentence(),
	}
	for _, test := range []struct {
		Name     string
		Input    string
		Expected map[string]string
		Err      error
	}{
		{"Normal 1", "hello,world", map[string]string{"hello": "world"}, nil},
		{"Normal 2", "hello,world,adrien,cabarbaye", map[string]string{"hello": "world", "adrien": "cabarbaye"}, nil},
		{"Normal 2.5", "hello, world, adrien, cabarbaye", map[string]string{"hello": "world", "adrien": "cabarbaye"}, nil},
		{"Normal 3", "hello,world,adrien,cabarbaye,", map[string]string{"hello": "world", "adrien": "cabarbaye"}, nil},
		{"Normal 4", "hello,,world,adrien,,,cabarbaye,,,", map[string]string{"hello": "world", "adrien": "cabarbaye"}, nil},
		{"Normal 5", "hello,world,this,value has spaces", map[string]string{"hello": "world", "this": "value has spaces"}, nil},
		{"Normal 6", "hello,,world,this,,,value has spaces,,,", map[string]string{"hello": "world", "this": "value has spaces"}, nil},
		{"Normal 7", "", map[string]string{}, nil},
		{"Normal 8", ",", map[string]string{}, nil},
		{"Normal 9", ",,,,,", map[string]string{}, nil},
		{"Normal 10", ",, ,,  ,", map[string]string{}, nil},
		{"Normal 11", ConvertMapToCommaSeparatedList[string, string](randomMap), randomMap, nil},
		{"Bad 1", "one", nil, commonerrors.ErrInvalid},
		{"Bad 1", "one, two, three", nil, commonerrors.ErrInvalid},
		{"Bad 2", "one element with spaces", nil, commonerrors.ErrInvalid},
		{"Bad 3", "one element with spaces and end comma,", nil, commonerrors.ErrInvalid},
		{"Bad 4", "one element with spaces and multiple end commas,,,", nil, commonerrors.ErrInvalid},
		{"Bad 5", ",,,one element with spaces and multiple end/beginning commas,,,", nil, commonerrors.ErrInvalid},
	} {
		t.Run(test.Name, func(t *testing.T) {
			pairs, err := ParseCommaSeparatedListToMap(test.Input)
			errortest.AssertError(t, err, test.Err)
			assert.True(t, maps.Equal(test.Expected, pairs))
		})
	}
}

func TestParseCommaSeparatedPairListToMap(t *testing.T) {
	randomMap := map[string]string{
		faker.Sentence(): faker.Sentence(),
		faker.Word():     faker.Paragraph(),
		faker.Name():     faker.Sentence(),
		faker.Sentence(): faker.Sentence(),
		faker.Word():     faker.Paragraph(),
		faker.Name():     faker.Sentence(),
		faker.Sentence(): faker.Sentence(),
		faker.Word():     faker.Paragraph(),
		faker.Name():     faker.Sentence(),
	}
	for _, test := range []struct {
		Name          string
		Input         string
		Expected      map[string]string
		Err           error
		PairSeparator string
	}{
		{"Normal 1", "hello=world", map[string]string{"hello": "world"}, nil, "="},
		{"Normal 2", "hello+world,adrien+cabarbaye", map[string]string{"hello": "world", "adrien": "cabarbaye"}, nil, "+"},
		{"Normal 2", "hello, world, adrien, cabarbaye", map[string]string{"hello": "world", "adrien": "cabarbaye"}, nil, ","},
		{"Normal 2.5", "hello= world, adrien = cabarbaye", map[string]string{"hello": "world", "adrien": "cabarbaye"}, nil, "="},
		{"Normal 3", "hello&world,adrien&cabarbaye,", map[string]string{"hello": "world", "adrien": "cabarbaye"}, nil, "&"},
		{"Normal 4", "hello%%world,,,,adrien%%cabarbaye,,,", map[string]string{"hello": "world", "adrien": "cabarbaye"}, nil, "%%"},
		{"Normal 5", "hello$$$world,this$$$value has spaces", map[string]string{"hello": "world", "this": "value has spaces"}, nil, "$$$"},
		{"Normal 6", "hello__world,,,this__value has spaces,,,", map[string]string{"hello": "world", "this": "value has spaces"}, nil, "__"},
		{"Normal 7", "", map[string]string{}, nil, "+"},
		{"Normal 8", ",", map[string]string{}, nil, "+"},
		{"Normal 9", ",,,,,", map[string]string{}, nil, "^"},
		{"Normal 10", ",, ,,  ,", map[string]string{}, nil, "+"},
		{"Normal 11", ConvertMapToCommaSeparatedPairsList[string, string](randomMap, "/"), randomMap, nil, "/"},
		{"Normal 12", ConvertMapToCommaSeparatedPairsList[string, string](randomMap, "  "), randomMap, nil, "  "},
		{"Bad 1", "one", nil, commonerrors.ErrInvalid, "+"},
		{"Bad 1", "one, two, three", nil, commonerrors.ErrInvalid, "+"},
		{"Bad 2", "one element with spaces", nil, commonerrors.ErrInvalid, "+"},
		{"Bad 3", "one element with spaces and end comma,", nil, commonerrors.ErrInvalid, "+"},
		{"Bad 4", "one element with spaces and multiple end commas,,,", nil, commonerrors.ErrInvalid, "+"},
		{"Bad 5", ",,,one element with spaces and multiple end/beginning commas,,,", nil, commonerrors.ErrInvalid, "="},
	} {
		t.Run(test.Name, func(t *testing.T) {
			pairs, err := ParseCommaSeparatedListOfPairsToMap(test.Input, test.PairSeparator)
			errortest.AssertError(t, err, test.Err)
			assert.True(t, maps.Equal(test.Expected, pairs))
		})
	}
}
