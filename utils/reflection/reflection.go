package reflection

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/ARMmbed/golang-utils/utils/commonerrors"
)

func GetUnexportedStructureField(structure interface{}, fieldName string) interface{} {
	return GetStructureField(fetchStructureField(structure, fieldName))
}

func GetStructureField(field reflect.Value) interface{} {
	if !field.IsValid() {
		return nil
	}
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
}
func SetUnexportedStructureField(structure interface{}, fieldName string, value interface{}) {
	SetStructureField(fetchStructureField(structure, fieldName), value)
}
func SetStructureField(field reflect.Value, value interface{}) {
	if !field.IsValid() {
		return
	}
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().
		Set(reflect.ValueOf(value))
}

func fetchStructureField(structure interface{}, fieldName string) reflect.Value {
	return reflect.ValueOf(structure).Elem().FieldByName(fieldName)
}

// Check if the given structure has a given field. The structure should be passed by reference.
// It returns an interface and a boolean, the field's content and a boolean denoting whether or not the field exists.
// If the boolean is false then there is no such field on the structure.
// If the boolean is true but the interface stores "" then the field exists but is not set.
// If the boolean is true and the interface is not emtpy, the field exists and is set.
func GetStructField(structure interface{}, FieldName string) (interface{}, bool) {
	Field := fetchStructureField(structure, FieldName)
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

// Attempts to set a field of a structure to the given vaule
// It returns nil or an error, in case the field doesn't exist on the structure
// or the value and the field have different types
func SetStructField(structure interface{}, FieldName string, value interface{}) error {
	ValueStructure := reflect.ValueOf(structure)
	Field := ValueStructure.Elem().FieldByName(FieldName)
	// Test field exists on structure
	if !Field.IsValid() {
		return fmt.Errorf("error with field [%v]: %w", FieldName, commonerrors.ErrInvalid)
	}

	//test field is settable
	if !Field.CanSet() {
		return fmt.Errorf("error with unsettable field [%v]: %w", FieldName, commonerrors.ErrUnsupported)
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
		if valueKind != reflect.Ptr { //value not ptr, field ptr
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

// Use reflection to find if a struct "inherits" from a certain type.
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
	if kind == reflect.Ptr {
		pointerType = parentType
		structType = parentType.Elem()
	} else if kind == reflect.Interface {
		pointerType = parentType
	} else if kind == reflect.Struct {
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
