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
	ZipFS
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

// NewZipFileSystem returns a filesystem over the contents of a .zip file.
// Warning: After use of the filesystem, it is crucial to close the zip file (zipFile) which has been opened from source for the entirety of the filesystem use session.
// fs corresponds to the filesystem to use to find the zip file.
func NewZipFileSystem(fs FS, source string, limits ILimits) (zipFs FS, zipFile File, err error) {
	wrapped, zipFile, err := newZipFSAdapterFromFilePath(fs, source, limits)
	if err != nil {
		return
	}
	zipFs = NewVirtualFileSystem(wrapped, ZipFS, IdentityPathConverterFunc)
	return
}

// NewZipFileSystemFromStandardFileSystem returns a zip filesystem similar to NewZipFileSystem but assumes the zip file described by source can be found on the standard file system.
func NewZipFileSystemFromStandardFileSystem(source string, limits ILimits) (FS, File, error) {
	return NewZipFileSystem(NewStandardFileSystem(), source, limits)
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
