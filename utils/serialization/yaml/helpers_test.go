package yaml

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

type TestStruct struct {
	Elem1 string
	Elem2 int64
	Elem3 string
	Elem4 bool
	Elem5 float64
}

func TestMarshalAndUnmarshal(t *testing.T) {
	test := TestStruct{
		Elem1: faker.Sentence(),
		Elem2: 456,
		Elem3: faker.Paragraph(),
		Elem4: true,
		Elem5: 0.415465464565464,
	}

	encoded, err := Marshal(&test)
	require.NoError(t, err)
	assert.Contains(t, string(encoded), "Elem1:")

	var decoded TestStruct
	err = Unmarshal(encoded, &decoded)
	require.NoError(t, err)
	assert.Equal(t, test, decoded)
}

func TestEncoderAndDecoder(t *testing.T) {
	test := TestStruct{
		Elem1: faker.Sentence(),
		Elem2: 456,
		Elem3: faker.Paragraph(),
		Elem4: true,
		Elem5: 0.415465545454464565464,
	}

	var buf bytes.Buffer
	encoder := NewEncoder(context.Background(), &buf)
	err := encoder.Encode(test)
	require.NoError(t, err)

	var decoded TestStruct
	decoder := NewDecoder(context.Background(), &buf)
	err = decoder.Decode(&decoded)
	require.NoError(t, err)
	assert.Equal(t, test, decoded)
}

func TestMarshalAndUnmarshallWithContext(t *testing.T) {
	test := TestStruct{
		Elem1: faker.Sentence(),
		Elem2: 456,
		Elem3: faker.Paragraph(),
		Elem4: true,
		Elem5: 0.415465464565464,
	}

	encoded, err := MarshalWithContext(context.Background(), &test)
	require.NoError(t, err)

	var decoded TestStruct
	err = UnmarshallWithContext(context.Background(), encoded, &decoded)
	require.NoError(t, err)
	assert.Equal(t, test, decoded)
}

func TestToJSON(t *testing.T) {
	input := []byte("name: value\ncount: 2\n")

	output, err := ToJSON(input)
	require.NoError(t, err)

	assert.JSONEq(t, `{"count":2,"name":"value"}`, string(output))
}

func TestToJSONInvalidYAML(t *testing.T) {
	_, err := ToJSON([]byte("name: [value\n"))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrMarshalling)
}
