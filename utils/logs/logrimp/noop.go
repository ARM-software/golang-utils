/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logrimp

import "github.com/go-logr/logr"

// NewNoopLogger returns a discarding.
func NewNoopLogger() logr.Logger {
	return logr.Discard()
}
