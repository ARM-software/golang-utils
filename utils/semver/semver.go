package semver

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

// CanonicalWithGoPrefix will return the canonical form of the version with the .MAJOR .MINOR and .PATCH whilst discarding build information
// It will return an error if the version is not valid semver (unlike Canonical() in golang.org/x/mod/semver)
// It will prepend 'v' if necessary for compatibility with 'golang.org/x/mod/semver'. Use Canonical() if you don't want the 'v' prefix
func CanonicalWithGoPrefix(v string) (canonical string, err error) {
	v = strings.TrimSpace(v)
	if v == "" {
		err = commonerrors.New(commonerrors.ErrUndefined, "no version was supplied")
		return
	}

	if v[0] != 'v' {
		v = fmt.Sprint("v", v)
	}

	canonical = semver.Canonical(v)
	if canonical == "" {
		err = commonerrors.Newf(commonerrors.ErrInvalid, "could not parse '%v' as a semantic version", v)
	}

	return
}

// Canonical will return the canonical form of the version with the .MAJOR .MINOR and .PATCH whilst discarding build information
// It will return an error if the version is not valid semver (unlike Canonical() in golang.org/x/mod/semver)
// Use CanonicalWithGoPrefix() if you want the 'v' prefix for compatibility with golang.org/x/mod/semver
func Canonical(v string) (canonical string, err error) {
	canonical, err = CanonicalWithGoPrefix(v)
	if err != nil {
		return
	}

	canonical = TrimGoPrefix(canonical)
	return
}

// TrimGoPrefix will trim the 'v' prefix from a string
func TrimGoPrefix(v string) string {
	if v == "" || v[0] != 'v' {
		return v
	}
	return v[1:]
}
