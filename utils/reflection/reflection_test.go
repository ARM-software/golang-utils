package reflection

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type structtest struct {
	Exported   string
	unexported string
}

func TestGetUnexportedField(t *testing.T) {
	testValue1 := "field1"
	testValue2 := "field2"
	test := &structtest{
		Exported:   testValue1,
		unexported: testValue2,
	}
	require.Equal(t, testValue1, GetUnexportedStructureField(test, "Exported"))
	require.Equal(t, testValue2, GetUnexportedStructureField(test, "unexported"))
}

func TestGetUnexportedInvalidField(t *testing.T) {
	testValue1 := "field1"
	testValue2 := "field2"
	test := &structtest{
		Exported:   testValue1,
		unexported: testValue2,
	}
	require.Nil(t, GetUnexportedStructureField(test, "Exported-Incorrect"))
	require.Nil(t, GetUnexportedStructureField(test, "unexported-Incorrect"))
}

func TestSetUnexportedField(t *testing.T) {
	testValue1 := "field1"
	testValue2 := "field2"
	test := &structtest{
		Exported:   "",
		unexported: "",
	}
	require.Zero(t, GetUnexportedStructureField(test, "Exported"))
	require.Zero(t, GetUnexportedStructureField(test, "unexported"))
	SetUnexportedStructureField(test, "Exported", testValue1)
	SetUnexportedStructureField(test, "unexported", testValue2)
	require.Equal(t, testValue1, GetUnexportedStructureField(test, "Exported"))
	require.Equal(t, testValue2, GetUnexportedStructureField(test, "unexported"))
}

func TestSetUnexportedFieldInvalid(t *testing.T) {
	testValue1 := "field1"
	testValue2 := "field2"
	test := &structtest{
		Exported:   "",
		unexported: "",
	}
	require.Zero(t, GetUnexportedStructureField(test, "Exported"))
	require.Zero(t, GetUnexportedStructureField(test, "unexported"))
	SetUnexportedStructureField(test, "Exported-Incorrect", testValue1)
	SetUnexportedStructureField(test, "unexported-Incorrect", testValue2)
	require.Zero(t, GetUnexportedStructureField(test, "Exported"))
	require.Zero(t, GetUnexportedStructureField(test, "unexported"))
}
