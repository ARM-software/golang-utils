package headers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-http-utils/headers"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/encoding/base64"
	"github.com/ARM-software/golang-utils/utils/http/headers/useragent"
	"github.com/ARM-software/golang-utils/utils/http/schemes"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

const (
	HeaderWebsocketProtocol   = "Sec-WebSocket-Protocol" //https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Sec-WebSocket-Protocol
	HeaderWebsocketVersion    = "Sec-WebSocket-Version"
	HeaderWebsocketKey        = "Sec-WebSocket-Key"
	HeaderWebsocketAccept     = "Sec-WebSocket-Accept"
	HeaderWebsocketExtensions = "Sec-WebSocket-Extensions"
	HeaderConnection          = "Connection"
	HeaderVersion             = "Version"
	HeaderAcceptVersion       = "Accept-Version"
	HeaderHost                = "Host" // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/host
	// https://greenbytes.de/tech/webdav/draft-ietf-httpapi-deprecation-header-latest.html#sunset
	HeaderSunset      = "Sunset"      // https://datatracker.ietf.org/doc/html/rfc8594
	HeaderDeprecation = "Deprecation" // https://datatracker.ietf.org/doc/html/draft-ietf-httpapi-deprecation-header-02
	HeaderLink        = headers.Link  // https://datatracker.ietf.org/doc/html/rfc8288
	// TUS Headers https://tus.io/protocols/resumable-upload#headers
	HeaderUploadOffset        = "Upload-Offset"
	HeaderTusVersion          = "Tus-Version"
	HeaderUploadLength        = "Upload-Length"
	HeaderTusResumable        = "Tus-Resumable"
	HeaderTusExtension        = "Tus-Extension"
	HeaderTusMaxSize          = "Tus-Max-Size"
	HeaderXHTTPMethodOverride = "X-HTTP-Method-Override"
	HeaderXWWWFormURLEncoded  = "application/x-www-form-urlencoded"
	HeaderContentType         = "Content-Type"
)

var (
	// SafeHeaders corresponds to headers which do not store personal data.
	SafeHeaders = []string{
		HeaderVersion,
		HeaderAcceptVersion,
		HeaderHost,
		HeaderSunset,
		HeaderDeprecation,
		HeaderLink,
		HeaderWebsocketVersion,
		HeaderWebsocketAccept,
		HeaderWebsocketExtensions,
		HeaderConnection,
		HeaderUploadOffset,
		HeaderTusVersion,
		HeaderUploadLength,
		HeaderTusResumable,
		HeaderTusExtension,
		HeaderTusMaxSize,
		HeaderXHTTPMethodOverride,
		headers.Accept,
		headers.AcceptCharset,
		headers.AcceptEncoding,
		headers.AcceptLanguage,
		headers.CacheControl,
		headers.ContentLength,
		headers.ContentMD5,
		headers.ContentType,
		headers.DoNotTrack,
		headers.IfMatch,
		headers.IfModifiedSince,
		headers.IfNoneMatch,
		headers.IfRange,
		headers.IfUnmodifiedSince,
		headers.MaxForwards,
		headers.Pragma,
		headers.Range,
		headers.Referer,
		headers.UserAgent,
		headers.TE,
		headers.Via,
		headers.Warning,
		headers.AcceptDatetime,
		headers.XRequestedWith,
		headers.AccessControlAllowOrigin,
		headers.AccessControlAllowMethods,
		headers.AccessControlAllowHeaders,
		headers.AccessControlAllowCredentials,
		headers.AccessControlExposeHeaders,
		headers.AccessControlMaxAge,
		headers.AccessControlRequestMethod,
		headers.AccessControlRequestHeaders,
		headers.AcceptPatch,
		headers.AcceptRanges,
		headers.Allow,
		headers.ContentEncoding,
		headers.ContentLanguage,
		headers.ContentLocation,
		headers.ContentDisposition,
		headers.ContentRange,
		headers.ETag,
		headers.Expires,
		headers.LastModified,
		headers.Link,
		headers.Location,
		headers.P3P,
		headers.ProxyAuthenticate,
		headers.Refresh,
		headers.RetryAfter,
		headers.Server,
		headers.TransferEncoding,
		headers.Upgrade,
		headers.Vary,
		headers.XPoweredBy,
		headers.XHTTPMethodOverride,
		headers.XRatelimitLimit,
		headers.XRatelimitRemaining,
		headers.XRatelimitReset,
	}
)

type Header struct {
	Key   string
	Value string
}

func (h *Header) String() string {
	return fmt.Sprintf("%v: %v", h.Key, h.Value)
}

type Headers map[string]Header

func (hs Headers) AppendHeader(key, value string) {
	hs.Append(&Header{
		Key:   key,
		Value: value,
	})
}

func (hs Headers) Append(h *Header) {
	hs[h.Key] = *h
}

func (hs Headers) Has(h *Header) bool {
	if h == nil {
		return false
	}
	return hs.HasHeader(h.Key)
}

func (hs Headers) HasHeader(key string) bool {
	_, found := hs[key]
	return found
}

func (hs Headers) Empty() bool {
	return len(hs) == 0
}

func (hs Headers) AppendToResponse(w http.ResponseWriter) {
	if hs != nil && !hs.Empty() {
		for k, v := range hs {
			w.Header().Set(k, v.Value)
		}
	}
}

func (hs Headers) AppendToRequest(r *http.Request) {
	if hs != nil && !hs.Empty() {
		for k, v := range hs {
			r.Header.Set(k, v.Value)
		}
	}
}

func NewHeaders() *Headers {
	return &Headers{}
}

// ParseAuthorizationHeader fetches the `Authorization` header and parses it.
func ParseAuthorizationHeader(r *http.Request) (string, string, error) {
	return ParseAuthorisationValue(FetchWebsocketAuthorisation(r))
}

// ParseAuthorisationValue determines the different element of a `Authorization` header value.
// and makes sure it has 2 parts  <scheme> <token>
func ParseAuthorisationValue(authHeader string) (scheme string, token string, err error) {
	if reflection.IsEmpty(authHeader) {
		err = commonerrors.New(commonerrors.ErrUndefined, "authorization header is not set")
		return
	}
	parts := strings.Fields(authHeader)
	if len(parts) != 2 {
		err = commonerrors.New(commonerrors.ErrInvalid, "`Authorization` header contains incorrect number of parts")
		return
	}
	scheme = parts[0]
	token = parts[1]
	err = checkSchemeSupport(scheme)
	return
}

func checkSchemeSupport(scheme string) (err error) {
	schemeStr := strings.TrimSpace(scheme)
	if schemeStr == "" {
		err = commonerrors.UndefinedVariable("authorisation scheme")
		return
	}
	_, found := collection.FindInSlice(false, schemes.HTTPAuthorisationSchemes, scheme)
	if !found {
		err = commonerrors.Newf(commonerrors.ErrUnsupported, "supported `Authorization` schemes are %v", schemes.HTTPAuthorisationSchemes)
	}
	return err
}

// FetchAuthorisation fetches the value of `Authorization` header.
func FetchAuthorisation(r *http.Request) string {
	if r == nil {
		return ""
	}
	authHeader := r.Header.Get(headers.Authorization)
	return authHeader
}

// FetchWebSocketSubProtocols fetches the values of `Sec-WebSocket-Protocol` header https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Sec-WebSocket-Protocol.
func FetchWebSocketSubProtocols(r *http.Request) (subProtocols []string) {
	if r == nil {
		return
	}
	subProtocolsHeaders := r.Header.Values(HeaderWebsocketProtocol)
	if len(subProtocolsHeaders) == 0 {
		return
	}
	for i := range subProtocolsHeaders {
		subProtocols = append(subProtocols, collection.ParseCommaSeparatedList(subProtocolsHeaders[i])...)
	}
	return
}

// FetchWebsocketAuthorisation tries to find the authorisation header values in the case of websocket
// It will look in the `Authorization` header but will also look at some workaround suggested [here](https://ably.com/blog/websocket-authentication#:~:text=While%20the%20WebSocket%20browser%20API,token%20in%20the%20request%20header) and [there](https://github.com/kubernetes/kubernetes/pull/47740)
// If found using the workarounds, it will set the Authorization header with the determined value
func FetchWebsocketAuthorisation(r *http.Request) (authorisationHeader string) {
	if r == nil {
		return
	}
	authorisationHeader = FetchAuthorisation(r)
	if !reflection.IsEmpty(authorisationHeader) {
		return
	}
	subProtocols := FetchWebSocketSubProtocols(r)
	if len(subProtocols) == 0 {
		return
	}
	i, found := collection.FindInSlice(false, subProtocols, headers.Authorization)
	if found {
		if i < len(subProtocols)-1 {
			authorisationHeader = subProtocols[i+1]
			if decoded, err := base64.DecodeString(context.Background(), authorisationHeader); err == nil {
				authorisationHeader = decoded
			}
			_ = SetAuthorisationIfNotPresent(r, authorisationHeader)
			return
		}
	}
	// see https://github.com/kubernetes/kubernetes/pull/47740
	_, found = collection.FindInSlice(false, subProtocols, "base64.binary.k8s.io")
	if found {
		for j := range subProtocols {
			token := strings.TrimPrefix(subProtocols[j], "base64url.bearer.authorization.k8s.io.")
			if token != subProtocols[j] {
				data, err := base64.DecodeString(context.Background(), token)
				if err == nil {
					authorisationHeader = data
					_ = SetAuthorisationIfNotPresent(r, authorisationHeader)
					return
				}
			}
		}

	}
	return
}

// SetAuthorisationIfNotPresent sets the value of the `Authorization` header if not already set.
func SetAuthorisationIfNotPresent(r *http.Request, authorisation string) (err error) {
	if strings.TrimSpace(FetchAuthorisation(r)) == "" {
		err = SetAuthorisation(r, authorisation)
	}
	return
}

// SetAuthorisation sets the value of the `Authorization` header.
func SetAuthorisation(r *http.Request, authorisation string) (err error) {
	if r == nil {
		err = commonerrors.UndefinedVariable("request")
		return
	}
	if reflection.IsEmpty(authorisation) {
		err = commonerrors.UndefinedVariable("authorisation value")
		return
	}
	r.Header.Set(headers.Authorization, authorisation)
	return
}

// SetAuthorisationToken defines the `Authorization` header.
func SetAuthorisationToken(r *http.Request, scheme, token string) (err error) {
	value, err := GenerateAuthorizationHeaderValue(scheme, token)
	if err != nil {
		return
	}
	err = SetAuthorisation(r, value)
	return
}

func GenerateAuthorizationHeaderValue(scheme string, token string) (value string, err error) {
	err = checkSchemeSupport(scheme)
	if err != nil {
		return
	}
	if reflection.IsEmpty(token) {
		err = commonerrors.UndefinedVariable("authorisation token")
		return
	}
	value = fmt.Sprintf("%s %s", strings.TrimSpace(scheme), token)
	return
}

// AddToUserAgent adds some information to the `User Agent`.
func AddToUserAgent(r *http.Request, elements ...string) (err error) {
	if r == nil {
		err = fmt.Errorf("%w: missing request", commonerrors.ErrUndefined)
		return
	}
	if reflection.IsEmpty(elements) {
		err = fmt.Errorf("%w: empty elements to add", commonerrors.ErrUndefined)
		return
	}
	r.Header.Set(headers.UserAgent, useragent.AddValuesToUserAgent(FetchUserAgent(r), elements...))
	return
}

// AddProductInformationToUserAgent adds some product information to the `User Agent`.
func AddProductInformationToUserAgent(r *http.Request, product, productVersion, comment string) (err error) {
	productStr, err := useragent.GenerateUserAgentValue(product, productVersion, comment)
	if err != nil {
		return
	}
	err = AddToUserAgent(r, productStr)
	return
}

// FetchUserAgent fetches the value of the `User-Agent` header.
func FetchUserAgent(r *http.Request) string {
	authHeader := r.UserAgent()
	return authHeader
}

// SetLocationHeaders sets the location errors for `POST` requests.
func SetLocationHeaders(w http.ResponseWriter, location string) {
	h := NewHeaders()
	h.AppendHeader(headers.Location, location)
	h.AppendHeader(headers.ContentLocation, location)
	h.AppendToResponse(w)
}

// SetContentLocationHeader sets the `Content-Location` header
func SetContentLocationHeader(w http.ResponseWriter, location string) {
	w.Header().Set(headers.ContentLocation, location)
}

// CreateLinkHeader creates a link header for a relation and mimetype
func CreateLinkHeader(link, relation, contentType string) string {
	return fmt.Sprintf("<%v>; rel=\"%v\"; type=\"%v\"", link, relation, contentType)
}

// SanitiseHeaders sanitises a collection of request headers not to include any with personal data
func SanitiseHeaders(requestHeader *http.Header) *Headers {
	if requestHeader == nil {
		return nil
	}
	aHeaders := NewHeaders()
	for i := range SafeHeaders {
		safeHeader := SafeHeaders[i]
		rHeader := requestHeader.Get(safeHeader)
		if !reflection.IsEmpty(rHeader) {
			aHeaders.AppendHeader(safeHeader, rHeader)
		}
	}

	return aHeaders
}
