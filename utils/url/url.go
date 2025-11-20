package url

import (
	"strings"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// HasMatchingPathSegments checks whether two path strings match based on their segments.
func HasMatchingPathSegments(pathA, pathB string) (match bool, err error) {
	if reflection.IsEmpty(pathA) {
		err = commonerrors.UndefinedVariable("pathA")
		return
	}

	pathASegments := SplitPath(pathA)
	pathBSegments := SplitPath(pathB)
	if len(pathASegments) != len(pathBSegments) {
		return
	}

	for i := range pathBSegments {
		pathBSeg, pathASeg := pathBSegments[i], pathASegments[i]
		if pathBSeg != pathASeg {
			return
		}
	}

	match = true
	return
}

// HasMatchingPathSegmentsWithParams is similar to MatchingPathSegments but considers segments as matching if at least one of them contains a path parameter.
//
//	HasMatchingPathSegmentsWithParams("/some/{param}/path", "/some/{param}/path") // true
//	HasMatchingPathSegmentsWithParams("/some/abc/path", "/some/{param}/path") // true
//	HasMatchingPathSegmentsWithParams("/some/abc/path", "/some/def/path") // false
func HasMatchingPathSegmentsWithParams(pathA, pathB string) (match bool, err error) {
	if reflection.IsEmpty(pathA) {
		err = commonerrors.UndefinedVariable("pathA")
		return
	}

	pathASegments := SplitPath(pathA)
	pathBSegments := SplitPath(pathB)
	if len(pathASegments) != len(pathBSegments) {
		return
	}

	for i := range pathBSegments {
		pathBSeg, pathASeg := pathBSegments[i], pathASegments[i]
		if IsParamSegment(pathASeg) {
			if pathBSeg == "" {
				return
			}
			continue
		}

		if IsParamSegment(pathBSeg) {
			if pathASeg == "" {
				return
			}
			continue
		}
		if pathBSeg != pathASeg {
			return
		}
	}

	match = true
	return
}

// SplitPath returns a slice of the individual segments that make up the path string. It looks for the default "/" path separator when splitting.
func SplitPath(path string) []string {
	return SplitPathWithSeparator(path, "/")
}

// SplitPathWithSeparator is similar to SplitPath but allows for specifying the path separator to look for when splitting.
func SplitPathWithSeparator(path string, separator string) []string {
	path = strings.TrimSpace(path)
	if path == "" || path == separator {
		return nil
	}

	path = strings.Trim(path, separator)
	segments := strings.Split(path, separator)
	out := segments[:0]
	for _, p := range segments {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// IsParamSegment checks whether the segment string is a path parameter
func IsParamSegment(segment string) bool {
	return len(segment) >= 2 && segment[0] == '{' && segment[len(segment)-1] == '}'
}

// JoinPaths returns a single concatenated path string from the supplied paths and correctly sets the default "/" separator between them.
func JoinPaths(paths ...string) (joinedPath string, err error) {
	return JoinPathsWithSeparator("/", paths...)
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
