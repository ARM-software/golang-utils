package yamltest

import (
	"bytes"
	"context"
	"path"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/filesystem"
	yamlserialization "github.com/ARM-software/golang-utils/utils/serialization/yaml"
	"github.com/ARM-software/golang-utils/utils/serialization/yaml/yamltest/nofast"
)

func TestYAMLMarshalling(t *testing.T) {
	testStruct0 := &nofast.TestingStruct{}
	require.NoError(t, faker.FakeData(testStruct0))

	testStruct1 := &nofast.TestingStruct{}
	assert.False(t, testStruct0.Equals(testStruct1))

	encoded, err := yamlserialization.Marshal(testStruct0)
	require.NoError(t, err)

	err = yamlserialization.Unmarshal(encoded, testStruct1)
	require.NoError(t, err)

	assert.True(t, testStruct0.Equals(testStruct1))
	assert.True(t, testStruct1.Equals(testStruct0))
}

func TestYAMLMarshallingEncoding(t *testing.T) {
	test := &nofast.TestingStruct{}
	require.NoError(t, faker.FakeData(test))

	var buf bytes.Buffer
	encoder := yamlserialization.NewEncoder(context.Background(), &buf)
	err := encoder.Encode(test)
	require.NoError(t, err)

	decoded := &nofast.TestingStruct{}
	decoder := yamlserialization.NewDecoder(context.Background(), &buf)
	err = decoder.Decode(decoded)
	require.NoError(t, err)

	assert.True(t, test.Equals(decoded))
	assert.True(t, decoded.Equals(test))
}

func TestYAMLMarshallingNil(t *testing.T) {
	value, err := yamlserialization.Marshal(nil)
	require.NoError(t, err)
	assert.NotEmpty(t, value)
}

func TestYAMLFixtures(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		assertFn func(t *testing.T, data []byte)
	}{
		{
			name:     "anchor as key with alias",
			fileName: "anchor_as_key_with_alias.yaml",
			assertFn: func(t *testing.T, data []byte) {
				actual := map[string]any{}
				require.NoError(t, yamlserialization.Unmarshal(data, &actual))
				assert.Equal(t, map[string]any{"foo": "bar", "bar": "quz"}, actual)
			},
		},
		{
			name:     "alias reuse",
			fileName: "alias_reuse.yaml",
			assertFn: func(t *testing.T, data []byte) {
				actual := map[string]any{}
				require.NoError(t, yamlserialization.Unmarshal(data, &actual))
				assert.Equal(t, map[string]any{
					"First occurrence":  "Foo",
					"Second occurrence": "Foo",
					"Override anchor":   "Bar",
					"Reuse anchor":      "Bar",
				}, actual)
			},
		},
		{
			name:     "yaml 1.1 bool compatibility",
			fileName: "yaml11_bool_compat.yaml",
			assertFn: func(t *testing.T, data []byte) {
				actual := map[string]bool{}
				require.NoError(t, yamlserialization.Unmarshal(data, &actual))
				assert.Equal(t, map[string]bool{"option": true}, actual)
			},
		},
		{
			name:     "anchor sequence alias",
			fileName: "anchor_sequence_alias.yaml",
			assertFn: func(t *testing.T, data []byte) {
				actual := map[string][]int{}
				require.NoError(t, yamlserialization.Unmarshal(data, &actual))
				assert.Equal(t, map[string][]int{"a": {1, 2}, "b": {1, 2}}, actual)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := filesystem.ReadFile(path.Join("testdata", test.fileName))
			require.NoError(t, err)
			test.assertFn(t, data)
		})
	}
}

func TestToJSONFixtures(t *testing.T) {
	data, err := filesystem.ReadFile(path.Join("testdata", "anchor_as_key_with_alias.yaml"))
	require.NoError(t, err)

	jsonData, err := yamlserialization.ToJSON(data)
	require.NoError(t, err)

	assert.JSONEq(t, `{"bar":"quz","foo":"bar"}`, string(jsonData))
}
