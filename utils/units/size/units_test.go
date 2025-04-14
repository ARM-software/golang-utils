package size

import (
	"fmt"
	"sort"
	"strconv"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
)

func TestSizes(t *testing.T) {
	sizes := maps.Values(DecimalSIUnits)
	sort.Float64s(sizes)
	for i := range sizes {
		if i > 0 {
			assert.Equal(t, sizes[i], 1000*sizes[i-1])
		}
	}
	sizes = maps.Values(BinarySIUnits)
	sort.Float64s(sizes)
	for i := range sizes {
		if i > 0 {
			assert.Equal(t, sizes[i], 1024*sizes[i-1])
		}
	}
	assert.Equal(t, KiB, float64(1024))
	assert.Equal(t, KiB, float64(1<<10))
	assert.Equal(t, GiB, float64(1<<30))
	assert.Equal(t, 10*GiB, float64(10<<30))
	assert.Equal(t, MiB, float64(1<<20))
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		name       string
		units      map[string]float64
		formatFunc func(float64, int) (string, error)
	}{
		{
			name:       "decimal SI",
			units:      DecimalSIUnits,
			formatFunc: FormatSizeAsDecimalSI,
		},
		{
			name:       "binary SI",
			units:      BinarySIUnits,
			formatFunc: FormatSizeAsBinarySI,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			for k, v := range test.units {
				t.Run(k, func(t *testing.T) {
					random, err := faker.RandomInt(0, 999, 2)
					require.NoError(t, err)
					valueRandom, err := strconv.ParseFloat(fmt.Sprintf("%v.%v", random[0], random[1]), 64)
					require.NoError(t, err)
					expectedValue := valueRandom * v
					valueStr1, err := test.formatFunc(expectedValue, -1)
					require.NoError(t, err)
					valueStr2 := fmt.Sprintf("%v %v", valueRandom, k)
					parsedValue1, err := ParseSize(valueStr1)
					require.NoError(t, err)
					parsedValue2, err := ParseSize(valueStr2)
					require.NoError(t, err)
					assert.Equal(t, expectedValue, parsedValue1)
					assert.Equal(t, expectedValue, parsedValue2)
				})
			}
		})
	}

}

func TestFormatSizeAsDecimalSI(t *testing.T) {
	valueStr, err := FormatSizeAsDecimalSI(-1605.0, -1)
	require.NoError(t, err)
	assert.Equal(t, "-1.605KB", valueStr)
	valueStr, err = FormatSizeAsDecimalSI(-1605.0, 0)
	require.NoError(t, err)
	assert.Equal(t, "-2KB", valueStr)
	valueStr, err = FormatSizeAsDecimalSI(-1605.0, 1)
	require.NoError(t, err)
	assert.Equal(t, "-1.6KB", valueStr)
	valueStr, err = FormatSizeAsDecimalSI(-1605.0, 2)
	require.NoError(t, err)
	assert.Equal(t, "-1.60KB", valueStr)
	valueStr, err = FormatSizeAsDecimalSI(1605.0, -1)
	require.NoError(t, err)
	assert.Equal(t, "1.605KB", valueStr)
	valueStr, err = FormatSizeAsDecimalSI(1605.0, 0)
	require.NoError(t, err)
	assert.Equal(t, "2KB", valueStr)
	valueStr, err = FormatSizeAsDecimalSI(1605.0, 1)
	require.NoError(t, err)
	assert.Equal(t, "1.6KB", valueStr)
	valueStr, err = FormatSizeAsDecimalSI(1605.0, 2)
	require.NoError(t, err)
	assert.Equal(t, "1.60KB", valueStr)
}

func TestFormatSizeAsBinarySI(t *testing.T) {
	valueStr, err := FormatSizeAsBinarySI(-1605.0, -1)
	require.NoError(t, err)
	assert.Equal(t, "-1.5673828125KiB", valueStr)
	valueStr, err = FormatSizeAsBinarySI(-1605.0, 0)
	require.NoError(t, err)
	assert.Equal(t, "-2KiB", valueStr)
	valueStr, err = FormatSizeAsBinarySI(-1605.0, 1)
	require.NoError(t, err)
	assert.Equal(t, "-1.6KiB", valueStr)
	valueStr, err = FormatSizeAsBinarySI(-1605.0, 2)
	require.NoError(t, err)
	assert.Equal(t, "-1.57KiB", valueStr)
	valueStr, err = FormatSizeAsBinarySI(1605.0, -1)
	require.NoError(t, err)
	assert.Equal(t, "1.5673828125KiB", valueStr)
	valueStr, err = FormatSizeAsBinarySI(1605.0, 0)
	require.NoError(t, err)
	assert.Equal(t, "2KiB", valueStr)
	valueStr, err = FormatSizeAsBinarySI(1605.0, 1)
	require.NoError(t, err)
	assert.Equal(t, "1.6KiB", valueStr)
	valueStr, err = FormatSizeAsBinarySI(1605.0, 2)
	require.NoError(t, err)
	assert.Equal(t, "1.57KiB", valueStr)
}
