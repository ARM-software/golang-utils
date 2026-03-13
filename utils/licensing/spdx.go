// Package licensing provides helpers for validating, normalising, and working
// with SPDX licence identifiers and expressions.
//
// The package builds on and is inspired by
// https://github.com/git-pkgs/spdx, which provides parsing,
// normalisation, and validation of SPDX licence expressions.
//
// It adds higher-level utilities commonly needed when handling licensing
// metadata in applications
package licensing

import (
	"fmt"
	"iter"
	"net/url"
	"slices"

	"github.com/git-pkgs/spdx"
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

var (
	// errSPDXInvalid is returned when a string is not a valid SPDX licence
	errSPDXInvalid = validation.NewError("validation_is_spdx_licence", "must be a valid SPDX licence")

	// IsSPDXLicence defines an ozzo-validation rule that ensures a string
	// contains a valid SPDX licence expression.
	//
	// This rule can be used with github.com/go-ozzo/ozzo-validation to validate
	// fields containing licence identifiers or expressions such as:
	//
	//   MIT
	//   Apache-2.0
	//   MIT OR Apache-2.0
	//   GPL-3.0-only WITH Classpath-exception-2.0
	//
	// Example:
	//
	//   validation.Field(&pkg.Licence, licensing.IsSPDXLicence)
	//
	// Validation internally relies on ValidateSPDXLicence.
	IsSPDXLicence = validation.NewStringRuleWithError(
		func(l string) bool { return ValidateSPDXLicence(l) == nil },
		errSPDXInvalid,
	)
)

// ValidateSPDXLicence validates that the provided string is a valid SPDX
// licence expression.
//
// The expression is parsed using an lenient SPDX parser which will try to identify licences even if they are not in their canonical form.
//
// Returns an error if:
//   - the expression is empty
//   - the expression cannot be parsed as a valid SPDX licence expression
//
// Example valid expressions:
//
//	MIT
//	Apache-2.0
//	MIT OR Apache-2.0
//	GPL-2.0-or-later
func ValidateSPDXLicence(licence string) error {
	if reflection.IsEmpty(licence) {
		return commonerrors.UndefinedVariable("licence expression")
	}
	_, err := spdx.Parse(licence)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed normalising SPDX expression")
	}
	return err
}

// NormaliseSPDXLicence converts an SPDX licence expression into its canonical
// SPDX representation.
//
// This function performs a lax normalisation using the SPDX library, allowing
// minor variations in input formatting while still producing a valid canonical
// SPDX expression.
//
// For example:
//
//	"apache 2"        → "Apache-2.0"
//	"mit or apache2"  → "MIT OR Apache-2.0"
//
// Returns the canonical SPDX expression or an error if the expression cannot
// be parsed or normalised.
func NormaliseSPDXLicence(expression string) (canonical string, err error) {
	if reflection.IsEmpty(expression) {
		err = commonerrors.UndefinedVariable("licence expression")
		return
	}
	canonical, err = spdx.NormalizeExpressionLax(expression) //nolint:misspell
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed normalising SPDX expression")
	}
	return
}

// SatisfiesLicensingConstraints determines whether a licence expression is
// compatible with a list of allowed licences.
//
// Note: The input licence expression and all entries in the allowed list are first
// normalised to their canonical SPDX form before evaluation.
//
// Behaviour:
//   - The expression may contain SPDX operators such as AND / OR.
//   - The function returns true if the licence expression satisfies at least
//     one licence in the allowed list according to SPDX semantics.
//
// Example:
//
//	licence     = "MIT OR Apache-2.0"
//	allowedList = ["MIT"]
//
// Result:
//
//	true
//
// This behaviour is similar to:
// https://pkg.go.dev/github.com/github/go-spdx/v2/spdxexp#Satisfies
func SatisfiesLicensingConstraints(licence string, allowedList []string) (pass bool, err error) {
	norm, err := NormaliseSPDXLicence(licence)
	if err != nil {
		return
	}
	allowed, err := collection.MapWithError[string, string](allowedList, NormaliseSPDXLicence)
	if err != nil {
		return
	}
	pass, err = spdx.Satisfies(norm, allowed)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "failed checking licence constraints")
	}
	return
}

// FetchLicenceURL returns the SPDX website URL corresponding to the provided
// SPDX licence identifier.
//
// The input licence is normalised before constructing the URL.
//
// Example:
//
//	MIT        → https://spdx.org/licenses/MIT.html
//	Apache-2.0 → https://spdx.org/licenses/Apache-2.0.html
//
// The input must represent a single licence identifier rather than a compound
// SPDX expression.
func FetchLicenceURL(spdxLicence *string) (licenceURL *url.URL, err error) {
	if reflection.IsEmpty(spdxLicence) {
		err = commonerrors.UndefinedVariable("licence")
		return
	}
	lStr := field.OptionalString(spdxLicence, "")
	l, err := NormaliseSPDXLicence(lStr)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "failed identifying SPDX licence [%v]", lStr)
		return
	}
	if !spdx.ValidLicense(l) {
		err = commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "not a valid SPDX licence [%v]", lStr)
		return
	}
	licenceURL, err = url.Parse(fmt.Sprintf("https://spdx.org/licenses/%v.html", l))
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "failed determining the licence's URL [%v]", lStr)
		return
	}
	return
}

// FetchLicenceURLs extracts all licences referenced in an SPDX licence
// expression and returns the SPDX reference URLs for each licence.
//
// Note: The expression is first normalised before extracting the licences.
// Moreover, operators such as AND, OR, and WITH are ignored when extracting licences
//
// Example:
//
//	expression: "MIT OR Apache-2.0"
//
//	returns:
//
//	https://spdx.org/licenses/MIT.html
//	https://spdx.org/licenses/Apache-2.0.html
func FetchLicenceURLs(expression string) (urls iter.Seq[url.URL], err error) {
	l, err := NormaliseSPDXLicence(expression)
	if err != nil {
		return
	}
	licences, err := spdx.ExtractLicenses(l)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed extracting the licences from the expression")
		return
	}
	urls = collection.MapSequenceRefWithError[string, url.URL](slices.Values(licences), FetchLicenceURL)
	return
}
