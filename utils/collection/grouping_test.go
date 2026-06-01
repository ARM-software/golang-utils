package collection

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCountBy(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	count := CountBy(numbers, func(number int) bool {
		return number%2 == 0
	})
	assert.Equal(t, 4, count)
	assert.Equal(t, 4, CountByRef(numbers, func(number *int) bool {
		return *number%2 == 0
	}))
	assert.Equal(t, 4, CountBySequence(slices.Values(numbers), func(number int) bool {
		return number%2 == 0
	}))
	assert.Equal(t, 4, CountByRefSequence(slices.Values(numbers), func(number *int) bool {
		return *number%2 == 0
	}))
}

func TestGroupBy(t *testing.T) {
	planets := []string{"Mercury", "Venus", "Earth", "Mars", "Jupiter", "Saturn", "Uranus", "Neptune"}
	grouped := GroupBy(planets, func(planet string) int {
		return len(planet)
	})
	require.Len(t, grouped, 4)
	assert.ElementsMatch(t, []string{"Mercury", "Jupiter", "Neptune"}, grouped[7])
	assert.ElementsMatch(t, []string{"Saturn", "Uranus"}, grouped[6])
	assert.ElementsMatch(t, []string{"Venus", "Earth"}, grouped[5])
	assert.ElementsMatch(t, []string{"Mars"}, grouped[4])

	groupedSequence := GroupBySequence(slices.Values(planets), func(planet string) int {
		return len(planet)
	})
	assert.Equal(t, grouped, groupedSequence)

	groupedRef := GroupByRef(planets, func(planet *string) int {
		return len(*planet)
	})
	assert.Equal(t, grouped, groupedRef)

	groupedRefSequence := GroupByRefSequence(slices.Values(planets), func(planet *string) int {
		return len(*planet)
	})
	assert.Equal(t, grouped, groupedRefSequence)
}

func TestFrequenciesBy(t *testing.T) {
	frequencies := FrequenciesBy([]string{"aa", "aA", "bb", "cc"}, func(element string) string {
		return strings.ToLower(element)
	})
	assert.Equal(t, map[string]int{"aa": 2, "bb": 1, "cc": 1}, frequencies)
	assert.Equal(t, frequencies, FrequenciesByRef([]string{"aa", "aA", "bb", "cc"}, func(element *string) string {
		return strings.ToLower(*element)
	}))
	assert.Equal(t, frequencies, FrequenciesBySequence(slices.Values([]string{"aa", "aA", "bb", "cc"}), func(element string) string {
		return strings.ToLower(element)
	}))
	assert.Equal(t, frequencies, FrequenciesByRefSequence(slices.Values([]string{"aa", "aA", "bb", "cc"}), func(element *string) string {
		return strings.ToLower(*element)
	}))
}

func TestFlatMap(t *testing.T) {
	numbers := []int{1, 2, 3}
	flattened := FlatMap(numbers, func(number int) []string {
		return []string{fmt.Sprintf("%d", number), fmt.Sprintf("%d", number)}
	})
	assert.Equal(t, []string{"1", "1", "2", "2", "3", "3"}, flattened)
	assert.Equal(t, flattened, FlatMapRef(numbers, func(number *int) *[]string {
		mapped := []string{fmt.Sprintf("%d", *number), fmt.Sprintf("%d", *number)}
		return &mapped
	}))
	assert.Equal(t, flattened, FlatMapSequence(slices.Values(numbers), func(number int) []string {
		return []string{fmt.Sprintf("%d", number), fmt.Sprintf("%d", number)}
	}))
	assert.Equal(t, flattened, FlatMapRefSequence(slices.Values(numbers), func(number *int) *[]string {
		mapped := []string{fmt.Sprintf("%d", *number), fmt.Sprintf("%d", *number)}
		return &mapped
	}))
}
