/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package mocks

import (
	"go.uber.org/mock/gomock"

	serialisation "github.com/ARM-software/golang-utils/utils/serialisation/mocks"
)

// MockIFrontMatterParser is a mock front matter parser.
type MockIFrontMatterParser = serialisation.MockIFrontMatterParser

// MockIFrontMatterParserMockRecorder records expected parser calls.
type MockIFrontMatterParserMockRecorder = serialisation.MockIFrontMatterParserMockRecorder

// NewMockIFrontMatterParser creates a new mock front matter parser.
func NewMockIFrontMatterParser(ctrl *gomock.Controller) *MockIFrontMatterParser {
	return serialisation.NewMockIFrontMatterParser(ctrl)
}
