/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filesystem

import (
	"embed"
	"os"

	"github.com/spf13/afero"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

type FilesystemType int

const (
	StandardFS FilesystemType = iota
	InMemoryFS
	Embed
	Custom
)

var (
	FileSystemTypes = []FilesystemType{StandardFS, InMemoryFS}
)

func NewInMemoryFileSystem() FS {
	return NewVirtualFileSystem(afero.NewMemMapFs(), InMemoryFS, IdentityPathConverterFunc)
}

func NewStandardFileSystem() FS {
	return NewVirtualFileSystem(NewExtendedOsFs(), StandardFS, IdentityPathConverterFunc)
}

func NewEmbedFileSystem(fs *embed.FS) (FS, error) {
	wrapped, err := newEmbedFSAdapter(fs)
	if err != nil {
		return nil, err
	}
	return NewVirtualFileSystem(wrapped, Embed, IdentityPathConverterFunc), nil
}

func NewFs(fsType FilesystemType) FS {
	switch fsType {
	case StandardFS:
		return NewStandardFileSystem()
	case InMemoryFS:
		return NewInMemoryFileSystem()
	}
	return NewStandardFileSystem()
}

// ConvertFileSystemError converts file system error into common errors
func ConvertFileSystemError(err error) error {
	if err == nil {
		return nil
	}
	if commonerrors.Any(err, os.ErrExist) || commonerrors.CorrespondTo(err, "file exists", "file already exists") {
		return commonerrors.ErrExists
	}
	return err
}
