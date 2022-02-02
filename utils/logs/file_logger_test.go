/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

func TestFileLogger(t *testing.T) {
	file, err := filesystem.TempFileInTempDir("test-filelog-*.log")
	require.Nil(t, err)

	err = file.Close()
	require.Nil(t, err)
	defer func() { _ = filesystem.Rm(file.Name()) }()

	loggers, err := CreateFileLogger(file.Name(), "Test")
	require.Nil(t, err)

	_testLog(t, loggers)
	err = filesystem.Rm(file.Name())
	require.Nil(t, err)
}
