/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package logs defines loggers for use in projects.
package logs

import (
	"log/slog"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/logs/logrimp"
)

// NewSlogLogger returns a logger which uses slog logger (standard library package)
func NewSlogLogger(slogL *slog.Logger, loggerSource string) (loggers Loggers, err error) {
	if slogL == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	return NewLogrLogger(logrimp.NewSlogLogger(slogL), loggerSource)
}
