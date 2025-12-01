package collection

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

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
			assert.Equal(t, test.expected, Range(test.start, test.stop, test.step))
			if reflection.IsEmpty(test.expected) {
				assert.Empty(t, slices.Collect(RangeSequence(test.start, test.stop, test.step)))
			} else {
				assert.Equal(t, test.expected, slices.Collect(RangeSequence(test.start, test.stop, test.step)))
			}

		})
	}
}
