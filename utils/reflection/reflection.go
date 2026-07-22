/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package reflection

import (
	"reflect"
	"strings"
	"unsafe"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	valueUtils "github.com/ARM-software/golang-utils/utils/value"
)

func GetUnexportedStructureField(structure any, fieldName string) any {
	return GetStructureField(fetchStructureField(structure, fieldName))
}

func GetStructureField(field reflect.Value) any {
	if !field.IsValid() {
		return nil
	}
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface() //nolint:gosec // this conversion is between types recommended by Go https://cs.opensource.google/go/go/+/master:src/reflect/value.go;l=2445
}
func SetUnexportedStructureField(structure any, fieldName string, value any) {
	SetStructureField(fetchStructureField(structure, fieldName), value)
}
func SetStructureField(field reflect.Value, value any) {
	if !field.IsValid() {
		return
	}
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(reflect.ValueOf(value)) //nolint:gosec // this conversion is between types recommended by Go https://cs.opensource.google/go/go/+/master:src/reflect/value.go;l=2445
}

func fetchStructureField(structure any, fieldName string) reflect.Value {
	return reflect.ValueOf(structure).Elem().FieldByName(fieldName)
}

// GetStructField checks if the given structure has a given field. The structure should be passed by reference.
// It returns an interface and a boolean, the field's content and a boolean denoting whether or not the field exists.
// If the boolean is false then there is no such field on the structure.
// If the boolean is true but the interface stores "" then the field exists but is not set.
// If the boolean is true and the interface is not empty, the field exists and is set.
func GetStructField(structure any, fieldName string) (any, bool) {
	Field := fetchStructureField(structure, fieldName)
	if !Field.IsValid() {
		return "", false
	}

	if Field.Type().Kind() == reflect.Pointer {
		if Field.IsNil() {
			return "", true
		}
		return Field.Elem().Interface(), true
	} else {
		return Field.Interface(), true
	}
}

// SetStructField attempts to set a field of a structure to the given value
// It returns nil or an error, in case the field doesn't exist on the structure
// or the value and the field have different types
func SetStructField(structure any, fieldName string, value any) error {
	ValueStructure := reflect.ValueOf(structure)
	Field := ValueStructure.Elem().FieldByName(fieldName)
	// Test field exists on structure
	if !Field.IsValid() {
		return commonerrors.Newf(commonerrors.ErrInvalid, "error with field [%v]", fieldName)
	}

	// test field is settable
	if !Field.CanSet() {
		return commonerrors.Newf(commonerrors.ErrUnsupported, "error with unsettable field [%v]", fieldName)
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
	if valueKind == reflect.Pointer {
		valueUnderlyingType = valueUnderlyingType.Elem()
	}
	fieldUnderlyingType := Field.Type()
	if fieldKind == reflect.Pointer {
		fieldUnderlyingType = fieldUnderlyingType.Elem()
	}

	// Check that the underlying types are the same (e.g. no int and string)
	if fieldUnderlyingType != valueUnderlyingType {
		return commonerrors.Newf(commonerrors.ErrConflict, "conflicting types, field [%v] and value [%v]", fieldKind, valueKind)
	}

	if fieldKind == reflect.Pointer {
		if valueKind != reflect.Pointer { // value not ptr, field ptr
			if Field.IsNil() {
				pointerToValue := reflect.New(valueReflectValueWrapper.Type())
				pointerToValue.Elem().Set(valueReflectValueWrapper)
				Field.Set(pointerToValue)
			} else {
				Field.Elem().Set(valueReflectValueWrapper)
			}
		}
	} else { // field not ptr, val ptr
		if valueKind == reflect.Pointer {
			Field.Set(valueReflectValueWrapper.Elem())
		}
	}
	// This means the field was updated without errors
	return nil
}

// MapPropertyValue returns the value stored under key when rv is a map whose
// key type can safely represent the supplied string key.
func MapPropertyValue(rv reflect.Value, key string) (reflect.Value, bool) {
	if rv.Kind() != reflect.Map {
		return reflect.Value{}, false
	}
	lookupKey, ok := MapLookupKey(rv.Type().Key(), key)
	if !ok {
		return reflect.Value{}, false
	}
	value := rv.MapIndex(lookupKey)
	if !value.IsValid() {
		return reflect.Value{}, false
	}
	return value, true
}

// StructPropertyValue returns the exported struct field named key when it can be
// accessed safely without panicking.
func StructPropertyValue(rv reflect.Value, key string) (reflect.Value, bool) {
	if rv.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}
	field, found := StructFieldByPropertyName(rv.Type(), key)
	if !found || !field.IsExported() {
		return reflect.Value{}, false
	}
	fieldValue, err := rv.FieldByIndexErr(field.Index)
	if err != nil || !fieldValue.IsValid() || !fieldValue.CanInterface() {
		return reflect.Value{}, false
	}
	return fieldValue, true
}

// StructFieldByPropertyName resolves key to an exported struct field using the
// Go field name first and then the `json` tag name when present.
func StructFieldByPropertyName(rt reflect.Type, key string) (reflect.StructField, bool) {
	for _, field := range reflect.VisibleFields(rt) {
		if !field.IsExported() {
			continue
		}
		if field.Name == key {
			return field, true
		}
		if tag, ok := jsonTagName(field); ok && tag == key {
			return field, true
		}
	}
	return reflect.StructField{}, false
}

// StructPropertyNames returns the exported property names exposed by rt using
// `json` tag names when present and Go field names otherwise.
func StructPropertyNames(rt reflect.Type) []string {
	if rt.Kind() != reflect.Struct {
		return nil
	}
	result := make([]string, 0)
	for _, field := range reflect.VisibleFields(rt) {
		if !field.IsExported() {
			continue
		}
		if tag, ok := jsonTagName(field); ok {
			result = append(result, tag)
			continue
		}
		result = append(result, field.Name)
	}
	return result
}

// MapLookupKey converts a string property name into a reflected map key value
// when the map key type is compatible with strings.
func MapLookupKey(keyType reflect.Type, key string) (reflect.Value, bool) {
	stringKey := reflect.ValueOf(key)
	if stringKey.Type().AssignableTo(keyType) {
		return stringKey, true
	}
	if stringKey.Type().ConvertibleTo(keyType) {
		return stringKey.Convert(keyType), true
	}
	if keyType.Kind() == reflect.Interface {
		return stringKey, true
	}
	return reflect.Value{}, false
}

func jsonTagName(field reflect.StructField) (string, bool) {
	tag := field.Tag.Get("json")
	if tag == "" || tag == "-" {
		return "", false
	}
	parts := strings.Split(tag, ",")
	if len(parts) == 0 || parts[0] == "" {
		return "", false
	}
	return parts[0], true
}

// InheritsFrom uses reflection to find if a struct "inherits" from a certain type.
// In other words it checks whether the struct embeds a struct of that type.
func InheritsFrom(object any, parentType reflect.Type) bool {
	if parentType == nil {
		return object == nil
	}
	r := reflect.ValueOf(object)
	t := r.Type()

	if t == parentType {
		return true
	}

	if r.Kind() == reflect.Pointer {
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
	switch kind {
	case reflect.Pointer:
		pointerType = parentType
		structType = parentType.Elem()
	case reflect.Interface:
		pointerType = parentType
	case reflect.Struct:
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
	return valueUtils.IsEmpty(value)
}

// IsNotEmpty checks whether a value is not empty. See IsEmpty for more details about what is considered empty.
func IsNotEmpty(value any) bool {
	return valueUtils.IsNotEmpty(value)
}

// IsNilInterface checks whether an interface value is nil even when it has been
// passed around as `any`.
func IsNilInterface(i any) bool {
	return valueUtils.IsNilInterface(i)
}

// ToStructPtr returns an instance of the pointer (interface) to the object obj.
func ToStructPtr(obj reflect.Value) (val any, err error) {
	if !obj.IsValid() {
		err = commonerrors.Newf(commonerrors.ErrUnsupported, "obj value [%v] is not valid", obj)
		return
	}

	vp := reflect.New(obj.Type())
	if !vp.CanInterface() || !obj.CanInterface() {
		err = commonerrors.Newf(commonerrors.ErrUnsupported, "cannot get the value of the object pointer of type %T", obj.Type())
		return
	}
	vp.Elem().Set(obj)
	val = vp.Interface()
	return
}
