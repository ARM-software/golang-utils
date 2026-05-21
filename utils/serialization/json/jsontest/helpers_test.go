package jsontest

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jsonserialization "github.com/ARM-software/golang-utils/utils/serialization/json"
	"github.com/ARM-software/golang-utils/utils/serialization/json/jsontest/easyjson"
	"github.com/ARM-software/golang-utils/utils/serialization/json/jsontest/ffjson"
	"github.com/ARM-software/golang-utils/utils/serialization/json/jsontest/nofast"
)

type equals interface {
	Equals(any) bool
}

func TestFastJSONMarshalling(t *testing.T) {
	tests := []struct {
		name  string
		test1 equals
		test2 equals
	}{
		{
			name:  "no fast json",
			test1: &nofast.TestingStruct{},
			test2: &nofast.TestingStruct{},
		},
		{
			name:  "ffjson",
			test1: &ffjson.TestingStruct{},
			test2: &ffjson.TestingStruct{},
		},
		{
			name:  "easyjson",
			test1: &easyjson.TestingStruct{},
			test2: &easyjson.TestingStruct{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testStruct0 := test.test1
			require.NoError(t, faker.FakeData(testStruct0))

			testStruct1 := test.test2
			assert.False(t, testStruct0.Equals(testStruct1))
			assert.False(t, testStruct1.Equals(testStruct0))

			encoded, err := jsonserialization.Marshal(testStruct0)
			require.NoError(t, err)

			err = jsonserialization.Unmarshal(encoded, testStruct1)
			require.NoError(t, err)

			assert.True(t, testStruct0.Equals(testStruct1))
			assert.True(t, testStruct1.Equals(testStruct0))
		})
	}
}

func TestJSONMarshallingNil(t *testing.T) {
	value, err := jsonserialization.Marshal(nil)
	require.NoError(t, err)
	assert.NotEmpty(t, value)
}
