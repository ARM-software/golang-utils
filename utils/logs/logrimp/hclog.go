/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logrimp

import (
	"github.com/evanphx/hclogr"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-hclog"
)

// NewHclogLogger returns a new HCLog logger.
func NewHclogLogger(logger hclog.Logger) logr.Logger {
	return hclogr.Wrap(logger)
}
