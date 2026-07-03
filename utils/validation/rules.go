package validation

// rules.go contains the general-purpose non-schema-specific helpers, especially
// `Is...` validators that extend or adapt ozzo-validation's stock rules.

import (
	"strconv"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/encoding/base64"
	"github.com/ARM-software/golang-utils/utils/url"
)

// IsPort validates whether a value is a port using is.Port from github.com/go-ozzo/ozzo-validation/v4.
// However, it supports all base go integer types not just strings.
var IsPort = validation.By(isPort)

func isPort(vRaw any) (err error) {
	if isString, str, isBytes, bs := validation.StringOrBytes(vRaw); isString {
		err = is.Port.Validate(str)
	} else if isBytes {
		err = is.Port.Validate(string(bs))
	} else if i, convErr := validation.ToInt(vRaw); convErr == nil {
		err = is.Port.Validate(strconv.FormatInt(i, 10))
	} else if u, convErr := validation.ToUint(vRaw); convErr == nil {
		err = is.Port.Validate(strconv.FormatUint(u, 10))
	} else {
		return commonerrors.Newf(commonerrors.ErrMarshalling, "unsupported type for port validation '%T'", vRaw)
	}

	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrInvalid, err, "")
		return
	}

	return
}

// IsBase64 validates whether a value is a base64 encoded string. It is similar to is.Base64 but more generic and robust although less performant.
var IsBase64 = validation.NewStringRuleWithError(base64.IsEncoded, is.ErrBase64)

// IsPathParameter validates whether a value is a valid path parameter of a url.
var IsPathParameter = validation.NewStringRule(isValidPathParameter, "invalid path parameter")

func isValidPathParameter(value string) bool {
	err := url.ValidatePathParameter(value)
	return err == nil
}

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

// LengthExact validates that a length-aware value has exactly n elements.
func LengthExact(n int) validation.Rule {
	return validation.Length(n, n)
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
