/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package reflection

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	valueUtils "github.com/ARM-software/golang-utils/utils/value"
)

// IValueConverter transforms a value from its runtime source type into a new target type.
type IValueConverter func(reflect.Type, reflect.Type, any) (any, error)

// Converter resolves the runtime source type for value and applies the provided converter.
func Converter(converter IValueConverter, to reflect.Type, value any) (any, error) {
	if converter == nil {
		return value, nil
	}

	var from reflect.Type
	if value != nil {
		from = reflect.TypeOf(value)
	}

	return converter(from, to, value)
}

// NewValueTypeConverter adapts a reflection-style converter into a plain value
// converter.
func NewValueTypeConverter(converter IValueConverter) valueUtils.IValueConverter {
	return valueUtils.ValueConverterFunc(func(_ context.Context, value any) (any, error) {
		return Converter(converter, reflect.TypeOf((*any)(nil)).Elem(), value)
	})
}

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

// MapPropertyValue returns the value stored under key when rv is a map whose key
// type can be matched safely from the supplied string key.
//
// It supports maps keyed by strings, named string types, and interface key
// types whose stored key values either are strings or implement
// [fmt.Stringer].
//
// The returned value is the reflected map element. The `found` flag reports
// whether a matching key exists and the element can be accessed safely.
//
// Example:
//
//	value, found := MapPropertyValue(reflect.ValueOf(map[string]any{"name": "alice"}), "name")
func MapPropertyValue(rv reflect.Value, key string) (reflect.Value, bool) {
	if rv.Kind() != reflect.Map {
		return reflect.Value{}, false
	}
	lookupKey, ok := MapLookupKey(rv.Type().Key(), key)
	if ok {
		value := rv.MapIndex(lookupKey)
		if value.IsValid() {
			return value, true
		}
	}
	if rv.Type().Key().Kind() == reflect.Interface {
		for _, existingKey := range rv.MapKeys() {
			if mapKeyStringValue(existingKey) == key {
				value := rv.MapIndex(existingKey)
				if value.IsValid() {
					return value, true
				}
			}
		}
	}
	return reflect.Value{}, false
}

// StructPropertyValue returns the exported struct property named key when it can
// be accessed safely without panicking.
//
// The property name may be either the Go field name or the `json` tag name when
// one is defined.
//
// The returned value is the reflected field value. The `found` flag reports
// whether a matching exported field exists and can be interfaced safely.
//
// Example:
//
//	value, found := StructPropertyValue(reflect.ValueOf(cfg), "name")
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
//
// Example:
//
//	field, found := StructFieldByPropertyName(reflect.TypeOf(cfg), "name")
func StructFieldByPropertyName(rt reflect.Type, key string) (reflect.StructField, bool) {
	rt, ok := indirectStructType(rt)
	if !ok {
		return reflect.StructField{}, false
	}
	return structFieldByNameOrTag(rt, key, "json")
}

// StructTypeHasFieldTagValue reports whether the exported field identified by
// key on rt defines tagName with expectedValue.
//
// Concept: use this when you have a struct type and want to check a specific tag
// key/value pair on one of its exported fields, for example whether a field has
// `mask:"redact"`.
//
// The field may be identified by its Go field name or by a property name taken
// from common serialisation tags such as `json` and `yaml`, for example the
// property name `name` resolving the field tagged `json:"name,omitempty"`.
//
// Example:
//
//	ok := StructTypeHasFieldTagValue(reflect.TypeOf(cfg), "Password", "mask", "redact")
func StructTypeHasFieldTagValue(rt reflect.Type, key, tagName, expectedValue string) bool {
	rt, ok := indirectStructType(rt)
	if !ok {
		return false
	}
	field, found := structFieldByNameOrTag(rt, key, lookupStructTagNames(tagName)...)
	if !found {
		return false
	}
	tagValue, ok := structTagValue(field, tagName)
	return ok && tagValue == expectedValue
}

// StructTypeHasFieldTag reports whether the exported field identified by key on
// rt defines tagName.
//
// Concept: use this when you only care whether a tag key exists on a field,
// regardless of its value.
//
// The field may be identified by its Go field name or by a property name taken
// from common serialisation tags such as `json` and `yaml`, for example the
// property name `name` resolving the field tagged `json:"name,omitempty"`.
//
// Example:
//
//	ok := StructTypeHasFieldTag(reflect.TypeOf(cfg), "Password", "mask")
func StructTypeHasFieldTag(rt reflect.Type, key, tagName string) bool {
	rt, ok := indirectStructType(rt)
	if !ok {
		return false
	}
	field, found := structFieldByNameOrTag(rt, key, lookupStructTagNames(tagName)...)
	if !found {
		return false
	}
	return structTagDefined(field, tagName)
}

// StructTypeHasFieldPropertyName reports whether the exported field identified
// by key is exposed under propertyName through either its `json` or `yaml` tag.
//
// Concept: use this when you want to check the external serialised property name
// of a field without naming a specific tag key.
//
// The field may be identified by its Go field name or by an existing `json` or
// `yaml` property name.
//
// Example:
//
//	ok := StructTypeHasFieldPropertyName(reflect.TypeOf(cfg), "Password", "password")
func StructTypeHasFieldPropertyName(rt reflect.Type, key, propertyName string) bool {
	rt, ok := indirectStructType(rt)
	if !ok {
		return false
	}
	field, found := structFieldByNameOrTag(rt, key, "json", "yaml")
	if !found {
		return false
	}
	for _, tagName := range []string{"json", "yaml"} {
		tagValue, ok := structTagValue(field, tagName)
		if ok && tagValue == propertyName {
			return true
		}
	}
	return false
}

// StructPropertyHasTagValue reports whether the exported field identified by key
// on rv defines tagName with expectedValue.
//
// Concept: use this when you have a runtime value rather than a type and want
// to check a specific tag key/value pair. Collection values return true when any
// contained item resolves to a matching struct field.
//
// rv may be a struct value, a non-nil pointer/interface resolving to a struct
// value, or a slice/array/map whose items are checked until any matching field
// is found.
//
// Example:
//
//	ok := StructPropertyHasTagValue(reflect.ValueOf(cfg), "password", "mask", "redact")
func StructPropertyHasTagValue(rv reflect.Value, key, tagName, expectedValue string) bool {
	for rv.IsValid() {
		if rv.Kind() != reflect.Pointer && rv.Kind() != reflect.Interface {
			break
		}
		if rv.IsNil() {
			return false
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return false
	}
	switch rv.Kind() {
	case reflect.Struct:
		return StructTypeHasFieldTagValue(rv.Type(), key, tagName, expectedValue)
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			if StructPropertyHasTagValue(rv.Index(i), key, tagName, expectedValue) {
				return true
			}
		}
	case reflect.Map:
		for _, mapKey := range rv.MapKeys() {
			if StructPropertyHasTagValue(rv.MapIndex(mapKey), key, tagName, expectedValue) {
				return true
			}
		}
	}
	return false
}

// StructPropertyHasTag reports whether the exported field identified by key on
// rv defines tagName.
//
// Concept: use this when you have a runtime value rather than a type and want
// to check only whether a tag key exists. Collection values return true when any
// contained item resolves to a matching struct field.
//
// rv may be a struct value, a non-nil pointer/interface resolving to a struct
// value, or a slice/array/map whose items are checked until any matching field
// is found.
//
// Example:
//
//	ok := StructPropertyHasTag(reflect.ValueOf(cfg), "password", "mask")
func StructPropertyHasTag(rv reflect.Value, key, tagName string) bool {
	for rv.IsValid() {
		if rv.Kind() != reflect.Pointer && rv.Kind() != reflect.Interface {
			break
		}
		if rv.IsNil() {
			return false
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return false
	}
	switch rv.Kind() {
	case reflect.Struct:
		return StructTypeHasFieldTag(rv.Type(), key, tagName)
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			if StructPropertyHasTag(rv.Index(i), key, tagName) {
				return true
			}
		}
	case reflect.Map:
		for _, mapKey := range rv.MapKeys() {
			if StructPropertyHasTag(rv.MapIndex(mapKey), key, tagName) {
				return true
			}
		}
	}
	return false
}

// StructPropertyNames returns the exported property names exposed by rt using
// `json` tag names when present and Go field names otherwise.
//
// Example:
//
//	names := StructPropertyNames(reflect.TypeOf(cfg))
func StructPropertyNames(rt reflect.Type) []string {
	rt, ok := indirectStructType(rt)
	if !ok {
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
// when the map key type is directly compatible with strings.
//
// Interface-typed maps may still require a fallback scan of existing keys when
// the stored dynamic key type is a named string type or another string-like
// implementation such as [fmt.Stringer]. See [MapPropertyValue].
func MapLookupKey(keyType reflect.Type, key string) (reflect.Value, bool) {
	stringKey := reflect.ValueOf(key)
	if stringKey.Type().AssignableTo(keyType) {
		return stringKey, true
	}
	if stringKey.Type().ConvertibleTo(keyType) {
		return stringKey.Convert(keyType), true
	}
	if keyType.Kind() == reflect.Interface && stringKey.Type().Implements(keyType) {
		return stringKey, true
	}
	return reflect.Value{}, false
}

func mapKeyStringValue(key reflect.Value) string {
	if !key.IsValid() {
		return ""
	}
	if key.Kind() == reflect.Interface {
		if key.IsNil() {
			return ""
		}
		key = key.Elem()
	}
	if key.Kind() == reflect.String {
		return key.String()
	}
	if key.CanInterface() {
		if stringer, ok := key.Interface().(fmt.Stringer); ok && stringer != nil {
			return stringer.String()
		}
	}
	return ""
}

func jsonTagName(field reflect.StructField) (string, bool) {
	return structTagValue(field, "json")
}

func structTagValue(field reflect.StructField, tagName string) (string, bool) {
	if tagName == "" {
		return "", false
	}
	tag := field.Tag.Get(tagName)
	if tag == "" || tag == "-" {
		return "", false
	}
	parts := strings.Split(tag, ",")
	if len(parts) == 0 || parts[0] == "" {
		return "", false
	}
	return parts[0], true
}

func structTagDefined(field reflect.StructField, tagName string) bool {
	if tagName == "" {
		return false
	}
	_, ok := field.Tag.Lookup(tagName)
	return ok
}

func structFieldByNameOrTag(rt reflect.Type, key string, tagNames ...string) (reflect.StructField, bool) {
	for _, field := range reflect.VisibleFields(rt) {
		if !field.IsExported() {
			continue
		}
		if field.Name == key {
			return field, true
		}
		for _, tagName := range tagNames {
			if tag, ok := structTagValue(field, tagName); ok && tag == key {
				return field, true
			}
		}
	}
	return reflect.StructField{}, false
}

func indirectStructType(rt reflect.Type) (reflect.Type, bool) {
	for rt != nil && (rt.Kind() == reflect.Pointer || rt.Kind() == reflect.Interface) {
		rt = rt.Elem()
	}
	if rt == nil || rt.Kind() != reflect.Struct {
		return nil, false
	}
	return rt, true
}

func lookupStructTagNames(tagNames ...string) []string {
	result := []string{"json", "yaml"}
	for _, tagName := range tagNames {
		if tagName == "" || stringSliceContains(result, tagName) {
			continue
		}
		result = append(result, tagName)
	}
	return result
}

func stringSliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
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
