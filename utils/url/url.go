package url

import (
	netUrl "net/url"
	"path"
	"regexp"
	"strings"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

const (
	defaultPathSeparator       = "/"
	minimumPathParameterLength = 3
)

// Section 3.3 of RFC3986 details valid characters for path segments (see https://datatracker.ietf.org/doc/html/rfc3986#section-3.3)
var validPathRegex = regexp.MustCompile(`^(?:[A-Za-z0-9._~\-!$&'()*+,;=:@{}]|%[0-9A-Fa-f]{2})+$`)

// PathSegmentMatcherFunc defines the signature for path segment matcher functions.
type PathSegmentMatcherFunc = func(segmentA, segmentB string) (match bool, err error)

// ValidatePathParameter checks whether a path parameter is valid. An error is returned if it is invalid.
// Version 3.1.0 of the OpenAPI spec provides some guidance for path parameter values (see https://spec.openapis.org/oas/v3.1.0.html#path-templating)
func ValidatePathParameter(parameter string) error {
	if !MatchesPathParameterSyntax(parameter) {
		return commonerrors.Newf(commonerrors.ErrInvalid, "parameter %q must not be empty, cannot contain only whitespaces, have a length greater than or equal to three, start with an opening brace, and end with a closing brace", parameter)
	}

	unescapedSegment, err := netUrl.PathUnescape(parameter)
	if err != nil {
		return commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "an error occurred during path unescaping for parameter %q", parameter)
	}

	if !validPathRegex.MatchString(unescapedSegment) {
		return commonerrors.Newf(commonerrors.ErrInvalid, "parameter %q unescaped to %q can only contain alphanumeric characters, dashes, underscores, and a single pair of braces", parameter, unescapedSegment)
	}

	return nil
}

// MatchesPathParameterSyntax checks whether the parameter string matches the syntax for a path parameter as described by the OpenAPI spec (see https://spec.openapis.org/oas/v3.0.0.html#path-templating).
func MatchesPathParameterSyntax(parameter string) bool {
	if reflection.IsEmpty(parameter) {
		return false
	}

	if len(parameter) < minimumPathParameterLength {
		return false
	}

	if !strings.HasPrefix(parameter, "{") || !strings.HasSuffix(parameter, "}") {
		return false
	}

	return strings.Count(parameter, "{") == 1 && strings.Count(parameter, "}") == 1
}

// HasMatchingPathSegments checks whether two path strings match based on their segments by doing a simple equality check on each path segment pair.
func HasMatchingPathSegments(pathA, pathB string) (match bool, err error) {
	return MatchingPathSegments(pathA, pathB, BasicEqualityPathSegmentMatcher)
}

// HasMatchingPathSegmentsWithParams is similar to HasMatchingPathSegments but also considers segments as matching if at least one of them contains a path parameter.
//
//	HasMatchingPathSegmentsWithParams("/some/{param}/path", "/some/{param}/path") // true
//	HasMatchingPathSegmentsWithParams("/some/abc/path", "/some/{param}/path") // true
//	HasMatchingPathSegmentsWithParams("/some/abc/path", "/some/def/path") // false
func HasMatchingPathSegmentsWithParams(pathA, pathB string) (match bool, err error) {
	return MatchingPathSegments(pathA, pathB, BasicEqualityPathSegmentWithParamMatcher)
}

// BasicEqualityPathSegmentMatcher is a PathSegmentMatcherFunc that performs direct string comparison of two path segments.
func BasicEqualityPathSegmentMatcher(segmentA, segmentB string) (match bool, err error) {
	match = segmentA == segmentB
	return
}

// BasicEqualityPathSegmentWithParamMatcher is a PathSegmentMatcherFunc that is similar to BasicEqualityPathSegmentMatcher but accounts for path parameter segments.
func BasicEqualityPathSegmentWithParamMatcher(segmentA, segmentB string) (match bool, err error) {
	if MatchesPathParameterSyntax(segmentA) {
		if errValidatePathASeg := ValidatePathParameter(segmentA); errValidatePathASeg != nil {
			err = commonerrors.WrapErrorf(commonerrors.ErrInvalid, errValidatePathASeg, "an error occurred while validating path parameter %q", segmentA)
			return
		}

		match = !reflection.IsEmpty(segmentB)
		return
	}

	if MatchesPathParameterSyntax(segmentB) {
		if errValidatePathBSeg := ValidatePathParameter(segmentB); errValidatePathBSeg != nil {
			err = commonerrors.WrapErrorf(commonerrors.ErrInvalid, errValidatePathBSeg, "an error occurred while validating path parameter %q", segmentB)
			return
		}

		match = !reflection.IsEmpty(segmentA)
		return
	}

	return BasicEqualityPathSegmentMatcher(segmentA, segmentB)
}

// MatchingPathSegments checks whether two path strings match based on their segments using the provided matcher function.
func MatchingPathSegments(pathA, pathB string, matcherFn PathSegmentMatcherFunc) (match bool, err error) {
	if reflection.IsEmpty(pathA) {
		err = commonerrors.UndefinedVariable("path A")
		return
	}

	if reflection.IsEmpty(pathB) {
		err = commonerrors.UndefinedVariable("path B")
		return
	}

	if matcherFn == nil {
		err = commonerrors.UndefinedVariable("segment matcher function")
		return
	}

	unescapedPathA, errPathASeg := netUrl.PathUnescape(pathA)
	if errPathASeg != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, errPathASeg, "an error occurred while unescaping path %q", pathA)
		return
	}

	unescapedPathB, errPathBSeg := netUrl.PathUnescape(pathB)
	if errPathBSeg != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, errPathBSeg, "an error occurred while unescaping path %q", pathB)
		return
	}

	pathASegments := SplitPath(unescapedPathA)
	pathBSegments := SplitPath(unescapedPathB)
	if len(pathASegments) != len(pathBSegments) {
		return
	}

	for i := range pathBSegments {
		match, err = matcherFn(pathASegments[i], pathBSegments[i])
		if err != nil {
			err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "an error occurred during execution of the matcher function for path segments %q and %q", pathASegments[i], pathBSegments[i])
			return
		}

		if !match {
			return
		}
	}

	match = true
	return
}

// SplitPath returns a slice containing the individual segments that make up the path string p.
// It looks for the default forward slash path separator when splitting.
func SplitPath(p string) []string {
	if reflection.IsEmpty(p) {
		return []string{}
	}

	p = path.Clean(p)
	p = strings.Trim(p, defaultPathSeparator)
	return collection.ParseListWithCleanup(p, defaultPathSeparator)
}
