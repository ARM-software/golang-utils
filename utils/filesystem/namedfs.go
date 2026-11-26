/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filesystem

import (
	"os"

	"github.com/spf13/afero"
)

type namedAferoFs struct {
	afero.Fs
	name string
}

func (fs *namedAferoFs) Name() string {
	return fs.name
}

// withFilesystemTypeName wraps an afero filesystem so Name reports the package
// FilesystemType string while preserving optional filesystem interfaces used by
// this package.
func withFilesystemTypeName(fs afero.Fs, fsType FilesystemType) afero.Fs {
	if fs == nil {
		return nil
	}

	return &namedAferoFs{
		Fs:   fs,
		name: fsType.String(),
	}
}

func (fs *namedAferoFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	if underlying, ok := fs.Fs.(afero.Lstater); ok {
		return underlying.LstatIfPossible(name)
	}
	if underlying, ok := fs.Fs.(IStater); ok {
		return underlying.LstatIfPossible(name)
	}
	fileInfo, err := fs.Fs.Stat(name)
	return fileInfo, false, err
}

func (fs *namedAferoFs) SymlinkIfPossible(oldname, newname string) error {
	if underlying, ok := fs.Fs.(afero.Linker); ok {
		return underlying.SymlinkIfPossible(oldname, newname)
	}
	if underlying, ok := fs.Fs.(ISymLinker); ok {
		return underlying.SymlinkIfPossible(oldname, newname)
	}
	return afero.ErrNoSymlink
}

func (fs *namedAferoFs) ReadlinkIfPossible(name string) (string, error) {
	if underlying, ok := fs.Fs.(afero.LinkReader); ok {
		return underlying.ReadlinkIfPossible(name)
	}
	if underlying, ok := fs.Fs.(ILinkReader); ok {
		return underlying.ReadlinkIfPossible(name)
	}
	return "", afero.ErrNoReadlink
}

func (fs *namedAferoFs) ChownIfPossible(name string, uid int, gid int) error {
	if underlying, ok := fs.Fs.(IChowner); ok {
		return underlying.ChownIfPossible(name, uid, gid)
	}
	return ErrChownNotImplemented
}

func (fs *namedAferoFs) LinkIfPossible(oldname, newname string) error {
	if underlying, ok := fs.Fs.(ILinker); ok {
		return underlying.LinkIfPossible(oldname, newname)
	}
	return ErrLinkNotImplemented
}
