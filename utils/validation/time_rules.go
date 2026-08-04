package validation

// time_rules.go contains validation helpers related to durations and timestamps,
// including both parsing helpers (`IsDuration`, `IsRFC3339Timestamp`) and threshold/
// equality rules for time-oriented values.

import (
	"reflect"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

const (
	errInvalidDurationDescription                = "must be a valid duration"
	errDurationBelowMinimumDescription           = "must be a valid duration greater than or equal to the minimum"
	errDurationBelowExclusiveMinimumDescription  = "must be a valid duration greater than the minimum"
	errDurationAboveMaximumDescription           = "must be a valid duration less than or equal to the maximum"
	errDurationAboveExclusiveMaximumDescription  = "must be a valid duration less than the maximum"
	errUnexpectedDurationDescription             = "must be the expected duration"
	errInvalidTimestampDescription               = "must be a valid timestamp"
	errTimestampBelowMinimumDescription          = "must be a valid timestamp greater than or equal to the minimum"
	errTimestampBelowExclusiveMinimumDescription = "must be a valid timestamp greater than the minimum"
	errTimestampAboveMaximumDescription          = "must be a valid timestamp less than or equal to the maximum"
	errTimestampAboveExclusiveMaximumDescription = "must be a valid timestamp less than the maximum"
	errUnexpectedTimestampDescription            = "must be the expected timestamp"
)

var (
	errInvalidDuration                = commonerrors.New(commonerrors.ErrInvalid, errInvalidDurationDescription)
	errDurationBelowMinimum           = commonerrors.New(commonerrors.ErrInvalid, errDurationBelowMinimumDescription)
	errDurationBelowExclusiveMinimum  = commonerrors.New(commonerrors.ErrInvalid, errDurationBelowExclusiveMinimumDescription)
	errDurationAboveMaximum           = commonerrors.New(commonerrors.ErrInvalid, errDurationAboveMaximumDescription)
	errDurationAboveExclusiveMaximum  = commonerrors.New(commonerrors.ErrInvalid, errDurationAboveExclusiveMaximumDescription)
	errUnexpectedDuration             = commonerrors.New(commonerrors.ErrInvalid, errUnexpectedDurationDescription)
	errInvalidTimestamp               = commonerrors.New(commonerrors.ErrInvalid, errInvalidTimestampDescription)
	errTimestampBelowMinimum          = commonerrors.New(commonerrors.ErrInvalid, errTimestampBelowMinimumDescription)
	errTimestampBelowExclusiveMinimum = commonerrors.New(commonerrors.ErrInvalid, errTimestampBelowExclusiveMinimumDescription)
	errTimestampAboveMaximum          = commonerrors.New(commonerrors.ErrInvalid, errTimestampAboveMaximumDescription)
	errTimestampAboveExclusiveMaximum = commonerrors.New(commonerrors.ErrInvalid, errTimestampAboveExclusiveMaximumDescription)
	errUnexpectedTimestamp            = commonerrors.New(commonerrors.ErrInvalid, errUnexpectedTimestampDescription)
	timestampNotEmptyRule             = TimestampExclusiveMinimum(time.Time{})
)

// DurationMinimum validates that a duration value is greater than or equal to min.
func DurationMinimum(min time.Duration) validation.Rule {
	return validation.By(func(value any) error {
		d, ok := durationValue(value)
		if !ok {
			return errInvalidDuration
		}
		if d < min {
			return errDurationBelowMinimum
		}
		return nil
	})
}

// DurationExclusiveMinimum validates that a duration value is strictly greater than min.
func DurationExclusiveMinimum(min time.Duration) validation.Rule {
	return validation.By(func(value any) error {
		d, ok := durationValue(value)
		if !ok {
			return errInvalidDuration
		}
		if d <= min {
			return errDurationBelowExclusiveMinimum
		}
		return nil
	})
}

// DurationMaximum validates that a duration value is less than or equal to max.
func DurationMaximum(max time.Duration) validation.Rule {
	return validation.By(func(value any) error {
		d, ok := durationValue(value)
		if !ok {
			return errInvalidDuration
		}
		if d > max {
			return errDurationAboveMaximum
		}
		return nil
	})
}

// DurationExclusiveMaximum validates that a duration value is strictly less than max.
func DurationExclusiveMaximum(max time.Duration) validation.Rule {
	return validation.By(func(value any) error {
		d, ok := durationValue(value)
		if !ok {
			return errInvalidDuration
		}
		if d >= max {
			return errDurationAboveExclusiveMaximum
		}
		return nil
	})
}

// DurationConst validates that a duration value is exactly equal to expected.
func DurationConst(expected time.Duration) validation.Rule {
	return validation.By(func(value any) error {
		d, ok := durationValue(value)
		if !ok {
			return errInvalidDuration
		}
		if d != expected {
			return errUnexpectedDuration
		}
		return nil
	})
}

// TimestampMinimum validates that a timestamp value is greater than or equal to min.
func TimestampMinimum(min time.Time) validation.Rule {
	return validation.By(func(value any) error {
		ts, ok := timestampValue(value)
		if !ok {
			return errInvalidTimestamp
		}
		if ts.Before(min) {
			return errTimestampBelowMinimum
		}
		return nil
	})
}

// TimestampExclusiveMinimum validates that a timestamp value is strictly after min.
func TimestampExclusiveMinimum(min time.Time) validation.Rule {
	return validation.By(func(value any) error {
		ts, ok := timestampValue(value)
		if !ok {
			return errInvalidTimestamp
		}
		if !ts.After(min) {
			return errTimestampBelowExclusiveMinimum
		}
		return nil
	})
}

// TimestampMaximum validates that a timestamp value is less than or equal to max.
func TimestampMaximum(max time.Time) validation.Rule {
	return validation.By(func(value any) error {
		ts, ok := timestampValue(value)
		if !ok {
			return errInvalidTimestamp
		}
		if ts.After(max) {
			return errTimestampAboveMaximum
		}
		return nil
	})
}

// TimestampExclusiveMaximum validates that a timestamp value is strictly before max.
func TimestampExclusiveMaximum(max time.Time) validation.Rule {
	return validation.By(func(value any) error {
		ts, ok := timestampValue(value)
		if !ok {
			return errInvalidTimestamp
		}
		if !ts.Before(max) {
			return errTimestampAboveExclusiveMaximum
		}
		return nil
	})
}

// TimestampConst validates that a timestamp value is exactly equal to expected.
func TimestampConst(expected time.Time) validation.Rule {
	return validation.By(func(value any) error {
		ts, ok := timestampValue(value)
		if !ok {
			return errInvalidTimestamp
		}
		if !ts.Equal(expected) {
			return errUnexpectedTimestamp
		}
		return nil
	})
}

// IsDuration validates whether a value is a valid Go duration.
//
// It accepts duration strings and byte slices, as well as `time.Duration`
// values and their pointer forms.
var IsDuration = validation.By(func(value any) error {
	if _, ok := durationValue(value); ok {
		return nil
	}
	return errInvalidDuration
})

// NilTimestampOrNotEmpty validates that a timestamp value is either nil or a
// non-zero valid timestamp.
var NilTimestampOrNotEmpty = validation.By(func(value any) error {
	_, isNil := validation.Indirect(value)
	if isNil {
		return nil
	}
	return timestampNotEmptyRule.Validate(value)
})

// IsRFC3339Timestamp validates whether a value is a valid RFC3339 timestamp.
//
// It accepts RFC3339/RFC3339Nano strings and byte slices, as well as
// `time.Time` values and their pointer forms.
var IsRFC3339Timestamp = validation.By(func(value any) error {
	if _, ok := timestampValue(value); ok {
		return nil
	}
	return errInvalidTimestamp
})

func durationValue(value any) (time.Duration, bool) {
	if value == nil {
		return 0, false
	}
	rv := reflect.ValueOf(value)
	for rv.IsValid() {
		kind := rv.Kind()
		if kind != reflect.Pointer && kind != reflect.Interface {
			break
		}
		if rv.IsNil() {
			return 0, false
		}
		rv = rv.Elem()
	}
	if rv.IsValid() {
		value = rv.Interface()
	}
	if d, ok := value.(time.Duration); ok {
		return d, true
	}
	if isString, str, isBytes, bs := validation.StringOrBytes(value); isString {
		d, err := time.ParseDuration(str)
		return d, err == nil
	} else if isBytes {
		d, err := time.ParseDuration(string(bs))
		return d, err == nil
	}
	return 0, false
}

func timestampValue(value any) (time.Time, bool) {
	if value == nil {
		return time.Time{}, false
	}
	rv := reflect.ValueOf(value)
	for rv.IsValid() {
		kind := rv.Kind()
		if kind != reflect.Pointer && kind != reflect.Interface {
			break
		}
		if rv.IsNil() {
			return time.Time{}, false
		}
		rv = rv.Elem()
	}
	if rv.IsValid() {
		value = rv.Interface()
	}
	if ts, ok := value.(time.Time); ok {
		return ts, true
	}
	if isString, str, isBytes, bs := validation.StringOrBytes(value); isString {
		if ts, err := time.Parse(time.RFC3339, str); err == nil {
			return ts, true
		}
		if ts, err := time.Parse(time.RFC3339Nano, str); err == nil {
			return ts, true
		}
	} else if isBytes {
		s := string(bs)
		if ts, err := time.Parse(time.RFC3339, s); err == nil {
			return ts, true
		}
		if ts, err := time.Parse(time.RFC3339Nano, s); err == nil {
			return ts, true
		}
	}
	return time.Time{}, false
}
