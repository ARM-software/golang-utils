package semver

import (
	"golang.org/x/mod/semver"
)

// SanitiseVersionMajor will return the major and minor parts of the version with the 'v' prefix
func SanitiseVersionMajorMinor(version string) (majorMinor string, err error) {
	version, err = CanonicalWithGoPrefix(version)
	if err != nil {
		return
	}

	majorMinor = TrimPrefix(semver.MajorMinor(version))
	return
}

// SanitiseVersionMajor will return the major part of the version with the 'v' prefix
func SanitiseVersionMajor(version string) (majorMinor string, err error) {
	version, err = CanonicalWithGoPrefix(version)
	if err != nil {
		return
	}

	majorMinor = TrimPrefix(semver.Major(version))
	return
}
