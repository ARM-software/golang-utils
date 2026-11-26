/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filesystem

import (
	"embed"
	"fmt"
	"io"
	"os"

	"github.com/spf13/afero"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/platform"
)

//go:generate go tool enumer -type=FilesystemType -text -json -yaml

type FilesystemType int

const (
	StandardFS FilesystemType = iota
	InMemoryFS
	Embed
	Custom
	ZipFS
	TarFS
	OsBasePathFS
	PermissionDeniedFS
)

var (
	FileSystemTypes = []FilesystemType{StandardFS, InMemoryFS, OsBasePathFS}
)

func NewInMemoryFileSystem() FS {
	return NewVirtualFileSystem(withFilesystemTypeName(afero.NewMemMapFs(), InMemoryFS), InMemoryFS, IdentityPathConverterFunc)
}

func NewStandardFileSystem() FS {
	return NewVirtualFileSystem(NewExtendedOsFs(), StandardFS, IdentityPathConverterFunc)
}

// NewOsBasePathFs returns a closeable filesystem rooted at basePath.
//
// The returned filesystem uses os.Root so all operations remain confined to
// basePath, including traversal through symbolic links. Paths are interpreted
// relative to basePath, while absolute paths already under basePath are mapped
// back to filesystem-relative paths.
//
// Use this filesystem when code should behave like a writable base-path
// filesystem on the host OS, but path resolution must stay safely confined to a
// specific directory tree. It is suited to operations on untrusted or variable
// paths where afero's lexical base-path restriction is not sufficient because
// symbolic links also need to be contained.
//
// The filesystem must be closed after use to release the underlying root handle.
func NewOsBasePathFs(basePath string) (ICloseableFS, error) {
	wrapped, err := newOsBasePathFSAdapter(basePath)
	if err != nil {
		return nil, err
	}

	return NewCloseableVirtualFileSystem(
		wrapped,
		OsBasePathFS,
		wrapped,
		fmt.Sprintf("os root `%v`", basePath),
		basePathPathConverter(basePath),
	), nil
}

func NewEmbedFileSystem(fs *embed.FS) (FS, error) {
	wrapped, err := newEmbedFSAdapter(fs)
	err = ConvertFileSystemError(err)
	if err != nil {
		return nil, err
	}
	// https://pkg.go.dev/embed#hdr-Directives The path separator is a forward slash, even on Windows systems.
	return NewVirtualFileSystemWithPathSeparator(withFilesystemTypeName(wrapped, Embed), Embed, IdentityPathConverterFunc, '/'), nil
}

// NewZipFileSystem returns a filesystem over the contents of a .zip file.
// Warning: After use of the filesystem, it is crucial to close the zip file (zipFile) which has been opened from source for the entirety of the filesystem use session.
// fs corresponds to the filesystem to use to find the zip file.
func NewZipFileSystem(fs FS, source string, limits ILimits) (zipFs ICloseableFS, zipFile File, err error) {
	wrapped, zipFile, err := newZipFSAdapterFromFilePath(fs, source, limits)
	err = ConvertFileSystemError(err)
	if err != nil {
		return
	}
	zipFs = NewCloseableVirtualFileSystem(withFilesystemTypeName(wrapped, ZipFS), ZipFS, zipFile, fmt.Sprintf(".zip file `%v`", source), IdentityPathConverterFunc)
	return
}

// NewZipFileSystemFromStandardFileSystem returns a zip filesystem similar to NewZipFileSystem but assumes the zip file described by source can be found on the standard file system.
func NewZipFileSystemFromStandardFileSystem(source string, limits ILimits) (ICloseableFS, File, error) {
	return NewZipFileSystem(NewStandardFileSystem(), source, limits)
}

// NewTarFileSystem returns a filesystem over the contents of a .tar file.
// Warning: After use of the filesystem, it is crucial to close the tar file (tarFile) which has been opened from source for the entirety of the filesystem use session.
// fs corresponds to the filesystem to use to find the tar file.
func NewTarFileSystem(fs FS, source string, limits ILimits) (squashFS ICloseableFS, tarFile File, err error) {
	wrapped, tarFile, err := newTarFSAdapterFromFilePath(fs, source, limits)
	err = ConvertFileSystemError(err)
	if err != nil {
		return
	}
	squashFS = NewCloseableVirtualFileSystem(withFilesystemTypeName(wrapped, TarFS), TarFS, tarFile, fmt.Sprintf(".tar file `%v`", source), IdentityPathConverterFunc)
	return
}

// NewTarFileSystemFromStandardFileSystem returns a tar filesystem similar to NewTarFileSystem but assumes the tar file described by source can be found on the standard file system.
func NewTarFileSystemFromStandardFileSystem(source string, limits ILimits) (ICloseableFS, File, error) {
	return NewTarFileSystem(NewStandardFileSystem(), source, limits)
}

func NewFs(fsType FilesystemType) FS {
	switch fsType {
	case StandardFS:
		return NewStandardFileSystem()
	case InMemoryFS:
		return NewInMemoryFileSystem()
	default:
		return NewStandardFileSystem()
	}
}

// ConvertFileSystemError converts file system error into common errors
func ConvertFileSystemError(err error) error {
	err = commonerrors.ConvertContextError(platform.ConvertError(err))
	switch {
	case err == nil:
		return nil
	case commonerrors.Any(err, commonerrors.ErrTimeout, commonerrors.ErrCancelled):
		return err
	case os.IsTimeout(err) || commonerrors.Any(err, os.ErrDeadlineExceeded) || commonerrors.CorrespondTo(err, "i/o timeout"):
		return commonerrors.WrapError(commonerrors.ErrTimeout, err, "")
	case commonerrors.Any(err, os.ErrExist, afero.ErrFileExists, afero.ErrDestinationExists) || os.IsExist(err) || commonerrors.CorrespondTo(err, "file exists", "file already exists"):
		return commonerrors.WrapError(commonerrors.ErrExists, err, "")
	case commonerrors.CorrespondTo(err, "bad file descriptor") || os.IsPermission(err) || commonerrors.Any(err, os.ErrPermission, os.ErrClosed, afero.ErrFileClosed, ErrPathNotExist, io.ErrClosedPipe):
		return commonerrors.WrapError(commonerrors.ErrConflict, err, "")
	case isPrivilegeError(err) || commonerrors.CorrespondTo(err, "required privilege is not held", "operation not permitted"):
		return commonerrors.WrapError(commonerrors.ErrForbidden, err, "")
	case os.IsNotExist(err) || commonerrors.Any(err, os.ErrNotExist, afero.ErrFileNotFound) || IsPathNotExist(err) || commonerrors.CorrespondTo(err, "No such file or directory"):
		return commonerrors.WrapError(commonerrors.ErrNotFound, err, "")
	case commonerrors.Any(err, os.ErrNoDeadline):
		return commonerrors.WrapError(commonerrors.ErrUnsupported, err, "file type does not support deadline")
	case commonerrors.Any(err, os.ErrInvalid):
		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "")
	case commonerrors.Any(err, afero.ErrOutOfRange):
		return commonerrors.WrapError(commonerrors.ErrOutOfRange, err, "")
	case commonerrors.Any(err, afero.ErrTooLarge):
		return commonerrors.WrapError(commonerrors.ErrTooLarge, err, "")
	case commonerrors.Any(err, ErrChownNotImplemented, ErrLinkNotImplemented):
		return commonerrors.WrapError(commonerrors.ErrNotImplemented, err, "")
	case commonerrors.Any(err, io.ErrUnexpectedEOF):
		// Do not add io.EOF as it is used to read files
		return commonerrors.WrapError(commonerrors.ErrEOF, err, "")
	case isNotSupportedError(err) || commonerrors.Any(err, afero.ErrNoSymlink, afero.ErrNoReadlink) || commonerrors.CorrespondTo(err, "not supported"):
		return commonerrors.WrapError(commonerrors.ErrUnsupported, err, "")
	}

	return err
}
