package url

import (
	"net"
	netUrl "net/url"
	stdpath "path"
	"slices"
	"strings"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"golang.org/x/net/idna"
)

var defaultSchemePorts = map[string]string{
	"ftp":   "21",
	"http":  "80",
	"https": "443",
	"ws":    "80",
	"wss":   "443",
}

// NormalisationOptions controls how a URL is canonicalised before comparison.
//
// The available switches are loosely inspired by the purell library.
//
// References:
//   - RFC 3986: https://datatracker.ietf.org/doc/html/rfc3986
//   - URL normalisation overview: https://en.wikipedia.org/wiki/URL_normalization
//   - purell: https://github.com/PuerkitoBio/purell
//   - purell flags: https://pkg.go.dev/github.com/PuerkitoBio/purell
//   - urlx: https://github.com/goware/urlx
//   - urlx defaults reference: https://github.com/goware/urlx/blob/dcd04f6df527b011eae76fadc25a5e957294722b/urlx.go#L130-L135
type NormalisationOptions struct {
	Clean               bool
	Decode              bool
	IgnoreHost          bool
	IgnoreScheme        bool
	IgnoreQuery         bool
	IgnoreFragment      bool
	IgnoreTrailingSlash bool
	IgnoreDefaultPort   bool
	LowercaseScheme     bool
	LowercaseHost       bool
	RemoveWWW           bool
	SortQuery           bool
	RemoveUserInfo      bool
}

// NormalisationOption configures [NormalisationOptions].
type NormalisationOption func(*NormalisationOptions) *NormalisationOptions

// DefaultCanonicalForm returns the package's default URL canonical form.
//
// The defaults are aligned broadly with goware/urlx's normalisation behaviour:
// schemes and host names are lower-cased, default ports are removed, query
// parameters are sorted, escape sequences are normalised, and the URL is passed
// through the package's clean-up pass.
//
// Reference:
//   - https://github.com/goware/urlx/blob/dcd04f6df527b011eae76fadc25a5e957294722b/urlx.go#L130-L135
func DefaultCanonicalForm() *NormalisationOptions {
	return &NormalisationOptions{
		Clean:             true,
		Decode:            true,
		IgnoreDefaultPort: true,
		LowercaseScheme:   true,
		LowercaseHost:     true,
		SortQuery:         true,
	}
}

// WithCanonicalOptions materialises the effective canonical-form options after
// applying the supplied functional options to the package defaults.
//
// Example:
//
//	options := WithCanonicalOptions(WithoutClean(), IgnoreFragment(), RemoveWWW())
//	// options.Clean == false, options.IgnoreFragment == true, options.RemoveWWW == true
func WithCanonicalOptions(options ...NormalisationOption) *NormalisationOptions {
	form := DefaultCanonicalForm()
	collection.ForEach(options, func(option NormalisationOption) {
		if option != nil {
			form = option(form)
		}
	})
	return form
}

// WithClean applies a clean-up pass broadly equivalent to purell's
// `FlagRemoveUnnecessaryHostDots`, `FlagRemoveDotSegments`,
// `FlagRemoveDuplicateSlashes`, `FlagUppercaseEscapes`,
// `FlagDecodeUnnecessaryEscapes`, `FlagEncodeNecessaryEscapes`,
// `FlagRemoveEmptyQuerySeparator`, and `FlagRemoveEmptyPortSeparator`, plus
// removal of empty query values.
func WithClean() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.Clean = true
		return options
	}
}

// WithoutClean disables the clean-up pass enabled by default.
func WithoutClean() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.Clean = false
		return options
	}
}

// WithDecode applies a decode-oriented pass roughly equivalent to purell's
// `FlagUppercaseEscapes`, `FlagDecodeUnnecessaryEscapes`, and
// `FlagEncodeNecessaryEscapes`.
func WithDecode() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.Decode = true
		return options
	}
}

// WithoutDecode disables the decode-oriented pass enabled by default.
func WithoutDecode() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.Decode = false
		return options
	}
}

// IgnoreScheme removes the URL scheme from the canonical form.
func IgnoreScheme() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.IgnoreScheme = true
		return options
	}
}

// WithoutIgnoringScheme preserves the URL scheme in the canonical form.
func WithoutIgnoringScheme() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.IgnoreScheme = false
		return options
	}
}

// IgnoreHost removes the URL host from the canonical form.
func IgnoreHost() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.IgnoreHost = true
		return options
	}
}

// WithoutIgnoringHost preserves the URL host in the canonical form.
func WithoutIgnoringHost() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.IgnoreHost = false
		return options
	}
}

// IgnoreQuery removes the query string from the canonical form.
func IgnoreQuery() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.IgnoreQuery = true
		return options
	}
}

// WithoutIgnoringQuery preserves the query string in the canonical form.
func WithoutIgnoringQuery() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.IgnoreQuery = false
		return options
	}
}

// IgnoreFragment removes the fragment from the canonical form.
func IgnoreFragment() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.IgnoreFragment = true
		return options
	}
}

// WithoutIgnoringFragment preserves the fragment in the canonical form.
func WithoutIgnoringFragment() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.IgnoreFragment = false
		return options
	}
}

// RemoveTrailingSlash strips trailing slashes from non-root paths.
func RemoveTrailingSlash() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.IgnoreTrailingSlash = true
		return options
	}
}

// IgnoreDefaultPort removes the default port for recognised schemes.
func IgnoreDefaultPort() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.IgnoreDefaultPort = true
		return options
	}
}

// WithoutIgnoringDefaultPort preserves explicit default ports.
func WithoutIgnoringDefaultPort() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.IgnoreDefaultPort = false
		return options
	}
}

// WithLowercaseScheme lower-cases the URL scheme in the canonical form.
func WithLowercaseScheme() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.LowercaseScheme = true
		return options
	}
}

// WithoutLowercaseScheme preserves the original scheme casing.
func WithoutLowercaseScheme() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.LowercaseScheme = false
		return options
	}
}

// WithLowercaseHost lower-cases the URL host in the canonical form.
func WithLowercaseHost() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.LowercaseHost = true
		return options
	}
}

// WithoutLowercaseHost preserves the original host casing.
func WithoutLowercaseHost() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.LowercaseHost = false
		return options
	}
}

// WithoutLower preserves the original scheme and host casing.
func WithoutLower() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.LowercaseScheme = false
		options.LowercaseHost = false
		return options
	}
}

// CaseSensitive preserves the original scheme and host casing.
func CaseSensitive() NormalisationOption {
	return WithoutLower()
}

// RemoveWWW removes a leading `www.` label from the URL host.
func RemoveWWW() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.RemoveWWW = true
		return options
	}
}

// WithSortQuery sorts query parameters in the canonical form.
func WithSortQuery() NormalisationOption {
	return withQuery(true)
}

// WithoutSortQuery preserves the original query parameter order.
func WithoutSortQuery() NormalisationOption {
	return withQuery(false)
}

func withQuery(sort bool) NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.IgnoreQuery = false
		options.SortQuery = sort
		return options
	}
}

// RemoveUserInfo removes any user info from the canonical form.
func RemoveUserInfo() NormalisationOption {
	return func(options *NormalisationOptions) *NormalisationOptions {
		if options == nil {
			options = DefaultCanonicalForm()
		}
		options.RemoveUserInfo = true
		return options
	}
}

// NormaliseURL canonicalises rawURL according to the supplied options.
//
// The behaviour is inspired by github.com/PuerkitoBio/purell.
//
// Example:
//
//	normalised, err := NormaliseURL("https://www.example.com:443/path?a=1&b=2", RemoveWWW())
//	// normalised == "https://example.com/path?a=1&b=2"
//
//	normalised, err = NormaliseURL("https://api.example.com/v1/users", IgnoreScheme(), IgnoreHost())
//	// normalised == "/v1/users"
//
// References:
//   - RFC 3986: https://datatracker.ietf.org/doc/html/rfc3986
//   - URL normalisation overview: https://en.wikipedia.org/wiki/URL_normalization
func NormaliseURL(rawURL string, options ...NormalisationOption) (normalised string, err error) {
	if reflection.IsEmpty(rawURL) {
		err = commonerrors.UndefinedVariable("URL")
		return
	}
	err = IsURI.Validate(rawURL)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "invalid URL %q", rawURL)
		return
	}
	var parsed *netUrl.URL
	parsed, err = parseCanonicalURL(rawURL)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "failed to parse URL %q", rawURL)
		return
	}

	form := WithCanonicalOptions(options...)
	if !form.LowercaseScheme {
		parsed.Scheme = originalScheme(rawURL, parsed.Scheme)
	}
	if form.LowercaseScheme {
		parsed.Scheme = strings.ToLower(parsed.Scheme)
	}
	parsed.Host = canonicalHost(parsed, form)
	parsed.Path = canonicalPath(parsed.Path, form)
	parsed.RawPath = ""

	if form.RemoveUserInfo {
		parsed.User = nil
	}
	if form.IgnoreHost {
		parsed.Host = ""
		parsed.User = nil
	}
	if form.IgnoreScheme {
		parsed.Scheme = ""
	}
	if form.IgnoreQuery {
		parsed.RawQuery = ""
		parsed.ForceQuery = false
	} else if form.SortQuery || form.Clean || form.Decode {
		rawQuery, subErr := canonicalQuery(parsed.RawQuery, form.SortQuery, form.Clean)
		if subErr != nil {
			err = commonerrors.WrapErrorf(commonerrors.ErrInvalid, subErr, "failed to canonicalise query for URL %q", rawURL)
			return
		}
		parsed.RawQuery = rawQuery
	}
	if form.Clean && reflection.IsEmpty(parsed.RawQuery) {
		parsed.ForceQuery = false
	}
	if form.IgnoreFragment {
		parsed.Fragment = ""
	}

	normalised = formatCanonicalURL(parsed)
	return
}

// CompareURLs canonicalises both URLs with the same normalisation options and
// reports whether the resulting forms are equal.
//
// Example:
//
//	match, err := CompareURLs("https://www.example.com/path", "https://example.com/path", RemoveWWW())
//	// match == true
//
//	match, err = CompareURLs("https://api.example.com/v1/users", "/v1/users", IgnoreScheme(), IgnoreHost())
//	// match == true
//
// References:
//   - RFC 3986: https://datatracker.ietf.org/doc/html/rfc3986
//   - URL normalisation overview: https://en.wikipedia.org/wiki/URL_normalization
func CompareURLs(url1, url2 string, options ...NormalisationOption) (match bool, err error) {
	left, err := NormaliseURL(url1, options...)
	if err != nil {
		return
	}
	right, err := NormaliseURL(url2, options...)
	if err != nil {
		return
	}
	match = left == right
	return
}

func canonicalHost(parsed *netUrl.URL, form *NormalisationOptions) string {
	hostname := parsed.Hostname()
	port := parsed.Port()
	if unicodeHost, err := idna.ToUnicode(hostname); err == nil {
		hostname = unicodeHost
	}
	if form != nil && form.Clean {
		hostname = cleanHostDots(hostname)
	}
	if form != nil && form.LowercaseHost {
		hostname = strings.ToLower(hostname)
	}
	if form != nil && form.RemoveWWW {
		hostname = strings.TrimPrefix(hostname, "www.")
	}
	if form != nil && form.IgnoreDefaultPort {
		if defaultPort, ok := defaultSchemePorts[strings.ToLower(parsed.Scheme)]; ok && port == defaultPort {
			port = ""
		}
	}
	if reflection.IsEmpty(port) {
		return hostname
	}
	return net.JoinHostPort(hostname, port)
}

func canonicalPath(urlPath string, form *NormalisationOptions) string {
	if reflection.IsEmpty(urlPath) {
		return urlPath
	}
	hasTrailingSlash := (strings.HasSuffix(urlPath, defaultPathSeparator) || strings.HasSuffix(urlPath, "/.") || strings.HasSuffix(urlPath, "/..")) && urlPath != defaultPathSeparator
	if form != nil && form.Clean {
		urlPath = stdpath.Clean(urlPath)
		if urlPath == "." {
			urlPath = ""
		}
		if hasTrailingSlash && reflection.IsNotEmpty(urlPath) && urlPath != defaultPathSeparator && !form.IgnoreTrailingSlash {
			urlPath += defaultPathSeparator
		}
	}
	if form == nil || !form.IgnoreTrailingSlash || urlPath == defaultPathSeparator {
		return urlPath
	}
	trimmed := strings.TrimRight(urlPath, defaultPathSeparator)
	if reflection.IsEmpty(trimmed) {
		return defaultPathSeparator
	}
	return trimmed
}

func canonicalQuery(rawQuery string, sort, removeEmptyValues bool) (query string, err error) {
	if reflection.IsEmpty(rawQuery) {
		return
	}
	parts := strings.Split(rawQuery, "&")
	type queryPart struct {
		key      string
		value    string
		hasValue bool
	}
	parsed, err := collection.MapWithError(parts, func(part string) (queryPart, error) {
		keyPart, valuePart, hasValue := strings.Cut(part, "=")
		key, err := netUrl.QueryUnescape(keyPart)
		if err != nil {
			return queryPart{}, err
		}
		value := ""
		if hasValue {
			value, err = netUrl.QueryUnescape(valuePart)
			if err != nil {
				return queryPart{}, err
			}
		}
		return queryPart{key: key, value: value, hasValue: hasValue}, nil
	})
	if err != nil {
		return
	}
	if removeEmptyValues {
		parsed = collection.Filter(parsed, func(part queryPart) bool {
			return reflection.IsNotEmpty(part.value)
		})
	}
	if sort {
		slices.SortFunc(parsed, func(left, right queryPart) int {
			if comparison := strings.Compare(left.key, right.key); comparison != 0 {
				return comparison
			}
			if comparison := strings.Compare(left.value, right.value); comparison != 0 {
				return comparison
			}
			if left.hasValue == right.hasValue {
				return 0
			}
			if left.hasValue {
				return -1
			}
			return 1
		})
	}
	encoded := collection.Map(parsed, func(part queryPart) string {
		entry := netUrl.QueryEscape(part.key)
		if !part.hasValue {
			return entry
		}
		return entry + "=" + netUrl.QueryEscape(part.value)
	})
	query = strings.Join(encoded, "&")
	return
}

func cleanHostDots(hostname string) string {
	if reflection.IsEmpty(hostname) {
		return hostname
	}
	return strings.Join(collection.Filter(strings.Split(hostname, "."), func(part string) bool {
		return reflection.IsNotEmpty(part)
	}), ".")
}

func originalScheme(rawURL, fallback string) string {
	separatorIndex := strings.Index(rawURL, ":")
	if separatorIndex <= 0 {
		return fallback
	}
	return rawURL[:separatorIndex]
}

func parseCanonicalURL(rawURL string) (parsed *netUrl.URL, err error) {
	parsed, err = netUrl.Parse(rawURL)
	if err != nil {
		return
	}
	if collection.AnyTrue(
		reflection.IsNotEmpty(parsed.Scheme),
		reflection.IsNotEmpty(parsed.Host),
		reflection.IsEmpty(parsed.Path),
		strings.HasPrefix(parsed.Path, defaultPathSeparator),
		!looksLikeAuthority(parsed.Path),
	) {
		return
	}
	parsed, err = netUrl.Parse("//" + rawURL)
	return
}

func looksLikeAuthority(value string) bool {
	segment := value
	if separatorIndex := strings.Index(segment, defaultPathSeparator); separatorIndex >= 0 {
		segment = segment[:separatorIndex]
	}
	hasAuthorityMarker := collection.AnyFunc([]string{".", ":", "@"}, func(token string) bool {
		return strings.Contains(segment, token)
	})
	return collection.AnyTrue(
		hasAuthorityMarker,
		segment == "localhost",
		strings.HasPrefix(segment, "["),
	)
}

func formatCanonicalURL(parsed *netUrl.URL) string {
	if parsed == nil {
		return ""
	}
	var builder strings.Builder
	if reflection.IsNotEmpty(parsed.Scheme) {
		builder.WriteString(parsed.Scheme)
		builder.WriteString(":")
	}
	if reflection.IsNotEmpty(parsed.Host) {
		builder.WriteString("//")
		if parsed.User != nil {
			builder.WriteString(parsed.User.String())
			builder.WriteString("@")
		}
		builder.WriteString(parsed.Host)
	}
	path := parsed.EscapedPath()
	if reflection.IsEmpty(path) && reflection.IsNotEmpty(parsed.Host) && (reflection.IsNotEmpty(parsed.RawQuery) || reflection.IsNotEmpty(parsed.Fragment)) {
		path = defaultPathSeparator
	}
	builder.WriteString(path)
	if parsed.ForceQuery || reflection.IsNotEmpty(parsed.RawQuery) {
		builder.WriteString("?")
		builder.WriteString(parsed.RawQuery)
	}
	if reflection.IsNotEmpty(parsed.Fragment) {
		builder.WriteString("#")
		builder.WriteString(parsed.EscapedFragment())
	}
	return builder.String()
}
