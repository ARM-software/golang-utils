/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/assert"
)

func TestFindRetryAfter(t *testing.T) {
	tests := []struct {
		expectedTime  time.Duration
		has           bool
		includeHeader bool
		retryAfter    string
	}{
		{
			expectedTime:  0,
			has:           false,
			includeHeader: false,
			retryAfter:    "",
		},
		{
			expectedTime:  0,
			has:           false,
			includeHeader: true,
			retryAfter:    "",
		},
		{
			expectedTime:  0,
			has:           false,
			includeHeader: true,
			retryAfter:    "blahaha",
		},
		{
			expectedTime:  0,
			has:           false,
			includeHeader: true,
			retryAfter:    "15s",
		},
		{
			expectedTime:  0,
			has:           true,
			includeHeader: true,
			retryAfter:    "0",
		},
		{
			expectedTime:  0,
			has:           true,
			includeHeader: true,
			retryAfter:    "-12",
		},
		{
			expectedTime:  150 * time.Second,
			has:           true,
			includeHeader: true,
			retryAfter:    "150",
		},
		{
			expectedTime:  0,
			has:           true,
			includeHeader: true,
			retryAfter:    "2019-10-12T07:20:50.52Z",
		},
		{
			expectedTime:  0,
			has:           true,
			includeHeader: true,
			retryAfter:    "Mon, 09 Mar 2020 08:13:24 GMT",
		},
		{
			expectedTime:  0,
			has:           true,
			includeHeader: true,
			retryAfter:    "Mon Jan 2 15:04:05 2006",
		},
		{
			expectedTime:  3 * time.Minute,
			has:           true,
			includeHeader: true,
			retryAfter:    time.Now().Add(3 * time.Minute).UTC().Format(time.RFC1123),
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("with RetryAfter %v", test.retryAfter), func(t *testing.T) {
			responseRecorder := httptest.NewRecorder()
			responseRecorder.Code = http.StatusTooManyRequests
			if test.includeHeader {
				responseRecorder.Header().Add(headers.RetryAfter, test.retryAfter)
			}
			wait, found := findRetryAfter(responseRecorder.Result())
			if test.has {
				assert.True(t, found)
				assert.GreaterOrEqual(t, wait, time.Duration(0))
				assert.LessOrEqual(t, wait, test.expectedTime)
			} else {
				assert.False(t, found)
				assert.Empty(t, wait)
			}
		})
	}

}

func TestRetryPolicyFactory(t *testing.T) {
	tests := []struct {
		expectedPolicy IRetryWaitPolicy
		config         *RetryPolicyConfiguration
		policy         string
	}{
		{
			expectedPolicy: &BasicRetryPolicy{},
			config:         DefaultNoRetryPolicyConfiguration(),
			policy:         "basic retry",
		},
		{
			expectedPolicy: &BasicRetryPolicy{},
			config:         DefaultBasicRetryPolicyConfiguration(),
			policy:         "basic retry",
		},
		{
			expectedPolicy: &BasicRetryPolicy{},
			config:         nil,
			policy:         "basic retry",
		},
		{
			expectedPolicy: &BasicRetryPolicy{
				RetryWaitPolicy: RetryWaitPolicy{
					ConsiderRetryAfter: true,
				},
			},
			config: &RetryPolicyConfiguration{
				Enabled: false,
			},
			policy: "basic retry with retry after",
		},
		{
			expectedPolicy: &BasicRetryPolicy{
				RetryWaitPolicy: RetryWaitPolicy{
					ConsiderRetryAfter: true,
				},
			},
			config: DefaultRobustRetryPolicyConfiguration(),
			policy: "basic retry with retry after",
		},
		{
			expectedPolicy: &ExponentialBackoffPolicy{
				RetryWaitPolicy: RetryWaitPolicy{
					ConsiderRetryAfter: false,
				},
			},
			config: &RetryPolicyConfiguration{
				Enabled:            true,
				RetryAfterDisabled: true,
				BackOffEnabled:     true,
			},
			policy: "exponential backoff with no retry-after",
		},
		{
			expectedPolicy: &ExponentialBackoffPolicy{
				RetryWaitPolicy: RetryWaitPolicy{
					ConsiderRetryAfter: true,
				},
			},
			config: DefaultExponentialBackoffRetryPolicyConfiguration(),
			policy: "exponential backoff with retry-after",
		},
		{
			expectedPolicy: &LinearBackoffPolicy{
				RetryWaitPolicy: RetryWaitPolicy{
					ConsiderRetryAfter: false,
				},
			},
			config: &RetryPolicyConfiguration{
				Enabled:              true,
				RetryAfterDisabled:   true,
				BackOffEnabled:       true,
				LinearBackOffEnabled: true,
			},
			policy: "linear backoff with no retry-after",
		},
		{
			expectedPolicy: &LinearBackoffPolicy{
				RetryWaitPolicy: RetryWaitPolicy{
					ConsiderRetryAfter: true,
				},
			},
			config: DefaultLinearBackoffRetryPolicyConfiguration(),
			policy: "linear backoff with retry-after",
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("assert policy %v", test.policy), func(t *testing.T) {
			assert.Equal(t, test.expectedPolicy, BackOffPolicyFactory(test.config))
		})
	}
}
