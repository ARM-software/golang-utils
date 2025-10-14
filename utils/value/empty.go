package value

import (
	"reflect"
	"strings"
)

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
