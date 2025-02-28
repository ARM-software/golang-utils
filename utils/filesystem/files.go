/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filesystem

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v3"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/spf13/afero"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/platform"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/resource"
	"github.com/ARM-software/golang-utils/utils/safeio"
)

var (
	ErrLinkNotImplemented  = commonerrors.New(commonerrors.ErrNotImplemented, "link not implemented")
	ErrChownNotImplemented = commonerrors.New(commonerrors.ErrNotImplemented, "chown not implemented")
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
	resourceInUse resource.ICloseableResource
	vfs           afero.Fs
	fsType        FilesystemType
	pathConverter func(path string) string
}

// NewVirtualFileSystem returns a virtual filesystem similarly to NewCloseableVirtualFileSystem
func NewVirtualFileSystem(vfs afero.Fs, fsType FilesystemType, pathConverter func(path string) string) FS {
	return NewCloseableVirtualFileSystem(vfs, fsType, nil, "", pathConverter)
}

// NewCloseableVirtualFileSystem returns a virtual filesystem which requires closing after use.
// It is a wrapper over afero.FS virtual filesystem and tends to define common filesystem utilities as the ones available in posix.
func NewCloseableVirtualFileSystem(vfs afero.Fs, fsType FilesystemType, resourceInUse io.Closer, resourceInUseDescription string, pathConverter func(path string) string) ICloseableFS {
	var resourceInUseByFs resource.ICloseableResource
	if resourceInUse == nil {
		resourceInUseByFs = resource.NewNonCloseableResource()
	} else {
		resourceInUseByFs = resource.NewCloseableResource(resourceInUse, resourceInUseDescription)
	}
	return &VFS{
		resourceInUse: resourceInUseByFs,
		vfs:           vfs,
		fsType:        fsType,
		pathConverter: pathConverter,
	}
}

func GetGlobalFileSystem() FS {
	return globalFileSystem
}

func GetType() FilesystemType {
	return globalFileSystem.GetType()
}

// checkWhetherUnderlyingResourceIsClosed checks whether the filesystem is in a working state when relying on a ICloseableResource.
func (fs *VFS) checkWhetherUnderlyingResourceIsClosed() error {
	if fs.resourceInUse.IsClosed() {
		return commonerrors.Newf(commonerrors.ErrCondition, "the resource this filesystem is based on [%v] has been closed", fs.resourceInUse.String())
	}
	return nil
}

// Walk walks  https://golang.org/pkg/path/filepath/#WalkDir
func (fs *VFS) Walk(root string, fn filepath.WalkFunc) error {
	return fs.WalkWithContext(context.Background(), root, fn)
}

func (fs *VFS) WalkWithContext(ctx context.Context, root string, fn filepath.WalkFunc) error {
	return fs.WalkWithContextAndExclusionPatterns(ctx, root, fn)
}

func (fs *VFS) WalkWithContextAndExclusionPatterns(ctx context.Context, root string, fn filepath.WalkFunc, exclusionPatterns ...string) (err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	root = filepath.Join(root, string(fs.PathSeparator()))
	exclusionRegex, err := NewExclusionRegexList(fs.PathSeparator(), exclusionPatterns...)
	if err != nil {
		return
	}
	if IsPathExcluded(root, exclusionRegex...) {
		// In this case, the whole tree is excluded and so, we can stop here
		return
	}
	info, err := fs.Lstat(root)
	if err != nil {
		err = fn(root, nil, err)
	} else {
		err = fs.walk(ctx, root, info, exclusionRegex, fn)
	}
	err = ConvertFileSystemError(err)
	if commonerrors.Any(err, filepath.SkipDir) {
		return nil
	}

	return err
}

// walks recursively descends path, calling fn.
func (fs *VFS) walk(ctx context.Context, path string, info os.FileInfo, exclusions []*regexp.Regexp, fn filepath.WalkFunc) (err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
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
	cleansedList, err := ExcludeFiles(items, exclusions)
	if err != nil {
		err = fn(path, info, err)
		if err != nil {
			return
		}
	}
	for i := range cleansedList {
		itemName := cleansedList[i]
		subErr := parallelisation.DetermineContextError(ctx)
		if subErr != nil {
			return subErr
		}
		filename := filepath.Join(path, itemName)
		fileInfo, subErr := fs.Lstat(filename)
		if subErr != nil {
			if err := fn(filename, fileInfo, subErr); err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			subErr = fs.walk(ctx, filename, fileInfo, exclusions, fn)
			if subErr != nil {
				if !fileInfo.IsDir() || subErr != filepath.SkipDir {
					return subErr
				}
			}
		}
	}
	return nil
}

func (fs *VFS) GetType() FilesystemType {
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
	current, err := os.Getwd()
	err = ConvertFileSystemError(err)
	return fs.ConvertFilePath(current), err
}

func Lstat(name string) (fileInfo os.FileInfo, err error) {
	return globalFileSystem.Lstat(name)
}

func (fs *VFS) Lstat(name string) (fileInfo os.FileInfo, err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	if correctobj, ok := fs.vfs.(IStater); ok {
		fileInfo, _, err = correctobj.LstatIfPossible(name)
		err = ConvertFileSystemError(err)
		return
	}
	fileInfo, err = fs.Stat(name)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrNotImplemented, err, "")
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
	err := fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return nil, err
	}
	return convertFile(func() (afero.File, error) { return fs.vfs.Open(name) }, func() error { return nil })
}

func OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return globalFileSystem.OpenFile(name, flag, perm)
}

func (fs *VFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	err := fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return nil, err
	}
	return convertFile(func() (afero.File, error) { return fs.vfs.OpenFile(name, flag, perm) }, func() error { return nil })
}

func CreateFile(name string) (File, error) {
	return globalFileSystem.CreateFile(name)
}
func (fs *VFS) CreateFile(name string) (File, error) {
	err := fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return nil, err
	}
	return convertFile(func() (afero.File, error) { return fs.vfs.Create(name) }, func() error { return nil })
}

func (fs *VFS) NewRemoteLockFile(id string, dirToLock string) ILock {
	return NewRemoteLockFile(fs, id, dirToLock)
}

func ReadFile(name string) ([]byte, error) {
	return globalFileSystem.ReadFile(name)
}

func (fs *VFS) ReadFile(filename string) (content []byte, err error) {
	return fs.readFileWithContextAndLimits(context.Background(), filename, NoLimits())
}

func ReadFileWithLimits(filename string, limits ILimits) ([]byte, error) {
	return globalFileSystem.ReadFileWithLimits(filename, limits)
}

func (fs *VFS) ReadFileWithLimits(filename string, limits ILimits) ([]byte, error) {
	return fs.readFileWithContextAndLimits(context.Background(), filename, limits)
}

func (fs *VFS) ReadFileWithContext(ctx context.Context, filename string) ([]byte, error) {
	return fs.readFileWithContextAndLimits(ctx, filename, NoLimits())
}

func (fs *VFS) ReadFileWithContextAndLimits(ctx context.Context, filename string, limits ILimits) ([]byte, error) {
	return fs.readFileWithContextAndLimits(ctx, filename, limits)
}

func (fs *VFS) readFileWithContextAndLimits(ctx context.Context, filename string, limits ILimits) (content []byte, err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if limits == nil {
		err = commonerrors.New(commonerrors.ErrUndefined, "missing file system limits definition")
		return
	}
	// Really similar to afero ioutils Read file but using our utilities instead.
	f, err := fs.GenericOpen(filename)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	content, err = fs.ReadFileContent(ctx, f, limits)
	return
}

// ReadFileContent reads the file content.
func ReadFileContent(ctx context.Context, file File) ([]byte, error) {
	return globalFileSystem.ReadFileContent(ctx, file, NoLimits())
}

func (fs *VFS) ReadFileContent(ctx context.Context, file File, limits ILimits) (content []byte, err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if file == nil {
		err = fmt.Errorf("%w: missing file definition", commonerrors.ErrUndefined)
		return
	}
	if limits == nil {
		err = fmt.Errorf("%w: missing file system limits definition", commonerrors.ErrUndefined)
		return
	}
	var bufferCapacity int64 = bytes.MinRead
	var max int64 = -1
	if limits.Apply() {
		max = limits.GetMaxFileSize()
	}
	fi, err := file.Stat()
	if err == nil {
		// Don't preallocate a huge buffer, just in case.
		fileSize := fi.Size()
		if fileSize < 1e9 {
			bufferCapacity += fileSize
		}
		if limits.Apply() && fileSize > max {
			err = fmt.Errorf("%w: file [%v] is bigger than allowed size [%vB]", commonerrors.ErrTooLarge, file.Name(), max)
			return
		}
	}

	content, err = safeio.ReadAtMost(ctx, file, max, bufferCapacity)
	return
}

// WriteFile writes data to a file
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	return globalFileSystem.WriteFile(filename, data, perm)
}

func (fs *VFS) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return fs.WriteFileWithContext(context.Background(), filename, data, perm)
}

func (fs *VFS) WriteFileWithContext(ctx context.Context, filename string, data []byte, perm os.FileMode) (err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	reader := bytes.NewReader(data)
	n, err := fs.WriteToFile(ctx, filename, reader, perm)
	if err != nil {
		return
	}
	if int(n) < len(data) {
		err = io.ErrShortWrite
	}
	return
}

func (fs *VFS) WriteToFile(ctx context.Context, filename string, reader io.Reader, perm os.FileMode) (written int64, err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	f, err := fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	written, err = safeio.CopyDataWithContext(ctx, reader, f)
	if err != nil {
		return
	}
	if written == 0 {
		err = fmt.Errorf("%w: no bytes were written", commonerrors.ErrEmpty)
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
	err := fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return nil, err
	}
	info, err := fs.vfs.Stat(name)
	err = ConvertFileSystemError(err)
	return info, err
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
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	name, err = afero.TempDir(fs.vfs, dir, prefix)
	err = ConvertFileSystemError(err)
	return
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
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	file, err := afero.TempFile(fs.vfs, dir, prefix)
	err = ConvertFileSystemError(err)
	if err != nil {
		return
	}
	f, err = convertToExtendedFile(file, func() error { return nil })
	if err != nil {
		return
	}
	// Changing permissions so that it is aligned with os.CreateFile permissions
	f, err = fs.changeFilePermissions(f, 0o666)
	return
}

func (fs *VFS) changeFilePermissions(f File, mode os.FileMode) (openF File, err error) {
	if f == nil {
		err = commonerrors.New(commonerrors.ErrUndefined, "missing file")
		return
	}
	err = ConvertFileSystemError(f.Close())
	if err != nil {
		return
	}
	err = fs.Chmod(f.Name(), mode)
	if err != nil {
		return
	}
	openF, err = fs.GenericOpen(f.Name())
	return
}

// TouchTempFile creates an empty file in `dir` and returns it filename
func TouchTempFile(dir string, prefix string) (filename string, err error) {
	return globalFileSystem.TouchTempFile(dir, prefix)
}

func (fs *VFS) TouchTempFile(dir string, prefix string) (filename string, err error) {
	file, err := fs.TempFile(dir, prefix)
	if file != nil {
		_ = file.Close()
		filename = file.Name()
	}
	return
}

func TempFileInTempDir(pattern string) (f File, err error) {
	return globalFileSystem.TempFileInTempDir(pattern)
}

func (fs *VFS) TempFileInTempDir(prefix string) (f File, err error) {
	return fs.TempFile("", prefix)
}

// TouchTempFileInTempDir creates an empty file in temp directory and returns it filename
func TouchTempFileInTempDir(prefix string) (filename string, err error) {
	return globalFileSystem.TouchTempFileInTempDir(prefix)
}

func (fs *VFS) TouchTempFileInTempDir(prefix string) (filename string, err error) {
	return fs.TouchTempFile("", prefix)
}

func CleanDir(dir string) (err error) {
	return globalFileSystem.CleanDir(dir)
}

func (fs *VFS) CleanDir(dir string) error {
	return fs.CleanDirWithContext(context.Background(), dir)
}

func (fs *VFS) CleanDirWithContext(ctx context.Context, dir string) error {
	return fs.CleanDirWithContextAndExclusionPatterns(ctx, dir)
}
func (fs *VFS) CleanDirWithContextAndExclusionPatterns(ctx context.Context, dir string, exclusionPatterns ...string) (err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
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
	files, err := fs.LsWithExclusionPatterns(dir, exclusionPatterns...)
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
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = fs.RemoveWithContext(ctx, filepath.Join(dir, f))
	return
}

// Exists checks if a file or folder exists
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
	err = ConvertFileSystemError(err)
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

func Rm(dir string) error {
	return globalFileSystem.Rm(dir)
}

func (fs *VFS) Rm(dir string) error {
	return fs.RemoveWithContext(context.Background(), dir)
}

// RemoveWithPrivileges removes any directory (equivalent to `sudo rm -rf`)
func RemoveWithPrivileges(ctx context.Context, dir string) error {
	return globalFileSystem.RemoveWithPrivileges(ctx, dir)
}

func (fs *VFS) RemoveWithPrivileges(ctx context.Context, dir string) (err error) {
	err = fs.RemoveWithContext(ctx, dir)
	if commonerrors.Any(err, nil, commonerrors.ErrTimeout, commonerrors.ErrCancelled) {
		return
	}
	currentUser, subErr := user.Current()
	if subErr != nil {
		err = fmt.Errorf("%w: cannot retrieve information about current user: %v", commonerrors.ErrUnexpected, subErr.Error())
		return
	}
	subErr = fs.ChangeOwnership(dir, currentUser)
	if subErr == nil {
		err = fs.RemoveWithContext(ctx, dir)
		if commonerrors.Any(err, nil, commonerrors.ErrTimeout, commonerrors.ErrCancelled) {
			return
		}
	}
	if correctobj, ok := fs.vfs.(IForceRemover); ok {
		err = ConvertFileSystemError(correctobj.ForceRemoveIfPossible(dir))
	}
	return
}

func (fs *VFS) RemoveWithContext(ctx context.Context, dir string) error {
	return fs.RemoveWithContextAndExclusionPatterns(ctx, dir)
}
func (fs *VFS) RemoveWithContextAndExclusionPatterns(ctx context.Context, dir string, exclusionPatterns ...string) (err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
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
		err = fs.CleanDirWithContextAndExclusionPatterns(ctx, dir, exclusionPatterns...)
	}
	if err != nil {
		return
	}
	isEmpty, err = fs.IsEmpty(dir) // Checking again if empty as some files may have been ignored.
	if err != nil {
		return
	}
	if isDir && !isEmpty {
		return
	} // some files may have been ignored and so, the removal should stop
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if IsPathExcludedFromPatterns(dir, fs.PathSeparator(), exclusionPatterns...) {
		return
	}
	err = ConvertFileSystemError(fs.vfs.Remove(dir))
	return
}

// IsFile states whether it is a file or not
func IsFile(path string) (result bool, err error) {
	return globalFileSystem.IsFile(path)
}
func (fs *VFS) IsFile(path string) (result bool, err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
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
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
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
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	if !fs.Exists(path) {
		err = fmt.Errorf("%w: path [%v]", commonerrors.ErrNotFound, path)
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
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
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
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	f, err := fs.vfs.Open(name)
	err = ConvertFileSystemError(err)
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
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	if dir == "" {
		return fmt.Errorf("missing path: %w", commonerrors.ErrUndefined)
	}
	if fs.Exists(dir) {
		return
	}
	err = ConvertFileSystemError(fs.vfs.MkdirAll(dir, perm))
	// Directory was maybe created by a different process/thread
	if err != nil && fs.Exists(dir) {
		err = nil
	}
	return
}

// FindAll finds all the files with extensions
func FindAll(dir string, extensions ...string) (files []string, err error) {
	return globalFileSystem.FindAll(dir, extensions...)
}
func (fs *VFS) FindAll(dir string, extensions ...string) (files []string, err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
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
	files, err = fs.Glob(filepath.Join(dir, "**", fmt.Sprintf("*.%v", strings.TrimPrefix(ext, "."))))
	return
}

// Glob returns the names of all files matching pattern with support for "doublestar" (aka globstar: **) patterns
// You can find the children with patterns such as: **/child*, grandparent/**/child?, **/parent/*, or even just ** by itself (which will return all files and directories recursively).
func Glob(pattern string) ([]string, error) {
	return GetGlobalFileSystem().Glob(pattern)
}

func (fs *VFS) Glob(pattern string) (matches []string, err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	matches, err = doublestar.GlobOS(fs, pattern)
	if commonerrors.Any(err, doublestar.ErrBadPattern) {
		err = fmt.Errorf("%w: %v", commonerrors.ErrInvalid, err.Error())
	}
	return
}

// Chmod changes the file mode of a filesystem item.
func Chmod(name string, mode os.FileMode) error {
	return GetGlobalFileSystem().Chmod(name, mode)
}

func (fs *VFS) Chmod(name string, mode os.FileMode) (err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	err = ConvertFileSystemError(fs.vfs.Chmod(name, mode))
	return
}

func (fs *VFS) ChmodRecursively(ctx context.Context, path string, mode os.FileMode) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if isFile, _ := fs.IsFile(path); isFile {
		err = fs.Chmod(path, mode)
		return
	}
	err = fs.WalkWithContext(ctx, path, func(subPath string, info os.FileInfo, subErr error) error {
		if subErr != nil {
			return subErr
		}
		return fs.Chmod(subPath, mode)
	})
	return
}

func (fs *VFS) Touch(path string) (err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	if fs.Exists(path) {
		now := time.Now().UTC()
		err = fs.Chtimes(path, now, now)
		return
	}
	if strings.TrimSpace(path) == "" {
		err = commonerrors.New(commonerrors.ErrUndefined, "empty path")
		return
	}
	if EndsWithPathSeparator(fs, path) {
		err = fs.MkDir(path)
		return
	}
	file, err := fs.CreateFile(path)
	if file != nil {
		_ = file.Close()
	}

	return
}

func (fs *VFS) Chtimes(name string, atime time.Time, mtime time.Time) (err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	err = ConvertFileSystemError(fs.vfs.Chtimes(name, atime, mtime))
	return
}

func (fs *VFS) ChangeOwnership(name string, owner *user.User) error {
	if owner == nil {
		return fmt.Errorf("%w: missing user definition", commonerrors.ErrUndefined)
	}
	uid, err := strconv.Atoi(owner.Uid)
	if err != nil {
		return fmt.Errorf("%w: cannot parse owner uid", commonerrors.ErrUnexpected)
	}
	gid, err := strconv.Atoi(owner.Gid)
	if err != nil {
		return fmt.Errorf("%w: cannot parse owner gid", commonerrors.ErrUnexpected)
	}
	return fs.Chown(name, uid, gid)
}

func (fs *VFS) ChangeOwnershipRecursively(ctx context.Context, path string, owner *user.User) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if isFile, _ := fs.IsFile(path); isFile {
		err = fs.ChangeOwnership(path, owner)
		return
	}
	if owner == nil {
		return fmt.Errorf("%w: missing user definition", commonerrors.ErrUndefined)
	}
	err = fs.WalkWithContext(ctx, path, func(subPath string, info os.FileInfo, subErr error) error {
		if subErr != nil {
			return subErr
		}
		return fs.ChangeOwnership(subPath, owner)
	})
	return
}

func (fs *VFS) ChownRecursively(ctx context.Context, path string, uid, gid int) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if isFile, _ := fs.IsFile(path); isFile {
		err = fs.Chown(path, uid, gid)
		return
	}
	err = fs.WalkWithContext(ctx, path, func(subPath string, info os.FileInfo, subErr error) error {
		if subErr != nil {
			return subErr
		}
		return fs.Chown(subPath, uid, gid)
	})
	return
}

func (fs *VFS) Chown(name string, uid, gid int) (err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	if correctobj, ok := fs.vfs.(IChowner); ok {
		err = ConvertFileSystemError(correctobj.ChownIfPossible(name, uid, gid))
		return
	}
	err = ConvertFileSystemError(fs.vfs.Chown(name, uid, gid))
	return
}

func (fs *VFS) FetchOwners(name string) (uid, gid int, err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	stat, err := fs.Stat(name)
	if err != nil {
		return
	}
	if stat == nil {
		err = fmt.Errorf("%w: missing file info", commonerrors.ErrUndefined)
		return
	}
	if reflection.IsEmpty(stat.Sys()) {
		err = commonerrors.ErrNotImplemented
		return
	}
	uid, gid, err = determineFileOwners(stat)
	return
}

func (fs *VFS) FetchFileOwner(name string) (owner *user.User, err error) {
	uid, _, err := fs.FetchOwners(name)
	if err != nil {
		return
	}
	owner, err = user.LookupId(strconv.Itoa(uid))
	if err != nil {
		err = fmt.Errorf("%w: user [`%v`] could not be found: %v", commonerrors.ErrNotFound, uid, err.Error())
	}
	return
}

func (fs *VFS) Link(oldname, newname string) (err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	if correctobj, ok := fs.vfs.(ILinker); ok {
		err = ConvertFileSystemError(correctobj.LinkIfPossible(oldname, newname))
		return
	}
	err = commonerrors.Newf(commonerrors.ErrNotImplemented, "cannot link `%v` to `%v`", oldname, newname)
	return
}

func (fs *VFS) Readlink(name string) (value string, err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	if correctobj, ok := fs.vfs.(ILinkReader); ok {
		value, err = correctobj.ReadlinkIfPossible(name)
		err = ConvertFileSystemError(err)
		return
	}
	err = commonerrors.Newf(commonerrors.ErrNotImplemented, "cannot read link `%v`", name)
	return
}

func (fs *VFS) Symlink(oldname string, newname string) (err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	if correctobj, ok := fs.vfs.(ISymLinker); ok {
		err = ConvertFileSystemError(correctobj.SymlinkIfPossible(oldname, newname))
		return
	}
	err = commonerrors.Newf(commonerrors.ErrNotImplemented, "cannot symlink `%v` to `%v`", oldname, newname)
	return
}

func Ls(dir string) (files []string, err error) {
	return globalFileSystem.Ls(dir)
}
func (fs *VFS) Ls(dir string) (names []string, err error) {
	return fs.LsWithExclusionPatterns(dir)
}

func (fs *VFS) LsWithExclusionPatterns(dir string, exclusionPatterns ...string) (names []string, err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	regexes, err := NewExclusionRegexList(fs.PathSeparator(), exclusionPatterns...)
	if err != nil {
		return nil, err
	}
	names, err = LsWithExclusionPatterns(fs, dir, regexes)
	return
}

// LsRecursive lists all files recursively, including subdirectories
func LsRecursive(ctx context.Context, dir string, includeDirectories bool) (files []string, err error) {
	return GetGlobalFileSystem().LsRecursive(ctx, dir, includeDirectories)
}

// LsRecursiveWithExclusionPatterns lists all files and directory (equivalent to ls) but exclude the ones matching the exclusion patterns.
func LsRecursiveWithExclusionPatterns(ctx context.Context, dir string, includeDirectories bool, exclusionPatterns ...string) (files []string, err error) {
	return GetGlobalFileSystem().LsRecursiveWithExclusionPatterns(ctx, dir, includeDirectories, exclusionPatterns...)
}

func (fs *VFS) LsRecursive(ctx context.Context, dir string, includeDirectories bool) (files []string, err error) {
	return fs.LsRecursiveWithExclusionPatternsAndLimits(ctx, dir, NoLimits(), includeDirectories)
}

func (fs *VFS) LsRecursiveWithExclusionPatterns(ctx context.Context, dir string, includeDirectories bool, exclusionPatterns ...string) (files []string, err error) {
	return fs.LsRecursiveWithExclusionPatternsAndLimits(ctx, dir, NoLimits(), includeDirectories, exclusionPatterns...)
}

func (fs *VFS) LsRecursiveWithExclusionPatternsAndLimits(ctx context.Context, dir string, limits ILimits, includeDirectories bool, exclusionPatterns ...string) (files []string, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if limits == nil {
		err = fmt.Errorf("%w: missing file system limits", commonerrors.ErrUndefined)
		return
	}

	fn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		currentDepth, err := FileTreeDepth(fs, dir, path)
		if err != nil {
			return err
		}

		if limits.Apply() && currentDepth >= limits.GetMaxDepth() {
			return filepath.SkipDir
		}

		if limits.Apply() && int64(len(files)) >= limits.GetMaxFileCount() {
			return fmt.Errorf("number of files exceeds the limit of %d: %w", limits.GetMaxFileCount(), commonerrors.ErrTooLarge)
		}

		if includeDirectories || !info.IsDir() {
			files = append(files, path)
		}

		return nil
	}

	err = fs.WalkWithContextAndExclusionPatterns(ctx, dir, fn, exclusionPatterns...)
	return
}

func LsWithExclusionPatterns(fs FS, dir string, regexes []*regexp.Regexp) (names []string, err error) {
	if isDir, subErr := fs.IsDir(dir); !isDir || subErr != nil {
		err = fmt.Errorf("path [%v] is not a directory: %w", dir, commonerrors.ErrInvalid)
		return
	}
	f, err := fs.GenericOpen(dir)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	AllNames, err := fs.LsFromOpenedDirectory(f)
	_ = f.Close()
	if err != nil {
		return
	}
	names, err = ExcludeFiles(AllNames, regexes)
	return
}

func (fs *VFS) LsFromOpenedDirectory(dir File) ([]string, error) {
	err := fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return nil, err
	}
	if dir == nil {
		return nil, fmt.Errorf("%w: nil directory", commonerrors.ErrUndefined)
	}
	return dir.Readdirnames(-1)
}

func (fs *VFS) LsRecursiveFromOpenedDirectory(ctx context.Context, dir File, includeDirectories bool) (files []string, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if dir == nil {
		return nil, fmt.Errorf("%w: supplied directory was undefined", commonerrors.ErrUndefined)
	}
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return nil, fmt.Errorf("%w: underlying resource is closed", commonerrors.ErrUndefined)
	}

	return fs.LsRecursive(ctx, dir.Name(), includeDirectories)
}

func (fs *VFS) Lls(dir string) (files []os.FileInfo, err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
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
	err := fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return nil, err
	}
	if dir == nil {
		return nil, fmt.Errorf("%w: nil directory", commonerrors.ErrUndefined)
	}
	return dir.Readdir(-1)
}

func (fs *VFS) ConvertToAbsolutePath(rootPath string, paths ...string) ([]string, error) {
	err := fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return nil, err
	}
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
	err := fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return nil, err
	}
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
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
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
	err = ConvertFileSystemError(fs.vfs.Rename(src, dest))
	if err == nil {
		return
	}
	// os.Rename() give error "invalid cross-device link" for Docker container with Volumes.
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
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
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
	return os.IsNotExist(err) || commonerrors.Any(err, ErrPathNotExist) || commonerrors.CorrespondTo(err, ErrPathNotExist.Error())
}

func (fs *VFS) moveFile(ctx context.Context, src string, dest string) (err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if src == dest {
		return
	}
	err = CopyBetweenFSWithExclusionRegexes(ctx, fs, src, fs, dest, []*regexp.Regexp{}, []*regexp.Regexp{})
	if err != nil {
		return
	}
	err = ConvertFileSystemError(fs.vfs.Remove(src))
	if err != nil {
		return
	}
	return
}

func (fs *VFS) FileHash(hashAlgo string, path string) (string, error) {
	return fs.FileHashWithContext(context.Background(), hashAlgo, path)
}

func (fs *VFS) FileHashWithContext(ctx context.Context, hashAlgo string, path string) (hash string, err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	hasher, err := NewFileHash(hashAlgo)
	if err != nil {
		return
	}
	hash, err = hasher.CalculateFileWithContext(ctx, fs, path)
	return
}

func Copy(src string, dest string) (err error) {
	return globalFileSystem.Copy(src, dest)
}
func (fs *VFS) Copy(src string, dest string) (err error) {
	return fs.CopyWithContext(context.Background(), src, dest)
}

// CopyToFile copies a file into another file
func CopyToFile(srcFile, destFile string) error {
	return globalFileSystem.CopyToFile(srcFile, destFile)
}

func (fs *VFS) CopyToFile(src string, dest string) error {
	return fs.CopyToFileWithContext(context.Background(), src, dest)
}

// CopyToFileWithContext copies a file into another file
func CopyToFileWithContext(ctx context.Context, srcFile, destFile string) error {
	return globalFileSystem.CopyToFileWithContext(ctx, srcFile, destFile)
}

func (fs *VFS) CopyToFileWithContext(ctx context.Context, src string, dest string) (err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	isFile, err := fs.IsFile(src)
	if err != nil {
		return
	}
	if !isFile {
		err = fmt.Errorf("%w: the copy source [%v] must be a file", commonerrors.ErrInvalid, src)
		return
	}
	if fs.Exists(dest) {
		isFile, err = fs.IsFile(dest)
		if err != nil {
			return
		}
		if !isFile {
			err = fmt.Errorf("%w: the copy destination [%v] must be a file", commonerrors.ErrInvalid, dest)
			return
		}
	} else if reflection.IsEmpty(dest) || EndsWithPathSeparator(fs, dest) {
		err = fmt.Errorf("%w: the copy destination [%v] must be a valid filename", commonerrors.ErrInvalid, dest)
		return
	}

	return fs.CopyWithContext(ctx, src, dest)
}

// CopyToDirectory copies a src to a directory  destDirectory which will be created as such if not present.
func CopyToDirectory(src, destDirectory string) error {
	return GetGlobalFileSystem().CopyToDirectory(src, destDirectory)
}

func (fs *VFS) CopyToDirectory(src, destDirectory string) error {
	return fs.CopyToDirectoryWithContext(context.Background(), src, destDirectory)
}

// CopyToDirectoryWithContext copies a src to a directory  destDirectory which will be created as such if not present.
func CopyToDirectoryWithContext(ctx context.Context, src, destDirectory string) error {
	return GetGlobalFileSystem().CopyToDirectoryWithContext(ctx, src, destDirectory)
}

func (fs *VFS) CopyToDirectoryWithContext(ctx context.Context, src, destDirectory string) (err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	err = fs.MkDir(destDirectory)
	if err != nil {
		return
	}
	err = fs.CopyWithContext(ctx, src, destDirectory)
	return
}

func (fs *VFS) CopyWithContext(ctx context.Context, src string, dest string) error {
	return fs.CopyWithContextAndExclusionPatterns(ctx, src, dest)
}

func (fs *VFS) CopyWithContextAndExclusionPatterns(ctx context.Context, src string, dest string, exclusionPatterns ...string) error {
	return CopyBetweenFSWithExclusionPatterns(ctx, fs, src, fs, dest, exclusionPatterns...)
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
	return CopyBetweenFSWithExclusionPatterns(ctx, srcFs, src, destFs, dest)
}
func CopyBetweenFSWithExclusionPatterns(ctx context.Context, srcFs FS, src string, destFs FS, dest string, exclusionPatterns ...string) (err error) {
	exclusionSrcFsRegexes, err := NewExclusionRegexList(srcFs.PathSeparator(), exclusionPatterns...)
	if err != nil {
		return
	}
	exclusionDestFsRegexes, err := NewExclusionRegexList(destFs.PathSeparator(), exclusionPatterns...)
	if err != nil {
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = CopyBetweenFSWithExclusionRegexes(ctx, srcFs, src, destFs, dest, exclusionSrcFsRegexes, exclusionDestFsRegexes)
	return
}

func CopyBetweenFSWithExclusionRegexes(ctx context.Context, srcFs FS, src string, destFs FS, dest string, exclusionSrcFsRegexes []*regexp.Regexp, exclusionDestFsRegexes []*regexp.Regexp) (err error) {
	if IsPathExcluded(src, exclusionSrcFsRegexes...) || IsPathExcluded(dest, exclusionDestFsRegexes...) {
		return
	}
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
	isSrcDir, err := srcFs.IsDir(src)
	if err != nil {
		return
	}
	destExists := destFs.Exists(dest)
	isDestDir := false
	if destExists {
		isDestDir, err = destFs.IsDir(dest)
		if err != nil {
			return
		}
	} else {
		if isSrcDir {
			isDestDir = true
			err = destFs.MkDir(dest)
		} else {
			if EndsWithPathSeparator(destFs, dest) { // check if dest defined as a folder i.e. ending with `/` or `\\`
				isDestDir = true
				err = destFs.MkDir(dest)
			} else {
				isDestDir = false
				err = destFs.MkDir(filepath.Dir(dest))
			}
		}
		if err != nil {
			return
		}
	}

	var dst string
	if !(isSrcDir && !destExists) && isDestDir {
		dst = filepath.Join(dest, filepath.Base(src))
	} else {
		dst = dest
	}
	if isSrcDir {
		err = copyFolderBetweenFSWithExclusionRegexes(ctx, srcFs, src, destFs, dst, exclusionSrcFsRegexes, exclusionDestFsRegexes)
	} else {
		err = copyFileBetweenFSWithExclusionPatternsWithExclusionRegexes(ctx, srcFs, src, destFs, dst, exclusionSrcFsRegexes, exclusionDestFsRegexes)
	}
	return
}

func copyFolderBetweenFSWithExclusionRegexes(ctx context.Context, srcFs FS, src string, destFs FS, dest string, exclusionSrcFsRegexes []*regexp.Regexp, exclusionDestFsRegexes []*regexp.Regexp) (err error) {
	if IsPathExcluded(src, exclusionSrcFsRegexes...) || IsPathExcluded(dest, exclusionDestFsRegexes...) {
		return
	}
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
		files, err := LsWithExclusionPatterns(srcFs, src, exclusionSrcFsRegexes)
		if err != nil {
			return err
		}
		for i := range files {
			srcPath := filepath.Join(src, files[i])
			err = CopyBetweenFSWithExclusionRegexes(ctx, srcFs, srcPath, destFs, dest, exclusionSrcFsRegexes, exclusionDestFsRegexes)
			if err != nil {
				return err
			}
		}
	}
	return
}

func copyFileBetweenFSWithExclusionPatternsWithExclusionRegexes(ctx context.Context, srcFs FS, src string, destFs FS, dest string, exclusionSrcFsRegexes []*regexp.Regexp, exclusionDestFsRegexes []*regexp.Regexp) (err error) {
	if IsPathExcluded(src, exclusionSrcFsRegexes...) || IsPathExcluded(dest, exclusionDestFsRegexes...) {
		return
	}
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
	_, err = safeio.CopyDataWithContext(ctx, inputFile, outputFile)
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
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
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

func SubDirectories(directory string) ([]string, error) {
	return globalFileSystem.SubDirectories(directory)
}

func (fs *VFS) SubDirectories(directory string) ([]string, error) {
	return fs.SubDirectoriesWithContext(context.Background(), directory)
}

func (fs *VFS) SubDirectoriesWithContext(ctx context.Context, directory string) ([]string, error) {
	return fs.SubDirectoriesWithContextAndExclusionPatterns(ctx, directory, "^[.].*$")
}
func (fs *VFS) SubDirectoriesWithContextAndExclusionPatterns(ctx context.Context, directory string, exclusionPatterns ...string) (directories []string, err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	files, err := afero.ReadDir(fs.vfs, directory)
	err = ConvertFileSystemError(err)
	if err != nil {
		return
	}
	regexes, err := NewExclusionRegexList(fs.PathSeparator(), exclusionPatterns...)
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
		if file.IsDir() && !IsPathExcluded(file.Name(), regexes...) {
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

func (fs *VFS) ListDirTreeWithContext(ctx context.Context, dirPath string, list *[]string) error {
	return fs.ListDirTreeWithContextAndExclusionPatterns(ctx, dirPath, list)
}
func (fs *VFS) ListDirTreeWithContextAndExclusionPatterns(ctx context.Context, dirPath string, list *[]string, exclusionPatterns ...string) (err error) {
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	regexes, err := NewExclusionRegexList(fs.PathSeparator(), exclusionPatterns...)
	if err != nil {
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = ListDirTreeWithContextAndExclusionPatterns(ctx, fs, dirPath, list, regexes)
	return
}

func ListDirTreeWithContextAndExclusionPatterns(ctx context.Context, fs FS, dirPath string, list *[]string, regexes []*regexp.Regexp) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if list == nil {
		err = fmt.Errorf("uninitialised input variable: %w", commonerrors.ErrInvalid)
		return
	}

	elements, err := LsWithExclusionPatterns(fs, dirPath, regexes)
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
			subErr = ListDirTreeWithContextAndExclusionPatterns(ctx, fs, path, list, regexes)
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
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
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
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
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
	err = fs.checkWhetherUnderlyingResourceIsClosed()
	if err != nil {
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if deletePath && platform.IsWindows() {
		// On linux and potentially macOS, the access/modification of files in a directory do not affect those times for the parent folder.
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

func (fs *VFS) Close() error {
	return ConvertFileSystemError(fs.resourceInUse.Close())
}

type extendedFile struct {
	afero.File
	onCloseCallBack func() error
}

func (f *extendedFile) Close() (err error) {
	err = f.File.Close()
	err = ConvertFileSystemError(err)
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
	err = ConvertFileSystemError(err)
	if err != nil {
		return
	}
	return convertToExtendedFile(file, onCloseCallBack)
}

func convertToExtendedFile(file afero.File, onCloseCallBack func() error) (File, error) {
	return &extendedFile{
		File: file,
		onCloseCallBack: func() error {
			return ConvertFileSystemError(onCloseCallBack())
		},
	}, nil
}

// ConvertToOSFile converts a file to a `os` implementation of a file for certain use-cases where functions have not moved to using `fs.File`.
func ConvertToOSFile(f File) (osFile *os.File) {
	if f == nil {
		return
	}
	osFile = os.NewFile(f.Fd(), f.Name())
	return
}
