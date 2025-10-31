package proxy

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-http-utils/headers"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	httpheaders "github.com/ARM-software/golang-utils/utils/http/headers"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/safecast"
	"github.com/ARM-software/golang-utils/utils/safeio"
)

// ProxyDisallowList describes headers which are not proxied back.
var ProxyDisallowList = []string{
	headers.AccessControlAllowOrigin,
	headers.AccessControlAllowMethods,
	headers.AccessControlAllowHeaders,
	headers.AccessControlExposeHeaders,
	headers.AccessControlMaxAge,
	headers.AccessControlAllowCredentials,
}

// ProxyRequest proxies a request to a new endpoint. The method can also be changed. Headers are sanitised during the process.
func ProxyRequest(r *http.Request, proxyMethod, endpoint string) (proxiedRequest *http.Request, err error) {
	if reflection.IsEmpty(r) {
		err = commonerrors.UndefinedVariable("request to proxy")
		return
	}
	ctx := r.Context()
	contentLength := determineRequestContentLength(r)
	h := httpheaders.FromRequest(r).AllowList(headers.Authorization)
	if reflection.IsEmpty(proxyMethod) {
		proxyMethod = http.MethodGet
	}
	proxiedRequest, err = http.NewRequestWithContext(ctx, proxyMethod, endpoint, r.Body)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not create a proxied request")
		return
	}

	if proxiedRequest.ContentLength <= 0 {
		if proxiedRequest.Body == nil || proxiedRequest.Body == http.NoBody {
			if contentLength > 0 {
				proxiedRequest.Body = r.Body
				proxiedRequest.GetBody = r.GetBody
			} else {
				proxiedRequest, err = http.NewRequestWithContext(ctx, proxyMethod, endpoint, convertBody(ctx, r.Body))
				if err != nil {
					err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not create a proxied request")
					return
				}
			}
		}

		if contentLength <= 0 {
			proxiedRequest, err = http.NewRequestWithContext(ctx, proxyMethod, endpoint, convertBody(ctx, r.Body))
			if err != nil {
				err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not create a proxied request")
				return
			}
		}

		if contentLength > 0 && proxiedRequest.ContentLength <= 0 {
			proxiedRequest.ContentLength = contentLength
			h.AppendHeader(headers.ContentLength, strconv.FormatInt(contentLength, 10))
		}
	}
	if contentLength > 0 && contentLength != proxiedRequest.ContentLength {
		err = commonerrors.Newf(commonerrors.ErrUnexpected, "proxied request does not have the same content length `%v` as original request `%v`", proxiedRequest.ContentLength, contentLength)
		return
	}
	h.AppendToRequest(proxiedRequest)
	return
}

func determineRequestContentLength(r *http.Request) int64 {
	if reflection.IsEmpty(r) {
		return -1
	}
	if r.ContentLength > 0 {
		return r.ContentLength
	}
	// Following what was done in https://github.com/luraproject/lura/blob/b9ad9ab654dd6149aeb58a5d6ffe731aba41717e/proxy/http.go#L99C1-L105C4
	v := r.Header.Values(headers.ContentLength)
	if len(v) == 1 && v[0] != "chunked" {
		if size, err := strconv.Atoi(v[0]); err == nil {
			return safecast.ToInt64(size)
		}
	}
	return -1
}

func convertBody(_ context.Context, body io.Reader) io.Reader {
	if body == nil || body == http.NoBody {
		return http.NoBody
	}
	switch v := body.(type) {
	case *bytes.Buffer:
		return body
	case *bytes.Reader:
		return body
	case *strings.Reader:
		return body
	default:
		// see example https://github.com/luraproject/lura/blob/b9ad9ab654dd6149aeb58a5d6ffe731aba41717e/proxy/http.go#L73
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(v)
		if err != nil {
			return http.NoBody
		}
		if b, ok := body.(io.ReadCloser); ok {
			_ = b.Close()
		}
		return buf
	}
}

// ProxyResponse proxies a response to a writer. Headers are sanitised and some headers such as CORS headers will be removed from the response.
func ProxyResponse(ctx context.Context, resp *http.Response, w http.ResponseWriter) (err error) {
	if w == nil {
		err = commonerrors.UndefinedVariable("response writer")
		return
	}
	if reflection.IsEmpty(resp) {
		err = commonerrors.UndefinedVariable("response")
		return
	}
	h := httpheaders.FromResponse(resp)
	h.Sanitise()

	var written int64
	_, err = safeio.CopyDataWithContext(ctx, resp.Body, w)
	if resp.Body != nil && resp.Body != http.NoBody {
		written, err = safeio.CopyDataWithContext(ctx, resp.Body, w)
		if err != nil {
			err = commonerrors.DescribeCircumstance(err, "failed copying response body")
		}
	}
	if written >= 0 {
		h.AppendHeader(headers.ContentLength, strconv.FormatInt(written, 10))
	}
	h.RemoveHeaders(ProxyDisallowList...)
	h.AppendToResponse(w)
	w.WriteHeader(resp.StatusCode)
	return
}
