/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package filesystem

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bmatcuk/doublestar/v3"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/spf13/afero"
	"go.uber.org/atomic"
	"golang.org/x/text/encoding/unicode"

	"github.com/ARM-software/golang-utils/utils/charset"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/platform"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

var (
	ErrLinkNotImplemented  = fmt.Errorf("link not implemented: %w", commonerrors.ErrNotImplemented)
	ErrChownNotImplemented = fmt.Errorf("chown not implemented: %w", commonerrors.ErrNotImplemented)
	ErrPathNotExist        = errors.New("readdirent: no such file or directory")
	globalFileSystem       = NewFs(StandardFS)
)

const (
	UnsetFileHandle = ^uint64(0)
)

type UsageStat struct {
	Total             uint64
	Free              uint64
	Used              uint64
	UsedPercent       float64
	InodesTotal       uint64
	InodesUsed        uint64
	InodesFree        uint64
	InodesUsedPercent float64
}

func (d *UsageStat) GetTotal() uint64              { return d.Total }
func (d *UsageStat) GetFree() uint64               { return d.Free }
func (d *UsageStat) GetUsed() uint64               { return d.Used }
func (d *UsageStat) GetUsedPercent() float64       { return d.UsedPercent }
func (d *UsageStat) GetInodesTotal() uint64        { return d.InodesTotal }
func (d *UsageStat) GetInodesUsed() uint64         { return d.InodesUsed }
func (d *UsageStat) GetInodesFree() uint64         { return d.Free }
func (d *UsageStat) GetInodesUsedPercent() float64 { return d.InodesUsedPercent }

func IdentityPathConverterFunc(path string) string {
	return path
}

type VFS struct {
	vfs           afero.Fs
	fsType        int
	pathConverter func(path string) string
}

func NewVirtualFileSystem(vfs afero.Fs, fsType int, pathConverter func(path string) string) FS {
	return &VFS{
		vfs:           vfs,
		fsType:        fsType,
		pathConverter: pathConverter,
	}
}

func GetGlobalFileSystem() FS {
	return globalFileSystem
}

func GetType() int {
	return globalFileSystem.GetType()
}

// Walk walks  https://golang.org/pkg/path/filepath/#WalkDir
func (fs *VFS) Walk(root string, fn filepath.WalkFunc) error {
	return fs.WalkWithContext(context.Background(), root, fn)
}

func (fs *VFS) WalkWithContext(ctx context.Context, root string, fn filepath.WalkFunc) error {
	info, err := fs.Lstat(root)
	if err != nil {
		err = fn(root, nil, err)
	} else {
		err = fs.walk(ctx, root, info, fn)
	}
	if commonerrors.Any(err, filepath.SkipDir) {
		return nil
	}

	return err
}

// walks recursively descends path, calling fn.
func (fs *VFS) walk(ctx context.Context, path string, info os.FileInfo, fn filepath.WalkFunc) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = fn(path, info, nil)
	if err != nil || !info.IsDir() {
		if commonerrors.Any(err, filepath.SkipDir) && info.IsDir() {
			err = nil
		}
		return
	}
	items, err := fs.Ls(path)
	if err != nil {
		err = fn(path, info, err)
		if err != nil {
			return
		}
	}
	for _, name := range items {
		subErr := parallelisation.DetermineContextError(ctx)
		if subErr != nil {
			return subErr
		}
		filename := filepath.Join(path, name)
		fileInfo, subErr := fs.Lstat(filename)
		if subErr != nil {
			if err := fn(filename, fileInfo, subErr); err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			subErr = fs.walk(ctx, filename, fileInfo, fn)
			if subErr != nil {
				if !fileInfo.IsDir() || subErr != filepath.SkipDir {
					return subErr
				}
			}
		}
	}
	return nil

}

func (fs *VFS) GetType() int {
	return fs.fsType
}

func (fs *VFS) ConvertFilePath(name string) string {
	return fs.pathConverter(name)
}

func TempDirectory() string {
	return globalFileSystem.TempDirectory()
}
func (fs *VFS) TempDirectory() string {
	return afero.GetTempDir(fs.vfs, "")
}

func CurrentDirectory() (string, error) {
	return globalFileSystem.CurrentDirectory()
}
func (fs *VFS) CurrentDirectory() (string, error) {
	return os.Getwd()
}

func Lstat(name string) (fileInfo os.FileInfo, err error) {
	return globalFileSystem.Lstat(name)
}

func (fs *VFS) Lstat(name string) (fileInfo os.FileInfo, err error) {
	if correctobj, ok := fs.vfs.(interface {
		LstatIfPossible(string) (os.FileInfo, bool, error)
	}); ok {
		fileInfo, _, err = correctobj.LstatIfPossible(name)
		return
	}
	fileInfo, err = fs.Stat(name)
	if err != nil {
		err = commonerrors.ErrNotImplemented
	}
	return
}

func (fs *VFS) Open(name string) (doublestar.File, error) {
	return fs.GenericOpen(name)
}

func GenericOpen(name string) (File, error) {
	return globalFileSystem.GenericOpen(name)
}
func (fs *VFS) GenericOpen(name string) (File, error) {
	return convertFile(func() (afero.File, error) { return fs.vfs.Open(name) }, func() error { return nil })
}

func OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return globalFileSystem.OpenFile(name, flag, perm)
}

func (fs *VFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return convertFile(func() (afero.File, error) { return fs.vfs.OpenFile(name, flag, perm) }, func() error { return nil })
}

func CreateFile(name string) (File, error) {
	return globalFileSystem.CreateFile(name)
}
func (fs *VFS) CreateFile(name string) (File, error) {
	return convertFile(func() (afero.File, error) { return fs.vfs.Create(name) }, func() error { return nil })
}

func (fs *VFS) NewRemoteLockFile(id string, dirToLock string) ILock {
	return NewRemoteLockFile(fs, id, dirToLock)
}

func ReadFile(name string) ([]byte, error) {
	return globalFileSystem.ReadFile(name)
}

func (fs *VFS) ReadFile(filename string) (content []byte, err error) {
	return fs.readFileWithLimits(filename, NoLimits())
}

func ReadFileWithLimits(filename string, limits ILimits) ([]byte, error) {
	return globalFileSystem.ReadFileWithLimits(filename, limits)
}

func (fs *VFS) ReadFileWithLimits(filename string, limits ILimits) ([]byte, error) {
	return fs.readFileWithLimits(filename, limits)
}

func (fs *VFS) readFileWithLimits(filename string, limits ILimits) (content []byte, err error) {
	if limits == nil {
		err = fmt.Errorf("%w: missing file system limits definition", commonerrors.ErrUndefined)
		return
	}
	// Really similar to afero iotutils Read file but using our utilities instead.
	f, err := fs.GenericOpen(filename)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	var bufferCapacity int64 = bytes.MinRead
	fi, err := f.Stat()
	if err == nil {
		// Don't preallocate a huge buffer, just in case.
		size := fi.Size()
		if size < 1e9 {
			bufferCapacity += size
		}
		if limits.Apply() && size > limits.GetMaxFileSize() {
			err = commonerrors.ErrEOF
			return
		}
	}
	buf := bytes.NewBuffer(make([]byte, 0, bufferCapacity))
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	if limits.Apply() {
		_, err = buf.ReadFrom(io.LimitReader(f, limits.GetMaxFileSize()))
	} else {
		_, err = buf.ReadFrom(f)
	}
	if commonerrors.Any(err, io.EOF, io.ErrUnexpectedEOF) {
		err = fmt.Errorf("%w: %v", commonerrors.ErrEOF, err.Error())
	}
	content = buf.Bytes()
	return
}

func (fs *VFS) WriteFile(filename string, data []byte, perm os.FileMode) (err error) {
	f, err := fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	n, err := f.Write(data)
	if err != nil {
		return
	}
	if n < len(data) {
		err = io.ErrShortWrite
		return
	}
	err = f.Close()
	return
}

func PathSeparator() rune {
	return globalFileSystem.PathSeparator()
}

func (fs *VFS) PathSeparator() rune {
	return os.PathSeparator
}

func Stat(name string) (os.FileInfo, error) {
	return globalFileSystem.Stat(name)
}

func (fs *VFS) Stat(name string) (os.FileInfo, error) {
	return fs.vfs.Stat(name)
}

func (fs *VFS) StatTimes(name string) (info FileTimeInfo, err error) {
	stat, err := fs.Stat(name)
	if err == nil || stat == nil {
		return DetermineFileTimes(stat)
	}
	return
}

func TempDir(dir string, prefix string) (name string, err error) {
	return globalFileSystem.TempDir(dir, prefix)
}

func (fs *VFS) TempDir(dir string, prefix string) (name string, err error) {
	return afero.TempDir(fs.vfs, dir, prefix)
}

func TempDirInTempDir(prefix string) (name string, err error) {
	return globalFileSystem.TempDirInTempDir(prefix)
}

func (fs *VFS) TempDirInTempDir(prefix string) (name string, err error) {
	return fs.TempDir("", prefix)
}
func TempFile(dir string, pattern string) (f File, err error) {
	return globalFileSystem.TempFile(dir, pattern)
}
func (fs *VFS) TempFile(dir string, prefix string) (f File, err error) {
	file, err := afero.TempFile(fs.vfs, dir, prefix)
	if err != nil {
		return
	}
	return convertToExtendedFile(file, func() error { return nil })
}

func TempFileInTempDir(pattern string) (f File, err error) {
	return globalFileSystem.TempFileInTempDir(pattern)
}

func (fs *VFS) TempFileInTempDir(prefix string) (f File, err error) {
	return fs.TempFile("", prefix)
}

func CleanDir(dir string) (err error) {
	return globalFileSystem.CleanDir(dir)
}

func (fs *VFS) CleanDir(dir string) error {
	return fs.CleanDirWithContext(context.Background(), dir)
}

func (fs *VFS) CleanDirWithContext(ctx context.Context, dir string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if dir == "" || !fs.Exists(dir) {
		return
	}
	empty, err := fs.IsEmpty(dir)
	if empty || err != nil {
		return
	}
	files, err := fs.Ls(dir)
	if err != nil {
		return
	}
	for i := range files {
		subErr := fs.removeFileWithContext(ctx, dir, files[i])
		if subErr != nil {
			err = subErr
			return
		}
	}
	return
}

func (fs *VFS) removeFileWithContext(ctx context.Context, dir, f string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = fs.RemoveWithContext(ctx, filepath.Join(dir, f))
	return
}

// Checks if a file or folder exists
func Exists(path string) bool {
	return globalFileSystem.Exists(path)
}
func (fs *VFS) Exists(path string) bool {
	fi, err := fs.Stat(path)
	if err != nil {
		if IsPathNotExist(err) {
			return false
		}
	}
	if fi == nil {
		return false
	}
	// Double check for directories as it was seen on Docker that Stat would work on Docker even if path does not exist
	if fi.IsDir() {
		return fs.checkDirExists(path)
	}
	return true
}

func (fs *VFS) checkDirExists(path string) (exist bool) {
	exist = false
	f, err := fs.vfs.Open(path)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, err = f.Readdirnames(1)
	if IsPathNotExist(err) {
		exist = false
	} else {
		exist = true
	}
	_ = f.Close()
	return
}

func Rm(dir string) (err error) {
	return globalFileSystem.Rm(dir)
}
func (fs *VFS) Rm(dir string) error {
	return fs.RemoveWithContext(context.Background(), dir)
}

func (fs *VFS) RemoveWithContext(ctx context.Context, dir string) (err error) {
	if dir == "" {
		return
	}
	if !fs.Exists(dir) {
		return
	}
	isDir, err := fs.IsDir(dir)
	if err != nil {
		return
	}
	isEmpty, err := fs.IsEmpty(dir)
	if err != nil {
		return
	}
	if isDir && !isEmpty {
		err = fs.CleanDirWithContext(ctx, dir)
	}
	if err != nil {
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	return fs.vfs.Remove(dir)
}

// States whether it is a file or not
func IsFile(path string) (result bool, err error) {
	return globalFileSystem.IsFile(path)
}
func (fs *VFS) IsFile(path string) (result bool, err error) {
	if !fs.Exists(path) {
		return
	}
	fi, err := fs.Stat(path)
	if err != nil {
		return
	}
	result = IsRegularFile(fi)
	return
}

func IsRegularFile(fi os.FileInfo) bool {
	if fi == nil {
		return false
	}
	return fi.Mode().IsRegular()
}

func (fs *VFS) IsLink(path string) (result bool, err error) {
	if !fs.Exists(path) {
		return
	}
	fi, err := fs.Lstat(path)
	if err != nil {
		return
	}
	result = IsSymLink(fi)
	return
}

func IsSymLink(fi os.FileInfo) bool {
	if fi == nil {
		return false
	}
	return fi.Mode()&os.ModeType == os.ModeSymlink
}

// IsDir states whether it is a directory or not
func IsDir(path string) (result bool, err error) {
	return globalFileSystem.IsDir(path)
}
func (fs *VFS) IsDir(path string) (result bool, err error) {
	if !fs.Exists(path) {
		return
	}
	fi, err := fs.Stat(path)
	if err != nil {
		return
	}
	result = IsDirectory(fi)
	return
}

func IsDirectory(fi os.FileInfo) bool {
	if fi == nil {
		return false
	}
	return fi.IsDir()
}

// IsEmpty checks whether a path is empty or not
func IsEmpty(name string) (empty bool, err error) {
	return globalFileSystem.IsEmpty(name)
}
func (fs *VFS) IsEmpty(name string) (empty bool, err error) {
	if !fs.Exists(name) {
		empty = true
		return
	}
	isFile, err := fs.IsFile(name)
	if err != nil {
		return
	}
	if isFile {
		return fs.isFileEmpty(name)
	}
	return fs.isDirEmpty(name)
}

func (fs *VFS) isFileEmpty(name string) (empty bool, err error) {
	fi, err := fs.Stat(name)
	if err != nil {
		return
	}
	empty = fi.Size() == 0
	return
}

func (fs *VFS) isDirEmpty(name string) (empty bool, err error) {
	f, err := fs.vfs.Open(name)
	if err != nil {
		return
	}
	defer func() {
		if f != nil {
			_ = f.Close()
		}
	}()
	_, err = f.Readdirnames(1)
	if commonerrors.Any(err, io.EOF, commonerrors.ErrEOF) || IsPathNotExist(err) {
		err = nil
		empty = true
		return
	}
	err = f.Close()
	return
}

// MkDir makes directory (equivalent to mkdir -p)
func MkDir(dir string) (err error) {
	return globalFileSystem.MkDir(dir)
}
func (fs *VFS) MkDir(dir string) (err error) {
	return fs.MkDirAll(dir, 0755)
}

func (fs *VFS) MkDirAll(dir string, perm os.FileMode) (err error) {
	if dir == "" {
		return fmt.Errorf("missing path: %w", commonerrors.ErrUndefined)
	}
	if fs.Exists(dir) {
		return
	}
	err = fs.vfs.MkdirAll(dir, perm)
	// Directory was maybe created by a different process/thread
	if err != nil && fs.Exists(dir) {
		err = nil
	}
	return
}

// ExcludeAll excludes files
func ExcludeAll(files []string, exclusionPatterns ...string) ([]string, error) {
	return globalFileSystem.ExcludeAll(files, exclusionPatterns...)
}

func (fs *VFS) ExcludeAll(files []string, exclusionPatterns ...string) ([]string, error) {
	regexes := []*regexp.Regexp{}
	patternsExtendedList := []string{}
	for _, p := range exclusionPatterns {
		if p != "" {
			patternsExtendedList = append(patternsExtendedList, p, fmt.Sprintf(".*/%v/.*", p), fmt.Sprintf(".*%v%v%v.*", fs.PathSeparator(), p, fs.PathSeparator()))
		}
	}
	for _, p := range patternsExtendedList {
		r, err := regexp.Compile(p)
		if err != nil {
			return nil, err
		}
		regexes = append(regexes, r)
	}
	cleansedList := []string{}
	for _, f := range files {
		if !isExcluded(f, regexes...) {
			cleansedList = append(cleansedList, f)
		}
	}
	return cleansedList, nil
}

func isExcluded(path string, exclusionPatterns ...*regexp.Regexp) bool {
	for _, p := range exclusionPatterns {
		if p.MatchString(path) {
			return true
		}
	}
	return false
}

// FindAll finds all the files with extensions
func FindAll(dir string, extensions ...string) (files []string, err error) {
	return globalFileSystem.FindAll(dir, extensions...)
}
func (fs *VFS) FindAll(dir string, extensions ...string) (files []string, err error) {
	files = []string{}
	if !fs.Exists(dir) {
		return
	}
	for _, ext := range extensions {
		foundFiles, err := fs.findAllOfExtension(dir, ext)
		if err != nil {
			return files, err
		}
		files = append(files, foundFiles...)
	}
	return
}
func (fs *VFS) findAllOfExtension(dir string, ext string) (files []string, err error) {
	if strings.HasPrefix(ext, ".") {
		ext = string(ext[1:])
	}
	return doublestar.GlobOS(fs, filepath.Join(dir, "**", "*."+ext))
}

func (fs *VFS) Chmod(name string, mode os.FileMode) error {
	return fs.vfs.Chmod(name, mode)
}

func (fs *VFS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return fs.vfs.Chtimes(name, atime, mtime)
}

func (fs *VFS) Chown(name string, uid, gid int) (err error) {
	if correctobj, ok := fs.vfs.(interface {
		ChownIfPossible(string, int, int) error
	}); ok {
		err = correctobj.ChownIfPossible(name, uid, gid)
		return
	}
	err = commonerrors.ErrNotImplemented
	return
}

func (fs *VFS) Link(oldname, newname string) (err error) {
	if correctobj, ok := fs.vfs.(interface {
		LinkIfPossible(string, string) error
	}); ok {
		err = correctobj.LinkIfPossible(oldname, newname)
		return
	}
	err = commonerrors.ErrNotImplemented
	return
}

func (fs *VFS) Readlink(name string) (value string, err error) {
	if correctobj, ok := fs.vfs.(interface {
		ReadlinkIfPossible(string) (string, error)
	}); ok {
		value, err = correctobj.ReadlinkIfPossible(name)
		return
	}
	err = commonerrors.ErrNotImplemented
	return
}

func (fs *VFS) Symlink(oldname string, newname string) (err error) {
	if correctobj, ok := fs.vfs.(interface {
		SymlinkIfPossible(string, string) error
	}); ok {
		err = correctobj.SymlinkIfPossible(oldname, newname)
		return
	}
	err = commonerrors.ErrNotImplemented
	return
}

func Ls(dir string) (files []string, err error) {
	return globalFileSystem.Ls(dir)
}
func (fs *VFS) Ls(dir string) (names []string, err error) {
	if isDir, err := fs.IsDir(dir); !isDir || err != nil {
		err = fmt.Errorf("path [%v] is not a directory: %w", dir, commonerrors.ErrInvalid)
		return nil, err
	}
	f, err := fs.GenericOpen(dir)
	if err != nil {
		return nil, err
	}
	names, err = fs.LsFromOpenedDirectory(f)
	_ = f.Close()
	return names, err
}

func (fs *VFS) LsFromOpenedDirectory(dir File) ([]string, error) {
	if dir == nil {
		return nil, fmt.Errorf("%w: nil directory", commonerrors.ErrUndefined)
	}
	return dir.Readdirnames(-1)
}

func (fs *VFS) Lls(dir string) (files []os.FileInfo, err error) {
	if isDir, err := fs.IsDir(dir); !isDir || err != nil {
		err = fmt.Errorf("path [%v] is not a directory: %w", dir, commonerrors.ErrInvalid)
		return nil, err
	}
	f, err := fs.GenericOpen(dir)
	if err != nil {
		return nil, err
	}
	files, err = fs.LlsFromOpenedDirectory(f)
	_ = f.Close()
	return
}

func (fs *VFS) LlsFromOpenedDirectory(dir File) ([]os.FileInfo, error) {
	if dir == nil {
		return nil, fmt.Errorf("%w: nil directory", commonerrors.ErrUndefined)
	}
	return dir.Readdir(-1)
}

func (fs *VFS) ConvertToAbsolutePath(rootPath string, paths ...string) ([]string, error) {
	basepath := fs.ConvertFilePath(rootPath)
	converted := make([]string, 0, len(paths))
	for i := range paths {
		path := paths[i]
		var abs string
		if filepath.IsAbs(path) {
			abs = fs.ConvertFilePath(path)
		} else {
			abs = fs.ConvertFilePath(filepath.Join(basepath, path))
		}
		converted = append(converted, abs)
	}
	return converted, nil
}

func (fs *VFS) ConvertToRelativePath(rootPath string, paths ...string) ([]string, error) {
	basepath := fs.ConvertFilePath(rootPath)
	converted := make([]string, 0, len(paths))
	for i := range paths {
		path := paths[i]
		relPath, err := filepath.Rel(basepath, fs.ConvertFilePath(path))
		if err != nil {
			return nil, err
		}
		converted = append(converted, relPath)
	}
	return converted, nil
}

// Move moves a file (equivalent to mv)
func Move(src string, dest string) error {
	return globalFileSystem.Move(src, dest)
}
func (fs *VFS) Move(src string, dest string) error {
	return fs.MoveWithContext(context.Background(), src, dest)
}

func (fs *VFS) MoveWithContext(ctx context.Context, src string, dest string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if src == dest {
		return
	}
	if !fs.Exists(src) {
		err = fmt.Errorf("path [%v] does not exist: %w", src, commonerrors.ErrNotFound)
		return
	}
	err = fs.MkDir(filepath.Dir(dest))
	if err != nil {
		return
	}
	err = fs.vfs.Rename(src, dest)
	if err == nil {
		return
	}
	//os.Rename() give error "invalid cross-device link" for Docker container with Volumes.
	isDir, err := fs.IsDir(src)
	if err != nil {
		return
	}
	if isDir {
		err = fs.moveFolder(ctx, src, dest)
	} else {
		err = fs.moveFile(ctx, src, dest)
	}
	return
}

func (fs *VFS) moveFolder(ctx context.Context, src string, dest string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = fs.MkDir(dest)
	if err != nil {
		return
	}
	empty, err := fs.IsEmpty(src)
	if err != nil {
		return
	}
	if !empty {
		files, err := fs.Ls(src)
		if err != nil {
			if IsPathNotExist(err) {
				return nil
			}
			return err
		}
		for i := range files {
			f := files[i]
			err = fs.MoveWithContext(ctx, filepath.Join(src, f), filepath.Join(dest, f))
			if err != nil {
				return err
			}
		}
	}
	err = fs.RemoveWithContext(ctx, src)
	return
}

func IsPathNotExist(err error) bool {
	if err == nil {
		return false
	}
	return os.IsNotExist(err) || commonerrors.Any(err, ErrPathNotExist)
}

func (fs *VFS) moveFile(ctx context.Context, src string, dest string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if src == dest {
		return
	}
	err = copyFileBetweenFS(ctx, fs, src, fs, dest)
	if err != nil {
		return
	}
	err = fs.vfs.Remove(src)
	if err != nil {
		return
	}
	return
}

func (fs *VFS) FileHash(hashAlgo string, path string) (hash string, err error) {
	hasher, err := NewFileHash(hashAlgo)
	if err != nil {
		return
	}
	hash, err = hasher.CalculateFile(fs, path)
	return
}

func Copy(src string, dest string) (err error) {
	return globalFileSystem.Copy(src, dest)
}
func (fs *VFS) Copy(src string, dest string) (err error) {
	return fs.CopyWithContext(context.Background(), src, dest)
}

func (fs *VFS) CopyWithContext(ctx context.Context, src string, dest string) (err error) {
	return CopyBetweenFS(ctx, fs, src, fs, dest)
}

func MoveBetweenFS(ctx context.Context, srcFs FS, src string, destFs FS, dest string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if srcFs == destFs && src == dest {
		return
	}
	err = CopyBetweenFS(ctx, srcFs, src, destFs, dest)
	if err != nil {
		return
	}
	return srcFs.RemoveWithContext(ctx, src)
}

func CopyBetweenFS(ctx context.Context, srcFs FS, src string, destFs FS, dest string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if srcFs == destFs && src == dest {
		return
	}
	if !srcFs.Exists(src) {
		err = fmt.Errorf("path [%v] does not exist: %w", src, commonerrors.ErrNotFound)
		return
	}
	err = destFs.MkDir(dest)
	if err != nil {
		return
	}

	isDir, err := srcFs.IsDir(src)
	if err != nil {
		return
	}
	dst := filepath.Join(dest, filepath.Base(src))
	if isDir {
		err = copyFolderBetweenFS(ctx, srcFs, src, destFs, dst)
	} else {
		err = copyFileBetweenFS(ctx, srcFs, src, destFs, dst)
	}
	return
}

func copyFolderBetweenFS(ctx context.Context, srcFs FS, src string, destFs FS, dest string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = destFs.MkDir(dest)
	if err != nil {
		return
	}
	empty, err := srcFs.IsEmpty(src)
	if err != nil {
		return
	}
	if !empty {
		files, err := srcFs.Ls(src)
		if err != nil {
			return err
		}
		for i := range files {
			srcPath := filepath.Join(src, files[i])
			err = CopyBetweenFS(ctx, srcFs, srcPath, destFs, dest)
			if err != nil {
				return err
			}
		}
	}
	return
}

func copyFileBetweenFS(ctx context.Context, srcFs FS, src string, destFs FS, dest string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	inputFile, err := srcFs.GenericOpen(src)
	if err != nil {
		return
	}
	defer func() { _ = inputFile.Close() }()
	outputFile, err := destFs.CreateFile(dest)
	if err != nil {

		return
	}
	defer func() { _ = outputFile.Close() }()
	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		return
	}
	err = inputFile.Close()
	if err != nil {
		return
	}
	err = outputFile.Close()
	if err != nil {
		return
	}
	return
}

func (fs *VFS) DiskUsage(name string) (usage DiskUsage, err error) {
	realPath := fs.pathConverter(name)
	du, err := disk.Usage(realPath)
	if err != nil {
		return
	}
	usage = &UsageStat{
		Total:             du.Total,
		Free:              du.Free,
		Used:              du.Used,
		UsedPercent:       du.UsedPercent,
		InodesTotal:       du.InodesTotal,
		InodesUsed:        du.InodesUsed,
		InodesFree:        du.InodesFree,
		InodesUsedPercent: du.InodesUsedPercent,
	}
	return
}

func GetFileSize(name string) (size int64, err error) {
	return globalFileSystem.GetFileSize(name)
}

func (fs *VFS) GetFileSize(name string) (size int64, err error) {
	info, err := fs.Stat(name)
	if err != nil {
		return
	}
	size = info.Size()
	return
}

func Zip(source string, destination string) error {
	return globalFileSystem.Zip(source, destination)
}

func (fs *VFS) Zip(source, destination string) error {
	return fs.ZipWithContext(context.Background(), source, destination)
}

func (fs *VFS) ZipWithContext(ctx context.Context, source, destination string) (err error) {
	return fs.ZipWithContextAndLimits(ctx, source, destination, NoLimits())
}

func (fs *VFS) ZipWithContextAndLimits(ctx context.Context, source, destination string, limits ILimits) (err error) {
	if limits == nil {
		err = fmt.Errorf("%w: missing file system limits", commonerrors.ErrUndefined)
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	file, err := fs.CreateFile(destination)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	// create a new zip archive
	w := zip.NewWriter(file)
	defer func() { _ = w.Close() }()

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if limits.Apply() && info.Size() > limits.GetMaxFileSize() {
			err = fmt.Errorf("%w: file [%v] is too big (%v B) and beyond limits (max: %v B)", commonerrors.ErrTooLarge, path, info.Size(), limits.GetMaxFileSize())
			return err
		}

		// Get the relative path
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		// If directory
		if info.IsDir() {
			if path == source {
				return nil
			}
			header := &zip.FileHeader{
				Name:     relPath + "/",
				Method:   zip.Deflate,
				Modified: info.ModTime(),
			}
			_, err = w.CreateHeader(header)
			return err
		}

		// if file
		src, err := fs.GenericOpen(path)
		if err != nil {
			return err
		}
		defer func() { _ = src.Close() }()

		// create file in archive
		relPath, err = filepath.Rel(source, path)
		if err != nil {
			return err
		}
		header := &zip.FileHeader{
			Name:     relPath,
			Method:   zip.Deflate,
			Modified: info.ModTime(),
		}
		dest, err := w.CreateHeader(header)
		if err != nil {
			return err
		}
		n, err := io.Copy(dest, src)
		if err != nil {
			return err
		}

		if info.Size() != n && !IsSymLink(info) {
			return fmt.Errorf("could not write the full file [%v] content (wrote %v/%v bytes)", relPath, n, info.Size())
		}
		return nil
	}
	err = fs.WalkWithContext(ctx, source, walker)

	if limits.Apply() {
		stat, subErr := file.Stat()
		if subErr != nil {
			return subErr
		}
		if stat.Size() > limits.GetMaxFileSize() {
			subErr = fmt.Errorf("%w: file [%v] is too big (%v B) and beyond limits (max: %v B)", commonerrors.ErrTooLarge, destination, stat.Size(), limits.GetMaxFileSize())
			return subErr
		}
	}
	return
}

// Prevents any ZipSlip (files outside extraction dirPath) https://snyk.io/research/zip-slip-vulnerability#go
func sanitiseZipExtractPath(fs FS, filePath string, destination string) (destPath string, err error) {
	destPath = filepath.Join(destination, filePath) // join cleans the destpath so we can check for ZipSlip
	if destPath == destination {
		return
	}
	if strings.HasPrefix(destPath, fmt.Sprintf("%v%v", destination, string(fs.PathSeparator()))) {
		return
	}
	if strings.HasPrefix(destPath, fmt.Sprintf("%v/", destination)) {
		return
	}
	err = fmt.Errorf("%w: zipslip security breach detected, file dirPath '%s' not in destination directory '%s'", commonerrors.ErrInvalidDestination, filePath, destination)
	return
}

func Unzip(source, destination string) ([]string, error) {
	return globalFileSystem.Unzip(source, destination)
}

func (fs *VFS) Unzip(source, destination string) ([]string, error) {
	return fs.UnzipWithContext(context.Background(), source, destination)
}

func (fs *VFS) UnzipWithContext(ctx context.Context, source string, destination string) (fileList []string, err error) {
	return fs.unzip(ctx, source, destination, NoLimits())
}
func (fs *VFS) UnzipWithContextAndLimits(ctx context.Context, source string, destination string, limits ILimits) (fileList []string, err error) {
	return fs.unzip(ctx, source, destination, limits)
}
func (fs *VFS) unzip(ctx context.Context, source string, destination string, limits ILimits) (fileList []string, err error) {
	if limits == nil {
		err = fmt.Errorf("%w: missing file system limits", commonerrors.ErrUndefined)
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	// List of file paths to return

	totalSizeOnDisk := atomic.NewUint64(0)

	info, err := fs.Lstat(source)
	if err != nil {
		return
	}
	f, err := fs.GenericOpen(source)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	zipFileSize := info.Size()

	if limits.Apply() && zipFileSize > limits.GetMaxFileSize() {
		err = fmt.Errorf("%w: zip file [%v] is too big (%v B) and beyond limits (max: %v B)", commonerrors.ErrTooLarge, source, zipFileSize, limits.GetMaxFileSize())
		return
	}

	zipReader, err := zip.NewReader(f, zipFileSize)
	if err != nil {
		return
	}

	// Clean the destination to find shortest dirPath
	destination = filepath.Clean(destination)
	err = fs.MkDir(destination)
	if err != nil {
		return
	}
	directoryInfo := map[string]os.FileInfo{}

	// For each file in the zip file
	for i := range zipReader.File {
		zippedFile := zipReader.File[i]
		subErr := parallelisation.DetermineContextError(ctx)
		if subErr != nil {
			return fileList, subErr
		}

		// Calculate file dirPath
		filePath, subErr := sanitiseZipExtractPath(fs, zippedFile.Name, destination)
		if subErr != nil {
			return fileList, subErr
		}

		// Keep list of files unzipped
		fileList = append(fileList, filePath)

		if zippedFile.FileInfo().IsDir() {
			// Create directory
			subErr = fs.MkDir(filePath)

			if subErr != nil {
				return fileList, fmt.Errorf("unable to create directory [%s]: %w", filePath, subErr)
			}
			// recording directory dirInfo to preserve timestamps
			directoryInfo[filePath] = zippedFile.FileInfo()
			// Nothing more to do for a directory, move to next zip file
			continue
		}

		// If a file create the dirPath into which to write the file
		directoryPath := filepath.Dir(filePath)
		subErr = fs.MkDir(directoryPath)
		if subErr != nil {
			return fileList, fmt.Errorf("unable to create directory '%s': %w", directoryPath, subErr)
		}

		fileSizeOnDisk, subErr := fs.unzipZipFile(ctx, filePath, zippedFile, limits)
		if subErr != nil {
			return fileList, subErr
		}
		totalSizeOnDisk.Add(uint64(fileSizeOnDisk))
		if limits.Apply() && totalSizeOnDisk.Load() > limits.GetMaxTotalSize() {
			return fileList, fmt.Errorf("%w: more than %v B of disk space was used while unzipping %v (%v B used already)", commonerrors.ErrTooLarge, limits.GetMaxTotalSize(), source, totalSizeOnDisk.Load())
		}
	}

	// Ensuring directory timestamps are preserved (this needs to be done after all the files have been created).
	for dirPath, dirInfo := range directoryInfo {
		subErr := parallelisation.DetermineContextError(ctx)
		if subErr != nil {
			return fileList, subErr
		}
		times := newDefaultTimeInfo(dirInfo)
		subErr = fs.Chtimes(dirPath, times.AccessTime(), times.ModTime())
		if subErr != nil {
			return fileList, fmt.Errorf("unable to set directory timestamp [%s]: %w", dirPath, subErr)
		}
	}

	return fileList, nil
}

// unzipZipFile unzips file to destination directory
func (fs *VFS) unzipZipFile(ctx context.Context, dest string, zippedFile *zip.File, limits ILimits) (fileSizeOnDisk int64, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	destinationPath, err := determineUnzippedFilepath(dest)
	if err != nil {
		return
	}

	destinationFile, err := fs.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zippedFile.Mode())
	if err != nil {
		err = fmt.Errorf("%w: unable to open file '%s': %v", commonerrors.ErrUnexpected, destinationPath, err.Error())
		return
	}
	defer func() { _ = destinationFile.Close() }()

	sourceFile, err := zippedFile.Open()
	if err != nil {
		err = fmt.Errorf("%w: unable to open zipped file '%s': %v", commonerrors.ErrUnsupported, zippedFile.Name, err.Error())
		return
	}
	defer func() { _ = sourceFile.Close() }()

	info := zippedFile.FileInfo()
	fileSizeOnDisk = info.Size()
	if limits.Apply() {
		if fileSizeOnDisk > limits.GetMaxFileSize() {
			err = fmt.Errorf("%w: zipped file [%v] is too big (%v B) and above max size (%v B)", commonerrors.ErrTooLarge, info.Name(), info.Size(), limits.GetMaxFileSize())
			return
		}
	}

	_, err = io.CopyN(destinationFile, sourceFile, fileSizeOnDisk)
	if err != nil {
		err = fmt.Errorf("copy of zipped file to '%s' failed: %w", destinationPath, err)
		return
	}
	err = destinationFile.Close()
	if err != nil {
		return
	}
	// Ensuring the timestamp is preserved.
	times := newDefaultTimeInfo(info)
	err = fs.Chtimes(destinationPath, times.AccessTime(), times.ModTime())
	// Nothing more to do for a directory, move to next zip file
	return
}

func determineUnzippedFilepath(destinationPath string) (string, error) {

	// See https://go-review.googlesource.com/c/go/+/75592/
	// Character encodings other than CP-437 and UTF-8
	//are not officially supported by the ZIP specification, pragmatically
	//the world has permitted use of them.
	//
	//When a non-standard encoding is used, it is the user's responsibility
	//to ensure that the target system is expecting the encoding used
	//(e.g., producing a ZIP file you know is used on a Chinese version of Windows).
	if utf8.ValidString(destinationPath) {
		return destinationPath, nil
	}
	// Nonetheless, instead of raising an error when non-UTF8 characters are present in filepath,
	// we try to guess the encoding and then, convert the string to UTF-8.
	encoding, charsetName, err := charset.DetectTextEncoding([]byte(destinationPath))
	if err != nil {
		return "", fmt.Errorf("%w: file path [%s] is not a valid utf-8 string and charset could not be detected: %v", commonerrors.ErrInvalid, destinationPath, err.Error())
	}
	convertedDestinationPath, err := charset.IconvString(destinationPath, encoding, unicode.UTF8)
	if err != nil {
		return "", fmt.Errorf("%w: file path [%s] is encoded using charset [%v] but could not be converted to valid utf-8: %v", commonerrors.ErrUnexpected, destinationPath, charsetName, err.Error())
		//If zip file paths must be accepted even when their encoding is unknown, or conversion to utf-8 failed, then the following can be done.
		//destinationPath = strings.ToValidUTF8(dest, charset.InvalidUTF8CharacterReplacement)
	}
	return convertedDestinationPath, err
}

func SubDirectories(directory string) ([]string, error) {
	return globalFileSystem.SubDirectories(directory)
}

func (fs *VFS) SubDirectories(directory string) ([]string, error) {
	return fs.SubDirectoriesWithContext(context.Background(), directory)
}

func (fs *VFS) SubDirectoriesWithContext(ctx context.Context, directory string) (directories []string, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	files, err := afero.ReadDir(fs.vfs, directory)
	if err != nil {
		return
	}

	for i := range files {
		subErr := parallelisation.DetermineContextError(ctx)
		if subErr != nil {
			err = subErr
			return
		}
		file := files[i]
		if file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
			directories = append(directories, file.Name())
		}
	}
	return
}

// ListDirTree returns a list of files and directories recursively available under specified path
func ListDirTree(dirPath string, list *[]string) error {
	return globalFileSystem.ListDirTree(dirPath, list)
}

func (fs *VFS) ListDirTree(dirPath string, list *[]string) error {
	return fs.ListDirTreeWithContext(context.Background(), dirPath, list)
}

func (fs *VFS) ListDirTreeWithContext(ctx context.Context, dirPath string, list *[]string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if list == nil {
		err = fmt.Errorf("uninitialised input variable: %w", commonerrors.ErrInvalid)
		return
	}

	elements, err := fs.Ls(dirPath)
	if err != nil {
		return err
	}

	for i := range elements {
		subErr := parallelisation.DetermineContextError(ctx)
		if subErr != nil {
			err = subErr
			return
		}
		path := filepath.Join(dirPath, elements[i])
		*list = append(*list, path)
		if isDir, _ := fs.IsDir(path); isDir {
			subErr = fs.ListDirTreeWithContext(ctx, path, list)
			if subErr != nil {
				err = subErr
				return
			}
		}
	}
	return nil
}

func (fs *VFS) GarbageCollect(root string, durationSinceLastAccess time.Duration) error {
	return fs.GarbageCollectWithContext(context.Background(), root, durationSinceLastAccess)
}

func (fs *VFS) GarbageCollectWithContext(ctx context.Context, root string, durationSinceLastAccess time.Duration) error {
	return fs.garbageCollect(ctx, durationSinceLastAccess, root, false)
}

func (fs *VFS) garbageCollectFile(ctx context.Context, durationSinceLastAccess time.Duration, path string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	info, err := fs.StatTimes(path)
	if err != nil {
		return
	}
	elapsedTime := time.Since(info.AccessTime())
	if elapsedTime > durationSinceLastAccess {
		err = fs.RemoveWithContext(ctx, path)
	}
	return
}

func (fs *VFS) garbageCollect(ctx context.Context, durationSinceLastAccess time.Duration, path string, deletePath bool) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if !fs.Exists(path) {
		return
	}
	if isDir, _ := fs.IsDir(path); isDir {
		return fs.garbageCollectDir(ctx, durationSinceLastAccess, path, deletePath)
	}
	return fs.garbageCollectFile(ctx, durationSinceLastAccess, path)
}

func (fs *VFS) garbageCollectDir(ctx context.Context, durationSinceLastAccess time.Duration, path string, deletePath bool) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if deletePath && platform.IsWindows() {
		// On linux and potentially MacOs, the access/modification of files in a directory do not affect those times for the parent folder.
		// Therefore, we cannot rely on the times of a directory to make a decision.
		// See https://superuser.com/questions/1039003/linux-how-does-file-modification-time-affect-directory-modification-time-and-di
		err = fs.garbageCollectFile(ctx, durationSinceLastAccess, path)
		if err != nil {
			return err
		}
		if !fs.Exists(path) {
			return nil
		}
	}
	files, err := fs.Ls(path)
	if err != nil {
		return
	}
	_, _ = parallelisation.Parallelise(files, func(arg interface{}) (interface{}, error) {
		file := filepath.Join(path, arg.(string))
		return nil, fs.garbageCollect(ctx, durationSinceLastAccess, file, true)
	}, nil)
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if empty, subErr := fs.IsEmpty(path); subErr == nil && empty && deletePath {
		err = fs.RemoveWithContext(ctx, path)
	}
	return
}

type extendedFile struct {
	afero.File
	onCloseCallBack func() error
}

func (f *extendedFile) Close() (err error) {
	err = f.File.Close()
	if err != nil {
		return
	}
	if f.onCloseCallBack != nil {
		err = f.onCloseCallBack()
	}
	return
}

func (f *extendedFile) Fd() (fd uintptr) {
	if correctFile, ok := retrieveSubFile(f.File).(interface {
		Fd() uintptr
	}); ok {
		fd = correctFile.Fd()
	} else {
		fd = uintptr(UnsetFileHandle)
	}
	return
}
func IsFileHandleUnset(fh uintptr) bool {
	return uint64(fh) == UnsetFileHandle
}
func retrieveSubFile(top interface{}) interface{} {
	// This is necessary because depending on the file abstraction it may be difficult to find out the real file handle.
	var actualfile = top
	for {
		subFile := reflection.GetUnexportedStructureField(actualfile, "File")
		if subFile == nil {
			break
		}
		actualfile = subFile
	}
	return actualfile
}
func convertFile(getFile func() (afero.File, error), onCloseCallBack func() error) (f File, err error) {
	file, err := getFile()
	if err != nil {
		return
	}
	return convertToExtendedFile(file, onCloseCallBack)
}

func convertToExtendedFile(file afero.File, onCloseCallBack func() error) (File, error) {
	return &extendedFile{
		File:            file,
		onCloseCallBack: onCloseCallBack,
	}, nil
}

// FilepathStem returns  the final path component, without its suffix.
func FilepathStem(fp string) string {
	return strings.TrimSuffix(filepath.Base(fp), filepath.Ext(fp))
}
