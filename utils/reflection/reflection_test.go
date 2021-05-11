package reflection

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type structtest struct {
	Exported   string
	unexported string
}

type TestTypeWithTitle struct {
	Name    string
	Address string
	Title   string
}

type TestTypeNoTitle struct {
	Name    string
	Address string
}

type TestTypeWithTitleAsPointer struct {
	Name    string
	Address string
	Title   *string
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

func TestGetStructField_Happy(t *testing.T) {
	// Given a structure that has a title field
	// It returns the title field and true
	test_structure := TestTypeWithTitle{
		Name:    "test_name",
		Address: "random_address",
		Title:   "test_title",
	}
	result, exists := GetStructField(&test_structure, "Title")
	assert.Equal(t, result, "test_title")
	assert.Equal(t, exists, true)
}

func TestGetStructField_NoTitle(t *testing.T) {
	// Given a structure that does not have a title field
	// It returns "" and false
	test_structure := TestTypeNoTitle{
		Name:    "test_name",
		Address: "random_address",
	}
	result, exists := GetStructField(&test_structure, "Title")
	assert.Equal(t, result, "")
	assert.Equal(t, exists, false)
}

func TestGetStructField_TitleNotSet(t *testing.T) {
	// Given a structure that has a title field which is not set
	// It returns the content of the field (i.e. "") and true
	test_structure := TestTypeWithTitle{
		Name:    "test_name",
		Address: "random_address",
	}
	result, exists := GetStructField(&test_structure, "Title")
	assert.Equal(t, result, "")
	assert.Equal(t, exists, true)
}

func TestGetStructField_TitleStringPtr(t *testing.T) {
	// Given a structure that has a title field which is not set
	// It returns the content of the field (i.e. "") and true
	title := "test_title"
	test_structure := TestTypeWithTitleAsPointer{
		Name:    "test_name",
		Address: "random_address",
		Title:   &title,
	}
	result, exists := GetStructField(&test_structure, "Title")
	assert.Equal(t, result, title)
	assert.Equal(t, exists, true)
}

func TestInheritsFrom(t *testing.T) {
	type A0 interface {
		test()
	}
	type A struct{ A0 }
	type B struct{ A }
	type C struct{ *A }
	type D struct {
		B
	}
	type E struct {
		C
	}
	type F struct {
		D
	}
	type G struct {
		D
		E
	}
	type H struct {
	}
	type I struct{ H }
	type J struct {
		B
		I
	}
	type K struct {
		I
		B
	}

	tests := []struct {
		object   interface{}
		inherits bool
		name     string
	}{
		{
			object:   A{},
			inherits: true,
			name:     "A",
		},
		{
			object:   B{},
			inherits: true,
			name:     "B",
		},
		{
			object: C{
				A: &A{},
			},
			inherits: true,
			name:     "C2",
		},
		{
			object:   D{},
			inherits: true,
			name:     "D",
		},
		{
			object: E{C{
				A: &A{},
			},
			},
			inherits: true,
			name:     "E",
		},
		{
			object:   F{},
			inherits: true,
			name:     "F",
		},
		{
			object:   G{},
			inherits: true,
			name:     "G",
		},
		{
			object:   H{},
			inherits: false,
			name:     "H",
		},
		{
			object:   I{},
			inherits: false,
			name:     "I",
		},
		{
			object:   J{},
			inherits: true,
			name:     "J",
		},
		{
			object:   K{},
			inherits: true,
			name:     "K",
		},
	}

	A0Type := reflect.TypeOf((*A0)(nil)).Elem()
	AType := reflect.TypeOf(A{})
	AstarType := reflect.TypeOf(&A{})
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("subtest #%v: struct %v", i, test.name), func(t *testing.T) {
			assert.Equal(t, test.inherits, InheritsFrom(test.object, A0Type))
			assert.Equal(t, test.inherits, InheritsFrom(test.object, AType))
			assert.Equal(t, test.inherits, InheritsFrom(test.object, AstarType))
		})
	}
}
