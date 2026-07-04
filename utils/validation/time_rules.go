package validation

// time_rules.go contains validation helpers related to durations and timestamps,
// including both parsing helpers (`IsDuration`, `IsTimestamp`) and threshold/
// equality rules for time-oriented values.

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

// DurationMinimum validates that a duration value is greater than or equal to min.
func DurationMinimum(min time.Duration) validation.Rule {
	return validation.By(func(value any) error {
		d, ok := durationValue(value)
		if !ok {
			return nil
		}
		if d < min {
			return commonerrors.New(commonerrors.ErrInvalid, "must be a valid duration greater than or equal to the minimum")
		}
		return nil
	})
}

// DurationExclusiveMinimum validates that a duration value is strictly greater than min.
func DurationExclusiveMinimum(min time.Duration) validation.Rule {
	return validation.By(func(value any) error {
		d, ok := durationValue(value)
		if !ok {
			return nil
		}
		if d <= min {
			return commonerrors.New(commonerrors.ErrInvalid, "must be a valid duration greater than the minimum")
		}
		return nil
	})
}

// DurationMaximum validates that a duration value is less than or equal to max.
func DurationMaximum(max time.Duration) validation.Rule {
	return validation.By(func(value any) error {
		d, ok := durationValue(value)
		if !ok {
			return nil
		}
		if d > max {
			return commonerrors.New(commonerrors.ErrInvalid, "must be a valid duration less than or equal to the maximum")
		}
		return nil
	})
}

// DurationExclusiveMaximum validates that a duration value is strictly less than max.
func DurationExclusiveMaximum(max time.Duration) validation.Rule {
	return validation.By(func(value any) error {
		d, ok := durationValue(value)
		if !ok {
			return nil
		}
		if d >= max {
			return commonerrors.New(commonerrors.ErrInvalid, "must be a valid duration less than the maximum")
		}
		return nil
	})
}

// DurationConst validates that a duration value is exactly equal to expected.
func DurationConst(expected time.Duration) validation.Rule {
	return validation.By(func(value any) error {
		d, ok := durationValue(value)
		if !ok {
			return nil
		}
		if d != expected {
			return commonerrors.New(commonerrors.ErrInvalid, "must be the expected duration")
		}
		return nil
	})
}

// TimestampMinimum validates that a timestamp value is greater than or equal to min.
func TimestampMinimum(min time.Time) validation.Rule {
	return validation.By(func(value any) error {
		ts, ok := timestampValue(value)
		if !ok {
			return nil
		}
		if ts.Before(min) {
			return commonerrors.New(commonerrors.ErrInvalid, "must be a valid timestamp greater than or equal to the minimum")
		}
		return nil
	})
}

// TimestampExclusiveMinimum validates that a timestamp value is strictly after min.
func TimestampExclusiveMinimum(min time.Time) validation.Rule {
	return validation.By(func(value any) error {
		ts, ok := timestampValue(value)
		if !ok {
			return nil
		}
		if !ts.After(min) {
			return commonerrors.New(commonerrors.ErrInvalid, "must be a valid timestamp greater than the minimum")
		}
		return nil
	})
}

// TimestampMaximum validates that a timestamp value is less than or equal to max.
func TimestampMaximum(max time.Time) validation.Rule {
	return validation.By(func(value any) error {
		ts, ok := timestampValue(value)
		if !ok {
			return nil
		}
		if ts.After(max) {
			return commonerrors.New(commonerrors.ErrInvalid, "must be a valid timestamp less than or equal to the maximum")
		}
		return nil
	})
}

// TimestampExclusiveMaximum validates that a timestamp value is strictly before max.
func TimestampExclusiveMaximum(max time.Time) validation.Rule {
	return validation.By(func(value any) error {
		ts, ok := timestampValue(value)
		if !ok {
			return nil
		}
		if !ts.Before(max) {
			return commonerrors.New(commonerrors.ErrInvalid, "must be a valid timestamp less than the maximum")
		}
		return nil
	})
}

// TimestampConst validates that a timestamp value is exactly equal to expected.
func TimestampConst(expected time.Time) validation.Rule {
	return validation.By(func(value any) error {
		ts, ok := timestampValue(value)
		if !ok {
			return nil
		}
		if !ts.Equal(expected) {
			return commonerrors.New(commonerrors.ErrInvalid, "must be the expected timestamp")
		}
		return nil
	})
}

// IsDuration validates whether a string or byte slice is a valid Go duration.
var IsDuration = validation.NewStringRule(func(value string) bool {
	_, err := time.ParseDuration(value)
	return err == nil
}, "must be a valid duration")

// IsTimestamp validates whether a string or byte slice is a valid RFC3339 timestamp.
var IsTimestamp = validation.NewStringRule(func(value string) bool {
	if _, err := time.Parse(time.RFC3339, value); err == nil {
		return true
	}
	if _, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return true
	}
	return false
}, "must be a valid timestamp")

func durationValue(value any) (time.Duration, bool) {
	if value == nil {
		return 0, false
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
