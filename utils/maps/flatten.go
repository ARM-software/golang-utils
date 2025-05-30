package maps

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/safecast"
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
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			result[prefix] = "true"
		} else {
			result[prefix] = "false"
		}
	case reflect.Int64:
		switch v.Type() {
		case reflect.TypeOf(time.Duration(5)):
			result[prefix] = v.Interface().(time.Duration).String()
		default:
			result[prefix] = strconv.FormatInt(safecast.ToInt64(v.Int()), 10)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		result[prefix] = strconv.FormatInt(safecast.ToInt64(v.Int()), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		result[prefix] = strconv.FormatUint(safecast.ToUint64(v.Uint()), 10)
	case reflect.Float64, reflect.Float32:
		result[prefix] = strconv.FormatFloat(v.Float(), 'g', -1, 64)
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
		result[prefix] = v.String()
	case reflect.Invalid:
		result[prefix] = ""
	default:
		if v.IsZero() {
			result[prefix] = ""
		} else {
			err = commonerrors.Newf(commonerrors.ErrUnknown, "unknown %v", v)
		}
	}
	return
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
