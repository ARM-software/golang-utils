/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import "github.com/go-logr/logr"

func NewNoopLogger(loggerSource string) (loggers Loggers, err error) {
	return NewLogrLogger(logr.Discard(), loggerSource)
}
