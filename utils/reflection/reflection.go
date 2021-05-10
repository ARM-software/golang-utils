package reflection

import (
	"reflect"
	"unsafe"
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
	ValueStructure := reflect.ValueOf(structure)

	Field := ValueStructure.Elem().FieldByName(FieldName)
	if !Field.IsValid() {
		return "", false
	}

	if Field.Type().Kind() == reflect.Ptr {
		return Field.Elem().Interface(), true
	} else {
		return Field.Interface(), true
	}
}
