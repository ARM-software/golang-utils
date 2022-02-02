/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
)

// Creates a logger to logrus logger (https://github.com/Sirupsen/logrus)
func NewLogrusLogger(logrusL *logrus.Logger, loggerSource string) (loggers Loggers, err error) {
	loggers = &GenericLoggers{
		Output: log.New(logrusL.WriterLevel(logrus.InfoLevel), fmt.Sprintf("[%v] ", loggerSource), log.LstdFlags),
		Error:  log.New(logrusL.WriterLevel(logrus.ErrorLevel), fmt.Sprintf("[%v] ", loggerSource), log.LstdFlags),
	}
	return
}
