package value

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

var errBoom = errors.New("boom")

type stringValue string

func (s stringValue) String() string {
	return "stringer:" + string(s)
}

type textValue string

func (t textValue) MarshalText() ([]byte, error) {
	return []byte("text:" + string(t)), nil
}

func (t textValue) String() string {
	return "stringer:" + string(t)
}

type failingTextValue struct{}

func (failingTextValue) MarshalText() ([]byte, error) {
	return nil, errBoom
}

type marshalOnlyValue string

func (m marshalOnlyValue) MarshalText() ([]byte, error) {
	return []byte("text:" + string(m)), nil
}

type nilStringer struct{}

func (*nilStringer) String() string {
	return "unexpected"
}

type plainStruct struct {
	Name  string
	Count int
}

type structWithPointerField struct {
	Name  string
	Value *string
}

func TestIdentityConverter(t *testing.T) {
	value, err := IdentityConverter.ConvertValue(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, 42, value)
}

func TestStringConverter(t *testing.T) {
	t.Run("matches flatten scalar formatting", func(t *testing.T) {
		require.NoError(t, faker.SetRandomMapAndSliceSize(1))

		generatedInt64, err := faker.RandomInt(1, 1000, 1)
		require.NoError(t, err)

		generatedUint64, err := faker.RandomInt(1, 1000, 1)
		require.NoError(t, err)

		generatedFloatParts, err := faker.RandomInt(1, 1000, 2)
		require.NoError(t, err)

		testCases := []struct {
			name     string
			input    any
			expected string
		}{
			{name: "bool true", input: true, expected: "true"},
			{name: "bool false", input: false, expected: "false"},
			{name: "int64", input: int64(generatedInt64[0]), expected: fmt.Sprintf("%d", int64(generatedInt64[0]))}, //nolint:gosec // G115: testing
			{name: "duration", input: 5*time.Second + 12*time.Millisecond, expected: "5.012s"},
			{name: "int", input: int(-7), expected: "-7"},                                                                                            //nolint:gosec // G115: testing
			{name: "int8", input: int8(-8), expected: "-8"},                                                                                          //nolint:gosec // G115: testing
			{name: "int16", input: int16(-16), expected: "-16"},                                                                                      //nolint:gosec // G115: testing
			{name: "int32", input: int32(-32), expected: "-32"},                                                                                      //nolint:gosec // G115: testing
			{name: "uint", input: uint(7), expected: "7"},                                                                                            //nolint:gosec // G115: testing
			{name: "uint8", input: uint8(8), expected: "8"},                                                                                          //nolint:gosec // G115: testing
			{name: "uint16", input: uint16(16), expected: "16"},                                                                                      //nolint:gosec // G115: testing
			{name: "uint32", input: uint32(32), expected: "32"},                                                                                      //nolint:gosec // G115: testing
			{name: "uint64", input: uint64(generatedUint64[0]), expected: fmt.Sprintf("%d", uint64(generatedUint64[0]))},                             //nolint:gosec // G115: testing
			{name: "float64", input: float64(generatedFloatParts[0]) / 10, expected: fmt.Sprintf("%g", float64(generatedFloatParts[0])/10)},          //nolint:gosec // G115: testing
			{name: "float32", input: float32(generatedFloatParts[1]) / 10, expected: fmt.Sprintf("%g", float64(float32(generatedFloatParts[1])/10))}, //nolint:gosec // G115: testing
			{name: "string", input: faker.Sentence(), expected: faker.Sentence()},
		}

		testCases[len(testCases)-1].input = testCases[len(testCases)-1].expected

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				value, err := StringConverter.ConvertValue(context.Background(), testCase.input)
				require.NoError(t, err)
				assert.Equal(t, testCase.expected, value)
			})
		}
	})

	t.Run("stringer", func(t *testing.T) {
		value, err := StringConverter.ConvertValue(context.Background(), stringValue("value"))
		require.NoError(t, err)
		assert.Equal(t, "stringer:value", value)
	})

	t.Run("pointer stringer preferred over pointer formatting", func(t *testing.T) {
		input := new(stringValue)
		*input = "value"

		value, err := StringConverter.ConvertValue(context.Background(), input)
		require.NoError(t, err)
		assert.Equal(t, "stringer:value", value)
	})

	t.Run("stringer preferred over text marshaler", func(t *testing.T) {
		value, err := StringConverter.ConvertValue(context.Background(), textValue("value"))
		require.NoError(t, err)
		assert.Equal(t, "stringer:value", value)
	})

	t.Run("text marshaler", func(t *testing.T) {
		value, err := StringConverter.ConvertValue(context.Background(), marshalOnlyValue("value"))
		require.NoError(t, err)
		assert.Equal(t, "text:value", value)
	})

	t.Run("pointer text marshaler preferred over pointer formatting", func(t *testing.T) {
		input := new(marshalOnlyValue)
		*input = "value"

		value, err := StringConverter.ConvertValue(context.Background(), input)
		require.NoError(t, err)
		assert.Equal(t, "text:value", value)
	})

	t.Run("marshal text error", func(t *testing.T) {
		_, err := StringConverter.ConvertValue(context.Background(), failingTextValue{})
		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrMarshalling)
		assert.ErrorContains(t, err, errBoom.Error())
	})

	t.Run("struct uses verbose formatting with hardcoded values", func(t *testing.T) {
		input := plainStruct{Name: "value", Count: 2}

		converted, err := StringConverter.ConvertValue(context.Background(), input)
		require.NoError(t, err)
		assert.Equal(t, "{Name:value Count:2}", converted)
	})

	t.Run("struct uses verbose formatting", func(t *testing.T) {
		input := plainStruct{Name: faker.Word(), Count: int(faker.RandomUnixTime())} //nolint:gosec // G115: testing

		converted, err := StringConverter.ConvertValue(context.Background(), input)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%+v", input), converted)
	})

	t.Run("pointer includes address and converted value with hardcoded values", func(t *testing.T) {
		input := &plainStruct{Name: "value", Count: 2}

		converted, err := StringConverter.ConvertValue(context.Background(), input)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%p -> {Name:value Count:2}", input), converted)
	})

	t.Run("pointer includes address and converted value", func(t *testing.T) {
		input := &plainStruct{Name: faker.Word(), Count: int(faker.RandomUnixTime())} //nolint:gosec // G115: testing

		converted, err := StringConverter.ConvertValue(context.Background(), input)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%p -> %v", input, fmt.Sprintf("%+v", *input)), converted)
	})

	t.Run("nil plain pointer falls back to nil", func(t *testing.T) {
		var input *plainStruct

		converted, err := StringConverter.ConvertValue(context.Background(), input)
		require.NoError(t, err)
		assert.Equal(t, "<nil>", converted)
	})

	t.Run("pointer to nil pointer includes outer pointer and nil value", func(t *testing.T) {
		var inner *plainStruct
		input := &inner

		converted, err := StringConverter.ConvertValue(context.Background(), input)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%p -> %v", input, "<nil>"), converted)
	})

	t.Run("pointer to struct with nil pointer field uses hardcoded struct formatting", func(t *testing.T) {
		input := &structWithPointerField{Name: "value"}

		converted, err := StringConverter.ConvertValue(context.Background(), input)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%p -> {Name:value Value:<nil>}", input), converted)
	})

	t.Run("pointer to struct with nil pointer field uses struct formatting", func(t *testing.T) {
		input := &structWithPointerField{Name: faker.Word()}

		converted, err := StringConverter.ConvertValue(context.Background(), input)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%p -> %v", input, fmt.Sprintf("%+v", *input)), converted)
	})

	t.Run("nil stringer pointer falls back", func(t *testing.T) {
		var value *nilStringer
		converted, err := StringConverter.ConvertValue(context.Background(), value)
		require.NoError(t, err)
		assert.Equal(t, "<nil>", converted)
	})
}
