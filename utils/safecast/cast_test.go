package safecast

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/ptr"
)

type testCase[C1 IConvertible, C2 IConvertible] struct {
	name         string
	ctype        string
	value        C1
	expected     C2
	testCaseFunc func(t *testing.T, tCase *testCase[C1, C2])
}

func TestCastingToInt(t *testing.T) {
	tests := []testCase[int64, int64]{
		{
			name:     "zero",
			ctype:    "int8",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int8(tCase.expected), ToInt8(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int8(tCase.expected), ptr.From(ToInt8Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "uint8",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint8(tCase.expected), ToUint8(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint8(tCase.expected), ptr.From(ToUint8Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "int8",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int8(tCase.expected), ToInt8(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int8(tCase.expected), ptr.From(ToInt8Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "uint8",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint8(tCase.expected), ToUint8(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint8(tCase.expected), ptr.From(ToUint8Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "int8",
			value:    -1,
			expected: -1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int8(tCase.expected), ToInt8(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int8(tCase.expected), ptr.From(ToInt8Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "uint8",
			value:    -1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint8(tCase.expected), ToUint8(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint8(tCase.expected), ptr.From(ToUint8Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "int8",
			value:    math.MaxInt8 + 1,
			expected: math.MaxInt8,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int8(tCase.expected), ToInt8(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int8(tCase.expected), ptr.From(ToInt8Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "uint8",
			value:    math.MaxInt8 + 1,
			expected: math.MaxInt8 + 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint8(tCase.expected), ToUint8(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint8(tCase.expected), ptr.From(ToUint8Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "uint8",
			value:    math.MaxUint8 + 1,
			expected: math.MaxUint8,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint8(tCase.expected), ToUint8(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint8(tCase.expected), ptr.From(ToUint8Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Min",
			ctype:    "int8",
			value:    math.MinInt8 - 1,
			expected: math.MinInt8,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int8(tCase.expected), ToInt8(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int8(tCase.expected), ptr.From(ToInt8Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Min",
			ctype:    "uint8",
			value:    math.MinInt8 - 1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint8(tCase.expected), ToUint8(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint8(tCase.expected), ptr.From(ToUint8Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "int16",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int16(tCase.expected), ToInt16(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int16(tCase.expected), ptr.From(ToInt16Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "uint16",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint16(tCase.expected), ToUint16(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint16(tCase.expected), ptr.From(ToUint16Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "int16",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int16(tCase.expected), ToInt16(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int16(tCase.expected), ptr.From(ToInt16Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "uint16",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint16(tCase.expected), ToUint16(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint16(tCase.expected), ptr.From(ToUint16Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "int16",
			value:    -1,
			expected: -1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int16(tCase.expected), ToInt16(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int16(tCase.expected), ptr.From(ToInt16Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "uint16",
			value:    -1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint16(tCase.expected), ToUint16(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint16(tCase.expected), ptr.From(ToUint16Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "int16",
			value:    math.MaxInt16 + 1,
			expected: math.MaxInt16,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int16(tCase.expected), ToInt16(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int16(tCase.expected), ptr.From(ToInt16Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "uint16",
			value:    math.MaxInt16 + 1,
			expected: math.MaxInt16 + 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint16(tCase.expected), ToUint16(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint16(tCase.expected), ptr.From(ToUint16Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "uint16",
			value:    math.MaxUint16 + 1,
			expected: math.MaxUint16,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint16(tCase.expected), ToUint16(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint16(tCase.expected), ptr.From(ToUint16Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Min",
			ctype:    "int16",
			value:    math.MinInt16 - 1,
			expected: math.MinInt16,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int16(tCase.expected), ToInt16(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int16(tCase.expected), ptr.From(ToInt16Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Min",
			ctype:    "uint16",
			value:    math.MinInt16 - 1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint16(tCase.expected), ToUint16(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint16(tCase.expected), ptr.From(ToUint16Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "int32",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int32(tCase.expected), ToInt32(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int32(tCase.expected), ptr.From(ToInt32Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "uint32",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint32(tCase.expected), ToUint32(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint32(tCase.expected), ptr.From(ToUint32Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "uint32",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint32(tCase.expected), ToUint32(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint32(tCase.expected), ptr.From(ToUint32Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "int32",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int32(tCase.expected), ToInt32(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int32(tCase.expected), ptr.From(ToInt32Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "uint32",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint32(tCase.expected), ToUint32(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint32(tCase.expected), ptr.From(ToUint32Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "int32",
			value:    -1,
			expected: -1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int32(tCase.expected), ToInt32(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int32(tCase.expected), ptr.From(ToInt32Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "uint32",
			value:    -1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint32(tCase.expected), ToUint32(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint32(tCase.expected), ptr.From(ToUint32Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "int32",
			value:    math.MaxInt32 + 1,
			expected: math.MaxInt32,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int32(tCase.expected), ToInt32(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int32(tCase.expected), ptr.From(ToInt32Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "uint32",
			value:    math.MaxInt32 + 1,
			expected: math.MaxInt32 + 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint32(tCase.expected), ToUint32(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint32(tCase.expected), ptr.From(ToUint32Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "uint32",
			value:    math.MaxUint32 + 1,
			expected: math.MaxUint32,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint32(tCase.expected), ToUint32(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint32(tCase.expected), ptr.From(ToUint32Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Min",
			ctype:    "int32",
			value:    math.MinInt32 - 1,
			expected: math.MinInt32,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int32(tCase.expected), ToInt32(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int32(tCase.expected), ptr.From(ToInt32Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Min",
			ctype:    "uint32",
			value:    math.MinInt32 - 1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint32(tCase.expected), ToUint32(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint32(tCase.expected), ptr.From(ToUint32Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "int",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int(tCase.expected), ToInt(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int(tCase.expected), ptr.From(ToIntRef(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "uint",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint(tCase.expected), ToUint(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint(tCase.expected), ptr.From(ToUintRef(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "int",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int(tCase.expected), ToInt(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int(tCase.expected), ptr.From(ToIntRef(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "uint",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint(tCase.expected), ToUint(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint(tCase.expected), ptr.From(ToUintRef(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "int",
			value:    -1,
			expected: -1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int(tCase.expected), ToInt(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, int(tCase.expected), ptr.From(ToIntRef(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "uint",
			value:    -1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint(tCase.expected), ToUint(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint(tCase.expected), ptr.From(ToUintRef(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "int64",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, tCase.expected, ToInt64(tCase.value))
				assert.Equal(r, tCase.expected, ptr.From(ToInt64Ref(ptr.To(tCase.value))))
			},
		},
		{
			name:     "zero",
			ctype:    "uint64",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint64(tCase.expected), ToUint64(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint64(tCase.expected), ptr.From(ToUint64Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "int64",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, tCase.expected, ToInt64(tCase.value))
				assert.Equal(r, tCase.expected, ptr.From(ToInt64Ref(ptr.To(tCase.value))))
			},
		},
		{
			name:     "1",
			ctype:    "uint64",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint64(tCase.expected), ToUint64(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint64(tCase.expected), ptr.From(ToUint64Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "int64",
			value:    -1,
			expected: -1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, tCase.expected, ToInt64(tCase.value))
				assert.Equal(r, tCase.expected, ptr.From(ToInt64Ref(ptr.To(tCase.value))))
			},
		},
		{
			name:     "-1",
			ctype:    "uint64",
			value:    -1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint64(tCase.expected), ToUint64(tCase.value))                      //nolint: gosec //G115: testing
				assert.Equal(r, uint64(tCase.expected), ptr.From(ToUint64Ref(ptr.To(tCase.value)))) //nolint: gosec //G115: testing
			},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("%v/%v", test.ctype, test.name), func(r *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic: %v", r)
				}
			}()
			test.testCaseFunc(r, &test)
		})
	}

	t.Run("float", func(t *testing.T) {
		t.Run("int8", func(t *testing.T) {
			assert.Equal(t, int8(-4), ToInt8(-4.6))                                    //nolint: gosec //G115: testing
			assert.Equal(t, int8(-4), ptr.From(ToInt8Ref(ptr.To(-4.6))))               //nolint: gosec //G115: testing
			assert.Equal(t, int8(4), ToInt8(4.6))                                      //nolint: gosec //G115: testing
			assert.Equal(t, int8(4), ptr.From(ToInt8Ref(ptr.To(4.6))))                 //nolint: gosec //G115: testing
			assert.Equal(t, uint8(0), ToUint8(-4.6))                                   //nolint: gosec //G115: testing
			assert.Equal(t, uint8(0), ptr.From(ToUint8Ref(ptr.To(-4.6))))              //nolint: gosec //G115: testing
			assert.Equal(t, uint8(4), ToUint8(4.6))                                    //nolint: gosec //G115: testing
			assert.Equal(t, uint8(4), ptr.From(ToUint8Ref(ptr.To(4.6))))               //nolint: gosec //G115: testing
			assert.Equal(t, int8(math.MaxInt8), ToInt8(256.4))                         //nolint: gosec //G115: testing
			assert.Equal(t, int8(math.MaxInt8), ptr.From(ToInt8Ref(ptr.To(256.4))))    //nolint: gosec //G115: testing
			assert.Equal(t, uint8(math.MaxUint8), ToUint8(256.4))                      //nolint: gosec //G115: testing
			assert.Equal(t, uint8(math.MaxUint8), ptr.From(ToUint8Ref(ptr.To(256.4)))) //nolint: gosec //G115: testing
		})
		t.Run("int16", func(t *testing.T) {
			assert.Equal(t, int16(-4), ToInt16(-4.6))                                               //nolint: gosec //G115: testing
			assert.Equal(t, int16(-4), ptr.From(ToInt16Ref(ptr.To(-4.6))))                          //nolint: gosec //G115: testing
			assert.Equal(t, int16(4), ToInt16(4.6))                                                 //nolint: gosec //G115: testing
			assert.Equal(t, int16(4), ptr.From(ToInt16Ref(ptr.To(4.6))))                            //nolint: gosec //G115: testing
			assert.Equal(t, uint16(0), ToUint16(-4.6))                                              //nolint: gosec //G115: testing
			assert.Equal(t, uint16(0), ptr.From(ToUint16Ref(ptr.To(-4.6))))                         //nolint: gosec //G115: testing
			assert.Equal(t, uint16(4), ToUint16(4.6))                                               //nolint: gosec //G115: testing
			assert.Equal(t, uint16(4), ptr.From(ToUint16Ref(ptr.To(4.6))))                          //nolint: gosec //G115: testing
			assert.Equal(t, int16(math.MaxInt16), ToInt16(40000.4))                                 //nolint: gosec //G115: testing
			assert.Equal(t, int16(math.MaxInt16), ptr.From(ToInt16Ref(ptr.To(40000.4))))            //nolint: gosec //G115: testing
			assert.Equal(t, int16(math.MaxInt16), ToInt16(ToFloat32(40000.4)))                      //nolint: gosec //G115: testing
			assert.Equal(t, int16(math.MaxInt16), ptr.From(ToInt16Ref(ptr.To(ToFloat32(40000.4))))) //nolint: gosec //G115: testing
			assert.Equal(t, int16(math.MinInt16), ToInt16(-32768.4))                                //nolint: gosec //G115: testing
			assert.Equal(t, int16(math.MinInt16), ptr.From(ToInt16Ref(ptr.To(-32768.4))))           //nolint: gosec //G115: testing
			assert.Equal(t, uint16(math.MaxUint16), ToUint16(70000.4))                              //nolint: gosec //G115: testing
			assert.Equal(t, uint16(math.MaxUint16), ptr.From(ToUint16Ref(ptr.To(70000.4))))         //nolint: gosec //G115: testing
		})
		t.Run("int32", func(t *testing.T) {
			assert.Equal(t, int32(-4), ToInt32(-4.6))                                                     //nolint: gosec //G115: testing
			assert.Equal(t, int32(-4), ptr.From(ToInt32Ref(ptr.To(-4.6))))                                //nolint: gosec //G115: testing
			assert.Equal(t, int32(4), ToInt32(4.6))                                                       //nolint: gosec //G115: testing
			assert.Equal(t, int32(4), ptr.From(ToInt32Ref(ptr.To(4.6))))                                  //nolint: gosec //G115: testing
			assert.Equal(t, uint32(0), ToUint32(-4.6))                                                    //nolint: gosec //G115: testing
			assert.Equal(t, uint32(0), ptr.From(ToUint32Ref(ptr.To(-4.6))))                               //nolint: gosec //G115: testing
			assert.Equal(t, uint32(4), ToUint32(4.6))                                                     //nolint: gosec //G115: testing
			assert.Equal(t, uint32(4), ptr.From(ToUint32Ref(ptr.To(4.6))))                                //nolint: gosec //G115: testing
			assert.Equal(t, int32(math.MaxInt32), ToInt32(2147483647.4))                                  //nolint: gosec //G115: testing
			assert.Equal(t, int32(math.MaxInt32), ptr.From(ToInt32Ref(ptr.To(2147483647.4))))             //nolint: gosec //G115: testing
			assert.Equal(t, int32(math.MaxInt32), ToInt32(ToFloat32(2147483647.4)))                       //nolint: gosec //G115: testing
			assert.Equal(t, int32(math.MaxInt32), ptr.From(ToInt32Ref(ptr.To(ToFloat32(2147483647.4)))))  //nolint: gosec //G115: testing
			assert.Equal(t, int32(math.MinInt32), ToInt32(ToFloat32(-2147483648.4)))                      //nolint: gosec //G115: testing
			assert.Equal(t, int32(math.MinInt32), ptr.From(ToInt32Ref(ptr.To(ToFloat32(-2147483648.4))))) //nolint: gosec //G115: testing
			assert.Equal(t, uint32(math.MaxUint32), ToUint32(4294967295.4))                               //nolint: gosec //G115: testing
			assert.Equal(t, uint32(math.MaxUint32), ptr.From(ToUint32Ref(ptr.To(4294967295.4))))          //nolint: gosec //G115: testing
		})
		t.Run("int", func(t *testing.T) {
			assert.Equal(t, -4, ToInt(-4.6))                                                         //nolint: gosec //G115: testing
			assert.Equal(t, -4, ptr.From(ToIntRef(ptr.To(-4.6))))                                    //nolint: gosec //G115: testing
			assert.Equal(t, 4, ToInt(4.6))                                                           //nolint: gosec //G115: testing
			assert.Equal(t, 4, ptr.From(ToIntRef(ptr.To(4.6))))                                      //nolint: gosec //G115: testing
			assert.Equal(t, uint(0), ToUint(-4.6))                                                   //nolint: gosec //G115: testing
			assert.Equal(t, uint(0), ptr.From(ToUintRef(ptr.To(-4.6))))                              //nolint: gosec //G115: testing
			assert.Equal(t, uint(4), ToUint(4.6))                                                    //nolint: gosec //G115: testing
			assert.Equal(t, uint(4), ptr.From(ToUintRef(ptr.To(4.6))))                               //nolint: gosec //G115: testing
			assert.Equal(t, math.MaxInt, ToInt(9223372036854775807.4))                               //nolint: gosec //G115: testing
			assert.Equal(t, math.MaxInt, ptr.From(ToIntRef(ptr.To(9223372036854775807.4))))          //nolint: gosec //G115: testing
			assert.Equal(t, uint(math.MaxUint), ToUint(18446744073709551615.4))                      //nolint: gosec //G115: testing
			assert.Equal(t, uint(math.MaxUint), ptr.From(ToUintRef(ptr.To(18446744073709551615.4)))) //nolint: gosec //G115: testing
			assert.Equal(t, math.MinInt, ToInt(-9223372036854775808.4))                              //nolint: gosec //G115: testing
			assert.Equal(t, math.MinInt, ptr.From(ToIntRef(ptr.To(-9223372036854775808.4))))         //nolint: gosec //G115: testing
			assert.Equal(t, uint(0), ToUint(-18446744073709551615.4))                                //nolint: gosec //G115: testing
			assert.Equal(t, uint(0), ptr.From(ToUintRef(ptr.To(-18446744073709551615.4))))           //nolint: gosec //G115: testing
		})
		t.Run("int64", func(t *testing.T) {
			assert.Equal(t, int64(-4), ToInt64(-4.6))                                                      //nolint: gosec //G115: testing
			assert.Equal(t, int64(-4), ptr.From(ToInt64Ref(ptr.To(-4.6))))                                 //nolint: gosec //G115: testing
			assert.Equal(t, int64(4), ToInt64(4.6))                                                        //nolint: gosec //G115: testing
			assert.Equal(t, int64(4), ptr.From(ToInt64Ref(ptr.To(4.6))))                                   //nolint: gosec //G115: testing
			assert.Equal(t, uint64(0), ToUint64(-4.6))                                                     //nolint: gosec //G115: testing
			assert.Equal(t, uint64(0), ptr.From(ToUint64Ref(ptr.To(-4.6))))                                //nolint: gosec //G115: testing
			assert.Equal(t, uint64(4), ToUint64(4.6))                                                      //nolint: gosec //G115: testing
			assert.Equal(t, uint64(4), ptr.From(ToUint64Ref(ptr.To(4.6))))                                 //nolint: gosec //G115: testing
			assert.Equal(t, int64(math.MaxInt64), ToInt64(9223372036854775807.4))                          //nolint: gosec //G115: testing
			assert.Equal(t, int64(math.MaxInt64), ptr.From(ToInt64Ref(ptr.To(9223372036854775807.4))))     //nolint: gosec //G115: testing
			assert.Equal(t, uint64(math.MaxUint64), ToUint64(18446744073709551616.4))                      //nolint: gosec //G115: testing
			assert.Equal(t, uint64(math.MaxUint64), ptr.From(ToUint64Ref(ptr.To(18446744073709551616.4)))) //nolint: gosec //G115: testing
			assert.Equal(t, int64(math.MinInt64), ToInt64(-9223372036854775808.4))                         //nolint: gosec //G115: testing
			assert.Equal(t, int64(math.MinInt64), ptr.From(ToInt64Ref(ptr.To(-9223372036854775808.4))))    //nolint: gosec //G115: testing
			assert.Equal(t, uint64(0), ToUint64(-18446744073709551616.4))                                  //nolint: gosec //G115: testing
			assert.Equal(t, uint64(0), ptr.From(ToUint64Ref(ptr.To(-18446744073709551616.4))))             //nolint: gosec //G115: testing
		})
		t.Run("float64", func(t *testing.T) {
			assert.Equal(t, float64(-4.6), ToFloat64(-4.6))                                                  //nolint: gosec //G115: testing
			assert.Equal(t, float64(-4.6), ptr.From(ToFloat64Ref(ptr.To(-4.6))))                             //nolint: gosec //G115: testing
			assert.Equal(t, float64(4.6), ToFloat64(4.6))                                                    //nolint: gosec //G115: testing
			assert.Equal(t, float64(4.6), ptr.From(ToFloat64Ref(ptr.To(4.6))))                               //nolint: gosec //G115: testing
			assert.Equal(t, float64(math.MaxInt64), ToFloat64(math.MaxInt64))                                //nolint: gosec //G115: testing
			assert.Equal(t, float64(math.MaxInt64), ptr.From(ToFloat64Ref(ptr.To(math.MaxInt64))))           //nolint: gosec //G115: testing
			assert.Equal(t, float64(math.MaxUint64), ToFloat64[uint64](math.MaxUint64))                      //nolint: gosec //G115: testing
			assert.Equal(t, float64(math.MaxUint64), ptr.From(ToFloat64Ref(ptr.To(uint64(math.MaxUint64))))) //nolint: gosec //G115: testing
			assert.Equal(t, float64(math.MinInt64), ToFloat64(math.MinInt64))                                //nolint: gosec //G115: testing
			assert.Equal(t, float64(math.MinInt64), ptr.From(ToFloat64Ref(ptr.To(math.MinInt64))))           //nolint: gosec //G115: testing
		})
		t.Run("float32", func(t *testing.T) {
			assert.Equal(t, float32(-4.6), ToFloat32(-4.6))                                                   //nolint: gosec //G115: testing
			assert.Equal(t, float32(-4.6), ptr.From(ToFloat32Ref(ptr.To(-4.6))))                              //nolint: gosec //G115: testing
			assert.Equal(t, float32(4.6), ToFloat32(4.6))                                                     //nolint: gosec //G115: testing
			assert.Equal(t, float32(4.6), ptr.From(ToFloat32Ref(ptr.To(4.6))))                                //nolint: gosec //G115: testing
			assert.Equal(t, float32(math.MaxFloat32), ToFloat32(math.MaxFloat64))                             //nolint: gosec //G115: testing
			assert.Equal(t, float32(math.MaxFloat32), ptr.From(ToFloat32Ref(ptr.To(math.MaxFloat64))))        //nolint: gosec //G115: testing
			assert.Equal(t, float32(math.MaxFloat32), ToFloat32[uint64](math.MaxUint64))                      //nolint: gosec //G115: testing
			assert.Equal(t, float32(math.MaxFloat32), ptr.From(ToFloat32Ref(ptr.To(uint64(math.MaxUint64))))) //nolint: gosec //G115: testing
			assert.Equal(t, float32(math.MinInt64), ToFloat32(math.MinInt64))                                 //nolint: gosec //G115: testing
			assert.Equal(t, float32(math.MinInt64), ptr.From(ToFloat32Ref(ptr.To(math.MinInt64))))            //nolint: gosec //G115: testing
		})
	})
}

type robustCastTestCase[F IConvertible, T IConvertible] struct {
	name         string
	ctype        string
	value        F
	expected     T
	testCaseFunc func(t *testing.T, tCase *robustCastTestCase[F, T])
}

func TestRobustCast(t *testing.T) {
	tests := []robustCastTestCase[int64, int64]{
		{
			name:     "zero",
			ctype:    "int8",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int8](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int8(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int8](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int8(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "uint8",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint8](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint8(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint8](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint8(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "int8",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int8](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int8(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int8](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int8(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "uint8",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint8](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint8(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint8](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint8(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "int8",
			value:    -1,
			expected: -1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int8](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int8(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int8](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int8(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "uint8",
			value:    -1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint8](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint8(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint8](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint8(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "int8",
			value:    math.MaxInt8 + 1,
			expected: math.MaxInt8,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int8](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int8(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int8](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int8(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "uint8",
			value:    math.MaxInt8 + 1,
			expected: math.MaxInt8 + 1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint8](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint8(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint8](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint8(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "uint8",
			value:    math.MaxUint8 + 1,
			expected: math.MaxUint8,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint8](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint8(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint8](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint8(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Min",
			ctype:    "int8",
			value:    math.MinInt8 - 1,
			expected: math.MinInt8,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int8](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int8(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int8](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int8(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Min",
			ctype:    "uint8",
			value:    math.MinInt8 - 1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint8](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint8(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint8](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint8(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "int16",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int16](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int16(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int16](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int16(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "uint16",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint16](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint16(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint16](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint16(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "int16",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int16](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int16(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int16](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int16(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "uint16",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint16](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint16(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint16](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint16(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "int16",
			value:    -1,
			expected: -1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int16](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int16(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int16](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int16(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "uint16",
			value:    -1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint16](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint16(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint16](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint16(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "int16",
			value:    math.MaxInt16 + 1,
			expected: math.MaxInt16,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int16](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int16(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int16](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int16(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "uint16",
			value:    math.MaxInt16 + 1,
			expected: math.MaxInt16 + 1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint16](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint16(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint16](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint16(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "uint16",
			value:    math.MaxUint16 + 1,
			expected: math.MaxUint16,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint16](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint16(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint16](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint16(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Min",
			ctype:    "int16",
			value:    math.MinInt16 - 1,
			expected: math.MinInt16,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int16](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int16(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int16](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int16(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Min",
			ctype:    "uint16",
			value:    math.MinInt16 - 1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint16](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint16(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint16](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint16(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "int32",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int32](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int32(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int32](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int32(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "uint32",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint32](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint32(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint32](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint32(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "uint32",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint32](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint32(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint32](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint32(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "int32",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int32](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int32(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int32](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int32(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "uint32",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint32](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint32(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint32](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint32(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "int32",
			value:    -1,
			expected: -1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int32](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int32(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int32](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int32(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "uint32",
			value:    -1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint32](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint32(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint32](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint32(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "int32",
			value:    math.MaxInt32 + 1,
			expected: math.MaxInt32,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int32](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int32(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int32](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int32(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "uint32",
			value:    math.MaxInt32 + 1,
			expected: math.MaxInt32 + 1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint32](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint32(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint32](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint32(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Max",
			ctype:    "uint32",
			value:    math.MaxUint32 + 1,
			expected: math.MaxUint32,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint32](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint32(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint32](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint32(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Min",
			ctype:    "int32",
			value:    math.MinInt32 - 1,
			expected: math.MinInt32,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int32](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int32(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int32](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int32(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "Min",
			ctype:    "uint32",
			value:    math.MinInt32 - 1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint32](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint32(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint32](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint32(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "int",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "uint",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "int",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "uint",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "int",
			value:    -1,
			expected: -1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, int(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, int](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, int(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "uint",
			value:    -1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "zero",
			ctype:    "int64",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int64](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, tCase.expected, c)

				cRef, err := RobustCastRef[int64, int64](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, tCase.expected, ptr.From(cRef))
			},
		},
		{
			name:     "zero",
			ctype:    "uint64",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint64](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint64(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint64](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint64(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "1",
			ctype:    "int64",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int64](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, tCase.expected, c)

				cRef, err := RobustCastRef[int64, int64](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, tCase.expected, ptr.From(cRef))
			},
		},
		{
			name:     "1",
			ctype:    "uint64",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint64](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint64(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint64](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint64(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
		{
			name:     "-1",
			ctype:    "int64",
			value:    -1,
			expected: -1,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, int64](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, tCase.expected, c)

				cRef, err := RobustCastRef[int64, int64](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, tCase.expected, ptr.From(cRef))
			},
		},
		{
			name:     "-1",
			ctype:    "uint64",
			value:    -1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *robustCastTestCase[int64, int64]) {
				c, err := RobustCast[int64, uint64](tCase.value)
				require.NoError(r, err)
				assert.Equal(r, uint64(tCase.expected), c) //nolint: gosec //G115: testing

				cRef, err := RobustCastRef[int64, uint64](ptr.To(tCase.value))
				require.NoError(r, err)
				assert.Equal(r, uint64(tCase.expected), ptr.From(cRef)) //nolint: gosec //G115: testing
			},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("%v/%v", test.ctype, test.name), func(r *testing.T) {
			test.testCaseFunc(r, &test)
		})
	}

	t.Run("nil ref", func(t *testing.T) {
		c, err := RobustCastRef[int64, int](nil)
		require.NoError(t, err)
		assert.Nil(t, c)
	})

	t.Run("float", func(t *testing.T) {
		t.Run("int8", func(t *testing.T) {
			c1, err := RobustCast[float64, int8](-4.6)
			require.NoError(t, err)
			assert.Equal(t, int8(-4), c1) //nolint: gosec //G115: testing

			c1r, err := RobustCastRef[float64, int8](ptr.To(-4.6))
			require.NoError(t, err)
			assert.Equal(t, int8(-4), ptr.From(c1r)) //nolint: gosec //G115: testing

			c2, err := RobustCast[float64, int8](4.6)
			require.NoError(t, err)
			assert.Equal(t, int8(4), c2) //nolint: gosec //G115: testing

			c2r, err := RobustCastRef[float64, int8](ptr.To(4.6))
			require.NoError(t, err)
			assert.Equal(t, int8(4), ptr.From(c2r)) //nolint: gosec //G115: testing

			c3, err := RobustCast[float64, uint8](-4.6)
			require.NoError(t, err)
			assert.Equal(t, uint8(0), c3) //nolint: gosec //G115: testing

			c3r, err := RobustCastRef[float64, uint8](ptr.To(-4.6))
			require.NoError(t, err)
			assert.Equal(t, uint8(0), ptr.From(c3r)) //nolint: gosec //G115: testing

			c4, err := RobustCast[float64, uint8](4.6)
			require.NoError(t, err)
			assert.Equal(t, uint8(4), c4) //nolint: gosec //G115: testing

			c4r, err := RobustCastRef[float64, uint8](ptr.To(4.6))
			require.NoError(t, err)
			assert.Equal(t, uint8(4), ptr.From(c4r)) //nolint: gosec //G115: testing

			c5, err := RobustCast[float64, int8](256.4)
			require.NoError(t, err)
			assert.Equal(t, int8(math.MaxInt8), c5) //nolint: gosec //G115: testing

			c5r, err := RobustCastRef[float64, int8](ptr.To(256.4))
			require.NoError(t, err)
			assert.Equal(t, int8(math.MaxInt8), ptr.From(c5r)) //nolint: gosec //G115: testing

			c6, err := RobustCast[float64, uint8](256.4)
			require.NoError(t, err)
			assert.Equal(t, uint8(math.MaxUint8), c6) //nolint: gosec //G115: testing

			c6r, err := RobustCastRef[float64, uint8](ptr.To(256.4))
			require.NoError(t, err)
			assert.Equal(t, uint8(math.MaxUint8), ptr.From(c6r)) //nolint: gosec //G115: testing
		})

		t.Run("int16", func(t *testing.T) {
			c1, err := RobustCast[float64, int16](-4.6)
			require.NoError(t, err)
			assert.Equal(t, int16(-4), c1) //nolint: gosec //G115: testing

			c1r, err := RobustCastRef[float64, int16](ptr.To(-4.6))
			require.NoError(t, err)
			assert.Equal(t, int16(-4), ptr.From(c1r)) //nolint: gosec //G115: testing

			c2, err := RobustCast[float64, int16](4.6)
			require.NoError(t, err)
			assert.Equal(t, int16(4), c2) //nolint: gosec //G115: testing

			c2r, err := RobustCastRef[float64, int16](ptr.To(4.6))
			require.NoError(t, err)
			assert.Equal(t, int16(4), ptr.From(c2r)) //nolint: gosec //G115: testing

			c3, err := RobustCast[float64, uint16](-4.6)
			require.NoError(t, err)
			assert.Equal(t, uint16(0), c3) //nolint: gosec //G115: testing

			c3r, err := RobustCastRef[float64, uint16](ptr.To(-4.6))
			require.NoError(t, err)
			assert.Equal(t, uint16(0), ptr.From(c3r)) //nolint: gosec //G115: testing

			c4, err := RobustCast[float64, uint16](4.6)
			require.NoError(t, err)
			assert.Equal(t, uint16(4), c4) //nolint: gosec //G115: testing

			c4r, err := RobustCastRef[float64, uint16](ptr.To(4.6))
			require.NoError(t, err)
			assert.Equal(t, uint16(4), ptr.From(c4r)) //nolint: gosec //G115: testing

			c5, err := RobustCast[float64, int16](40000.4)
			require.NoError(t, err)
			assert.Equal(t, int16(math.MaxInt16), c5) //nolint: gosec //G115: testing

			c5r, err := RobustCastRef[float64, int16](ptr.To(40000.4))
			require.NoError(t, err)
			assert.Equal(t, int16(math.MaxInt16), ptr.From(c5r)) //nolint: gosec //G115: testing

			f32 := ToFloat32(40000.4)
			c6, err := RobustCast[float32, int16](f32)
			require.NoError(t, err)
			assert.Equal(t, int16(math.MaxInt16), c6) //nolint: gosec //G115: testing

			c6r, err := RobustCastRef[float32, int16](ptr.To(f32))
			require.NoError(t, err)
			assert.Equal(t, int16(math.MaxInt16), ptr.From(c6r)) //nolint: gosec //G115: testing

			c7, err := RobustCast[float64, int16](-32768.4)
			require.NoError(t, err)
			assert.Equal(t, int16(math.MinInt16), c7) //nolint: gosec //G115: testing

			c7r, err := RobustCastRef[float64, int16](ptr.To(-32768.4))
			require.NoError(t, err)
			assert.Equal(t, int16(math.MinInt16), ptr.From(c7r)) //nolint: gosec //G115: testing

			c8, err := RobustCast[float64, uint16](70000.4)
			require.NoError(t, err)
			assert.Equal(t, uint16(math.MaxUint16), c8) //nolint: gosec //G115: testing

			c8r, err := RobustCastRef[float64, uint16](ptr.To(70000.4))
			require.NoError(t, err)
			assert.Equal(t, uint16(math.MaxUint16), ptr.From(c8r)) //nolint: gosec //G115: testing
		})

		t.Run("int32", func(t *testing.T) {
			c1, err := RobustCast[float64, int32](-4.6)
			require.NoError(t, err)
			assert.Equal(t, int32(-4), c1) //nolint: gosec //G115: testing

			c1r, err := RobustCastRef[float64, int32](ptr.To(-4.6))
			require.NoError(t, err)
			assert.Equal(t, int32(-4), ptr.From(c1r)) //nolint: gosec //G115: testing

			c2, err := RobustCast[float64, int32](4.6)
			require.NoError(t, err)
			assert.Equal(t, int32(4), c2) //nolint: gosec //G115: testing

			c2r, err := RobustCastRef[float64, int32](ptr.To(4.6))
			require.NoError(t, err)
			assert.Equal(t, int32(4), ptr.From(c2r)) //nolint: gosec //G115: testing

			c3, err := RobustCast[float64, uint32](-4.6)
			require.NoError(t, err)
			assert.Equal(t, uint32(0), c3) //nolint: gosec //G115: testing

			c3r, err := RobustCastRef[float64, uint32](ptr.To(-4.6))
			require.NoError(t, err)
			assert.Equal(t, uint32(0), ptr.From(c3r)) //nolint: gosec //G115: testing

			c4, err := RobustCast[float64, uint32](4.6)
			require.NoError(t, err)
			assert.Equal(t, uint32(4), c4) //nolint: gosec //G115: testing

			c4r, err := RobustCastRef[float64, uint32](ptr.To(4.6))
			require.NoError(t, err)
			assert.Equal(t, uint32(4), ptr.From(c4r)) //nolint: gosec //G115: testing

			c5, err := RobustCast[float64, int32](2147483647.4)
			require.NoError(t, err)
			assert.Equal(t, int32(math.MaxInt32), c5) //nolint: gosec //G115: testing

			c5r, err := RobustCastRef[float64, int32](ptr.To(2147483647.4))
			require.NoError(t, err)
			assert.Equal(t, int32(math.MaxInt32), ptr.From(c5r)) //nolint: gosec //G115: testing

			fMax := ToFloat32(2147483647.4)
			c6, err := RobustCast[float32, int32](fMax)
			require.NoError(t, err)
			assert.Equal(t, int32(math.MaxInt32), c6) //nolint: gosec //G115: testing

			c6r, err := RobustCastRef[float32, int32](ptr.To(fMax))
			require.NoError(t, err)
			assert.Equal(t, int32(math.MaxInt32), ptr.From(c6r)) //nolint: gosec //G115: testing

			fMin := ToFloat32(-2147483648.4)
			c7, err := RobustCast[float32, int32](fMin)
			require.NoError(t, err)
			assert.Equal(t, int32(math.MinInt32), c7) //nolint: gosec //G115: testing

			c7r, err := RobustCastRef[float32, int32](ptr.To(fMin))
			require.NoError(t, err)
			assert.Equal(t, int32(math.MinInt32), ptr.From(c7r)) //nolint: gosec //G115: testing

			c8, err := RobustCast[float64, uint32](4294967295.4)
			require.NoError(t, err)
			assert.Equal(t, uint32(math.MaxUint32), c8) //nolint: gosec //G115: testing

			c8r, err := RobustCastRef[float64, uint32](ptr.To(4294967295.4))
			require.NoError(t, err)
			assert.Equal(t, uint32(math.MaxUint32), ptr.From(c8r)) //nolint: gosec //G115: testing
		})

		t.Run("int", func(t *testing.T) {
			c1, err := RobustCast[float64, int](-4.6)
			require.NoError(t, err)
			assert.Equal(t, -4, c1) //nolint: gosec //G115: testing

			c1r, err := RobustCastRef[float64, int](ptr.To(-4.6))
			require.NoError(t, err)
			assert.Equal(t, -4, ptr.From(c1r)) //nolint: gosec //G115: testing

			c2, err := RobustCast[float64, int](4.6)
			require.NoError(t, err)
			assert.Equal(t, 4, c2) //nolint: gosec //G115: testing

			c2r, err := RobustCastRef[float64, int](ptr.To(4.6))
			require.NoError(t, err)
			assert.Equal(t, 4, ptr.From(c2r)) //nolint: gosec //G115: testing

			c3, err := RobustCast[float64, uint](-4.6)
			require.NoError(t, err)
			assert.Equal(t, uint(0), c3) //nolint: gosec //G115: testing

			c3r, err := RobustCastRef[float64, uint](ptr.To(-4.6))
			require.NoError(t, err)
			assert.Equal(t, uint(0), ptr.From(c3r)) //nolint: gosec //G115: testing

			c4, err := RobustCast[float64, uint](4.6)
			require.NoError(t, err)
			assert.Equal(t, uint(4), c4) //nolint: gosec //G115: testing

			c4r, err := RobustCastRef[float64, uint](ptr.To(4.6))
			require.NoError(t, err)
			assert.Equal(t, uint(4), ptr.From(c4r)) //nolint: gosec //G115: testing

			c5, err := RobustCast[float64, int](9223372036854775807.4)
			require.NoError(t, err)
			assert.Equal(t, math.MaxInt, c5) //nolint: gosec //G115: testing

			c5r, err := RobustCastRef[float64, int](ptr.To(9223372036854775807.4))
			require.NoError(t, err)
			assert.Equal(t, math.MaxInt, ptr.From(c5r)) //nolint: gosec //G115: testing

			c6, err := RobustCast[float64, uint](18446744073709551615.4)
			require.NoError(t, err)
			assert.Equal(t, uint(math.MaxUint), c6) //nolint: gosec //G115: testing

			c6r, err := RobustCastRef[float64, uint](ptr.To(18446744073709551615.4))
			require.NoError(t, err)
			assert.Equal(t, uint(math.MaxUint), ptr.From(c6r)) //nolint: gosec //G115: testing

			c7, err := RobustCast[float64, int](-9223372036854775808.4)
			require.NoError(t, err)
			assert.Equal(t, math.MinInt, c7) //nolint: gosec //G115: testing

			c7r, err := RobustCastRef[float64, int](ptr.To(-9223372036854775808.4))
			require.NoError(t, err)
			assert.Equal(t, math.MinInt, ptr.From(c7r)) //nolint: gosec //G115: testing

			c8, err := RobustCast[float64, uint](-18446744073709551615.4)
			require.NoError(t, err)
			assert.Equal(t, uint(0), c8) //nolint: gosec //G115: testing

			c8r, err := RobustCastRef[float64, uint](ptr.To(-18446744073709551615.4))
			require.NoError(t, err)
			assert.Equal(t, uint(0), ptr.From(c8r)) //nolint: gosec //G115: testing
		})

		t.Run("int64", func(t *testing.T) {
			c1, err := RobustCast[float64, int64](-4.6)
			require.NoError(t, err)
			assert.Equal(t, int64(-4), c1) //nolint: gosec //G115: testing

			c1r, err := RobustCastRef[float64, int64](ptr.To(-4.6))
			require.NoError(t, err)
			assert.Equal(t, int64(-4), ptr.From(c1r)) //nolint: gosec //G115: testing

			c2, err := RobustCast[float64, int64](4.6)
			require.NoError(t, err)
			assert.Equal(t, int64(4), c2) //nolint: gosec //G115: testing

			c2r, err := RobustCastRef[float64, int64](ptr.To(4.6))
			require.NoError(t, err)
			assert.Equal(t, int64(4), ptr.From(c2r)) //nolint: gosec //G115: testing

			c3, err := RobustCast[float64, uint64](-4.6)
			require.NoError(t, err)
			assert.Equal(t, uint64(0), c3) //nolint: gosec //G115: testing

			c3r, err := RobustCastRef[float64, uint64](ptr.To(-4.6))
			require.NoError(t, err)
			assert.Equal(t, uint64(0), ptr.From(c3r)) //nolint: gosec //G115: testing

			c4, err := RobustCast[float64, uint64](4.6)
			require.NoError(t, err)
			assert.Equal(t, uint64(4), c4) //nolint: gosec //G115: testing

			c4r, err := RobustCastRef[float64, uint64](ptr.To(4.6))
			require.NoError(t, err)
			assert.Equal(t, uint64(4), ptr.From(c4r)) //nolint: gosec //G115: testing

			c5, err := RobustCast[float64, int64](9223372036854775807.4)
			require.NoError(t, err)
			assert.Equal(t, int64(math.MaxInt64), c5) //nolint: gosec //G115: testing

			c5r, err := RobustCastRef[float64, int64](ptr.To(9223372036854775807.4))
			require.NoError(t, err)
			assert.Equal(t, int64(math.MaxInt64), ptr.From(c5r)) //nolint: gosec //G115: testing

			c6, err := RobustCast[float64, uint64](18446744073709551616.4)
			require.NoError(t, err)
			assert.Equal(t, uint64(math.MaxUint64), c6) //nolint: gosec //G115: testing

			c6r, err := RobustCastRef[float64, uint64](ptr.To(18446744073709551616.4))
			require.NoError(t, err)
			assert.Equal(t, uint64(math.MaxUint64), ptr.From(c6r)) //nolint: gosec //G115: testing

			c7, err := RobustCast[float64, int64](-9223372036854775808.4)
			require.NoError(t, err)
			assert.Equal(t, int64(math.MinInt64), c7) //nolint: gosec //G115: testing

			c7r, err := RobustCastRef[float64, int64](ptr.To(-9223372036854775808.4))
			require.NoError(t, err)
			assert.Equal(t, int64(math.MinInt64), ptr.From(c7r)) //nolint: gosec //G115: testing

			c8, err := RobustCast[float64, uint64](-18446744073709551616.4)
			require.NoError(t, err)
			assert.Equal(t, uint64(0), c8) //nolint: gosec //G115: testing

			c8r, err := RobustCastRef[float64, uint64](ptr.To(-18446744073709551616.4))
			require.NoError(t, err)
			assert.Equal(t, uint64(0), ptr.From(c8r)) //nolint: gosec //G115: testing
		})

		t.Run("float64", func(t *testing.T) {
			c1, err := RobustCast[float64, float64](-4.6)
			require.NoError(t, err)
			assert.Equal(t, float64(-4.6), c1) //nolint: gosec //G115: testing

			c1r, err := RobustCastRef[float64, float64](ptr.To(-4.6))
			require.NoError(t, err)
			assert.Equal(t, float64(-4.6), ptr.From(c1r)) //nolint: gosec //G115: testing

			c2, err := RobustCast[float64, float64](4.6)
			require.NoError(t, err)
			assert.Equal(t, float64(4.6), c2) //nolint: gosec //G115: testing

			c2r, err := RobustCastRef[float64, float64](ptr.To(4.6))
			require.NoError(t, err)
			assert.Equal(t, float64(4.6), ptr.From(c2r)) //nolint: gosec //G115: testing

			c3, err := RobustCast[int64, float64](math.MaxInt64)
			require.NoError(t, err)
			assert.Equal(t, float64(math.MaxInt64), c3) //nolint: gosec //G115: testing

			c3r, err := RobustCastRef[int64, float64](ptr.To(int64(math.MaxInt64)))
			require.NoError(t, err)
			assert.Equal(t, float64(math.MaxInt64), ptr.From(c3r)) //nolint: gosec //G115: testing

			c4, err := RobustCast[uint64, float64](math.MaxUint64)
			require.NoError(t, err)
			assert.Equal(t, float64(math.MaxUint64), c4) //nolint: gosec //G115: testing

			c4r, err := RobustCastRef[uint64, float64](ptr.To(uint64(math.MaxUint64)))
			require.NoError(t, err)
			assert.Equal(t, float64(math.MaxUint64), ptr.From(c4r)) //nolint: gosec //G115: testing

			c5, err := RobustCast[int64, float64](math.MinInt64)
			require.NoError(t, err)
			assert.Equal(t, float64(math.MinInt64), c5) //nolint: gosec //G115: testing

			c5r, err := RobustCastRef[int64, float64](ptr.To(int64(math.MinInt64)))
			require.NoError(t, err)
			assert.Equal(t, float64(math.MinInt64), ptr.From(c5r)) //nolint: gosec //G115: testing
		})

		t.Run("float32", func(t *testing.T) {
			c1, err := RobustCast[float64, float32](-4.6)
			require.NoError(t, err)
			assert.Equal(t, float32(-4.6), c1) //nolint: gosec //G115: testing

			c1r, err := RobustCastRef[float64, float32](ptr.To(-4.6))
			require.NoError(t, err)
			assert.Equal(t, float32(-4.6), ptr.From(c1r)) //nolint: gosec //G115: testing

			c2, err := RobustCast[float64, float32](4.6)
			require.NoError(t, err)
			assert.Equal(t, float32(4.6), c2) //nolint: gosec //G115: testing

			c2r, err := RobustCastRef[float64, float32](ptr.To(4.6))
			require.NoError(t, err)
			assert.Equal(t, float32(4.6), ptr.From(c2r)) //nolint: gosec //G115: testing

			c3, err := RobustCast[float64, float32](math.MaxFloat64)
			require.NoError(t, err)
			assert.Equal(t, float32(math.MaxFloat32), c3) //nolint: gosec //G115: testing

			c3r, err := RobustCastRef[float64, float32](ptr.To(math.MaxFloat64))
			require.NoError(t, err)
			assert.Equal(t, float32(math.MaxFloat32), ptr.From(c3r)) //nolint: gosec //G115: testing

			c4, err := RobustCast[uint64, float32](math.MaxUint64)
			require.NoError(t, err)
			assert.Equal(t, float32(math.MaxFloat32), c4) //nolint: gosec //G115: testing

			c4r, err := RobustCastRef[uint64, float32](ptr.To(uint64(math.MaxUint64)))
			require.NoError(t, err)
			assert.Equal(t, float32(math.MaxFloat32), ptr.From(c4r)) //nolint: gosec //G115: testing

			c5, err := RobustCast[int64, float32](math.MinInt64)
			require.NoError(t, err)
			assert.Equal(t, float32(math.MinInt64), c5) //nolint: gosec //G115: testing

			c5r, err := RobustCastRef[int64, float32](ptr.To(int64(math.MinInt64)))
			require.NoError(t, err)
			assert.Equal(t, float32(math.MinInt64), ptr.From(c5r)) //nolint: gosec //G115: testing
		})
	})
}
