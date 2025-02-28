package filesystem

import (
	"fmt"
	"regexp"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

func ExcludeFiles(files []string, regexes []*regexp.Regexp) (cleansedList []string, err error) {
	cleansedList = []string{}
	for i := range files {
		f := files[i]
		if !IsPathExcluded(f, regexes...) {
			cleansedList = append(cleansedList, f)
		}
	}
	return
}

func NewExclusionRegexList(pathSeparator rune, exclusionPatterns ...string) ([]*regexp.Regexp, error) {
	var regexes []*regexp.Regexp
	var patternsExtendedList []string
	for i := range exclusionPatterns {
		pattern := exclusionPatterns[i]

		if !reflection.IsEmpty(pattern) {
			patternsExtendedList = append(patternsExtendedList, pattern, fmt.Sprintf(".*/%v/.*", pattern), fmt.Sprintf(".*%v%v%v.*", pathSeparator, pattern, pathSeparator))
		}
	}
	for i := range patternsExtendedList {
		r, err := regexp.Compile(patternsExtendedList[i])
		if err != nil {
			return nil, commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "could not compile pattern [%v]", patternsExtendedList[i])
		}
		regexes = append(regexes, r)
	}
	return regexes, nil
}

func IsPathExcludedFromPatterns(path string, pathSeparator rune, exclusionPatterns ...string) bool {
	regexes, err := NewExclusionRegexList(pathSeparator, exclusionPatterns...)
	if err != nil {
		return false
	}
	return IsPathExcluded(path, regexes...)
}

func IsPathExcluded(path string, exclusionPatterns ...*regexp.Regexp) bool {
	for i := range exclusionPatterns {
		if exclusionPatterns[i].MatchString(path) {
			return true
		}
	}
	return false
}

// ExcludeAll excludes files
func ExcludeAll(files []string, exclusionPatterns ...string) ([]string, error) {
	return globalFileSystem.ExcludeAll(files, exclusionPatterns...)
}

func (fs *VFS) ExcludeAll(files []string, exclusionPatterns ...string) ([]string, error) {
	regexes, err := NewExclusionRegexList(fs.PathSeparator(), exclusionPatterns...)
	if err != nil {
		return nil, err
	}
	return ExcludeFiles(files, regexes)
}
