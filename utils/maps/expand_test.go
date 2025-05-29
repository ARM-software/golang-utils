package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpand(t *testing.T) {
	cases := []struct {
		Map         map[string]string
		Key         string
		Output      any
		NoKeyOutput any
	}{
		{
			Map: map[string]string{
				"foo": "bar",
				"bar": "baz",
			},
			Key:    "foo",
			Output: "bar",
			NoKeyOutput: map[string]any{
				"foo": "bar",
				"bar": "baz",
			},
		},

		{
			Map: map[string]string{
				"foo.0": "one",
				"foo.1": "two",
			},
			Key: "foo",
			Output: []any{
				"one",
				"two",
			},
		},

		{
			Map: map[string]string{
				"foo.0": "one",
				"foo.1": "two",
				"foo.2": "three",
			},
			Key: "foo",
			Output: []any{
				"one",
				"two",
				"three",
			},
		},

		{
			Map: map[string]string{
				"foo.0": "one",
				"foo.1": "two",
				"foo.2": "three",
			},
			Key: "foo",
			Output: []any{
				"one",
				"two",
				"three",
			},
		},

		{
			Map: map[string]string{
				"foo.0.name":    "bar",
				"foo.0.port":    "3000",
				"foo.0.enabled": "true",
			},
			Key: "foo",
			Output: []any{
				map[string]any{
					"name":    "bar",
					"port":    "3000",
					"enabled": true,
				},
			},
		},

		{
			Map: map[string]string{
				"foo.0.name":    "bar",
				"foo.0.ports.0": "1",
				"foo.0.ports.1": "2",
			},
			Key: "foo",
			Output: []any{
				map[string]any{
					"name": "bar",
					"ports": []any{
						"1",
						"2",
					},
				},
			},
		},
		{
			Map: map[string]string{
				"list_of_map.0.a": "1",
				"list_of_map.1.b": "2",
				"list_of_map.1.c": "3",
			},
			Key: "list_of_map",
			Output: []any{
				map[string]any{
					"a": "1",
				},
				map[string]any{
					"b": "2",
					"c": "3",
				},
			},
			NoKeyOutput: map[string]any{
				"list_of_map": []any{
					map[string]any{"a": "1"},
					map[string]any{
						"b": "2",
						"c": "3",
					},
				},
			},
		},

		{
			Map: map[string]string{
				"map_of_list.list2.0": "c",
				"map_of_list.list1.0": "a",
				"map_of_list.list1.1": "b",
			},
			Key: "map_of_list",
			Output: map[string]any{
				"list1": []any{"a", "b"},
				"list2": []any{"c"},
			},
		},
		{
			Map: map[string]string{
				"struct.0.name": "hello",
			},
			Key: "struct",
			Output: []any{
				map[string]any{
					"name": "hello",
				},
			},
		},
		{
			Map: map[string]string{
				"struct.0.name":      "hello",
				"struct.0.set.0.key": "value",
			},
			Key: "struct",
			Output: []any{
				map[string]any{
					"name": "hello",
					"set": []any{
						map[string]any{
							"key": "value",
						},
					},
				},
			},
		},
		{
			Map: map[string]string{
				"struct.0b.name": "hello",
				"struct.0b.key":  "value",
			},
			Key: "struct",
			Output: map[string]any{
				"0b": map[string]any{
					"name": "hello",
					"key":  "value",
				},
			},
		},
	}

	t.Run("ExpandPrefixed", func(t *testing.T) {
		for i := range cases {
			tc := cases[i]
			t.Run(tc.Key, func(t *testing.T) {
				actual, err := ExpandPrefixed(tc.Map, tc.Key)
				require.NoError(t, err)
				assert.Equal(t, tc.Output, actual)
			})
		}
	})
	t.Run("Expand", func(t *testing.T) {
		for i := range cases {
			tc := cases[i]
			t.Run(tc.Key, func(t *testing.T) {
				actual, err := Expand(tc.Map)
				require.NoError(t, err)
				if tc.NoKeyOutput == nil {
					assert.Equal(t, map[string]any{tc.Key: tc.Output}, actual)
				} else {
					assert.Equal(t, tc.NoKeyOutput, actual)
				}

			})
		}
	})
}
