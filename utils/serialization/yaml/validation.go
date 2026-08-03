package yaml

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

// IsValidYAML validates that a string or byte slice contains syntactically
// valid YAML.
var IsValidYAML = validation.By(isValidYAML)

func isValidYAML(value any) error {
	if isString, str, isBytes, bs := validation.StringOrBytes(value); isString {
		_, err := ToJSON([]byte(str))
		if err == nil {
			return nil
		}
		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "must be valid YAML")
	} else if isBytes {
		_, err := ToJSON(bs)
		if err == nil {
			return nil
		}
		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "must be valid YAML")
	}
	return commonerrors.Newf(commonerrors.ErrMarshalling, "unsupported type for YAML validation '%T'", value)
}
