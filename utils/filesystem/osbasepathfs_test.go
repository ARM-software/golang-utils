/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filesystem

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestOsBasePathFs_BasicOperations(t *testing.T) {
	baseDir := t.TempDir()
	standardFs := NewTestFs(t, StandardFS)

	fs, err := NewOsBasePathFs(baseDir)
	require.NoError(t, err)
	defer func() { require.NoError(t, fs.Close()) }()

	require.NoError(t, fs.MkDirAll("nested", 0o755))

	file, err := fs.CreateFile("nested/file.txt")
	require.NoError(t, err)

	_, err = file.WriteString("hello world")
	require.NoError(t, err)
	require.Equal(t, FilePathClean(fs, "nested/file.txt"), file.Name())
	require.NoError(t, file.Close())

	content, err := standardFs.ReadFile(FilePathJoin(standardFs, baseDir, "nested", "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(content))

	fileInfo, err := fs.Stat(FilePathJoin(standardFs, baseDir, "nested", "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "file.txt", fileInfo.Name())

	openedFile, err := fs.GenericOpen("nested/file.txt")
	require.NoError(t, err)
	assert.Equal(t, FilePathClean(fs, "nested/file.txt"), openedFile.Name())
	require.NoError(t, openedFile.Close())
}

func TestOsBasePathFs_BlocksPathEscapes(t *testing.T) {
	baseDir := t.TempDir()
	standardFs := NewTestFs(t, StandardFS)
	outsideFile := FilePathJoin(standardFs, t.TempDir(), "outside.txt")
	require.NoError(t, standardFs.WriteFile(outsideFile, []byte("outside"), 0o644))

	fs, err := NewOsBasePathFs(baseDir)
	require.NoError(t, err)
	defer func() { require.NoError(t, fs.Close()) }()

	_, err = fs.CreateFile("../escape.txt")
	errortest.AssertError(t, err, commonerrors.ErrNotFound)

	_, err = fs.Open(outsideFile)
	errortest.AssertError(t, err, commonerrors.ErrNotFound)

	require.NoError(t, fs.MkDirAll("safe", 0o755))
	_, err = fs.CreateFile(FilePathJoin(standardFs, baseDir, "safe", "inside.txt"))
	require.NoError(t, err)
	assert.False(t, standardFs.Exists(FilePathJoin(standardFs, baseDir, "escape.txt")))
}

func TestOsBasePathFs_SymlinkUsesFilesystemPaths(t *testing.T) {
	printWarningOnWindows(t)

	baseDir := t.TempDir()
	standardFs := NewTestFs(t, StandardFS)
	fs, err := NewOsBasePathFs(baseDir)
	require.NoError(t, err)
	defer func() { require.NoError(t, fs.Close()) }()

	require.NoError(t, fs.MkDirAll("dir", 0o755))
	require.NoError(t, standardFs.WriteFile(FilePathJoin(standardFs, baseDir, "dir", "target.txt"), []byte("payload"), 0o644))

	err = fs.Symlink("dir/target.txt", "dir/link.txt")
	if err != nil {
		skipIfLinksNotSupported(t, ConvertFileSystemError(err))
	}

	opened, err := fs.Open("dir/link.txt")
	require.NoError(t, err)
	require.NoError(t, opened.Close())

	target, err := fs.Readlink("dir/link.txt")
	require.NoError(t, err)
	assert.Equal(t, FilePathClean(fs, "dir/target.txt"), target)

	outsideTarget := FilePathJoin(standardFs, t.TempDir(), "outside-target.txt")
	require.NoError(t, standardFs.WriteFile(outsideTarget, []byte("outside"), 0o644))
	err = fs.Symlink(outsideTarget, "dir/outside-link.txt")
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
	assert.False(t, standardFs.Exists(FilePathJoin(standardFs, baseDir, "dir", "outside-link.txt")))
}
