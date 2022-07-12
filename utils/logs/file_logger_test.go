/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

func TestFileLogger(t *testing.T) {
	file, err := filesystem.TempFileInTempDir("test-filelog-*.log")
	require.Nil(t, err)

	err = file.Close()
	require.NoError(t, err)

	defer func() { _ = filesystem.Rm(file.Name()) }()

	empty, err := filesystem.IsEmpty(file.Name())
	require.NoError(t, err)
	assert.True(t, empty)

	loggers, err := NewFileLogger(file.Name(), "Test")
	require.NoError(t, err)

	testLog(t, loggers)

	empty, err = filesystem.IsEmpty(file.Name())
	require.NoError(t, err)
	assert.False(t, empty)

	err = filesystem.Rm(file.Name())
	require.NoError(t, err)
}
