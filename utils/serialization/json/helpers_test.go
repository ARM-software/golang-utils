package json

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

func TestMarshalling(t *testing.T) {
	test := TestStruct{
		Elem1: faker.Sentence(),
		Elem2: 456,
		Elem3: faker.Paragraph(),
		Elem4: true,
		Elem5: 0.415465464565464,
	}

	encoded, err := Marshal(&test)
	require.NoError(t, err)

	var decoded TestStruct
	err = Unmarshal(encoded, &decoded)
	require.NoError(t, err)
	assert.Equal(t, test, decoded)
}

func TestMarshallingEncoding(t *testing.T) {
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

func TestToYAML(t *testing.T) {
	input := []byte(`{"name":"value","count":2}`)

	output, err := ToYAML(input)
	require.NoError(t, err)

	assert.Contains(t, string(output), "name: value")
	assert.Contains(t, string(output), "count: 2")
}

func TestToYAMLIntegrationInspiredByKubernetesSigsYAML(t *testing.T) {
	// Tests inspired by https://github.com/kubernetes-sigs/yaml/blob/master/yaml_test.go
	tests := map[string]struct {
		json             string
		expectedContains []string
	}{
		"string value": {
			json:             `{"t":"a"}`,
			expectedContains: []string{"t: a"},
		},
		"boolean value": {
			json:             `{"t":true}`,
			expectedContains: []string{"t: true"},
		},
		"array": {
			json:             `[{"t":"a"}]`,
			expectedContains: []string{"- t: a"},
		},
		"large integer": {
			json:             `{"t":9007199254740993}`,
			expectedContains: []string{"t: 9007199254740993"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := ToYAML([]byte(test.json))
			require.NoError(t, err)
			for _, expected := range test.expectedContains {
				assert.Contains(t, string(output), expected)
			}
		})
	}
}

func TestToYAMLInvalidJSON(t *testing.T) {
	_, err := ToYAML([]byte(`{"name":`))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrMarshalling)
}
