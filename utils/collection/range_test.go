package collection

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/safecast"
)

func testRange[T safecast.IConvertable](t *testing.T, start, stop T, step *T, expected []T) {
	t.Helper()
	assert.Equal(t, expected, Range[T](start, stop, step))
	if reflection.IsEmpty(expected) {
		assert.Empty(t, slices.Collect(RangeSequence(start, stop, step)))
	} else {
		assert.Equal(t, expected, slices.Collect(RangeSequence(start, stop, step)))
	}
}
func TestRange(t *testing.T) {
	tests := []struct {
		start    int
		stop     int
		step     *int
		expected []int
	}{

		{2, 5, nil, []int{2, 3, 4}},
		{5, 2, nil, []int{}}, // empty, since stop < start
		{2, 10, field.ToOptionalInt(2), []int{2, 4, 6, 8}},
		{0, 10, field.ToOptionalInt(3), []int{0, 3, 6, 9}},
		{1, 10, field.ToOptionalInt(3), []int{1, 4, 7}},
		{10, 2, field.ToOptionalInt(-2), []int{10, 8, 6, 4}},
		{5, -1, field.ToOptionalInt(-1), []int{5, 4, 3, 2, 1, 0}},
		{0, -5, field.ToOptionalInt(-2), []int{0, -2, -4}},
		{0, 5, nil, []int{0, 1, 2, 3, 4}},
		{0, 5, field.ToOptionalInt(0), []int{}},
		{2, 2, field.ToOptionalInt(1), []int{}},
		{2, 2, field.ToOptionalInt(-1), []int{}},
	}

	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("[%v,%v,%v]", test.start, test.stop, test.step), func(t *testing.T) {
			testRange(t, test.start, test.stop, test.step, test.expected)
		})
	}
	t.Run("different types", func(t *testing.T) {
		testRange[uint64](t, 1, 10, field.ToOptionalUint64(3), []uint64{1, 4, 7})
		testRange[float32](t, 1, 10, field.ToOptionalFloat32(3), []float32{1, 4, 7})
		testRange[int8](t, 1, 10, field.ToOptional[int8](3), []int8{1, 4, 7})
	})
}

func Test_determineRangeLength(t *testing.T) {
	assert.Equal(t, 3, determineRangeLength[int](2, 5, 1))
	assert.Equal(t, 4, determineRangeLength[int](2, 10, 2))
	assert.Equal(t, 4, determineRangeLength[int](10, 2, -2))
	assert.Equal(t, 3, determineRangeLength[int](0, -5, -2))

	assert.Equal(t, 3, determineRangeLength[int64](2, 5, 1))
	assert.Equal(t, 4, determineRangeLength[int64](2, 10, 2))
	assert.Equal(t, 4, determineRangeLength[int64](10, 2, -2))
	assert.Equal(t, 3, determineRangeLength[int64](0, -5, -2))

	assert.Equal(t, 3, determineRangeLength[uint64](2, 5, 1))
	assert.Equal(t, 4, determineRangeLength[uint64](2, 10, 2))
	assert.Equal(t, 3, determineRangeLength[uint32](2, 5, 1))
	assert.Equal(t, 4, determineRangeLength[uint32](2, 10, 2))

	assert.Equal(t, 3, determineRangeLength[float64](2, 5, 1))
	assert.Equal(t, 4, determineRangeLength[float64](2, 10, 2))
	assert.Equal(t, 4, determineRangeLength[float64](10, 2, -2))
	assert.Equal(t, 3, determineRangeLength[float64](0, -5, -2))
}
