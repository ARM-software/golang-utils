/*
 * Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/safeio"
)

type FieldObjectForTest struct {
	// Name of the field name in the request which has failed validation.
	FieldName string `json:"fieldName"`
	// Field name, possibly including the path of the field which caused the error.
	FieldPath *string `json:"fieldPath,omitempty"`
	// A human-readable message, which should provide an explanation and possible corrective actions.
	Message string `json:"message"`
}

type ErrorResponseForTest struct {
	// Fields in the request that failed validation [Optional].
	Fields []FieldObjectForTest `json:"fields,omitempty"`
	// HTTP Status Code
	HTTPStatusCode int32 `json:"httpStatusCode"`
	// A human-readable message, which should provide an explanation and possible corrective actions.
	Message string `json:"message"`
	// Request ID that could be used to identify the error in logs.
	RequestID string `json:"requestID"`
}

func (e *ErrorResponseForTest) ToMap() (map[string]any, error) {
	toSerialize := map[string]any{}
	if e.Fields != nil {
		toSerialize["fields"] = e.Fields
	}
	toSerialize["httpStatusCode"] = e.HTTPStatusCode
	toSerialize["message"] = e.Message
	toSerialize["requestID"] = e.RequestID
	return toSerialize, nil
}

func (e *ErrorResponseForTest) UnmarshalJSON(data []byte) (err error) {
	requiredProperties := []string{
		"httpStatusCode",
		"message",
		"requestId",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err
	}

	for _, requiredProperty := range requiredProperties {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varErrorResponse := ErrorResponseForTest{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varErrorResponse)

	if err != nil {
		return err
	}

	*e = varErrorResponse

	return err
}

func extractError(ctx context.Context, resp *http.Response) (message string, err error) {
	if resp == nil {
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	errorResponse := ErrorResponseForTest{}
	b, err := safeio.ReadAll(ctx, resp.Body)
	if err != nil {
		return
	}
	d := json.NewDecoder(bytes.NewBuffer(b))
	// The following is to match the OpenAPI client behaviour https://github.com/OpenAPITools/openapi-generator/issues/21446
	d.DisallowUnknownFields()
	err = d.Decode(&errorResponse)
	message = string(b)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrMarshalling, err, "error occurred when unmarshalling response")
	}
	return
}

func TestIsAPICallSuccessful(t *testing.T) {
	t.Run("api call successful", func(t *testing.T) {
		resp := http.Response{StatusCode: 200}
		isSuccessful := IsCallSuccessful(&resp)
		assert.True(t, isSuccessful)
	})

	t.Run("api call unsuccessful", func(t *testing.T) {
		resp := http.Response{StatusCode: 400}
		isSuccessful := IsCallSuccessful(&resp)
		assert.False(t, isSuccessful)
	})

	t.Run("api call returns nothing", func(t *testing.T) {
		resp := http.Response{}
		isSuccessful := IsCallSuccessful(&resp)
		assert.False(t, isSuccessful)
	})
}

func TestCheckAPICallSuccess(t *testing.T) {
	t.Run("context cancelled", func(t *testing.T) {
		errMessage := "context cancelled"
		parentCtx := context.Background()
		ctx, cancelCtx := context.WithCancel(parentCtx)
		cancelCtx()
		respBody := http.Response{Body: io.NopCloser(bytes.NewReader(nil))}
		actualErr := CheckAPICallSuccess(ctx, errMessage, extractError, &respBody, errors.New(errMessage))
		errortest.AssertError(t, actualErr, commonerrors.ErrCancelled)

	})

	t.Run("api call not successful", func(t *testing.T) {
		errMessage := "client error"
		parentCtx := context.Background()
		resp := http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(bytes.NewReader([]byte("{\"message\": \"client error\",\"requestId\": \"761761721\"}")))}
		actualErr := CheckAPICallSuccess(parentCtx, errMessage, extractError, &resp, errors.New(errMessage))
		errortest.AssertError(t, actualErr, commonerrors.ErrInvalid)
	})

	t.Run("api call not successful", func(t *testing.T) {
		errMessage := "client error"
		parentCtx := context.Background()
		resp := http.Response{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(bytes.NewReader([]byte("{\"message\": \"client error\",\"requestId\": \"761761721\"}")))}
		actualErr := CheckAPICallSuccess(parentCtx, errMessage, extractError, &resp, errors.New(errMessage))
		errortest.AssertError(t, actualErr, commonerrors.ErrUnavailable)
	})

	t.Run("api call not successful", func(t *testing.T) {
		errMessage := "client error"
		parentCtx := context.Background()
		resp := http.Response{StatusCode: http.StatusUnauthorized, Body: io.NopCloser(bytes.NewReader([]byte("{\"message\": \"client error\",\"requestId\": \"761761721\"}")))}
		actualErr := CheckAPICallSuccess(parentCtx, errMessage, extractError, &resp, errors.New(errMessage))
		errortest.AssertError(t, actualErr, commonerrors.ErrUnauthorised)
	})

	t.Run("api call not successful (no JSON response)", func(t *testing.T) {
		errMessage := "response error"
		parentCtx := context.Background()
		resp := http.Response{StatusCode: http.StatusForbidden, Body: io.NopCloser(bytes.NewReader([]byte("<html><head><title>403 Forbidden</title></head></html>")))}
		actualErr := CheckAPICallSuccess(parentCtx, errMessage, extractError, &resp, errors.New("403 Forbidden"))
		expectedErr := "response error (403): <html><head><title>403 Forbidden</title></head></html>; 403 Forbidden"
		assert.Contains(t, actualErr.Error(), expectedErr)
		errortest.AssertError(t, actualErr, commonerrors.ErrForbidden)
	})

	t.Run("no context error, api call successful", func(t *testing.T) {
		errMessage := "no error"
		parentCtx := context.Background()
		resp := http.Response{StatusCode: 200}
		err := CheckAPICallSuccess(parentCtx, errMessage, extractError, &resp, errors.New(errMessage))
		assert.NoError(t, err)
	})
}

func TestCallAndCheckSuccess(t *testing.T) {
	t.Run("context cancelled", func(t *testing.T) {
		errMessage := "context cancelled"
		parentCtx := context.Background()
		ctx, cancelCtx := context.WithCancel(parentCtx)
		cancelCtx()
		_, actualErr := CallAndCheckSuccess(ctx, errMessage, extractError,
			func(ctx context.Context) (*struct{}, *http.Response, error) {
				return nil, &http.Response{Body: io.NopCloser(bytes.NewReader(nil))}, nil
			})
		errortest.AssertError(t, actualErr, commonerrors.ErrCancelled)
	})

	t.Run("api call not successful", func(t *testing.T) {
		errMessage := "client error"
		parentCtx := context.Background()
		_, actualErr := CallAndCheckSuccess(parentCtx, errMessage, extractError,
			func(ctx context.Context) (*struct{}, *http.Response, error) {
				resp := http.Response{StatusCode: 400, Body: io.NopCloser(bytes.NewReader([]byte("{\"message\": \"client error\",\"requestId\": \"761761721\"}")))}
				return nil, &resp, errors.New(errMessage)
			})
		errortest.AssertError(t, actualErr, commonerrors.ErrInvalid)
	})

	t.Run("api call successful, marshalling failed due to missing required field in response", func(t *testing.T) {
		expectedErrorMessage := ErrorResponseForTest{
			Fields: []FieldObjectForTest{{
				FieldName: faker.Name(),
				FieldPath: field.ToOptionalString(faker.Name()),
				Message:   faker.Sentence(),
			}},
			HTTPStatusCode: 200,
			Message:        faker.Sentence(),
			RequestID:      faker.UUIDDigit(),
		}
		response, err := expectedErrorMessage.ToMap()
		require.NoError(t, err)
		delete(response, "message")

		reducedResponse, err := json.Marshal(response)
		require.NoError(t, err)

		errorResponse := ErrorResponseForTest{}
		errM := errorResponse.UnmarshalJSON(reducedResponse)
		require.Error(t, errM)

		parentCtx := context.Background()
		_, err = CallAndCheckSuccess[ErrorResponseForTest](parentCtx, "test", extractError,
			func(ctx context.Context) (*ErrorResponseForTest, *http.Response, error) {
				return &errorResponse, &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(reducedResponse))}, errM
			})
		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrMarshalling)
	})

	t.Run("api call successful, strict marshalling failed but recovery", func(t *testing.T) {
		expectedErrorMessage := ErrorResponseForTest{
			Fields: []FieldObjectForTest{{
				FieldName: faker.Name(),
				FieldPath: field.ToOptionalString(faker.Name()),
				Message:   faker.Sentence(),
			}},
			HTTPStatusCode: 200,
			Message:        faker.Sentence(),
			RequestID:      faker.UUIDDigit(),
		}
		response, err := expectedErrorMessage.ToMap()
		require.NoError(t, err)
		response[faker.Word()] = faker.Name()
		response[faker.Word()] = faker.Sentence()
		response[faker.Word()] = faker.Paragraph()
		response[faker.Word()] = faker.UUIDDigit()
		extendedResponse, err := json.Marshal(response)
		require.NoError(t, err)
		errMessage := "no error"
		parentCtx := context.Background()
		result, err := CallAndCheckSuccess[ErrorResponseForTest](parentCtx, errMessage, extractError,
			func(ctx context.Context) (*ErrorResponseForTest, *http.Response, error) {
				return &ErrorResponseForTest{}, &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(extendedResponse))}, errors.New(errMessage)
			})
		require.NoError(t, err)
		assert.Equal(t, expectedErrorMessage, *result)
	})

	t.Run("api call successful, strict marshalling failed but recovery, with raw response", func(t *testing.T) {
		expectedErrorMessage := ErrorResponseForTest{
			Fields: []FieldObjectForTest{{
				FieldName: faker.Name(),
				FieldPath: field.ToOptionalString(faker.Name()),
				Message:   faker.Sentence(),
			}},
			HTTPStatusCode: 200,
			Message:        faker.Sentence(),
			RequestID:      faker.UUIDDigit(),
		}
		response, err := expectedErrorMessage.ToMap()
		require.NoError(t, err)
		response[faker.Word()] = faker.Name()
		response[faker.Word()] = faker.Sentence()
		response[faker.Word()] = faker.Paragraph()
		response[faker.Word()] = faker.UUIDDigit()
		extendedResponse, err := json.Marshal(response)
		require.NoError(t, err)
		errMessage := "no error"
		parentCtx := context.Background()
		result, resp, err := CallAndCheckSuccessAndReturnRawResponse[ErrorResponseForTest](parentCtx, errMessage, extractError,
			func(ctx context.Context) (*ErrorResponseForTest, *http.Response, error) {
				return &ErrorResponseForTest{}, &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(extendedResponse))}, errors.New(errMessage)
			})
		defer func() {
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}
		}()
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Body)
		assert.Equal(t, expectedErrorMessage, *result)
		bodyC, err := safeio.ReadAll(context.Background(), resp.Body)
		require.NoError(t, err)
		assert.NotEmpty(t, bodyC)
		require.NoError(t, resp.Body.Close())
	})

	t.Run("api call successful, empty response", func(t *testing.T) {
		errMessage := "no error"
		parentCtx := context.Background()
		_, err := CallAndCheckSuccess(parentCtx, errMessage, extractError,
			func(ctx context.Context) (*struct{}, *http.Response, error) {
				return &struct{}{}, &http.Response{StatusCode: 200}, errors.New(errMessage)
			})
		errortest.AssertError(t, err, commonerrors.ErrMarshalling)
		errortest.AssertErrorDescription(t, err, "API call was successful but an error occurred during response marshalling")
	})

	t.Run("api call successful, broken response decode", func(t *testing.T) {
		errMessage := "no error"
		parentCtx := context.Background()
		_, err := CallAndCheckSuccess(parentCtx, errMessage, extractError,
			func(ctx context.Context) (*struct{}, *http.Response, error) {
				return &struct{}{}, &http.Response{StatusCode: 200}, nil
			})
		errortest.AssertError(t, err, commonerrors.ErrMarshalling)
		errortest.AssertErrorDescription(t, err, "unmarshalled response is empty")
	})
}

func TestGenericCallAndCheckSuccess(t *testing.T) {
	t.Run("context cancelled", func(t *testing.T) {
		errMessage := "context cancelled"
		parentCtx := context.Background()
		ctx, cancelCtx := context.WithCancel(parentCtx)
		cancelCtx()
		_, actualErr := GenericCallAndCheckSuccess(ctx, errMessage, extractError,
			func(ctx context.Context) (*struct{}, *http.Response, error) {
				return nil, &http.Response{Body: io.NopCloser(bytes.NewReader(nil))}, errors.New(errMessage)
			})
		errortest.AssertError(t, actualErr, commonerrors.ErrCancelled)
	})

	t.Run("api call not successful", func(t *testing.T) {
		errMessage := "client error"
		parentCtx := context.Background()
		_, actualErr := GenericCallAndCheckSuccess(parentCtx, errMessage, extractError,
			func(ctx context.Context) (*struct{}, *http.Response, error) {
				resp := http.Response{StatusCode: 400, Body: io.NopCloser(bytes.NewReader([]byte("{\"message\": \"client error\",\"requestId\": \"761761721\"}")))}
				return nil, &resp, errors.New(errMessage)
			})
		errortest.AssertError(t, actualErr, commonerrors.ErrInvalid)
	})

	t.Run("api call successful but error marshalling", func(t *testing.T) {
		errMessage := "no error"
		parentCtx := context.Background()
		_, err := GenericCallAndCheckSuccess(parentCtx, errMessage, extractError,
			func(ctx context.Context) (any, *http.Response, error) {
				tmp := struct {
					test string
				}{
					test: faker.Word(),
				}
				return &tmp, &http.Response{StatusCode: 200}, errors.New(errMessage)
			})
		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrMarshalling)
	})

	t.Run("api call successful, empty response", func(t *testing.T) {
		errMessage := "response error"
		parentCtx := context.Background()
		_, err := GenericCallAndCheckSuccess(parentCtx, errMessage, extractError,
			func(ctx context.Context) (*struct{}, *http.Response, error) {
				return &struct{}{}, &http.Response{StatusCode: 200}, errors.New(errMessage)
			})
		errortest.AssertError(t, err, commonerrors.ErrMarshalling)
	})

	t.Run("api call successful, incorrect response", func(t *testing.T) {
		parentCtx := context.Background()
		_, err := GenericCallAndCheckSuccess(parentCtx, "response error", extractError,
			func(ctx context.Context) (struct{ Blah string }, *http.Response, error) {
				return struct{ Blah string }{Blah: "fsadsfs"}, &http.Response{StatusCode: 200}, nil
			})
		errortest.AssertError(t, err, commonerrors.ErrConflict)
	})

	t.Run("api call successful, incorrect response, returned raw response", func(t *testing.T) {
		parentCtx := context.Background()
		_, resp, err := GenericCallAndCheckSuccessAndReturnRawResponse(parentCtx, "response error", extractError,
			func(ctx context.Context) (struct{ Blah string }, *http.Response, error) {
				return struct{ Blah string }{Blah: "fsadsfs"}, &http.Response{StatusCode: 200}, nil
			})
		defer func() {
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}
		}()
		errortest.AssertError(t, err, commonerrors.ErrConflict)
		require.NotNil(t, resp)
		require.Nil(t, resp.Body)
	})
}
