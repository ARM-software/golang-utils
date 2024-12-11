package semver

import (
	"golang.org/x/mod/semver"
)

// SanetiseVersionMajor will return the major and minor parts of the version with the 'v' prefix
func SanetiseVersionMajorMinor(version string) (majorMinor string, err error) {
	version, err = CanonicalPrefix(version)
	if err != nil {
		return
	}

	majorMinor = TrimPrefix(semver.MajorMinor(version))
	return
}

// SanetiseVersionMajor will return the major part of the version with the 'v' prefix
func SanetiseVersionMajor(version string) (majorMinor string, err error) {
	version, err = CanonicalPrefix(version)
	if err != nil {
		return
	}

	majorMinor = TrimPrefix(semver.Major(version))
	return
}
