package maps

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
	valueUtils "github.com/ARM-software/golang-utils/utils/value"
)

// Flatten takes a structure and turns into a flat maps[string]string.
//
// Within the "thing" parameter, only primitive values are allowed. Structs are
// not supported. Therefore, it can only be slices, maps, primitives, and
// any combination of those together.
//
// See the tests for examples of what inputs are turned into.
func Flatten(thing map[string]any) (result Map, err error) {
	result = make(map[string]string)

	for k, raw := range thing {
		subErr := flatten(result, k, reflect.ValueOf(raw))
		if subErr != nil {
			err = subErr
			return
		}
	}

	return
}

func flatten(result map[string]string, prefix string, v reflect.Value) (err error) {
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Bool:
		return flattenStringValue(result, prefix, v.Interface())
	case reflect.Int64:
		switch v.Type() {
		case reflect.TypeOf(time.Duration(5)):
			return flattenStringValue(result, prefix, v.Interface())
		default:
			return flattenStringValue(result, prefix, v.Interface())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return flattenStringValue(result, prefix, v.Interface())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return flattenStringValue(result, prefix, v.Interface())
	case reflect.Float64, reflect.Float32:
		return flattenStringValue(result, prefix, v.Interface())
	case reflect.Map:
		err = flattenMap(result, prefix, v)
		if err != nil {
			return err
		}
	case reflect.Slice, reflect.Array:
		err = flattenSlice(result, prefix, v)
		if err != nil {
			return err
		}
	case reflect.Interface:
	case reflect.Struct:
		switch v.Type().String() {
		case "time.Time":
			result[prefix] = v.Interface().(time.Time).Format(time.RFC3339Nano)
			return
		default:
			err = flattenStruct(result, prefix, v)
			if err != nil {
				return
			}
		}
	case reflect.String:
		return flattenStringValue(result, prefix, v.Interface())
	case reflect.Invalid:
		result[prefix] = ""
	case reflect.Pointer:
		if v.IsNil() {
			return
		}
		err = flatten(result, prefix, v.Elem())
	default:
		if v.IsZero() {
			result[prefix] = ""
		} else {
			err = commonerrors.Newf(commonerrors.ErrUnknown, "unknown value '%v'", v)
		}
	}
	return
}

func flattenStringValue(result map[string]string, prefix string, raw any) error {
	converted, err := valueUtils.StringConverter.ConvertValue(context.Background(), raw)
	if err != nil {
		return err
	}
	result[prefix] = converted.(string)
	return nil
}

func flattenMap(result Map, prefix string, v reflect.Value) (err error) {
	for _, k := range v.MapKeys() {
		if k.Kind() == reflect.Interface {
			k = k.Elem()
		}

		if k.Kind() != reflect.String {
			err = commonerrors.Newf(commonerrors.ErrInvalid, "%s: maps key is not string: %s", prefix, k)
			return

		}

		keyString := k.String()
		subPrefix := ""
		if reflection.IsEmpty(keyString) {
			subPrefix = prefix
		} else {
			subPrefix = fmt.Sprintf("%s%s%s", prefix, separator, k.String())
		}
		subErr := flatten(result, subPrefix, v.MapIndex(k))
		if subErr != nil {
			err = subErr
			return
		}
	}
	return
}

func flattenSlice(result Map, prefix string, v reflect.Value) (err error) {
	prefix += separator

	for i := 0; i < v.Len(); i++ {
		subErr := flatten(result, fmt.Sprintf("%s%d", prefix, i), v.Index(i))
		if subErr != nil {
			err = subErr
			return
		}
	}
	return
}

func flattenStruct(result Map, prefix string, v reflect.Value) (err error) {
	prefix += separator
	ty := v.Type()
	for i := 0; i < ty.NumField(); i++ {
		subErr := flatten(result, fmt.Sprintf("%s%s", prefix, ty.Field(i).Name), v.Field(i))
		if subErr != nil {
			err = subErr
			return
		}
	}
	return
}
