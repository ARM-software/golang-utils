package maps

import (
	"strconv"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapFlattenExpand(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			testCase := map[string]any{
				"test1": map[string]any{
					"test1.1": faker.DomainName(),
					"test2": map[string]any{
						faker.Name(): faker.Paragraph(),
						"test3": map[string]any{
							faker.Word():      faker.Phonenumber(),
							"test3.1":         5.54,
							"some time":       time.Now().UTC(),
							"some float":      45454.454545812,
							faker.UUIDDigit(): faker.RandomUnixTime(),
							faker.Password():  time.Now().UTC(),
							"test4": map[string]any{
								"test5": map[string]time.Duration{
									faker.DomainName(): 5 * time.Hour,
								},
							},
						},
					},
				},
			}

			flattened, err := Flatten(testCase)
			require.NoError(t, err)
			expanded, err := Expand(flattened)
			require.NoError(t, err)
			flattened, err = Flatten(testCase)
			require.NoError(t, err)
			expanded2, err := Expand(flattened)
			require.NoError(t, err)
			assert.NotEmpty(t, expanded)
			assert.Equal(t, expanded, expanded2)
		})
	}
}
