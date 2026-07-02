// Package validation provides additional validation rules built on top of
// ozzo-validation and its `is` helpers.
//
// It extends the standard rule set with reusable and project-specific
// validators while remaining fully compatible with the ozzo-validation API.
//
// Upstream projects:
//   - ozzo-validation: https://github.com/go-ozzo/ozzo-validation
//   - ozzo-validation/is: https://github.com/go-ozzo/ozzo-validation/tree/master/is
//
// Documentation:
//   - https://go-ozzo.github.io/ozzo-validation/
package validation

import (
	"strconv"

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
