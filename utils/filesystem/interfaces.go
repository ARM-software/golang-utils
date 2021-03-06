/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/bmatcuk/doublestar/v3"
	"github.com/spf13/afero"

	"github.com/ARM-software/golang-utils/utils/config"
)

//go:generate mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE IFileHash,Chowner,Linker,File,DiskUsage,FileTimeInfo,ILock,ILimits,FS

// For reference.
//https://stackoverflow.com/questions/1761607/what-is-the-fastest-hash-algorithm-to-check-if-two-files-are-equal
type IFileHash interface {
	Calculate(f File) (string, error)
	CalculateFile(fs FS, path string) (string, error)
	GetType() string
}

// Optional interface. It is only implemented by the
// filesystems saying so.
type Chowner interface {
	ChownIfPossible(string, int, int) error
}

// Optional interface. It is only implemented by the
// filesystems saying so.
type Linker interface {
	LinkIfPossible(string, string) error
}

type File interface {
	afero.File
	Fd() uintptr
}

type DiskUsage interface {
	GetTotal() uint64
	GetFree() uint64
	GetUsed() uint64
	GetUsedPercent() float64
	GetInodesTotal() uint64
	GetInodesUsed() uint64
	GetInodesFree() uint64
	GetInodesUsedPercent() float64
}

type FileTimeInfo interface {
	ModTime() time.Time
	AccessTime() time.Time
	ChangeTime() time.Time
	BirthTime() time.Time
	HasChangeTime() bool
	HasBirthTime() bool
	HasAccessTime() bool
}

// ILimits defines general FileSystemLimits for actions performed on the filesystem
type ILimits interface {
	config.IServiceConfiguration
	// Apply states whether the limit should be applied
	Apply() bool
	// GetMaxFileSize returns the maximum size in byte a file can have on a file system
	GetMaxFileSize() int64
	// GetMaxTotalSize returns the maximum size in byte a location can have on a file system (whether it is a file or a folder)
	GetMaxTotalSize() uint64
}

// ILock defines a generic lock using the file system.
// FIXME it should be noted that despite being possible to use the lock with an in-memory filesystem, it should be avoided at all cost.
// The implementation of the in-memory FS used (afero) has shown several thread safety issues (e.g. https://github.com/spf13/afero/issues/298) and therefore, should not be used for scenarios involving concurrency until it is fixed.
type ILock interface {
	// Lock locks the lock. This call will wait (i.e. block) until the lock is available.
	Lock(ctx context.Context) error
	// LockWithTimeout tries to lock the lock until the timeout expires. If the timeout expires, this method will return commonerror.ErrTimeout.
	LockWithTimeout(ctx context.Context, timeout time.Duration) error
	// TryLock attempts to lock the lock instantly. This method will return commonerrors.ErrLocked immediately if the lock cannot be acquired straight away.
	TryLock(ctx context.Context) error
	// IsStale determines whether a lock is stale (the owner forgot to release it or is dead) or not.
	IsStale() bool
	// Unlock releases the lock. This takes precedence over any current lock.
	Unlock(ctx context.Context) error
	// ReleaseIfStale forces the lock to be released if it is considered as stale.
	ReleaseIfStale(ctx context.Context) error
	// MakeStale makes the lock stale. This is mostly for testing purposes.
	MakeStale(ctx context.Context) error
}

// FS defines all the methods a file system should provide.
type FS interface {
	// Open opens a file. The following is for being able to use doublestar
	Open(name string) (doublestar.File, error)
	// GenericOpen opens a file for reading. It opens the named file with specified flag (O_RDONLY etc.).
	// See os.Open()
	GenericOpen(name string) (File, error)
	// OpenFile opens a file using the given flags and the given mode.
	// OpenFile is the generalized open call
	// most users will use GenericOpen or Create instead.
	// See os.OpenFile
	OpenFile(name string, flag int, perm os.FileMode) (File, error)
	// CreateFile creates a file.
	CreateFile(name string) (File, error)
	// PathSeparator returns the path separator character.
	PathSeparator() rune
	Stat(name string) (os.FileInfo, error)
	Lstat(name string) (os.FileInfo, error)
	// StatTimes returns file time information.
	StatTimes(name string) (FileTimeInfo, error)
	// GetType returns the type of the file system.
	GetType() int
	// CleanDir removes all the files in a directory (equivalent rm -rf .../*)
	CleanDir(dir string) (err error)
	// CleanDirWithContext removes all the files in a directory (equivalent rm -rf .../*)
	CleanDirWithContext(ctx context.Context, dir string) (err error)
	// Exists checks whether a file or folder exists
	Exists(path string) bool
	// Rm removes directory (equivalent to rm -r)
	Rm(dir string) (err error)
	// RemoveWithContext removes directory (equivalent to rm -r)
	RemoveWithContext(ctx context.Context, dir string) (err error)
	// IsFile states whether it is a file or not
	IsFile(path string) (result bool, err error)
	// IsDir states whether it is a directory or not
	IsDir(path string) (result bool, err error)
	// IsLink states whether it is a link or not
	IsLink(path string) (result bool, err error)
	// IsEmpty checks whether a path is empty or not
	IsEmpty(name string) (empty bool, err error)
	// MkDir makes directory (equivalent to mkdir -p)
	MkDir(dir string) (err error)
	// MkDirAll makes directory (equivalent to mkdir -p)
	MkDirAll(dir string, perm os.FileMode) (err error)
	// ExcludeAll returns the list without the path matching the exclusion patterns.
	ExcludeAll(files []string, exclusionPatterns ...string) ([]string, error)
	// FindAll finds all the files with extensions
	FindAll(dir string, extensions ...string) (files []string, err error)
	// Walk walks  the file tree rooted at root, calling fn for each file or
	// directory in the tree, including root. See https://golang.org/pkg/path/filepath/#WalkDir
	Walk(root string, fn filepath.WalkFunc) error
	// WalkWithContext walks  the file tree rooted at root, calling fn for each file or
	// directory in the tree, including root. See https://golang.org/pkg/path/filepath/#WalkDir
	WalkWithContext(ctx context.Context, root string, fn filepath.WalkFunc) error
	// Ls lists all files and directory (equivalent to ls)
	Ls(dir string) (files []string, err error)
	// LsFromOpenedDirectory lists all files and directory (equivalent to ls)
	LsFromOpenedDirectory(dir File) (files []string, err error)
	// Lls lists all files and directory (equivalent to ls -l)
	Lls(dir string) (files []os.FileInfo, err error)
	// LlsFromOpenedDirectory lists all files and directory (equivalent to ls -l)
	LlsFromOpenedDirectory(dir File) (files []os.FileInfo, err error)
	// Copy copies files and directory (equivalent to cp -r)
	Copy(src string, dest string) (err error)
	// CopyWithContext copies files and directory (equivalent to cp -r)
	CopyWithContext(ctx context.Context, src string, dest string) (err error)
	// Move moves a file (equivalent to mv)
	Move(src string, dest string) (err error)
	// MoveWithContext moves a file (equivalent to mv)
	MoveWithContext(ctx context.Context, src string, dest string) (err error)
	// TempDir creates a temp directory
	TempDir(dir string, prefix string) (name string, err error)
	// TempDirInTempDir creates a temp directory in temp directory.
	TempDirInTempDir(prefix string) (name string, err error)
	// TempFile creates a temp file
	TempFile(dir string, pattern string) (f File, err error)
	// TempFileInTempDir creates a temp file in temp directory.
	TempFileInTempDir(pattern string) (f File, err error)
	// TempDirectory returns the temp directory.
	TempDirectory() string
	// CurrentDirectory returns current directory.
	CurrentDirectory() (string, error)
	// ReadFile reads a file and return its content.
	ReadFile(filename string) ([]byte, error)
	// ReadFileWithLimits reads a file and return its content. Nonetheless, it stops with EOF after FileSystemLimits are exceeded.
	ReadFileWithLimits(filename string, limits ILimits) ([]byte, error)
	// WriteFile writes data to a file named by filename.
	// If the file does not exist, WriteFile creates it with permissions perm;
	// otherwise WriteFile truncates it before writing.
	WriteFile(filename string, data []byte, perm os.FileMode) error
	// GarbageCollect runs the Garbage collector on the filesystem (removes any file which has not been accessed for a certain duration)
	GarbageCollect(root string, durationSinceLastAccess time.Duration) error
	// GarbageCollectWithContext runs the Garbage collector on the filesystem (removes any file which has not been accessed for a certain duration)
	GarbageCollectWithContext(ctx context.Context, root string, durationSinceLastAccess time.Duration) error
	// Chmod changes the mode of the named file to mode.
	Chmod(name string, mode os.FileMode) error
	// Chtimes changes the access and modification times of the named file
	Chtimes(name string, atime time.Time, mtime time.Time) error
	// Chown changes the numeric uid and gid of the named file.
	Chown(name string, uid, gid int) error
	// Link creates newname as a hard link to the oldname file
	Link(oldname, newname string) error
	// Readlink returns the destination of the named symbolic link.
	Readlink(name string) (string, error)
	// Symlink creates newname as a symbolic link to oldname.
	Symlink(oldname string, newname string) error
	// DiskUsage determines Disk usage
	DiskUsage(name string) (DiskUsage, error)
	// GetFileSize gets file size
	GetFileSize(filename string) (int64, error)
	// SubDirectories returns a list of all subdirectories (which are not hidden) names
	SubDirectories(directory string) ([]string, error)
	// SubDirectoriesWithContext returns a list of all subdirectories (which are not hidden)
	SubDirectoriesWithContext(ctx context.Context, directory string) ([]string, error)
	// ListDirTree lists the content of directory recursively
	ListDirTree(dirPath string, list *[]string) error
	// ListDirTreeWithContext lists the content of directory recursively
	ListDirTreeWithContext(ctx context.Context, dirPath string, list *[]string) error
	// ConvertFilePath gets FS file path instead of real file path. In most cases, returned file path
	// should be identical however this may not be true for some particular file systems e.g. for base FS, file path
	// returned will have any base prefix removed.
	ConvertFilePath(name string) string
	// ConvertToRelativePath converts a list of paths to relative paths
	ConvertToRelativePath(rootPath string, paths ...string) ([]string, error)
	// ConvertToAbsolutePath converts a list of paths to relative paths
	ConvertToAbsolutePath(rootPath string, paths ...string) ([]string, error)
	// NewRemoteLockFile creates a lock file on a remote location (NFS)
	NewRemoteLockFile(id string, dirToLock string) ILock
	// Zip compresses a file tree (source) into a zip file (destination)
	Zip(source string, destination string) error
	// ZipWithContext compresses a file tree (source) into a zip file (destination)
	ZipWithContext(ctx context.Context, source string, destination string) error
	// 	ZipWithContextAndLimits(ctx context.Context, source string, destination string) error compresses a file tree (source) into a zip file (destination) .Nonetheless, if FileSystemLimits are exceeded, an error will be returned and the process will be stopped.
	// It is however the responsibility of the caller to clean any partially created zipped archive if error occurs.
	ZipWithContextAndLimits(ctx context.Context, source string, destination string, limits ILimits) error
	// Unzip decompresses a source zip archive into the destination
	Unzip(source string, destination string) ([]string, error)
	// UnzipWithContext decompresses a source zip archive into the destination
	UnzipWithContext(ctx context.Context, source string, destination string) ([]string, error)
	// UnzipWithContextAndLimits decompresses a source zip archive into the destination. Nonetheless, if FileSystemLimits are exceeded, an error will be returned and the process will be stopped.
	// It is however the responsibility of the caller to clean any partially unzipped archive if error occurs.
	UnzipWithContextAndLimits(ctx context.Context, source string, destination string, limits ILimits) (fileList []string, err error)
	// FileHash calculates file hash
	FileHash(hashAlgo string, path string) (string, error)
}
