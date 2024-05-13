/*
 * Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"log"
	"os"
)

// NewPipeLogger will log messages without appending any prefix. Can be used when re-logging some other log output to prevent duplication of the log source etc.
func NewPipeLogger() (loggers Loggers, err error) {
	loggers = &GenericLoggers{
		Output: log.New(os.Stdout, "", 0),
		Error:  log.New(os.Stderr, "", 0),
	}
	return
}
