/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package reflection

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func GetUnexportedStructureField(structure interface{}, fieldName string) interface{} {
	return GetStructureField(fetchStructureField(structure, fieldName))
}

func GetStructureField(field reflect.Value) interface{} {
	if !field.IsValid() {
		return nil
	}
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())). //nolint:gosec // this conversion is between types recommended by Go https://cs.opensource.google/go/go/+/master:src/reflect/value.go;l=2445
										Elem().
										Interface()
}
func SetUnexportedStructureField(structure interface{}, fieldName string, value interface{}) {
	SetStructureField(fetchStructureField(structure, fieldName), value)
}
func SetStructureField(field reflect.Value, value interface{}) {
	if !field.IsValid() {
		return
	}
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())). //nolint:gosec // this conversion is between types recommended by Go https://cs.opensource.google/go/go/+/master:src/reflect/value.go;l=2445
										Elem().
										Set(reflect.ValueOf(value))
}

func fetchStructureField(structure interface{}, fieldName string) reflect.Value {
	return reflect.ValueOf(structure).Elem().FieldByName(fieldName)
}

// GetStructField checks if the given structure has a given field. The structure should be passed by reference.
// It returns an interface and a boolean, the field's content and a boolean denoting whether or not the field exists.
// If the boolean is false then there is no such field on the structure.
// If the boolean is true but the interface stores "" then the field exists but is not set.
// If the boolean is true and the interface is not empty, the field exists and is set.
func GetStructField(structure interface{}, fieldName string) (interface{}, bool) {
	Field := fetchStructureField(structure, fieldName)
	if !Field.IsValid() {
		return "", false
	}

	if Field.Type().Kind() == reflect.Ptr {
		if Field.IsNil() {
			return "", true
		}
		return Field.Elem().Interface(), true
	} else {
		return Field.Interface(), true
	}
}

// SetStructField attempts to set a field of a structure to the given vaule
// It returns nil or an error, in case the field doesn't exist on the structure
// or the value and the field have different types
func SetStructField(structure interface{}, fieldName string, value interface{}) error {
	ValueStructure := reflect.ValueOf(structure)
	Field := ValueStructure.Elem().FieldByName(fieldName)
	// Test field exists on structure
	if !Field.IsValid() {
		return fmt.Errorf("error with field [%v]: %w", fieldName, commonerrors.ErrInvalid)
	}

	// test field is settable
	if !Field.CanSet() {
		return fmt.Errorf("error with unsettable field [%v]: %w", fieldName, commonerrors.ErrUnsupported)
	}

	// Helper variables
	valueReflectValueWrapper := reflect.ValueOf(value)
	valueKind := valueReflectValueWrapper.Type().Kind()
	fieldKind := Field.Type().Kind()

	// Value and field have the same type
	if valueKind == fieldKind {
		Field.Set(valueReflectValueWrapper)
		return nil
	}

	// helpers for determining whether the field and the value have the same underlying types
	valueUnderlyingType := reflect.TypeOf(value)
	if valueKind == reflect.Ptr {
		valueUnderlyingType = valueUnderlyingType.Elem()
	}
	fieldUnderlyingType := Field.Type()
	if fieldKind == reflect.Ptr {
		fieldUnderlyingType = fieldUnderlyingType.Elem()
	}

	// Check that the underlying types are the same (e.g. no int and string)
	if fieldUnderlyingType != valueUnderlyingType {
		return fmt.Errorf("conflicting types, field [%v] and value [%v]: %w", fieldKind, valueKind, commonerrors.ErrConflict)
	}

	if fieldKind == reflect.Ptr {
		if valueKind != reflect.Ptr { // value not ptr, field ptr
			if Field.IsNil() {
				pointerToValue := reflect.New(valueReflectValueWrapper.Type())
				pointerToValue.Elem().Set(valueReflectValueWrapper)
				Field.Set(pointerToValue)
			} else {
				Field.Elem().Set(valueReflectValueWrapper)
			}
		}
	} else { // field not ptr, val ptr
		if valueKind == reflect.Ptr {
			Field.Set(valueReflectValueWrapper.Elem())
		}
	}
	// This means the field was updated without errors
	return nil
}

// InheritsFrom uses reflection to find if a struct "inherits" from a certain type.
// In other words it checks whether the struct embeds a struct of that type.
func InheritsFrom(object interface{}, parentType reflect.Type) bool {
	if parentType == nil {
		return object == nil
	}
	r := reflect.ValueOf(object)
	t := r.Type()

	if t == parentType {
		return true
	}

	if r.Kind() == reflect.Ptr {
		if r.IsNil() {
			return false
		}
		r = r.Elem()
		if InheritsFrom(r.Interface(), parentType) {
			return true
		}
	}

	if r.Kind() == reflect.Interface {
		return r.Type().Implements(parentType)
	}
	if r.Kind() != reflect.Struct {
		return false
	}

	var (
		structType  reflect.Type
		pointerType reflect.Type
	)
	kind := parentType.Kind()
	switch {
	case kind == reflect.Ptr:
		pointerType = parentType
		structType = parentType.Elem()
	case kind == reflect.Interface:
		pointerType = parentType
	case kind == reflect.Struct:
		structType = parentType
	}

	if pointerType != nil && (t.AssignableTo(pointerType) || t.ConvertibleTo(pointerType)) {
		return true
	}
	if structType != nil && (t.AssignableTo(structType) || t.ConvertibleTo(structType)) {
		return true
	}

	for i := 0; i < r.NumField(); i++ {
		f := r.Field(i)
		if f.Type() == parentType {
			return true
		}
		fieldType := f.Type()
		if pointerType != nil && (fieldType.AssignableTo(pointerType) || fieldType.ConvertibleTo(pointerType)) {
			return true
		}
		if structType != nil && (fieldType.AssignableTo(structType) || fieldType.ConvertibleTo(structType)) {
			return true
		}

		if f.CanInterface() && InheritsFrom(f.Interface(), parentType) {
			return true
		}
	}
	return false
}

// IsEmpty checks whether a value is empty i.e. "", nil, 0, [], {}, false, etc.
// For Strings, a string is considered empty if it is "" or if it only contains whitespaces
func IsEmpty(value any) bool {
	if value == nil {
		return true
	}
	if valueStr, ok := value.(string); ok {
		return len(strings.TrimSpace(valueStr)) == 0
	}
	if valueStrPtr, ok := value.(*string); ok {
		if valueStrPtr == nil {
			return true
		}
		return len(strings.TrimSpace(*valueStrPtr)) == 0
	}
	if valueBool, ok := value.(bool); ok {
		// if set to true, then value is not empty
		return !valueBool
	}
	objValue := reflect.ValueOf(value)
	switch objValue.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return objValue.Len() == 0
	case reflect.Ptr:
		if objValue.IsNil() {
			return true
		}
		deref := objValue.Elem().Interface()
		return IsEmpty(deref)
	default:
		zero := reflect.Zero(objValue.Type())
		return reflect.DeepEqual(value, zero.Interface())
	}
}

// ToStructPtr returns an instance of the pointer (interface) to the object obj.
func ToStructPtr(obj reflect.Value) (val interface{}, err error) {
	if !obj.IsValid() {
		err = fmt.Errorf("%w: obj value [%v] is not valid", commonerrors.ErrUnsupported, obj)
		return
	}

	vp := reflect.New(obj.Type())
	if !vp.CanInterface() || !obj.CanInterface() {
		err = fmt.Errorf("%w: cannot get the value of the object pointer of type %T", commonerrors.ErrUnsupported, obj.Type())
		return
	}
	vp.Elem().Set(obj)
	val = vp.Interface()
	return
}
