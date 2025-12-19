package maps

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMerge_EmptyMapsReturnsEmptyMap(t *testing.T) {
	tests := []struct {
		merge map[int]int
	}{
		{Merge[int, int]()},
		{Merge[int, int](nil)},
		{Merge[int, int](nil, nil)},
		{Merge[int, int](nil, nil, nil, nil)},
	}
	for i := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := tests[i].merge
			require.NotNil(t, got)
			assert.Empty(t, got)
		})
	}
}

func TestMerge_NilMapsAreIgnored(t *testing.T) {
	got := Merge[string, int](nil, map[string]int{"a": 1}, nil)

	expected := map[string]int{"a": 1}
	assert.Equal(t, expected, got)
}

func TestMerge_MergesDistinctKeys(t *testing.T) {
	m1 := map[string]int{"a": 1}
	m2 := map[string]int{"b": 2}
	m3 := map[string]int{"c": 3}

	got := Merge[string, int](m1, m2, m3)
	expected := map[string]int{"a": 1, "b": 2, "c": 3}

	assert.Equal(t, expected, got)
}

func TestMerge_LaterMapsOverrideEarlierOnConflicts(t *testing.T) {
	m1 := map[string]int{"a": 1, "b": 1}
	m2 := map[string]int{"b": 2}
	m3 := map[string]int{"a": 3}

	got := Merge[string, int](m1, m2, m3)
	expected := map[string]int{"a": 3, "b": 2}

	assert.Equal(t, expected, got)
}

func TestMerge_ResultIsIndependentFromInputs(t *testing.T) {
	m1 := map[string]int{"a": 1}
	m2 := map[string]int{"b": 2}

	got := Merge[string, int](m1, m2)

	// Mutate inputs after merge.
	m1["a"] = 100
	m2["b"] = 200
	m1["c"] = 300

	// Ensure result didn't change.
	expected := map[string]int{"a": 1, "b": 2}
	assert.Equal(t, expected, got)

	// Mutate result and ensure it doesn't affect inputs (sanity).
	got["a"] = 999
	assert.NotEqualf(t, 999, m1["a"], "expected input map not to be affected by result mutation")
}

func TestMerge_WorksWithStructValues(t *testing.T) {
	type V struct {
		N int
		S string
	}

	m1 := map[string]V{"a": {N: 1, S: "one"}}
	m2 := map[string]V{"b": {N: 2, S: "two"}, "a": {N: 10, S: "ten"}}

	got := Merge[string, V](m1, m2)
	expected := map[string]V{
		"a": {N: 10, S: "ten"}, // overridden
		"b": {N: 2, S: "two"},
	}

	assert.Equal(t, expected, got)
}

func TestMerge_DoesNotPanicOnAllNil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			require.Failf(t, "unexpected panic", "%v", r)
		}
	}()

	got := Merge[string, int](nil, nil)
	require.NotNil(t, got)
	assert.Empty(t, got)
}
