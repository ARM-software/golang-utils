package semver

import (
	"fmt"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"golang.org/x/mod/semver"
)

// CanonicalPrefix will return the canonical form of the version with the .MAJOR .MINOR and .PATCH whilst discarding build information
// It will return an error if the version is not valid semver (unlike Canonical() in golang.org/x/mod/semver)
// It will prepend 'v' if necessary for compatibility with 'golang.org/x/mod/semver'. Use Canonical() if you don't want the 'v' prefix
func CanonicalPrefix(v string) (canonical string, err error) {
	if v == "" {
		err = fmt.Errorf("%w: no version was supplied", commonerrors.ErrUndefined)
		return
	}

	if v[0] != 'v' {
		v = fmt.Sprint("v", v)
	}

	canonical = semver.Canonical(v)
	if canonical == "" {
		err = fmt.Errorf("%w: could not parse '%v' as a semantic version", commonerrors.ErrInvalid, v)
	}

	return
}

// Canonical will return the canonical form of the version with the .MAJOR .MINOR and .PATCH whilst discarding build information
// It will return an error if the version is not valid semver (unlike Canonical() in golang.org/x/mod/semver)
// Use CanonicalPrefix() if you want the 'v' prefix for compatibility with golang.org/x/mod/semver
func Canonical(v string) (canonical string, err error) {
	if v == "" {
		err = fmt.Errorf("%w: no version was supplied", commonerrors.ErrUndefined)
		return
	}

	canonical, err = CanonicalPrefix(v)
	if err != nil {
		return
	}

	canonical = TrimPrefix(canonical)
	return
}

// TrimPrefix will trim the 'v' prefix from a string
func TrimPrefix(v string) string {
	if v == "" || v[0] != 'v' {
		return v
	}
	return v[1:]
}
