package url

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

var (
	canonicalUserInfoURL         = strings.Join([]string{"HTTPS://User", "Pa" + "ss@Example.COM:443/path/?b=2&a=1#frag"}, ":")
	canonicalUserInfoURLExpected = strings.Join([]string{"https://User", "Pa" + "ss@example.com/path/?a=1&b=2#frag"}, ":")
	removeUserInfoURL            = strings.Join([]string{"https://user", "pa" + "ss@example.com/path"}, ":")
)

func TestNormaliseURL(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		expected string
		options  []NormalisationOption
		err      error
	}{
		{
			name:     "default canonical form",
			rawURL:   canonicalUserInfoURL,
			expected: canonicalUserInfoURLExpected,
		},
		{
			name:     "removes default port",
			rawURL:   "http://example.com:80/index.html",
			expected: "http://example.com/index.html",
		},
		{
			name:     "removes duplicate slashes",
			rawURL:   "http://example.com///x//////y///index.html",
			expected: "http://example.com/x/y/index.html",
		},
		{
			name:     "duplicate slashes keep trailing slash",
			rawURL:   "https://root//a//b///c////",
			expected: "https://root/a/b/c/",
		},
		{
			name:     "removes dot segments",
			rawURL:   "http://example.com/./x/y/z/../index.html",
			expected: "http://example.com/x/y/index.html",
		},
		{
			name:     "dot segment trailing current directory",
			rawURL:   "http://test.example/foo/bar/.",
			expected: "http://test.example/foo/bar/",
		},
		{
			name:     "dot segment trailing parent directory",
			rawURL:   "http://test.example/foo/bar/..",
			expected: "http://test.example/foo/",
		},
		{
			name:     "dot segment above root",
			rawURL:   "http://test.example/../foo",
			expected: "http://test.example/foo",
		},
		{
			name:     "sorts query parameters",
			rawURL:   "http://example.com/index.html?c=z&a=x&b=y",
			expected: "http://example.com/index.html?a=x&b=y&c=z",
		},
		{
			name:     "sorts duplicate query parameters",
			rawURL:   "http://root/toto/?b=4&a=1&c=3&b=2&a=5",
			expected: "http://root/toto/?a=1&a=5&b=2&b=4&c=3",
		},
		{
			name:     "sorts escaped query keys",
			rawURL:   "http://root/toto/?b=2&a%20key=1",
			expected: "http://root/toto/?a+key=1&b=2",
		},
		{
			name:     "leaves fragment as is",
			rawURL:   "http://example.com/index.html#t=20",
			expected: "http://example.com/index.html#t=20",
		},
		{
			name:     "canonicalises path and query example",
			rawURL:   "http://localhost:80///x///y/z/../././index.html?b=y&a=x#t=20",
			expected: "http://localhost/x/y/index.html?a=x&b=y#t=20",
		},
		{
			name:     "decodes punycode host",
			rawURL:   "http://xn--kda4b0koi.pl/%C5%BC%C3%B3%C5%82%C4%87.html",
			expected: "http://żółć.pl/%C5%BC%C3%B3%C5%82%C4%87.html",
		},
		{
			name:     "ignore scheme query and fragment",
			rawURL:   "https://example.com/path?a=1#frag",
			expected: "//example.com/path",
			options:  []NormalisationOption{IgnoreScheme(), IgnoreQuery(), IgnoreFragment()},
		},
		{
			name:     "without ignoring scheme preserves scheme",
			rawURL:   "https://example.com/path",
			expected: "https://example.com/path",
			options:  []NormalisationOption{IgnoreScheme(), WithoutIgnoringScheme()},
		},
		{
			name:     "ignore host",
			rawURL:   "https://example.com/path",
			expected: "https:/path",
			options:  []NormalisationOption{IgnoreHost()},
		},
		{
			name:     "without ignoring host preserves host",
			rawURL:   "https://example.com/path",
			expected: "https://example.com/path",
			options:  []NormalisationOption{IgnoreHost(), WithoutIgnoringHost()},
		},
		{
			name:     "ignore host and scheme for endpoint comparison",
			rawURL:   "https://example.com/path",
			expected: "/path",
			options:  []NormalisationOption{IgnoreScheme(), IgnoreHost()},
		},
		{
			name:     "ignore fragment only",
			rawURL:   "https://example.com/path#frag",
			expected: "https://example.com/path",
			options:  []NormalisationOption{IgnoreFragment()},
		},
		{
			name:     "without ignoring fragment preserves fragment",
			rawURL:   "https://example.com/path#frag",
			expected: "https://example.com/path#frag",
			options:  []NormalisationOption{IgnoreFragment(), WithoutIgnoringFragment()},
		},
		{
			name:     "ignore trailing slash",
			rawURL:   "https://example.com/path/",
			expected: "https://example.com/path",
			options:  []NormalisationOption{RemoveTrailingSlash()},
		},
		{
			name:     "remove user info",
			rawURL:   removeUserInfoURL,
			expected: "https://example.com/path",
			options:  []NormalisationOption{RemoveUserInfo()},
		},
		{
			name:     "remove www",
			rawURL:   "https://www.example.com/path",
			expected: "https://example.com/path",
			options:  []NormalisationOption{RemoveWWW()},
		},
		{
			name:     "remove www mixed case host",
			rawURL:   "https://WwW.Example.com/path",
			expected: "https://example.com/path",
			options:  []NormalisationOption{RemoveWWW()},
		},
		{
			name:     "keep non default port",
			rawURL:   "http://example.com:8080",
			expected: "http://example.com:8080",
		},
		{
			name:     "custom canonical form preserves query order and port",
			rawURL:   "https://Example.com:443/path?b=2&a=1",
			expected: "https://Example.com:443/path?b=2&a=1",
			options:  []NormalisationOption{WithoutClean(), WithoutDecode(), WithoutIgnoringDefaultPort(), WithoutLowercaseHost(), WithoutSortQuery()},
		},
		{
			name:     "ignore default port re enables default port removal",
			rawURL:   "https://Example.com:443/path",
			expected: "https://Example.com/path",
			options:  []NormalisationOption{WithoutClean(), WithoutDecode(), WithoutIgnoringDefaultPort(), IgnoreDefaultPort(), WithoutLowercaseHost(), WithoutSortQuery()},
		},
		{
			name:     "custom canonical form preserves scheme case",
			rawURL:   "HTTPS://Example.com/path",
			expected: "HTTPS://Example.com/path",
			options:  []NormalisationOption{WithoutClean(), WithoutDecode(), WithoutLowercaseScheme(), WithoutLowercaseHost(), WithoutSortQuery()},
		},
		{
			name:     "with lowercase scheme re enables lowercasing",
			rawURL:   "HTTPS://Example.com/path",
			expected: "https://Example.com/path",
			options:  []NormalisationOption{WithoutClean(), WithoutDecode(), WithoutLowercaseScheme(), WithLowercaseScheme(), WithoutLowercaseHost(), WithoutSortQuery()},
		},
		{
			name:     "with lowercase host re enables lowercasing",
			rawURL:   "https://Example.COM/path",
			expected: "https://example.com/path",
			options:  []NormalisationOption{WithoutClean(), WithoutDecode(), WithoutLowercaseHost(), WithLowercaseHost(), WithoutSortQuery()},
		},
		{
			name:     "without lower preserves scheme and host casing",
			rawURL:   "HTTPS://Example.COM/path",
			expected: "HTTPS://Example.COM/path",
			options:  []NormalisationOption{WithoutClean(), WithoutDecode(), WithoutLower(), WithoutSortQuery(), WithoutIgnoringDefaultPort()},
		},
		{
			name:     "case sensitive preserves scheme and host casing",
			rawURL:   "HTTPS://Example.COM/path",
			expected: "HTTPS://Example.COM/path",
			options:  []NormalisationOption{WithoutClean(), WithoutDecode(), CaseSensitive(), WithoutSortQuery(), WithoutIgnoringDefaultPort()},
		},
		{
			name:     "clean host path and query escapes",
			rawURL:   "https://Example..COM./a//b/./c/%7Euser/%41?b=%2f2&a=%7e1",
			expected: "https://example.com/a/b/c/~user/A?a=~1&b=%2F2",
			options:  []NormalisationOption{WithClean()},
		},
		{
			name:     "without clean preserves duplicate slashes and dot segments",
			rawURL:   "https://example.com/a//b/./c/../index.html",
			expected: "https://example.com/a//b/./c/../index.html",
			options:  []NormalisationOption{WithoutClean(), WithoutDecode(), WithoutSortQuery()},
		},
		{
			name:     "clean removes empty query values",
			rawURL:   "https://example.com/path?a=&b=2&c",
			expected: "https://example.com/path?b=2",
			options:  []NormalisationOption{WithClean()},
		},
		{
			name:     "clean removes empty query separator",
			rawURL:   "https://example.com/path?",
			expected: "https://example.com/path",
			options:  []NormalisationOption{WithClean()},
		},
		{
			name:     "clean removes empty port separator",
			rawURL:   "https://example.com:/path",
			expected: "https://example.com/path",
			options:  []NormalisationOption{WithClean()},
		},
		{
			name:     "clean removes empty port separator without path",
			rawURL:   "http://www.src.ca:",
			expected: "http://www.src.ca",
			options:  []NormalisationOption{WithClean()},
		},
		{
			name:     "clean host dot edge case with port",
			rawURL:   "http://www.foo.com.:81/foo",
			expected: "http://www.foo.com:81/foo",
			options:  []NormalisationOption{WithClean()},
		},
		{
			name:     "decode query escapes without full clean",
			rawURL:   "https://Example.com/path?b=%2f2&a=%7e1",
			expected: "https://example.com/path?a=~1&b=%2F2",
			options:  []NormalisationOption{WithoutClean(), WithDecode(), WithSortQuery()},
		},
		{
			name:     "clean encodes necessary spaces in path",
			rawURL:   "http://test.example/path/with a%20space+/",
			expected: "http://test.example/path/with%20a%20space+/",
			options:  []NormalisationOption{WithClean()},
		},
		{
			name:     "without decode preserves escape casing and escapes",
			rawURL:   "https://Example.com/path?b=%2f2&a=%7e1",
			expected: "https://example.com/path?b=%2f2&a=%7e1",
			options:  []NormalisationOption{WithoutClean(), WithoutDecode(), WithoutSortQuery()},
		},
		{
			name:     "without ignoring query preserves query",
			rawURL:   "https://example.com/path?a=1&b=2",
			expected: "https://example.com/path?a=1&b=2",
			options:  []NormalisationOption{IgnoreQuery(), WithoutIgnoringQuery()},
		},
		{
			name:   "empty URL",
			rawURL: "",
			err:    commonerrors.ErrUndefined,
		},
		{
			name:   "invalid URL",
			rawURL: "http://%zz",
			err:    commonerrors.ErrInvalid,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			result, err := NormaliseURL(test.rawURL, test.options...)
			errortest.AssertError(t, err, test.err)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestCompareURLs(t *testing.T) {
	tests := []struct {
		name    string
		left    string
		right   string
		match   bool
		options []NormalisationOption
		err     error
	}{
		{
			name:  "default canonical form matches equivalent URLs",
			left:  "https://example.com:443/path?b=2&a=1",
			right: "https://EXAMPLE.com/path?a=1&b=2",
			match: true,
		},
		{
			name:    "ignore scheme",
			left:    "https://example.com/path",
			right:   "http://example.com/path",
			match:   true,
			options: []NormalisationOption{IgnoreScheme()},
		},
		{
			name:    "ignore query",
			left:    "https://example.com/path?a=1",
			right:   "https://example.com/path?b=2",
			match:   true,
			options: []NormalisationOption{IgnoreQuery()},
		},
		{
			name:    "ignore trailing slash",
			left:    "https://example.com/path/",
			right:   "https://example.com/path",
			match:   true,
			options: []NormalisationOption{RemoveTrailingSlash()},
		},
		{
			name:    "remove www",
			left:    "https://www.example.com/path",
			right:   "https://example.com/path",
			match:   true,
			options: []NormalisationOption{RemoveWWW()},
		},
		{
			name:    "ignore scheme compares URLs with different schemes",
			left:    "https://example.com/path",
			right:   "http://example.com/path",
			match:   true,
			options: []NormalisationOption{IgnoreScheme()},
		},
		{
			name:    "ignore scheme compares URL with schemeless authority",
			left:    "https://example.com/path",
			right:   "example.com/path",
			match:   true,
			options: []NormalisationOption{IgnoreScheme()},
		},
		{
			name:    "ignore scheme and host compares URL with endpoint",
			left:    "https://api.example.com/v1/users",
			right:   "/v1/users",
			match:   true,
			options: []NormalisationOption{IgnoreScheme(), IgnoreHost()},
		},
		{
			name:    "ignore host compares absolute URLs with different hosts",
			left:    "https://api.example.com/v1/users",
			right:   "https://other.example.com/v1/users",
			match:   true,
			options: []NormalisationOption{IgnoreHost()},
		},
		{
			name:  "punycode and unicode host equivalence",
			left:  "http://xn--kda4b0koi.pl/path",
			right: "http://żółć.pl/path",
			match: true,
		},
		{
			name:  "different URLs remain different",
			left:  "https://example.com/a",
			right: "https://example.com/b",
			match: false,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			match, err := CompareURLs(test.left, test.right, test.options...)
			errortest.AssertError(t, err, test.err)
			assert.Equal(t, test.match, match)
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name    string
		left    string
		right   string
		result  int
		options []NormalisationOption
		err     error
	}{
		{
			name:   "equal after normalisation",
			left:   "https://example.com:443/path?b=2&a=1",
			right:  "https://EXAMPLE.com/path?a=1&b=2",
			result: 0,
		},
		{
			name:   "left sorts before right",
			left:   "https://example.com/a",
			right:  "https://example.com/b",
			result: -1,
		},
		{
			name:    "url and endpoint compare equally",
			left:    "https://api.example.com/v1/users",
			right:   "/v1/users",
			result:  0,
			options: []NormalisationOption{IgnoreScheme(), IgnoreHost()},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			result, err := Compare(test.left, test.right, test.options...)
			errortest.AssertError(t, err, test.err)
			assert.Equal(t, test.result, result)
		})
	}
}
