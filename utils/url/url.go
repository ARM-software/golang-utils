package url

import (
	"strings"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

const defaultPathSeparator = "/"

// IsParamSegment checks whether the segment string is a path parameter
func IsParamSegment(segment string) bool {
	return len(segment) >= 2 && segment[0] == '{' && segment[len(segment)-1] == '}'
}

// HasMatchingPathSegments checks whether two path strings match based on their segments.
func HasMatchingPathSegments(pathA, pathB string) (bool, error) {
	return hasMatchingPathSegments(pathA, pathB, func(segmentA, segmentB string) bool {
		return segmentA == segmentB
	})
}

// HasMatchingPathSegmentsWithParams is similar to HasMatchingPathSegments but also considers segments as matching if at least one of them contains a path parameter.
//
//	HasMatchingPathSegmentsWithParams("/some/{param}/path", "/some/{param}/path") // true
//	HasMatchingPathSegmentsWithParams("/some/abc/path", "/some/{param}/path") // true
//	HasMatchingPathSegmentsWithParams("/some/abc/path", "/some/def/path") // false
func HasMatchingPathSegmentsWithParams(pathA, pathB string) (bool, error) {
	return hasMatchingPathSegments(pathA, pathB, func(pathASeg, pathBSeg string) bool {
		switch {
		case IsParamSegment(pathASeg):
			return pathBSeg != ""
		case IsParamSegment(pathBSeg):
			return pathASeg != ""
		default:
			return pathASeg == pathBSeg
		}
	})
}

func hasMatchingPathSegments(pathA, pathB string, segmentMatcherFn func(segmentA, segmentB string) bool) (match bool, err error) {
	if reflection.IsEmpty(pathA) {
		err = commonerrors.UndefinedVariable("path A")
		return
	}

	if reflection.IsEmpty(pathB) {
		err = commonerrors.UndefinedVariable("path B")
		return
	}

	if segmentMatcherFn == nil {
		err = commonerrors.UndefinedVariable("segment matcher function")
		return
	}

	pathASegments := SplitPath(pathA)
	pathBSegments := SplitPath(pathB)
	if len(pathASegments) != len(pathBSegments) {
		return
	}

	for i := range pathBSegments {
		if !segmentMatcherFn(pathASegments[i], pathBSegments[i]) {
			return
		}
	}

	match = true
	return
}

// SplitPath returns a slice containing the individual segments that make up the path string. It looks for the default "/" path separator when splitting.
func SplitPath(path string) []string {
	return SplitPathWithSeparator(path, defaultPathSeparator)
}

// SplitPathWithSeparator is similar to SplitPath but allows for specifying the path separator to look for when splitting.
func SplitPathWithSeparator(path string, separator string) []string {
	path = strings.TrimSpace(path)
	if reflection.IsEmpty(path) || path == separator {
		return nil
	}

	path = strings.Trim(path, separator)
	segments := strings.Split(path, separator)
	out := segments[:0]
	for _, p := range segments {
		if !reflection.IsEmpty(p) {
			out = append(out, p)
		}
	}
	return out
}

// JoinPaths returns a single concatenated path string from the supplied paths and correctly sets the default "/" separator between them.
func JoinPaths(paths ...string) (joinedPath string, err error) {
	return JoinPathsWithSeparator(defaultPathSeparator, paths...)
}

// JoinPathsWithSeparator is similar to JoinPaths but allows for specifying the path separator to use.
func JoinPathsWithSeparator(separator string, paths ...string) (joinedPath string, err error) {
	if paths == nil {
		err = commonerrors.UndefinedVariable("paths")
		return
	}
	if len(paths) == 0 {
		return
	}
	if len(paths) == 1 {
		joinedPath = paths[0]
		return
	}

	if reflection.IsEmpty(separator) {
		separator = defaultPathSeparator
	}

	joinedPath = paths[0]
	for _, p := range paths[1:] {
		pathAHasSlashSuffix := strings.HasSuffix(joinedPath, separator)
		pathBHasSlashPrefix := strings.HasPrefix(p, separator)

		switch {
		case pathAHasSlashSuffix && pathBHasSlashPrefix:
			joinedPath += p[1:]
		case !pathAHasSlashSuffix && !pathBHasSlashPrefix:
			joinedPath += separator + p
		default:
			joinedPath += p
		}
	}

	return
}
