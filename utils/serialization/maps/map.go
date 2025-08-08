package maps

import (
	"reflect"
	"time"

	"github.com/go-viper/mapstructure/v2"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/maps"
)

// ToMapFromPointer is like ToMap but deals with a pointer.
func ToMapFromPointer[T any](o T) (m map[string]string, err error) {
	if reflect.TypeOf(o) == nil {
		err = commonerrors.UndefinedVariable("pointer")
		return
	}
	if reflect.TypeOf(o).Kind() != reflect.Ptr {
		err = commonerrors.Newf(commonerrors.ErrInvalid, "expected a pointer and got %T", o)
		return
	}
	if reflect.ValueOf(o).IsNil() {
		err = commonerrors.UndefinedVariable("pointer")
		return
	}
	mapAny := map[string]any{}
	err = mapstructureDecoder(o, &mapAny)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrMarshalling, err, "failed to serialise object")
		return
	}
	m, err = maps.Flatten(mapAny)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrMarshalling, err, "failed to flatten map")
	}
	return
}

// ToMap converts a struct to a flat map using (mapstructure)[https://github.com/go-viper/mapstructure]
func ToMap[T any](o *T) (m map[string]string, err error) {
	if o == nil {
		err = commonerrors.UndefinedVariable("object")
		return
	}
	m, err = ToMapFromPointer[*T](o)
	return
}

// FromMapToPointer is like FromMap but deals with a pointer.
func FromMapToPointer[T any](m map[string]string, o T) (err error) {
	if reflect.TypeOf(o) == nil {
		err = commonerrors.UndefinedVariable("pointer")
		return
	}
	if reflect.TypeOf(o).Kind() != reflect.Ptr {
		err = commonerrors.Newf(commonerrors.ErrInvalid, "expected a pointer and got %T", o)
		return
	}
	if reflect.ValueOf(o).IsNil() {
		err = commonerrors.UndefinedVariable("pointer")
		return
	}
	if len(m) == 0 {
		return
	}
	expandedMap, err := maps.Expand(m)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrMarshalling, err, "failed to expand the map")
		return
	}

	err = mapstructureDecoder(expandedMap, o)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrMarshalling, err, "failed to deserialise upload request")
	}
	return
}

// FromMap deserialises a flatten map into a struct using (mapstructure)[https://github.com/go-viper/mapstructure]
func FromMap[T any](m map[string]string, o *T) (err error) {
	if o == nil {
		err = commonerrors.UndefinedVariable("object")
		return
	}
	err = FromMapToPointer[*T](m, o)
	return
}

func timeHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		switch {
		case t == reflect.TypeOf(time.Time{}):
			return toTime(f, t, data)
		case f == reflect.TypeOf(time.Time{}) || f == reflect.TypeOf(&time.Time{}):
			return fromTime(f, t, data)
		default:
			return data, nil
		}
	}
}

func fromTime(f, t reflect.Type, data any) (any, error) {
	switch f {
	case reflect.TypeOf(time.Time{}):
		subtime := data.(time.Time)
		value := subtime.Format(time.RFC3339Nano)
		return convertTo(value, data, t)
	case reflect.TypeOf(&time.Time{}):
		subtime := data.(*time.Time)
		if subtime == nil {
			return nil, nil
		}
		value := subtime.Format(time.RFC3339Nano)
		return convertTo(value, data, t)
	default:
		return data, nil
	}
}

func convertTo(value string, rawValue any, t reflect.Type) (any, error) {
	switch t.Kind() {
	case reflect.String:
		return value, nil
	case reflect.Map:
		return map[string]any{"": value}, nil
	case reflect.Slice:
		return []string{value}, nil
	default:
		return rawValue, nil
	}
}

func toTime(f reflect.Type, t reflect.Type, data any) (any, error) {
	if t != reflect.TypeOf(time.Time{}) {
		return data, nil
	}

	switch f.Kind() {
	case reflect.String:
		return time.Parse(time.RFC3339Nano, data.(string))
	case reflect.Float64:
		return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
	case reflect.Int64:
		return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
	default:
		return data, nil
	}
}

func mapstructureDecoder(input, result any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			timeHookFunc(), mapstructure.StringToTimeDurationHookFunc(), mapstructure.StringToURLHookFunc(), mapstructure.StringToIPHookFunc()),
		Result: result,
	})
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}
