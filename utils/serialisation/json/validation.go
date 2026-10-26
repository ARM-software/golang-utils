package json

import (
	stdjson "encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

// IsValidJSON validates that a string or byte slice contains syntactically
// valid JSON.
var IsValidJSON = validation.By(isValidJSON)

func isValidJSON(value any) error {
	if isString, str, isBytes, bs := validation.StringOrBytes(value); isString {
		if stdjson.Valid([]byte(str)) {
			return nil
		}
		return commonerrors.Newf(commonerrors.ErrInvalid, "must be valid JSON")
	} else if isBytes {
		if stdjson.Valid(bs) {
			return nil
		}
		return commonerrors.Newf(commonerrors.ErrInvalid, "must be valid JSON")
	}
	return commonerrors.Newf(commonerrors.ErrMarshalling, "unsupported type for JSON validation '%T'", value)
}
