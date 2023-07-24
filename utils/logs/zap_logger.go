/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package logs defines loggers for use in projects.
package logs

import (
	"go.uber.org/zap"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/logs/logrimp"
)

const (
	syncError    = "invalid argument"    // sync error can happen on Linux (sync /dev/stderr: invalid argument) see https://github.com/uber-go/zap/issues/328
	syncErrorMac = "bad file descriptor" // sync error can happen on MacOs (sync /dev/stderr: invalid argument) see https://github.com/uber-go/zap/issues/328

)

// NewZapLogger returns a logger which uses zap logger (https://github.com/uber-go/zap)
func NewZapLogger(zapL *zap.Logger, loggerSource string) (loggers Loggers, err error) {
	if zapL == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	return NewLogrLoggerWithClose(logrimp.NewZapLogger(zapL), loggerSource, func() error {
		err := zapL.Sync()
		// handling this error https://github.com/uber-go/zap/issues/328
		if commonerrors.CorrespondTo(err, syncError, syncErrorMac) {
			return nil
		}
		return err
	})
}
