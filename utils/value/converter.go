package value

import (
	"context"
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

// IValueConverter converts a value before it is consumed by a caller.
//
// Implementations may use ctx to observe cancellation, access request-scoped
// metadata, or carry other conversion-time state. The returned value replaces
// the original input. A non-nil error reports that conversion failed and should
// usually stop the caller's processing.
type IValueConverter interface {
	ConvertValue(ctx context.Context, value any) (any, error)
}

// ValueConverterFunc adapts a function into [IValueConverter].
//
// A nil function behaves like [IdentityConverter]: it returns the input value
// unchanged and reports no error.
type ValueConverterFunc func(ctx context.Context, value any) (any, error)

// ConvertValue applies f to value.
func (f ValueConverterFunc) ConvertValue(ctx context.Context, value any) (any, error) {
	if f == nil {
		return value, nil
	}
	return f(ctx, value)
}

// IdentityConverter returns values unchanged.
//
// It ignores ctx and always succeeds.
var IdentityConverter IValueConverter = ValueConverterFunc(func(_ context.Context, value any) (any, error) {
	return value, nil
})

// StringConverter converts values to strings.
//
// Nil interface and pointer values are rendered as `<nil>`. For non-nil
// values, it prefers [fmt.Stringer], then [encoding.TextMarshaler]. If neither
// applies, scalar values use the same formatting as flattening, structs use `%+v`,
// pointers use `<pointer> -> <converted pointed value>`, and everything else
// falls back to [fmt.Sprint]. Errors returned by
// [encoding.TextMarshaler.MarshalText] are propagated to the caller.
var StringConverter IValueConverter = ValueConverterFunc(func(_ context.Context, value any) (any, error) {
	return stringifyValue(value)
})

func stringifyValue(value any) (string, error) {
	if isNilMethodReceiver(value) {
		return fmt.Sprint(nil), nil
	}

	rv := reflect.ValueOf(value)

	if stringer, ok := value.(fmt.Stringer); ok {
		return stringer.String(), nil
	}
	if textMarshaler, ok := value.(encoding.TextMarshaler); ok {
		text, err := textMarshaler.MarshalText()
		if err != nil {
			return "", commonerrors.WrapError(commonerrors.ErrMarshalling, err, "failed to marshal value as text")
		}
		return string(text), nil
	}

	switch rv.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool()), nil
	case reflect.Int64:
		if rv.Type() == reflect.TypeOf(time.Duration(0)) {
			return value.(time.Duration).String(), nil
		}
		return strconv.FormatInt(rv.Int(), 10), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return strconv.FormatInt(rv.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64), nil
	case reflect.String:
		return rv.String(), nil
	}

	if rv.Kind() == reflect.Struct {
		return fmt.Sprintf("%+v", value), nil
	}
	if rv.Kind() == reflect.Pointer {
		converted, err := stringifyValue(rv.Elem().Interface())
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%p -> %v", value, converted), nil
	}

	return fmt.Sprint(value), nil
}

func isNilMethodReceiver(value any) bool {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return true
	}
	switch rv.Kind() {
	case reflect.Interface, reflect.Pointer:
		return rv.IsNil()
	default:
		return false
	}
}
