package maps

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/field"
)

var randomNumber = faker.RandomUnixTime()

func TestFlatten(t *testing.T) {
	cases := []struct {
		Input  map[string]any
		Output map[string]string
	}{
		{
			Input: map[string]any{
				"foo": "bar",
				"bar": "baz",
			},
			Output: map[string]string{
				"foo": "bar",
				"bar": "baz",
			},
		},
		{
			Input: map[string]any{
				"foo": "bar",
				"bar": field.ToOptionalString("baz"),
			},
			Output: map[string]string{
				"foo": "bar",
				"bar": "baz",
			},
		},

		{
			Input: map[string]any{
				"foo": []string{
					"one",
					"two",
				},
			},
			Output: map[string]string{
				"foo.0": "one",
				"foo.1": "two",
			},
		},

		{
			Input: map[string]any{
				"foo": []map[any]any{
					map[any]any{
						"name":    "bar",
						"port":    3000,
						"enabled": true,
					},
				},
			},
			Output: map[string]string{
				"foo.0.name":    "bar",
				"foo.0.port":    "3000",
				"foo.0.enabled": "true",
			},
		},

		{
			Input: map[string]any{
				"foo": []map[any]any{
					map[any]any{
						"name": "bar",
						"ports": []string{
							"1",
							"2",
						},
					},
				},
			},
			Output: map[string]string{
				"foo.0.name":    "bar",
				"foo.0.ports.0": "1",
				"foo.0.ports.1": "2",
			},
		},
	}

	for i := range cases {
		test := cases[i]
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			result, err := Flatten(test.Input)
			require.NoError(t, err)
			assert.Equal(t, test.Output, result.AsMap())
		})
	}
}

func TestFlatten2(t *testing.T) {
	now := time.Now().UTC()
	cases := []struct {
		Input  map[string]any
		Output map[string]string
	}{
		{
			Input: map[string]any{
				"foo": "bar",
				"bar": "baz",
			},
			Output: map[string]string{
				"foo": "bar",
				"bar": "baz",
			},
		},
		{
			Input: map[string]any{
				"foo": []string{
					"one",
					"two",
				},
			},
			Output: map[string]string{
				"foo.0": "one",
				"foo.1": "two",
			},
		},
		{
			Input: map[string]any{
				"foo": []map[any]any{
					{
						"name":    "bar",
						"port":    3000,
						"enabled": true,
					},
				},
			},
			Output: map[string]string{
				"foo.0.name":    "bar",
				"foo.0.port":    "3000",
				"foo.0.enabled": "true",
			},
		},
		{
			Input: map[string]any{
				"foo": []map[any]any{
					{
						"name": "bar",
						"ports": []string{
							"1",
							"2",
						},
					},
				},
			},
			Output: map[string]string{
				"foo.0.name":    "bar",
				"foo.0.ports.0": "1",
				"foo.0.ports.1": "2",
			},
		},
		{
			Input: map[string]any{
				"foo": struct {
					Name string
					Age  int
				}{
					"astaxie",
					30,
				},
			},
			Output: map[string]string{
				"foo.Name": "astaxie",
				"foo.Age":  "30",
			},
		},
		{
			Input: map[string]any{
				"foo": struct {
					Name string
					Age  int
					Test int64
				}{
					"astaxie",
					30,
					randomNumber,
				},
			},
			Output: map[string]string{
				"foo.Name": "astaxie",
				"foo.Age":  "30",
				"foo.Test": fmt.Sprintf("%d", randomNumber),
			},
		},
		{
			Input: map[string]any{
				"foo": struct {
					SomeTime time.Time
				}{
					now,
				},
			},
			Output: map[string]string{
				"foo.SomeTime": now.UTC().Format(time.RFC3339Nano),
			},
		},
		{
			Input: map[string]any{
				"foo": struct {
					SomeDuration time.Duration
				}{
					56 * time.Minute,
				},
			},
			Output: map[string]string{
				"foo.SomeDuration": (56 * time.Minute).String(),
			},
		},
	}

	for i := range cases {
		test := cases[i]
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			result, err := Flatten(test.Input)
			require.NoError(t, err)
			assert.Equal(t, test.Output, result.AsMap())
		})
	}
}
