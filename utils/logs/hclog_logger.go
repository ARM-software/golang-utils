/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package logs defines loggers for use in projects.
package logs

import (
	"github.com/hashicorp/go-hclog"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/logs/logrimp"
)

// NewHclogLogger returns a logger which uses hclog logger (https://github.com/hashicorp/go-hclog)
func NewHclogLogger(hclogL hclog.Logger, loggerSource string) (loggers Loggers, err error) {
	if hclogL == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	return NewLogrLogger(logrimp.NewHclogLogger(hclogL), loggerSource)
}

// NewHclogWrapper returns an hclog logger from a Loggers logger
func NewHclogWrapper(loggers Loggers) (hclogL hclog.Logger, err error) {
	if loggers == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	intercept := hclog.NewInterceptLogger(nil)

	info, err := NewInfoWriterFromLoggers(loggers)
	if err != nil {
		return
	}
	errL, err := NewErrorWriterFromLoggers(loggers)
	if err != nil {
		return
	}

	sinkErr := hclog.NewSinkAdapter(&hclog.LoggerOptions{
		Level:       hclog.Warn,
		Output:      errL,
		DisableTime: true,
	})
	sinkInfo := hclog.NewSinkAdapter(&hclog.LoggerOptions{
		Level:       hclog.Info,
		Output:      info,
		DisableTime: true,
	})

	intercept.RegisterSink(sinkErr)
	intercept.RegisterSink(sinkInfo)

	hclogL = intercept
	return
}
