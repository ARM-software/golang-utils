package validation

import (
	"reflect"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	valueutils "github.com/ARM-software/golang-utils/utils/value"
)

var (
	// ErrZeroRequired reports that the value must be the zero value for its type.
	ErrZeroRequired = validation.NewError("validation_zero_required", "must be zero")
	// ErrNotZeroRequired reports that the value must not be the zero value for its type.
	ErrNotZeroRequired = validation.NewError("validation_not_zero_required", "must not be zero")
	// ErrTrueRequired reports that the value must be true.
	ErrTrueRequired = validation.NewError("validation_true_required", "must be true")
	// ErrFalseRequired reports that the value must be false.
	ErrFalseRequired = validation.NewError("validation_false_required", "must be false")
	// ErrEmptyRequired reports that the value must be empty.
	ErrEmptyRequired = validation.NewError("validation_empty_required", "must be empty")
	// ErrNotEmptyRequired reports that the value must not be empty.
	ErrNotEmptyRequired = validation.NewError("validation_not_empty_required", "must not be empty")
	// ErrNilRequired reports that the value must be nil.
	ErrNilRequired = validation.NewError("validation_nil_required", "must be nil")
	// ErrNotNilRequired reports that the value must not be nil.
	ErrNotNilRequired = validation.NewError("validation_not_nil_required", "must not be nil")
)

// IsZero validates that a value is the zero value for its type.
var IsZero = validation.By(func(value any) error {
	if !isZeroValue(value) {
		return ErrZeroRequired
	}
	return nil
})

// IsNotZero validates that a value is not the zero value for its type.
var IsNotZero = validation.By(func(value any) error {
	if isZeroValue(value) {
		return ErrNotZeroRequired
	}
	return nil
})

// IsTrue validates that a value is true.
var IsTrue = validation.By(func(value any) error {
	if valueBool, ok := value.(bool); ok && valueBool {
		return nil
	}
	return ErrTrueRequired
})

// IsFalse validates that a value is false.
var IsFalse = validation.By(func(value any) error {
	if valueBool, ok := value.(bool); ok && !valueBool {
		return nil
	}
	return ErrFalseRequired
})

// IsEmpty validates that a value is empty according to golang-utils empty semantics.
var IsEmpty = validation.By(func(value any) error {
	if !isEmptyValue(value) {
		return ErrEmptyRequired
	}
	return nil
})

// IsNotEmpty validates that a value is not empty according to golang-utils empty semantics.
var IsNotEmpty = validation.By(func(value any) error {
	if isEmptyValue(value) {
		return ErrNotEmptyRequired
	}
	return nil
})

// NotEmpty validates that a value is not empty according to the repository's
// reflection-based emptiness semantics.
//
// Example: `NotEmpty()` rejects `"   "`.
func NotEmpty() validation.Rule {
	return IsNotEmpty
}

// IsNil validates that a value is nil.
var IsNil = validation.Nil.ErrorObject(ErrNilRequired)

// IsNotNil validates that a value is not nil.
var IsNotNil = validation.NotNil.ErrorObject(ErrNotNilRequired)

// IsNotNilAndNotEmpty validates that a value is both non-nil and non-empty.
var IsNotNilAndNotEmpty = validation.By(func(value any) error {
	if err := IsNotNil.Validate(value); err != nil {
		return err
	}
	v, isNil := validation.Indirect(value)
	if isNil || isEmptyValue(v) {
		return ErrNotEmptyRequired
	}
	return nil
})

// IsNilOrNotEmpty validates that a value is either nil or not empty.
var IsNilOrNotEmpty = validation.By(func(value any) error {
	if isNilValue(value) {
		return nil
	}
	if isEmptyValue(value) {
		return validation.ErrNilOrNotEmpty
	}
	return nil
})

func isNilValue(value any) bool {
	_, isNil := validation.Indirect(value)
	return isNil
}

func isEmptyValue(value any) bool {
	return valueutils.IsEmpty(value)
}

func isNotEmptyValue(value any) bool {
	return !isEmptyValue(value)
}

// IsRequiredLegacy preserves the pre-v4.4 ozzo Required semantics where zero-valued
// structs other than time.Time{} are still considered present.
var IsRequiredLegacy = validation.By(func(value any) error {
	value, isNil := validation.Indirect(value)
	if isNil || isEmptyLegacy(value) {
		return validation.ErrRequired
	}
	return nil
})

// IsRequired preserves the legacy struct-friendly required semantics for callers
// that still expect the pre-v4.4 ozzo behaviour.
var IsRequired = IsRequiredLegacy

func isZeroValue(value any) bool {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return true
	}
	if rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return true
		}
		return isZeroValue(rv.Elem().Interface())
	}
	return rv.IsZero()
}

func isEmptyLegacy(value any) bool {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String, reflect.Array, reflect.Map, reflect.Slice:
		return rv.Len() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Invalid:
		return true
	case reflect.Interface, reflect.Pointer:
		if rv.IsNil() {
			return true
		}
		return isEmptyLegacy(rv.Elem().Interface())
	case reflect.Struct:
		if t, ok := value.(time.Time); ok && t.IsZero() {
			return true
		}
	}
	return false
}
