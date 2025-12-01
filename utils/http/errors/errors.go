package errors

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// ExtractAPIErrorDescriptionFunc defines a function which can extract an error message from a API response.
type ExtractAPIErrorDescriptionFunc func(ctx context.Context, resp *http.Response) (message string, err error)

// FormatAPIErrorToGo formats an API error into a Go error.
// errorContext corresponds to the description of what led to the error e.g. `Failed adding a user`. This is to add further details about the error.
// resp corresponds to the HTTP response from a certain endpoint. Note: the body of such response is not closed by this function.
// clientErr corresponds to the error which may be returned by the HTTP client when calling the endpoint.
func FormatAPIErrorToGo(ctx context.Context, errorContext string, resp *http.Response, clientErr error, errorExtract ExtractAPIErrorDescriptionFunc) (err error) {
	statusCode := 0
	errorMessage := strings.Builder{}
	respErr := commonerrors.ErrUnexpected
	if resp != nil {
		statusCode = resp.StatusCode
		respErr = MapErrorToHTTPResponseCode(statusCode)
		errorDetails, subErr := errorExtract(ctx, resp)
		if commonerrors.Ignore(subErr, commonerrors.ErrMarshalling) != nil {
			err = commonerrors.Join(commonerrors.New(respErr, errorContext), subErr)
			return
		}
		if !reflection.IsEmpty(errorDetails) {
			_, _ = errorMessage.WriteString(errorDetails)
		}
		_ = resp.Body.Close()
	}
	if respErr == nil {
		if clientErr == nil {
			return
		} else {
			respErr = commonerrors.ErrUnexpected
		}
	}
	extra := ""
	if clientErr != nil {
		extra = clientErr.Error()
	}
	errMsgBuilder := strings.Builder{}
	if !reflection.IsEmpty(errorContext) {
		errMsgBuilder.WriteString(errorContext)
	}
	if !reflection.IsEmpty(statusCode) {
		if errMsgBuilder.Len() > 0 {
			errMsgBuilder.WriteString(" ")
		}
		errMsgBuilder.WriteString(fmt.Sprintf("(%d)", statusCode))
	}
	errorDetails := errorMessage.String()
	if !reflection.IsEmpty(errorDetails) {
		if errMsgBuilder.Len() > 0 {
			errMsgBuilder.WriteString(": ")
		}
		errMsgBuilder.WriteString(errorDetails)
	}
	if !reflection.IsEmpty(extra) {
		if errMsgBuilder.Len() > 0 {
			errMsgBuilder.WriteString("; ")
		}
		errMsgBuilder.WriteString(extra)
	}

	err = commonerrors.New(respErr, errMsgBuilder.String())
	return
}

// MapErrorToHTTPResponseCode maps a response status code to a common error.
func MapErrorToHTTPResponseCode(statusCode int) error {
	if statusCode < http.StatusBadRequest {
		return nil
	}
	switch statusCode {
	case http.StatusBadRequest:
		return commonerrors.ErrInvalid
	case http.StatusUnauthorized:
		return commonerrors.ErrUnauthorised
	case http.StatusPaymentRequired:
		return commonerrors.ErrUnknown
	case http.StatusForbidden:
		return commonerrors.ErrForbidden
	case http.StatusNotFound:
		return commonerrors.ErrNotFound
	case http.StatusMethodNotAllowed:
		return commonerrors.ErrNotFound
	case http.StatusNotAcceptable:
		return commonerrors.ErrUnsupported
	case http.StatusProxyAuthRequired:
		return commonerrors.ErrUnauthorised
	case http.StatusRequestTimeout:
		return commonerrors.ErrTimeout
	case http.StatusConflict:
		return commonerrors.ErrConflict
	case http.StatusGone:
		return commonerrors.ErrNotFound
	case http.StatusLengthRequired:
		return commonerrors.ErrInvalid
	case http.StatusPreconditionFailed:
		return commonerrors.ErrCondition
	case http.StatusRequestEntityTooLarge:
		return commonerrors.ErrTooLarge
	case http.StatusRequestURITooLong:
		return commonerrors.ErrTooLarge
	case http.StatusUnsupportedMediaType:
		return commonerrors.ErrUnsupported
	case http.StatusRequestedRangeNotSatisfiable:
		return commonerrors.ErrOutOfRange
	case http.StatusExpectationFailed:
		return commonerrors.ErrUnsupported
	case http.StatusTeapot:
		return commonerrors.ErrUnknown
	case http.StatusMisdirectedRequest:
		return commonerrors.ErrUnsupported
	case http.StatusUnprocessableEntity:
		return commonerrors.ErrMarshalling
	case http.StatusLocked:
		return commonerrors.ErrLocked
	case http.StatusFailedDependency:
		return commonerrors.ErrFailed
	case http.StatusTooEarly:
		return commonerrors.ErrUnexpected
	case http.StatusUpgradeRequired:
		return commonerrors.ErrUnsupported
	case http.StatusPreconditionRequired:
		return commonerrors.ErrCondition
	case http.StatusTooManyRequests:
		return commonerrors.ErrUnavailable
	case http.StatusRequestHeaderFieldsTooLarge:
		return commonerrors.ErrTooLarge
	case http.StatusUnavailableForLegalReasons:
		return commonerrors.ErrUnavailable

	case http.StatusInternalServerError:
		return commonerrors.ErrUnexpected
	case http.StatusNotImplemented:
		return commonerrors.ErrNotImplemented
	case http.StatusBadGateway:
		return commonerrors.ErrUnavailable
	case http.StatusServiceUnavailable:
		return commonerrors.ErrUnavailable
	case http.StatusGatewayTimeout:
		return commonerrors.ErrTimeout
	case http.StatusHTTPVersionNotSupported:
		return commonerrors.ErrUnsupported
	case http.StatusVariantAlsoNegotiates:
		return commonerrors.ErrUnexpected
	case http.StatusInsufficientStorage:
		return commonerrors.ErrUnexpected
	case http.StatusLoopDetected:
		return commonerrors.ErrUnexpected
	case http.StatusNotExtended:
		return commonerrors.ErrUnexpected
	case http.StatusNetworkAuthenticationRequired:
		return commonerrors.ErrUnauthorised
	default:
		return commonerrors.ErrUnexpected
	}
}
