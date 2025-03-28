/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package reflection

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
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
	testStructure := TestTypeWithTitle{
		Name:    "test_name",
		Address: "random_address",
		Title:   "test_title",
	}
	result, exists := GetStructField(&testStructure, "Title")
	assert.Equal(t, result, "test_title")
	assert.Equal(t, exists, true)
}

func TestGetStructField_NoTitle(t *testing.T) {
	// Given a structure that does not have a title field
	// It returns "" and false
	testStructure := TestTypeNoTitle{
		Name:    "test_name",
		Address: "random_address",
	}
	result, exists := GetStructField(&testStructure, "Title")
	assert.Equal(t, result, "")
	assert.Equal(t, exists, false)
}

func TestGetStructField_TitleNotSet(t *testing.T) {
	// Given a structure that has a title field which is not set
	// It returns the content of the field (i.e. "") and true
	testStructure := TestTypeWithTitle{
		Name:    "test_name",
		Address: "random_address",
	}
	result, exists := GetStructField(&testStructure, "Title")
	assert.Equal(t, result, "")
	assert.Equal(t, exists, true)
}

func TestGetStructField_TitleStringPtr(t *testing.T) {
	// Given a structure that has a title field which is not set
	// It returns the content of the field (i.e. "") and true
	title := "test_title"
	testStructure := TestTypeWithTitleAsPointer{
		Name:    "test_name",
		Address: "random_address",
		Title:   &title,
	}
	result, exists := GetStructField(&testStructure, "Title")
	assert.Equal(t, result, title)
	assert.Equal(t, exists, true)
}

func TestGetStructField_TitlePtrNotSet(t *testing.T) {
	// Given a structure that has a title field pointer which is not set
	// It returns the content of the field (i.e. "") and true
	testStructure := TestTypeWithTitleAsPointer{
		Name:    "test_name",
		Address: "random_address",
	}
	result, exists := GetStructField(&testStructure, "Title")
	assert.Equal(t, "", result)
	assert.Equal(t, exists, true)
}

func TestSetStructField_FieldPtrValueNotPtr(t *testing.T) {
	// Given a structure that has the given field
	// The field is a pointer and the value is not a pointer
	// It returns no errors and it updates the structure's field to the value
	title := "test_title"
	newTitle := "NEW_title"
	testStructure := TestTypeWithTitleAsPointer{
		Name:    "test_name",
		Address: "random_address",
		Title:   &title,
	}
	err := SetStructField(&testStructure, "Title", newTitle)

	assert.Equal(t, *testStructure.Title, newTitle)
	assert.Nil(t, err)
}

func TestSetStructField_FieldPtrUnsetValueNotPtr(t *testing.T) {
	// Given a structure that has the given field
	// The field is a nil pointer and the value is not a pointer
	// It returns no errors and it updates the structure's field to the value
	newTitle := "NEW_title"
	testStructure := TestTypeWithTitleAsPointer{
		Name:    "test_name",
		Address: "random_address",
	}
	err := SetStructField(&testStructure, "Title", newTitle)
	assert.Equal(t, *testStructure.Title, newTitle)
	assert.Nil(t, err)
}

func TestSetStructField_FieldPtrValuePtr(t *testing.T) {
	// Given a structure that has the given field
	// The field and the value are pointers
	// It returns no errors and it updates the structure's field to the value
	title := "test_title"
	newTitle := "NEW_title"
	testStructure := TestTypeWithTitleAsPointer{
		Name:    "test_name",
		Address: "random_address",
		Title:   &title,
	}
	err := SetStructField(&testStructure, "Title", &newTitle)

	assert.Equal(t, *testStructure.Title, newTitle)
	assert.Nil(t, err)
}

func TestSetStructField_FieldNotPtrValuePtr(t *testing.T) {
	// Given a structure that has the given field
	// The field is not a pointer but the value is a pointer
	// It returns no errors and it updates the structure's field to the value
	newTitle := "NEW_title"
	testStructure := TestTypeWithTitle{
		Name:    "test_name",
		Address: "random_address",
		Title:   "test_title",
	}
	err := SetStructField(&testStructure, "Title", &newTitle)

	assert.Equal(t, testStructure.Title, newTitle)
	assert.Nil(t, err)
}

func TestSetStructField_FieldNotPtrValueNotPtr(t *testing.T) {
	// Given a structure that has the given field
	// The field and the value are not pointers
	// It returns no errors and it updates the structure's field to the value
	newTitle := "NEW_title"
	testStructure := TestTypeWithTitle{
		Name:    "test_name",
		Address: "random_address",
		Title:   "test_title",
	}
	err := SetStructField(&testStructure, "Title", "NEW_title")

	assert.Equal(t, testStructure.Title, newTitle)
	assert.Nil(t, err)
}

func TestSetStructField_InvalidField(t *testing.T) {
	// Given a structure that doesn't have the given field
	// It returns an error saying that the field is invalid
	testStructure := TestTypeNoTitle{
		Name:    "test_name",
		Address: "random_address",
	}
	err := SetStructField(&testStructure, "Title", "NEW_title")

	assert.NotNil(t, err)
	assert.Equal(t, err, fmt.Errorf("error with field [%v]: %w", "Title", commonerrors.ErrInvalid))
}

func TestSetStructField_UnsettableField(t *testing.T) {
	// Given a structure that has an unexported field
	// It returns an error saying that the field is unsettable
	// And the field is not updated
	testStructure := structtest{
		Exported:   "settable_field",
		unexported: "unsettable_field",
	}
	err := SetStructField(&testStructure, "unexported", "NEW_title")

	assert.NotNil(t, err)
	assert.Equal(t, err, fmt.Errorf("error with unsettable field [%v]: %w", "unexported", commonerrors.ErrUnsupported))
	assert.NotEqual(t, testStructure.unexported, "NEW_title")
	assert.Equal(t, testStructure.unexported, "unsettable_field")
}

func TestSetStructField_FieldAndValueDifferentTypes(t *testing.T) {
	// Given a structure that has the given field
	// The field and the value are of different types
	// It returns an error saying
	title := "test_title"
	testStructure := TestTypeWithTitle{
		Name:    "test_name",
		Address: "random_address",
		Title:   title,
	}
	err := SetStructField(&testStructure, "Title", 133)

	assert.NotNil(t, err)
	assert.Equal(t, err, fmt.Errorf("conflicting types, field [%v] and value [%v]: %w", reflect.ValueOf(testStructure).FieldByName("Title").Type().Kind(), reflect.TypeOf(123), commonerrors.ErrConflict))
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

func TestIsEmpty(t *testing.T) {
	type testInterface interface {
	}
	var testEmptyPtr testInterface
	aFilledChannel := make(chan struct{}, 1)
	aFilledChannel <- struct{}{}
	tests := []struct {
		value                  interface{}
		isEmpty                bool
		differsFromAssertEmpty bool
	}{
		{
			value:   nil,
			isEmpty: true,
		},
		{
			value:   0,
			isEmpty: true,
		},
		{
			value:   uint(0),
			isEmpty: true,
		},
		{
			value:   float64(0),
			isEmpty: true,
		},
		{
			value:   "",
			isEmpty: true,
		},
		{
			value:                  "                                   ",
			isEmpty:                true,
			differsFromAssertEmpty: true,
		},
		{
			value:   (*string)(nil),
			isEmpty: true,
		},
		{
			value:   field.ToOptionalString(""),
			isEmpty: true,
		},
		{
			value:                  field.ToOptionalString("                                   "),
			isEmpty:                true,
			differsFromAssertEmpty: true,
		},
		{
			value:   false,
			isEmpty: true,
		},
		{
			value:   []string{},
			isEmpty: true,
		},
		{
			value:   []int64{},
			isEmpty: true,
		},
		{
			value:   []int64{int64(0)},
			isEmpty: false,
		},
		{
			value:   "blah",
			isEmpty: false,
		},
		{
			value:   1,
			isEmpty: false,
		},
		{
			value:   true,
			isEmpty: false,
		},
		{
			value:   testEmptyPtr,
			isEmpty: true,
		},
		{
			value:   map[string]string{},
			isEmpty: true,
		},
		{
			value:   map[string]interface{}{},
			isEmpty: true,
		},
		{
			value:   map[string]interface{}{"foo": "bar"},
			isEmpty: false,
		},
		{
			value:   time.Time{},
			isEmpty: true,
		},
		{
			value:   time.Now(),
			isEmpty: false,
		},
		{
			value:   make(chan struct{}),
			isEmpty: true,
		},
		{
			value:   aFilledChannel,
			isEmpty: false,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("subtest #%v (%v)", i, test.value), func(t *testing.T) {
			assert.Equal(t, test.isEmpty, IsEmpty(test.value))
			if test.isEmpty && !test.differsFromAssertEmpty {
				assert.Empty(t, test.value)
			} else {
				assert.NotEmpty(t, test.value)
			}
		})
	}
}

func TestToStructPtr(t *testing.T) {
	vInt := 15
	vStr := faker.Sentence()
	var vPrt *string
	vBytes := []byte(faker.Sentence())
	vBool := false
	vFloat := 150.454
	vMap := map[string]string{faker.Word(): faker.Sentence()}
	vArray := []string{faker.Word(), faker.Sentence(), faker.Name()}
	vStruct := struct {
		Test  string
		Test2 int
		test3 *string
	}{
		Test:  vStr,
		Test2: vInt,
		test3: vPrt,
	}
	var vEmpty interface{}
	type testInterface interface {
	}
	var vInterface testInterface

	tests := []struct {
		input          interface{}
		expectedOutput interface{}
		expectedError  error
	}{
		{
			input:          nil,
			expectedOutput: nil,
			expectedError:  commonerrors.ErrUnsupported,
		},
		{
			input:          vInt,
			expectedOutput: &vInt,
			expectedError:  nil,
		},
		{
			input:          vStr,
			expectedOutput: &vStr,
			expectedError:  nil,
		},
		{
			input:          vPrt,
			expectedOutput: &vPrt,
			expectedError:  nil,
		},
		{
			input:          vBytes,
			expectedOutput: &vBytes,
			expectedError:  nil,
		},
		{
			input:          vBool,
			expectedOutput: &vBool,
			expectedError:  nil,
		},
		{
			input:          vFloat,
			expectedOutput: &vFloat,
			expectedError:  nil,
		},
		{
			input:          vMap,
			expectedOutput: &vMap,
			expectedError:  nil,
		},
		{
			input:          vArray,
			expectedOutput: &vArray,
			expectedError:  nil,
		},
		{
			input:          vStruct,
			expectedOutput: &vStruct,
			expectedError:  nil,
		},
		{
			input:          vEmpty,
			expectedOutput: nil,
			expectedError:  commonerrors.ErrUnsupported,
		},
		{
			input:          vInterface,
			expectedOutput: nil,
			expectedError:  commonerrors.ErrUnsupported,
		},
		{
			input:          reflect.ValueOf(vStruct).FieldByName("Test"),
			expectedOutput: &vStruct.Test,
			expectedError:  nil,
		},
		{
			input:          reflect.ValueOf(vStruct).FieldByName("test3"),
			expectedOutput: nil,
			expectedError:  commonerrors.ErrUnsupported,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			var err error
			var input reflect.Value
			if v, ok := test.input.(reflect.Value); ok {
				input = v
			} else {
				input = reflect.ValueOf(test.input)
			}
			actualVal, err := ToStructPtr(input)
			if test.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, commonerrors.Any(err, test.expectedError))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedOutput, actualVal)
			}
		})
	}
}
