package url

import (
	"testing"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/stretchr/testify/assert"
)

func TestUrl_IsParamSegment(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		assert.True(t, IsParamSegment("{abc}"))
	})

	t.Run("false", func(t *testing.T) {
		assert.False(t, IsParamSegment("abc"))
	})
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

func TestUrl_JoinPaths(t *testing.T) {
	tests := []struct {
		name   string
		paths  []string
		result string
		error  error
	}{
		{
			"empty paths",
			[]string{},
			"",
			nil,
		},
		{
			"undefined paths",
			nil,
			"",
			commonerrors.ErrUndefined,
		},
		{
			"one path",
			[]string{"abc/123"},
			"abc/123",
			nil,
		},
		{
			"two paths",
			[]string{"abc/123", "def/456"},
			"abc/123/def/456",
			nil,
		},
		{
			"two paths with leading and trailing slashes",
			[]string{"abc/123/", "/def/456"},
			"abc/123/def/456",
			nil,
		},
		{
			"multiple paths",
			[]string{"abc/123", "def/456", "zzz", "{param1}", "999"},
			"abc/123/def/456/zzz/{param1}/999",
			nil,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			joinedPaths, err := JoinPaths(test.paths...)

			errortest.AssertError(t, err, test.error)
			assert.Equal(t, test.result, joinedPaths)
		})
	}
}

func TestUrl_JoinPathsWithSeparator(t *testing.T) {
	tests := []struct {
		name      string
		paths     []string
		separator string
		result    string
		error     error
	}{
		{
			"empty paths",
			[]string{},
			"/",
			"",
			nil,
		},
		{
			"undefined paths",
			nil,
			"~",
			"",
			commonerrors.ErrUndefined,
		},
		{
			"one path with custom separator",
			[]string{"abc/123"},
			"|",
			"abc/123",
			nil,
		},
		{
			"two paths with custom separator",
			[]string{"abc/123", "def/456"},
			"|",
			"abc/123|def/456",
			nil,
		},
		{
			"two paths with empty separator",
			[]string{"abc/123", "def/456"},
			"",
			"abc/123/def/456",
			nil,
		},
		{
			"two paths with whitespace separator",
			[]string{"abc/123", "def/456"},
			"  ",
			"abc/123/def/456",
			nil,
		},
		{
			"two paths with leading and trailing slashes and custom separator",
			[]string{"abc/123/", "/def/456"},
			"#",
			"abc/123/#/def/456",
			nil,
		},
		{
			"multiple paths with custom separator",
			[]string{"abc/123", "def/456", "zzz", "{param1}", "999"},
			"~",
			"abc/123~def/456~zzz~{param1}~999",
			nil,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			joinedPaths, err := JoinPathsWithSeparator(test.separator, test.paths...)

			errortest.AssertError(t, err, test.error)
			assert.Equal(t, test.result, joinedPaths)
		})
	}
}

func TestUrl_SplitPath(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		result []string
		error  error
	}{
		{
			"empty path",
			"",
			[]string{},
			nil,
		},
		{
			"path with one segment",
			"abc",
			[]string{"abc"},
			nil,
		},
		{
			"path with two segments",
			"abc/123",
			[]string{"abc", "123"},
			nil,
		},
		{
			"path with multiple segments",
			"abc/123/def/456",
			[]string{"abc", "123", "def", "456"},
			nil,
		},
		{
			"path with multiple segments including param segment",
			"abc/123/def/456/zzz/{param1}/999",
			[]string{"abc", "123", "def", "456", "zzz", "{param1}", "999"},
			nil,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			segments := SplitPath(test.path)

			if test.result != nil {
				for i, s := range segments {
					assert.Equal(t, test.result[i], s)
				}
			} else {
				assert.Nil(t, segments)
			}
		})
	}
}

func TestUrl_SplitPathWithSeparator(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		separator string
		result    []string
		error     error
	}{
		{
			"empty path",
			"",
			"",
			[]string{},
			nil,
		},
		{
			"path with one segment and custom separator",
			"abc",
			"|",
			[]string{"abc"},
			nil,
		},
		{
			"path with two segments and custom separator",
			"abc|123",
			"|",
			[]string{"abc", "123"},
			nil,
		},
		{
			"path with multiple segments and custom separator",
			"abc~123~def~456",
			"~",
			[]string{"abc", "123", "def", "456"},
			nil,
		},
		{
			"path with multiple segments including param segment and custom separator",
			"abc~123~def~456~zzz~{param1}~999",
			"~",
			[]string{"abc", "123", "def", "456", "zzz", "{param1}", "999"},
			nil,
		},
		{
			"path with multiple segments including param segment and custom separator with other separators",
			"abc~123/def~456~zzz~{param1}|999",
			"~",
			[]string{"abc", "123/def", "456", "zzz", "{param1}|999"},
			nil,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			segments := SplitPathWithSeparator(test.path, test.separator)

			if test.result != nil {
				for i, s := range segments {
					assert.Equal(t, test.result[i], s)
				}
			} else {
				assert.Nil(t, segments)
			}
		})
	}
}
