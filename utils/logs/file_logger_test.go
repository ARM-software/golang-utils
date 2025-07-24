/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/filesystem"
	sizeUnits "github.com/ARM-software/golang-utils/utils/units/size"
)

func TestFileLogger(t *testing.T) {
	defer goleak.VerifyNone(t)
	var tests = []struct {
		loggerCreationFunc func(path string) (Loggers, error)
	}{
		{
			loggerCreationFunc: func(path string) (Loggers, error) { return NewFileLogger(path, "Test") },
		},
		{
			loggerCreationFunc: func(path string) (Loggers, error) { return NewFileOnlyLogger(path, "Test") },
		},
		{
			loggerCreationFunc: func(path string) (Loggers, error) {
				return NewRollingFilesLogger(path, "Test", WithMaxFileSize(sizeUnits.MiB), WithMaxBackups(2), WithMaxAge(time.Second))
			},
		},
		{
			loggerCreationFunc: func(path string) (Loggers, error) {
				return NewRollingFilesLogger(path, "Test")
			},
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("logger %v", i), func(t *testing.T) {
			defer goleak.VerifyNone(t)
			file, err := filesystem.TouchTempFileInTempDir("test-filelog-*.log")
			require.NoError(t, err)

			defer func() { _ = filesystem.Rm(file) }()

			empty, err := filesystem.IsEmpty(file)
			require.NoError(t, err)
			assert.True(t, empty)

			loggers, err := test.loggerCreationFunc(file)
			require.NoError(t, err)

			testLog(t, loggers)

			empty, err = filesystem.IsEmpty(file)
			require.NoError(t, err)
			assert.False(t, empty)

			content, err := filesystem.ReadFile(file)
			require.NoError(t, err)
			fmt.Println(string(content))

			err = filesystem.Rm(file)
			require.NoError(t, err)
		})
	}

}
