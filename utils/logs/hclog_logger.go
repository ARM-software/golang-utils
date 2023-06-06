/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package logs defines loggers for use in projects.
package logs

import (
	"github.com/evanphx/hclogr"
	"github.com/hashicorp/go-hclog"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

// NewHclogLogger returns a logger which uses hclog logger (https://github.com/hashicorp/go-hclog
func NewHclogLogger(hclogL hclog.Logger, loggerSource string) (loggers Loggers, err error) {
	if hclogL == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	return NewLogrLogger(hclogr.Wrap(hclogL), loggerSource)
}
