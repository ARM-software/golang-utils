/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/afero"

	"github.com/ARM-software/golang-utils/utils/reflection"
)

var (
	_ afero.Fs    = (*osBasePathFs)(nil)
	_ IStater     = (*osBasePathFs)(nil)
	_ IChowner    = (*osBasePathFs)(nil)
	_ ILinker     = (*osBasePathFs)(nil)
	_ ILinkReader = (*osBasePathFs)(nil)
	_ ISymLinker  = (*osBasePathFs)(nil)
)

// osBasePathFs wraps os.Root and adapts it to afero.Fs plus the optional
// filesystem interfaces used by this package.
// FIXME replace this adapter with native afero support if
// https://github.com/spf13/afero/pull/724 lands and
// https://github.com/spf13/afero/issues/463 is closed.
type osBasePathFs struct {
	root     *os.Root
	basePath string
}

type osBasePathFile struct {
	afero.File
	name string
}

func (f *osBasePathFile) Name() string {
	return f.name
}

func displayOsBasePath(name string) string {
	if reflection.IsEmpty(name) {
		return "."
	}

	return filepath.Clean(name)
}

func newOsBasePathFSAdapter(basePath string) (*osBasePathFs, error) {
	cleanBasePath := filepath.Clean(basePath)
	root, err := os.OpenRoot(cleanBasePath)
	if err != nil {
		return nil, ConvertFileSystemError(err)
	}

	return &osBasePathFs{
		root:     root,
		basePath: cleanBasePath,
	}, nil
}

func (fs *osBasePathFs) Name() string {
	return OsBasePathFS.String()
}

func (fs *osBasePathFs) Close() error {
	return fs.root.Close()
}

func (fs *osBasePathFs) Create(name string) (afero.File, error) {
	_, file, err := pathResultOp(fs, "create", name, fs.root.Create)
	if err != nil {
		return nil, err
	}

	return &osBasePathFile{File: file, name: displayOsBasePath(name)}, nil
}

func (fs *osBasePathFs) Mkdir(name string, perm os.FileMode) error {
	return pathErrorOp(fs, "mkdir", name, func(cleanName string) error {
		return fs.root.Mkdir(cleanName, perm)
	})
}

func (fs *osBasePathFs) MkdirAll(name string, perm os.FileMode) error {
	return pathErrorOp(fs, "mkdir", name, func(cleanName string) error {
		return fs.root.MkdirAll(cleanName, perm)
	})
}

func (fs *osBasePathFs) Open(name string) (afero.File, error) {
	_, file, err := pathResultOp(fs, "open", name, fs.root.Open)
	if err != nil {
		return nil, err
	}

	return &osBasePathFile{File: file, name: displayOsBasePath(name)}, nil
}

func (fs *osBasePathFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	_, file, err := pathResultOp(fs, "openfile", name, func(cleanName string) (*os.File, error) {
		return fs.root.OpenFile(cleanName, flag, perm)
	})
	if err != nil {
		return nil, err
	}

	return &osBasePathFile{File: file, name: displayOsBasePath(name)}, nil
}

func (fs *osBasePathFs) Remove(name string) error {
	return pathErrorOp(fs, "remove", name, fs.root.Remove)
}

func (fs *osBasePathFs) RemoveAll(path string) error {
	return pathErrorOp(fs, "remove_all", path, fs.root.RemoveAll)
}

func (fs *osBasePathFs) Rename(oldname, newname string) error {
	return linkErrorOp(fs, "rename", oldname, newname, fs.root.Rename)
}

func (fs *osBasePathFs) Stat(name string) (os.FileInfo, error) {
	_, fileInfo, err := pathResultOp(fs, "stat", name, fs.root.Stat)
	return fileInfo, err
}

func (fs *osBasePathFs) Chmod(name string, mode os.FileMode) error {
	return pathErrorOp(fs, "chmod", name, func(cleanName string) error {
		return fs.root.Chmod(cleanName, mode)
	})
}

func (fs *osBasePathFs) Chown(name string, uid, gid int) error {
	return pathErrorOp(fs, "chown", name, func(cleanName string) error {
		return fs.root.Chown(cleanName, uid, gid)
	})
}

func (fs *osBasePathFs) Chtimes(name string, atime, mtime time.Time) error {
	return pathErrorOp(fs, "chtimes", name, func(cleanName string) error {
		return fs.root.Chtimes(cleanName, atime, mtime)
	})
}

func (fs *osBasePathFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	_, fileInfo, err := pathResultOp(fs, "lstat", name, fs.root.Lstat)
	if err != nil {
		return nil, true, err
	}

	return fileInfo, true, nil
}

func (fs *osBasePathFs) Lstat(name string) (os.FileInfo, error) {
	fileInfo, _, err := fs.LstatIfPossible(name)
	return fileInfo, err
}

func (fs *osBasePathFs) ChownIfPossible(name string, uid int, gid int) error {
	return fs.Chown(name, uid, gid)
}

func (fs *osBasePathFs) LinkIfPossible(oldname, newname string) error {
	return linkErrorOp(fs, "link", oldname, newname, fs.root.Link)
}

func (fs *osBasePathFs) Link(oldname, newname string) error {
	return fs.LinkIfPossible(oldname, newname)
}

func (fs *osBasePathFs) ReadlinkIfPossible(name string) (string, error) {
	cleanName, linkTarget, err := pathResultOp(fs, "readlink", name, fs.root.Readlink)
	if err != nil {
		return "", err
	}

	resolvedTarget := filepath.Clean(filepath.Join(filepath.Dir(cleanName), linkTarget))
	if !filepath.IsLocal(resolvedTarget) {
		return "", &os.PathError{Op: "readlink", Path: name, Err: os.ErrNotExist}
	}

	return filepath.Join(fs.basePath, resolvedTarget), nil
}

func (fs *osBasePathFs) Readlink(name string) (string, error) {
	return fs.ReadlinkIfPossible(name)
}

func (fs *osBasePathFs) SymlinkIfPossible(oldname, newname string) error {
	cleanOldName, cleanNewName, err := cleanLinkPaths(fs, oldname, newname)
	if err != nil {
		return &os.LinkError{Op: "symlink", Old: oldname, New: newname, Err: err}
	}

	linkTarget, err := filepath.Rel(filepath.Dir(cleanNewName), cleanOldName)
	if err != nil {
		return &os.LinkError{Op: "symlink", Old: oldname, New: newname, Err: err}
	}

	return fs.root.Symlink(linkTarget, cleanNewName)
}

func (fs *osBasePathFs) Symlink(oldname, newname string) error {
	return fs.SymlinkIfPossible(oldname, newname)
}

func normaliseOsBasePath(basePath, name string) (string, bool) {
	if reflection.IsEmpty(name) {
		return ".", true
	}
	if strings.TrimSpace(name) == "" {
		return name, false
	}

	cleanName := filepath.Clean(name)
	if !reflection.IsEmpty(filepath.VolumeName(cleanName)) || filepath.IsAbs(cleanName) {
		relPath, err := filepath.Rel(basePath, cleanName)
		if err != nil {
			return name, false
		}
		cleanName = relPath
	}

	if !filepath.IsLocal(cleanName) && cleanName != "." {
		return name, false
	}

	return cleanName, true
}

func basePathPathConverter(basePath string) func(path string) string {
	cleanBasePath := filepath.Clean(basePath)
	return func(path string) string {
		converted, ok := normaliseOsBasePath(cleanBasePath, path)
		if ok {
			return converted
		}
		return path
	}
}

func (fs *osBasePathFs) rootPath(name string) (string, error) {
	cleanName, ok := normaliseOsBasePath(fs.basePath, name)
	if !ok {
		return name, os.ErrNotExist
	}
	return cleanName, nil
}

func pathErrorOp(fs *osBasePathFs, op, name string, fn func(string) error) error {
	cleanName, err := fs.rootPath(name)
	if err != nil {
		return &os.PathError{Op: op, Path: name, Err: err}
	}
	return fn(cleanName)
}

func pathResultOp[T any](fs *osBasePathFs, op, name string, fn func(string) (T, error)) (string, T, error) {
	var zero T

	cleanName, err := fs.rootPath(name)
	if err != nil {
		return "", zero, &os.PathError{Op: op, Path: name, Err: err}
	}

	value, err := fn(cleanName)
	if err != nil {
		return cleanName, zero, err
	}

	return cleanName, value, nil
}

func linkErrorOp(fs *osBasePathFs, op, oldname, newname string, fn func(string, string) error) error {
	cleanOldName, cleanNewName, err := cleanLinkPaths(fs, oldname, newname)
	if err != nil {
		return &os.LinkError{Op: op, Old: oldname, New: newname, Err: err}
	}

	return fn(cleanOldName, cleanNewName)
}

func cleanLinkPaths(fs *osBasePathFs, oldname, newname string) (string, string, error) {
	cleanOldName, err := fs.rootPath(oldname)
	if err != nil {
		return "", "", err
	}

	cleanNewName, err := fs.rootPath(newname)
	if err != nil {
		return "", "", err
	}

	return cleanOldName, cleanNewName, nil
}
