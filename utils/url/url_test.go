package url

import (
	"fmt"
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
		t.Run(fmt.Sprintf("JoinPaths_%v", test.name), func(t *testing.T) {
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
		t.Run(fmt.Sprintf("JoinPathsWithSeparator_%v", test.name), func(t *testing.T) {
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
		t.Run(fmt.Sprintf("JoinPaths_%v", test.name), func(t *testing.T) {
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
		t.Run(fmt.Sprintf("JoinPaths_%v", test.name), func(t *testing.T) {
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
