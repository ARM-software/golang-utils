package safecast

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cast[C1 IConvertable, C2 IConvertable](t *testing.T, castFunc func(i C1) C2, value C1, expected C2) {

	require.NotNil(t, castFunc)
	assert.Equal(t, expected, castFunc(value))
}

type testCase[C1 IConvertable, C2 IConvertable] struct {
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
				assert.Equal(r, int8(tCase.expected), ToInt8(tCase.value))
			},
		},
		{
			name:     "zero",
			ctype:    "uint8",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint8(tCase.expected), ToUint8(tCase.value))
			},
		},
		{
			name:     "1",
			ctype:    "int8",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int8(tCase.expected), ToInt8(tCase.value))
			},
		},
		{
			name:     "1",
			ctype:    "uint8",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint8(tCase.expected), ToUint8(tCase.value))
			},
		},
		{
			name:     "-1",
			ctype:    "int8",
			value:    -1,
			expected: -1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int8(tCase.expected), ToInt8(tCase.value))
			},
		},
		{
			name:     "-1",
			ctype:    "uint8",
			value:    -1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint8(tCase.expected), ToUint8(tCase.value))
			},
		},
		{
			name:     "Max",
			ctype:    "int8",
			value:    math.MaxInt8 + 1,
			expected: math.MaxInt8,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int8(tCase.expected), ToInt8(tCase.value))
			},
		},
		{
			name:     "Max",
			ctype:    "uint8",
			value:    math.MaxInt8 + 1,
			expected: math.MaxInt8 + 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint8(tCase.expected), ToUint8(tCase.value))
			},
		},
		{
			name:     "Max",
			ctype:    "uint8",
			value:    math.MaxUint8 + 1,
			expected: math.MaxUint8,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint8(tCase.expected), ToUint8(tCase.value))
			},
		},
		{
			name:     "Min",
			ctype:    "int8",
			value:    math.MinInt8 - 1,
			expected: math.MinInt8,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int8(tCase.expected), ToInt8(tCase.value))
			},
		},
		{
			name:     "Min",
			ctype:    "uint8",
			value:    math.MinInt8 - 1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint8(tCase.expected), ToUint8(tCase.value))
			},
		},
		{
			name:     "zero",
			ctype:    "int16",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int16(tCase.expected), ToInt16(tCase.value))
			},
		},
		{
			name:     "zero",
			ctype:    "uint16",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint16(tCase.expected), ToUint16(tCase.value))
			},
		},
		{
			name:     "1",
			ctype:    "int16",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int16(tCase.expected), ToInt16(tCase.value))
			},
		},
		{
			name:     "1",
			ctype:    "uint16",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint16(tCase.expected), ToUint16(tCase.value))
			},
		},
		{
			name:     "-1",
			ctype:    "int16",
			value:    -1,
			expected: -1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int16(tCase.expected), ToInt16(tCase.value))
			},
		},
		{
			name:     "-1",
			ctype:    "uint16",
			value:    -1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint16(tCase.expected), ToUint16(tCase.value))
			},
		},
		{
			name:     "Max",
			ctype:    "int16",
			value:    math.MaxInt16 + 1,
			expected: math.MaxInt16,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int16(tCase.expected), ToInt16(tCase.value))
			},
		},
		{
			name:     "Max",
			ctype:    "uint16",
			value:    math.MaxInt16 + 1,
			expected: math.MaxInt16 + 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint16(tCase.expected), ToUint16(tCase.value))
			},
		},
		{
			name:     "Max",
			ctype:    "uint16",
			value:    math.MaxUint16 + 1,
			expected: math.MaxUint16,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint16(tCase.expected), ToUint16(tCase.value))
			},
		},
		{
			name:     "Min",
			ctype:    "int16",
			value:    math.MinInt16 - 1,
			expected: math.MinInt16,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int16(tCase.expected), ToInt16(tCase.value))
			},
		},
		{
			name:     "Min",
			ctype:    "uint16",
			value:    math.MinInt16 - 1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint16(tCase.expected), ToUint16(tCase.value))
			},
		},
		{
			name:     "zero",
			ctype:    "int32",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int32(tCase.expected), ToInt32(tCase.value))
			},
		},
		{
			name:     "zero",
			ctype:    "uint32",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint32(tCase.expected), ToUint32(tCase.value))
			},
		},
		{
			name:     "zero",
			ctype:    "uint32",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint32(tCase.expected), ToUint32(tCase.value))
			},
		},
		{
			name:     "1",
			ctype:    "int32",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int32(tCase.expected), ToInt32(tCase.value))
			},
		},
		{
			name:     "1",
			ctype:    "uint32",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint32(tCase.expected), ToUint32(tCase.value))
			},
		},
		{
			name:     "-1",
			ctype:    "int32",
			value:    -1,
			expected: -1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int32(tCase.expected), ToInt32(tCase.value))
			},
		},
		{
			name:     "-1",
			ctype:    "uint32",
			value:    -1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint32(tCase.expected), ToUint32(tCase.value))
			},
		},
		{
			name:     "Max",
			ctype:    "int32",
			value:    math.MaxInt32 + 1,
			expected: math.MaxInt32,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int32(tCase.expected), ToInt32(tCase.value))
			},
		},
		{
			name:     "Max",
			ctype:    "uint32",
			value:    math.MaxInt32 + 1,
			expected: math.MaxInt32 + 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint32(tCase.expected), ToUint32(tCase.value))
			},
		},
		{
			name:     "Max",
			ctype:    "uint32",
			value:    math.MaxUint32 + 1,
			expected: math.MaxUint32,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint32(tCase.expected), ToUint32(tCase.value))
			},
		},
		{
			name:     "Min",
			ctype:    "int32",
			value:    math.MinInt32 - 1,
			expected: math.MinInt32,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int32(tCase.expected), ToInt32(tCase.value))
			},
		},
		{
			name:     "Min",
			ctype:    "uint32",
			value:    math.MinInt32 - 1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint32(tCase.expected), ToUint32(tCase.value))
			},
		},
		{
			name:     "zero",
			ctype:    "int",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int(tCase.expected), ToInt(tCase.value))
			},
		},
		{
			name:     "zero",
			ctype:    "uint",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint(tCase.expected), ToUint(tCase.value))
			},
		},
		{
			name:     "1",
			ctype:    "int",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int(tCase.expected), ToInt(tCase.value))
			},
		},
		{
			name:     "1",
			ctype:    "uint",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint(tCase.expected), ToUint(tCase.value))
			},
		},
		{
			name:     "-1",
			ctype:    "int",
			value:    -1,
			expected: -1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int(tCase.expected), ToInt(tCase.value))
			},
		},
		{
			name:     "-1",
			ctype:    "uint",
			value:    -1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint(tCase.expected), ToUint(tCase.value))
			},
		},
		{
			name:     "zero",
			ctype:    "int64",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int64(tCase.expected), ToInt64(tCase.value))
			},
		},
		{
			name:     "zero",
			ctype:    "uint64",
			value:    0,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint64(tCase.expected), ToUint64(tCase.value))
			},
		},
		{
			name:     "1",
			ctype:    "int64",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int64(tCase.expected), ToInt64(tCase.value))
			},
		},
		{
			name:     "1",
			ctype:    "uint64",
			value:    1,
			expected: 1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint64(tCase.expected), ToUint64(tCase.value))
			},
		},
		{
			name:     "-1",
			ctype:    "int64",
			value:    -1,
			expected: -1,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, int64(tCase.expected), ToInt64(tCase.value))
			},
		},
		{
			name:     "-1",
			ctype:    "uint64",
			value:    -1,
			expected: 0,
			testCaseFunc: func(r *testing.T, tCase *testCase[int64, int64]) {
				assert.Equal(r, uint64(tCase.expected), ToUint64(tCase.value))
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
			assert.Equal(t, int8(-4), ToInt8(-4.6))
			assert.Equal(t, int8(4), ToInt8(4.6))
			assert.Equal(t, uint8(0), ToUint8(-4.6))
			assert.Equal(t, uint8(4), ToUint8(4.6))
			assert.Equal(t, int8(math.MaxInt8), ToInt8(256.4))
			assert.Equal(t, uint8(math.MaxUint8), ToUint8(256.4))
		})
		t.Run("int16", func(t *testing.T) {
			assert.Equal(t, int16(-4), ToInt16(-4.6))
			assert.Equal(t, int16(4), ToInt16(4.6))
			assert.Equal(t, uint16(0), ToUint16(-4.6))
			assert.Equal(t, uint16(4), ToUint16(4.6))
			assert.Equal(t, int16(math.MaxInt16), ToInt16(40000.4))
			assert.Equal(t, int16(math.MaxInt16), ToInt16(float32(40000.4)))
			assert.Equal(t, int16(math.MinInt16), ToInt16(-32768.4))
			assert.Equal(t, uint16(math.MaxUint16), ToUint16(70000.4))
		})
		t.Run("int32", func(t *testing.T) {
			assert.Equal(t, int32(-4), ToInt32(-4.6))
			assert.Equal(t, int32(4), ToInt32(4.6))
			assert.Equal(t, uint32(0), ToUint32(-4.6))
			assert.Equal(t, uint32(4), ToUint32(4.6))
			assert.Equal(t, int32(math.MaxInt32), ToInt32(2147483647.4))
			assert.Equal(t, int32(math.MaxInt32), ToInt32(float32(2147483647.4)))
			assert.Equal(t, int32(math.MinInt32), ToInt32(float32(-2147483648.4)))
			assert.Equal(t, uint32(math.MaxUint32), ToUint32(4294967295.4))
		})
		t.Run("int", func(t *testing.T) {
			assert.Equal(t, -4, ToInt(-4.6))
			assert.Equal(t, 4, ToInt(4.6))
			assert.Equal(t, uint(0), ToUint(-4.6))
			assert.Equal(t, uint(4), ToUint(4.6))
			assert.Equal(t, math.MaxInt, ToInt(9223372036854775807.4))
			assert.Equal(t, uint(math.MaxUint), ToUint(18446744073709551615.4))
			assert.Equal(t, math.MinInt, ToInt(-9223372036854775808.4))
			assert.Equal(t, uint(0), ToUint(-18446744073709551615.4))
		})
		t.Run("int64", func(t *testing.T) {
			assert.Equal(t, int64(-4), ToInt64(-4.6))
			assert.Equal(t, int64(4), ToInt64(4.6))
			assert.Equal(t, uint64(0), ToUint64(-4.6))
			assert.Equal(t, uint64(4), ToUint64(4.6))
			assert.Equal(t, int64(math.MaxInt64), ToInt64(9223372036854775807.4))
			assert.Equal(t, uint64(math.MaxUint64), ToUint64(18446744073709551616.4))
			assert.Equal(t, int64(math.MinInt64), ToInt64(-9223372036854775808.4))
			assert.Equal(t, uint64(0), ToUint64(-18446744073709551616.4))
		})
	})
}
