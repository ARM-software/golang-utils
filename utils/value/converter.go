package value

import (
	"context"
	"encoding"
	"fmt"
	"reflect"
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
var IdentityConverter IValueConverter = ValueConverterFunc(func(ctx context.Context, value any) (any, error) {
	_ = ctx
	return value, nil
})

// StringConverter converts values to strings.
//
// It prefers [fmt.Stringer] when available, then [encoding.TextMarshaler], and
// otherwise falls back to [fmt.Sprint].
//
// Nil pointer and interface receivers do not have custom methods invoked; they
// are rendered as `<nil>` instead. When [encoding.TextMarshaler] returns an
// error, that error is propagated to the caller.
var StringConverter IValueConverter = ValueConverterFunc(func(ctx context.Context, value any) (any, error) {
	_ = ctx
	if stringer, ok := value.(fmt.Stringer); ok && !isNilMethodReceiver(value) {
		return stringer.String(), nil
	}
	if textMarshaler, ok := value.(encoding.TextMarshaler); ok && !isNilMethodReceiver(value) {
		text, err := textMarshaler.MarshalText()
		if err != nil {
			return nil, err
		}
		return string(text), nil
	}
	if isNilMethodReceiver(value) {
		return fmt.Sprint(nil), nil
	}
	return fmt.Sprint(value), nil
})

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
