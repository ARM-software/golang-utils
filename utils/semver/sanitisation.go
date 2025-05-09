package semver

import (
	"golang.org/x/mod/semver"
)

// SanitiseVersionMajorMinor will return the major and minor parts of the version with the 'v' prefix
func SanitiseVersionMajorMinor(version string) (majorMinor string, err error) {
	version, err = CanonicalWithGoPrefix(version)
	if err != nil {
		return
	}

	majorMinor = TrimGoPrefix(semver.MajorMinor(version))
	return
}

// SanitiseVersionMajor will return the major part of the version with the 'v' prefix
func SanitiseVersionMajor(version string) (majorMinor string, err error) {
	version, err = CanonicalWithGoPrefix(version)
	if err != nil {
		return
	}

	majorMinor = TrimGoPrefix(semver.Major(version))
	return
}
