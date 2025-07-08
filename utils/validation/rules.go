package validation

import (
	"reflect"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

// IsPort validates whether a value is a port using is.Port from github.com/go-ozzo/ozzo-validation/v4.
// However it supports all base go integer types not just strings.
func IsPort() validation.Rule {
	return validation.By(func(vRaw any) (err error) {
		switch val := reflect.ValueOf(vRaw); val.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			err = is.Port.Validate(strconv.FormatInt(val.Int(), 10))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			err = is.Port.Validate(strconv.FormatUint(val.Uint(), 10))
		case reflect.String:
			err = is.Port.Validate(val.String())
		case reflect.Slice:
			if b, ok := vRaw.([]byte); ok {
				err = is.Port.Validate(string(b))
			}
		default:
			return commonerrors.Newf(commonerrors.ErrMarshalling, "unsupported type for port validation '%T'", vRaw)
		}

		if err != nil {
			err = commonerrors.WrapError(commonerrors.ErrInvalid, err, "")
			return
		}

		return
	})
}
