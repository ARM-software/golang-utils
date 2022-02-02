/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package filesystem

import (
	"errors"
	"os"

	"github.com/spf13/afero"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

const (
	StandardFS int = iota
	InMemoryFS
)

var (
	FileSystemTypes = []int{StandardFS, InMemoryFS}
)

func NewInMemoryFileSystem() FS {
	return NewVirtualFileSystem(afero.NewMemMapFs(), InMemoryFS, IdentityPathConverterFunc)
}

func NewStandardFileSystem() FS {
	return NewVirtualFileSystem(NewExtendedOsFs(), StandardFS, IdentityPathConverterFunc)
}

func NewFs(fsType int) FS {
	switch fsType {
	case StandardFS:
		return NewStandardFileSystem()
	case InMemoryFS:
		return NewInMemoryFileSystem()
	}
	return NewStandardFileSystem()
}

// Converts file system error into common errors
func ConvertFileSytemError(err error) error {
	if err == nil {
		return nil
	}
	if commonerrors.Any(err, os.ErrExist, errors.New("file exists"), errors.New("file already exists")) {
		return commonerrors.ErrExists
	}
	return err
}
