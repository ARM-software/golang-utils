package maps

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapContains(t *testing.T) {
	cases := []struct {
		Input  map[string]string
		Key    string
		Result bool
	}{
		{
			Input: map[string]string{
				"foo": "bar",
				"bar": "nope",
			},
			Key:    "foo",
			Result: true,
		},

		{
			Input: map[string]string{
				"foo": "bar",
				"bar": "nope",
			},
			Key:    "baz",
			Result: false,
		},
	}

	for _, tc := range cases {
		actual := Map(tc.Input).Contains(tc.Key)
		assert.Equal(t, tc.Result, actual)
	}
}

func TestMapDelete(t *testing.T) {
	m, err := Flatten(map[string]any{
		"foo": "bar",
		"routes": []map[string]string{
			{
				"foo": "bar",
			},
		},
	})
	require.NoError(t, err)

	m.Delete("routes")

	expected := Map(map[string]string{"foo": "bar"})
	assert.Equal(t, expected, m)
}

func TestMapKeys(t *testing.T) {
	cases := []struct {
		Input  map[string]string
		Output []string
	}{
		{
			Input: map[string]string{
				"foo":       "bar",
				"bar.#":     "bar",
				"bar.0.foo": "bar",
				"bar.0.baz": "bar",
			},
			Output: []string{
				"bar",
				"foo",
			},
		},
	}

	for _, tc := range cases {
		actual := Map(tc.Input).Keys()

		// Sort so we have a consistent view of the output
		sort.Strings(actual)
		assert.Equal(t, tc.Output, actual)
	}
}

func TestMapMerge(t *testing.T) {
	cases := []struct {
		One    map[string]string
		Two    map[string]string
		Result map[string]string
	}{
		{
			One: map[string]string{
				"foo": "bar",
				"bar": "nope",
			},
			Two: map[string]string{
				"bar": "baz",
				"baz": "buz",
			},
			Result: map[string]string{
				"foo": "bar",
				"bar": "baz",
				"baz": "buz",
			},
		},
	}

	for _, tc := range cases {
		Map(tc.One).Merge(tc.Two)
		assert.Equal(t, tc.One, tc.Result)
	}
}
