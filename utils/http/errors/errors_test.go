package errors

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http/httptest"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/safeio"
)
import "net/http"

func TestErrorMapping(t *testing.T) {
	require.NoError(t, MapErrorToHTTPResponseCode(http.StatusAccepted))
	errortest.AssertError(t, MapErrorToHTTPResponseCode(http.StatusBadRequest), commonerrors.ErrInvalid)
	errorRange := collection.RangeSequence(http.StatusBadRequest, 600, nil)
	for s := range errorRange {
		err := MapErrorToHTTPResponseCode(s)
		require.Error(t, err)
		assert.True(t, commonerrors.IsCommonError(err))
	}
	require.NoError(t, MapErrorToHTTPResponseCode(http.StatusContinue))
}

func extractErrorMessage(ctx context.Context, resp *http.Response) (message string, err error) {
	if resp == nil || resp.Body == nil {
		return "", commonerrors.ErrUndefined
	}
	b, err := safeio.ReadAll(ctx, resp.Body)
	err = commonerrors.Ignore(err, commonerrors.ErrEmpty, commonerrors.ErrUndefined)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func TestFormatAPIErrorToGo(t *testing.T) {

	tests := []struct {
		errorContext  string
		responseBody  string
		responseCode  int
		clientErr     error
		expectedError error
	}{
		{
			responseCode:  http.StatusAccepted,
			clientErr:     nil,
			expectedError: nil,
		},
		{
			errorContext:  faker.Sentence(),
			responseBody:  "",
			responseCode:  http.StatusAccepted,
			clientErr:     nil,
			expectedError: nil,
		},
		{
			errorContext:  faker.Sentence(),
			responseBody:  "",
			responseCode:  http.StatusAccepted,
			clientErr:     commonerrors.ErrInvalid,
			expectedError: commonerrors.ErrUnexpected,
		},
		{
			errorContext:  faker.Sentence(),
			responseBody:  "    ",
			responseCode:  http.StatusBadRequest,
			clientErr:     commonerrors.ErrNotFound,
			expectedError: commonerrors.ErrInvalid,
		},
		{
			errorContext:  faker.Sentence(),
			responseBody:  faker.Sentence(),
			responseCode:  http.StatusBadRequest,
			clientErr:     commonerrors.ErrNotFound,
			expectedError: commonerrors.ErrInvalid,
		},
		{
			errorContext:  faker.Sentence(),
			responseBody:  faker.Sentence(),
			responseCode:  http.StatusServiceUnavailable,
			clientErr:     commonerrors.ErrConflict,
			expectedError: commonerrors.ErrUnavailable,
		},
		{
			errorContext:  faker.Sentence(),
			responseBody:  faker.Sentence(),
			responseCode:  http.StatusServiceUnavailable,
			clientErr:     nil,
			expectedError: commonerrors.ErrUnavailable,
		},
		{
			responseBody:  faker.Sentence(),
			responseCode:  http.StatusInsufficientStorage,
			clientErr:     nil,
			expectedError: commonerrors.ErrUnexpected,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			w := httptest.NewRecorder()
			w.WriteHeader(test.responseCode)
			if w.Body != nil {
				_, err := w.Body.WriteString(test.responseBody)
				require.NoError(t, err)
			}

			resp := w.Result()
			returnedError := FormatAPIErrorToGo(context.Background(), test.errorContext, resp, test.clientErr, extractErrorMessage)
			errortest.AssertError(t, returnedError, test.expectedError)
			if test.expectedError != nil {
				if test.clientErr != nil {
					errortest.AssertErrorDescription(t, returnedError, test.clientErr.Error())
				}
				if !reflection.IsEmpty(test.errorContext) {
					errortest.AssertErrorDescription(t, returnedError, test.errorContext)
				}
				if !reflection.IsEmpty(test.responseBody) {
					errortest.AssertErrorDescription(t, returnedError, test.responseBody)
				}
			}
		})
	}
}
