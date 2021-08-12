/*
 * Copyright (C) 2020-2021 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"io/ioutil"
	"log"
)

func NewNoopLogger(loggerSource string) (loggers Loggers, err error) {
	loggers = &GenericLoggers{
		Output: log.New(ioutil.Discard, "", 0),
		Error:  log.New(ioutil.Discard, "", 0),
	}
	return
}
