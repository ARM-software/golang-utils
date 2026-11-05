/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logrimp

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

// NewZapLogger returns a new zap logger
func NewZapLogger(logger *zap.Logger) logr.Logger {
	return zapr.NewLogger(logger)
}
