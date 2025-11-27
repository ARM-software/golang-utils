package url

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestUrl_MatchesPathParameterSyntax(t *testing.T) {
	tests := []struct {
		name      string
		parameter string
		result    bool
	}{
		{
			"valid",
			"{abc}",
			true,
		},
		{
			"with encoded underscore",
			"{abc%5F1}", // unescaped as '{abc_1}'
			true,
		},
		{
			"only whitespace",
			"  ",
			false,
		},
		{
			"missing opening brace",
			"abc}",
			false,
		},
		{
			"missing closing brace",
			"{abc",
			false,
		},
		{
			"missing both braces",
			"abc",
			false,
		},
		{
			"contains multiple braces",
			"{{abc}}",
			false,
		},
		{
			"with encoded asterisk",
			"{abc%2A123}", // unescaped as '{abc*123}'
			true,
		},
		{
			"with encoded space",
			"{%20abc%20}", // unescaped as '{ abc }'
			true,
		},
		{
			"with valid special characters",
			"{abc$123.zzz~999}",
			true,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, MatchesPathParameterSyntax(test.parameter))
		})
	}
}

func TestUrl_ValidatePathParameter(t *testing.T) {
	tests := []struct {
		name      string
		parameter string
		err       error
	}{
		{
			"valid",
			"{abc}",
			nil,
		},
		{
			"with valid special characters",
			"{abc.-_+$@!123(a)}",
			nil,
		},
		{
			"with encoded underscore",
			"{abc%5F1}", // unescaped as '{abc_1}'
			nil,
		},
		{
			"missing opening brace",
			"abc}",
			commonerrors.ErrInvalid,
		},
		{
			"missing closing brace",
			"{abc",
			commonerrors.ErrInvalid,
		},
		{
			"missing both braces",
			"abc",
			commonerrors.ErrInvalid,
		},
		{
			"contains multiple braces",
			"{{abc}}",
			commonerrors.ErrInvalid,
		},
		{
			"with encoded asterisk",
			"{abc%2A123}", // unescaped as '{abc*123}'
			nil,
		},
		{
			"with encoded hash",
			"{abc%23123}", // unescaped as '{abc#123}'
			commonerrors.ErrInvalid,
		},
		{
			"with encoded space",
			"{%20abc%20}", // unescaped as '{ abc }'
			commonerrors.ErrInvalid,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			errortest.AssertError(t, ValidatePathParameter(test.parameter), test.err)
		})
	}
}

func TestUrl_HasMatchingPathSegments(t *testing.T) {
	tests := []struct {
		name   string
		pathA  string
		pathB  string
		result bool
		err    error
	}{
		{
			"empty pathA",
			"",
			"abc/123",
			false,
			commonerrors.ErrUndefined,
		},
		{
			"empty pathB",
			"abc/123",
			"",
			false,
			commonerrors.ErrUndefined,
		},
		{
			"identical paths",
			"abc/123",
			"abc/123",
			true,
			nil,
		},
		{
			"identical paths with multiple segments",
			"abc/123/def/456/zzz",
			"abc/123/def/456/zzz",
			true,
			nil,
		},
		{
			"root paths",
			"/",
			"/",
			true,
			nil,
		},
		{
			"paths with different segment values",
			"abc/123",
			"abc/456",
			false,
			nil,
		},
		{
			"paths with different lengths",
			"abc/123",
			"abc/123/456",
			false,
			nil,
		},
		{
			"path with trailing slashes",
			"/abc/123/",
			"abc/123",
			true,
			nil,
		},
		{
			"paths with repeated slashes",
			"//abc///123/",
			"abc//123/////",
			true,
			nil,
		},
		{
			"path with valid encoding",
			"abc/123%5F456", // unescaped as 'abc/123_456'
			"abc/123_456",
			true,
			nil,
		},
		{
			"path with invalid encoding",
			"abc/%$#%*123",
			"abc/123",
			false,
			commonerrors.ErrUnexpected,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			match, err := HasMatchingPathSegments(test.pathA, test.pathB)
			errortest.AssertError(t, err, test.err)
			assert.Equal(t, test.result, match)
		})
	}
}

func TestUrl_HasMatchingPathSegmentsWithParams(t *testing.T) {
	tests := []struct {
		name   string
		pathA  string
		pathB  string
		result bool
		err    error
	}{
		{
			"empty pathA",
			"",
			"abc/123",
			false,
			commonerrors.ErrUndefined,
		},
		{
			"empty pathB",
			"abc/123",
			"",
			false,
			commonerrors.ErrUndefined,
		},
		{
			"identical paths",
			"abc/123",
			"abc/123",
			true,
			nil,
		},
		{
			"identical paths with repeated slashes",
			"abc///123//",
			"//abc/123///",
			true,
			nil,
		},
		{
			"identical paths with multiple segments",
			"abc/123/def/456/zzz",
			"abc/123/def/456/zzz",
			true,
			nil,
		},
		{
			"path with parameter segment",
			"/abc/{id}/123",
			"/abc/123/123",
			true,
			nil,
		},
		{
			"both paths with matching parameter segments",
			"/abc/{param}/123",
			"/abc/{param}/123",
			true,
			nil,
		},
		{
			"both paths with different parameter segments",
			"/abc/{id}/123",
			"/abc/{val}/123",
			true,
			nil,
		},
		{
			"paths with different segments",
			"/abc/123/xyz",
			"/def/123/zzz",
			false,
			nil,
		},
		{
			"paths with different segments with parameter",
			"/abc/{param}/123",
			"/def/123/zzz",
			false,
			nil,
		},
		{
			"paths with different lengths and params",
			"/abc/{param}",
			"/abc/{param}/123",
			false,
			nil,
		},
		{
			"path with valid encoding in parameter segment",
			"abc/{param%2D1}", // unescaped as 'abc/{param-1}'
			"abc/123",
			true,
			nil,
		},
		{
			"path with invalid encoding in parameter segment",
			"abc/{%$#%*param}",
			"abc/123",
			false,
			commonerrors.ErrUnexpected,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			match, err := HasMatchingPathSegmentsWithParams(test.pathA, test.pathB)
			errortest.AssertError(t, err, test.err)
			assert.Equal(t, test.result, match)
		})
	}
}

func TestUrl_MatchingPathSegments(t *testing.T) {
	tests := []struct {
		name      string
		pathA     string
		pathB     string
		matcherFn PathSegmentMatcherFunc
		result    bool
		err       error
	}{
		{
			"empty pathA",
			"",
			"abc/123",
			BasicEqualityPathSegmentMatcher,
			false,
			commonerrors.ErrUndefined,
		},
		{
			"empty pathB",
			"abc/123",
			"",
			BasicEqualityPathSegmentMatcher,
			false,
			commonerrors.ErrUndefined,
		},
		{
			"path with valid encoding",
			"abc/123%5F456", // unescaped as 'abc/123_456'
			"abc/123_456",
			BasicEqualityPathSegmentMatcher,
			true,
			nil,
		},
		{
			"path with invalid encoding",
			"abc/%$#%*123",
			"abc/123",
			BasicEqualityPathSegmentMatcher,
			false,
			commonerrors.ErrUnexpected,
		},
		{
			"paths with different segments with parameter",
			"/abc/{param}/123",
			"/def/123/zzz",
			BasicEqualityPathSegmentWithParamMatcher,
			false,
			nil,
		},
		{
			"paths with different lengths and params",
			"/abc/{param}",
			"/abc/{param}/123",
			BasicEqualityPathSegmentWithParamMatcher,
			false,
			nil,
		},
		{
			"matching paths when using a custom matcher function",
			"/abc/||zzz||/123",
			"/abc/||{param}||/123",
			func(segmentA string, segmentB string) (match bool, err error) {
				segmentA = strings.Trim(segmentA, "|")
				segmentB = strings.Trim(segmentB, "|")
				return BasicEqualityPathSegmentWithParamMatcher(segmentA, segmentB)
			},
			true,
			nil,
		},
		{
			"non-matching paths when using a custom matcher function",
			"/abc/##zzz||/123",
			"/abc/||{param}##/123",
			func(segmentA string, segmentB string) (match bool, err error) {
				segmentA = strings.Trim(segmentA, "#")
				segmentB = strings.Trim(segmentB, "|")
				return BasicEqualityPathSegmentWithParamMatcher(segmentA, segmentB)
			},
			false,
			nil,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			match, err := MatchingPathSegments(test.pathA, test.pathB, test.matcherFn)
			errortest.AssertError(t, err, test.err)
			assert.Equal(t, test.result, match)
		})
	}
}

func TestUrl_SplitPath(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		result []string
	}{
		{
			"empty path",
			"",
			[]string{},
		},
		{
			"root path",
			"/",
			[]string{"/"},
		},
		{
			"root path with repeated slashes",
			"///",
			[]string{"/"},
		},
		{
			"path with one segment",
			"abc",
			[]string{"abc"},
		},
		{
			"path with two segments",
			"abc/123",
			[]string{"abc", "123"},
		},
		{
			"path with multiple segments",
			"abc/123/def/456",
			[]string{"abc", "123", "def", "456"},
		},
		{
			"path with multiple segments including param segment",
			"abc/123/def/456/zzz/{param1}/999",
			[]string{"abc", "123", "def", "456", "zzz", "{param1}", "999"},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			segments := SplitPath(test.path)

			for i, s := range segments {
				assert.Equal(t, test.result[i], s)
			}
		})
	}
}
