// Package filesystem defines utilities with regards to filesystem access./*
package filesystem

import (
	"context"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/bmatcuk/doublestar/v3"
	"github.com/spf13/afero"

	"github.com/ARM-software/golang-utils/utils/config"
)

//go:generate mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE IFileHash,IChowner,ILinker,File,DiskUsage,FileTimeInfo,ILock,ILimits,FS,ICloseableFS,IForceRemover,IStater,ILinkReader,ISymLinker

// IFileHash defines a file hash.
// For reference.
// https://stackoverflow.com/questions/1761607/what-is-the-fastest-hash-algorithm-to-check-if-two-files-are-equal
type IFileHash interface {
	Calculate(f File) (string, error)
	CalculateWithContext(ctx context.Context, f File) (string, error)
	CalculateFile(fs FS, path string) (string, error)
	CalculateFileWithContext(ctx context.Context, fs FS, path string) (string, error)
	GetType() string
}

// IForceRemover is an Optional interface. It is only implemented by the
// filesystems saying so.
type IForceRemover interface {
	// ForceRemoveIfPossible will remove an item with escalated permissions if need be (equivalent to sudo rm -rf)
	ForceRemoveIfPossible(name string) error
}

// IChowner is an Optional interface. It is only implemented by the
// filesystems saying so.
type IChowner interface {
	// ChownIfPossible will change the file ownership
	ChownIfPossible(name string, uid int, gid int) error
}

// ILinker is an Optional interface. It is only implemented by the
// filesystems saying so.
type ILinker interface {
	// LinkIfPossible creates a hard link between oldname and new name.
	LinkIfPossible(oldname, newname string) error
}

// ILinkReader is an Optional interface. It is only implemented by the
// filesystems saying so.
type ILinkReader interface {
	ReadlinkIfPossible(string) (string, error)
}

// IStater is an Optional interface. It is only implemented by the
// filesystems saying so.
type IStater interface {
	// LstatIfPossible returns file information about an item.
	LstatIfPossible(string) (os.FileInfo, bool, error)
}

// ISymLinker is an Optional interface. It is only implemented by the
// filesystems saying so.
type ISymLinker interface {
	SymlinkIfPossible(string, string) error
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
	// GetMaxFileCount returns the maximum files in byte a file can have on a file system
	GetMaxFileCount() int64
	// GetMaxDepth returns the maximum depth of directory tree allowed (set to a negative number if disabled)
	GetMaxDepth() int64
	// ApplyRecursively specifies whether recursive action should be applied or not e.g. expand zip within zips during unzipping
	ApplyRecursively() bool
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
// Note: When an API accepting exclusion patterns, it means its processing will not be applied on path matching an exclusion pattern
// An exclusion pattern correspond to a regex string following the syntax defined in the [regexp](https://pkg.go.dev/regexp) module.
type FS interface {
	// Open opens a file. The following is for being able to use `doublestar`. Use GenericOpen instead.
	Open(name string) (doublestar.File, error)
	// GenericOpen opens a file for reading. It opens the named file with specified flag (O_RDONLY etc.).
	// See os.Open()
	GenericOpen(name string) (File, error)
	// OpenFile opens a file using the given flags and the given mode.
	// OpenFile is the generalised open call
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
	GetType() FilesystemType
	// CleanDir removes all the files in a directory (equivalent rm -rf .../*)
	CleanDir(dir string) (err error)
	// CleanDirWithContext removes all the files in a directory (equivalent rm -rf .../*)
	CleanDirWithContext(ctx context.Context, dir string) (err error)
	// CleanDirWithContextAndExclusionPatterns removes all the files in a directory (equivalent rm -rf .../*) unless they match some exclusion pattern.
	CleanDirWithContextAndExclusionPatterns(ctx context.Context, dir string, exclusionPatterns ...string) (err error)
	// Exists checks whether a file or folder exists
	Exists(path string) bool
	// Rm removes directory (equivalent to rm -rf)
	Rm(dir string) (err error)
	// RemoveWithContext removes directory (equivalent to rm -rf)
	RemoveWithContext(ctx context.Context, dir string) (err error)
	// RemoveWithContextAndExclusionPatterns removes directory (equivalent to rm -rf) unless they match some exclusion pattern.
	RemoveWithContextAndExclusionPatterns(ctx context.Context, dir string, exclusionPatterns ...string) (err error)
	// RemoveWithPrivileges removes a directory even if it is not owned by user (equivalent to sudo rm -rf). It expects the current user to be a superuser.
	RemoveWithPrivileges(ctx context.Context, dir string) (err error)
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
	// Glob returns the paths of all files matching pattern with support for "doublestar" (aka globstar: **) patterns.
	Glob(pattern string) ([]string, error)
	// FindAll finds all the files with extensions
	FindAll(dir string, extensions ...string) (files []string, err error)
	// Walk walks  the file tree rooted at root, calling fn for each file or
	// directory in the tree, including root. See https://golang.org/pkg/path/filepath/#WalkDir
	Walk(root string, fn filepath.WalkFunc) error
	// WalkWithContext walks  the file tree rooted at root, calling fn for each file or
	// directory in the tree, including root. See https://golang.org/pkg/path/filepath/#WalkDir
	WalkWithContext(ctx context.Context, root string, fn filepath.WalkFunc) error
	// WalkWithContextAndExclusionPatterns walks through the file tree rooted at root, calling fn for each file or
	// directory in the tree as long as they do not match an exclusion pattern.
	WalkWithContextAndExclusionPatterns(ctx context.Context, root string, fn filepath.WalkFunc, exclusionPatterns ...string) error
	// Ls lists all files and directory (equivalent to ls)
	Ls(dir string) (files []string, err error)
	// LsRecursive lists all files recursively, including subdirectories
	LsRecursive(ctx context.Context, dir string, includeDirectories bool) (files []string, err error)
	// LsWithExclusionPatterns lists all files and directory (equivalent to ls) but exclude the ones matching the exclusion patterns.
	LsWithExclusionPatterns(dir string, exclusionPatterns ...string) (files []string, err error)
	// LsRecursiveWithExclusionPatterns lists all files recursively, including subdirectories but exclude the ones matching the exclusion patterns.
	LsRecursiveWithExclusionPatterns(ctx context.Context, dir string, includeDirectories bool, exclusionPatterns ...string) (files []string, err error)
	// LsRecursiveWithExclusionPatternsAndLimits lists all files recursively, including subdirectories but exclude the ones matching the exclusion patterns and add some limits for recursion
	LsRecursiveWithExclusionPatternsAndLimits(ctx context.Context, dir string, limit ILimits, includeDirectories bool, exclusionPatterns ...string) (files []string, err error)
	// LsFromOpenedDirectory lists all files and directories (equivalent to ls)
	LsFromOpenedDirectory(dir File) (files []string, err error)
	// LsRecursiveFromOpenedDirectory lists all files recursively
	LsRecursiveFromOpenedDirectory(ctx context.Context, dir File, includeDirectories bool) (files []string, err error)
	// Lls lists all files and directory (equivalent to ls -l)
	Lls(dir string) (files []os.FileInfo, err error)
	// LlsFromOpenedDirectory lists all files and directory (equivalent to ls -l)
	LlsFromOpenedDirectory(dir File) (files []os.FileInfo, err error)
	// CopyToFile copies a file to another file.
	CopyToFile(srcFile, destFile string) error
	// CopyToFileWithContext copies a file to another file similarly to CopyToFile.
	CopyToFileWithContext(ctx context.Context, srcFile, destFile string) error
	// CopyToDirectory copies a src to a directory  destDirectory which will be created as such if not present.
	CopyToDirectory(src, destDirectory string) error
	// CopyToDirectoryWithContext copies a src to a directory similarly to CopyToDirectory.
	CopyToDirectoryWithContext(ctx context.Context, src, destDirectory string) error
	// Copy copies files and directory (equivalent to [POSIX cp -r](https://www.unix.com/man-page/posix/1P/cp/) or DOS `copy` or `shutil.copy()`/`shutil.copytree()` in [python](https://docs.python.org/3/library/shutil.html#shutil.copy))
	// It should be noted that although the behaviour should match `cp -r` in most cases, there may be some corner cases in which the behaviour of Copy may differ slightly.
	// For instance, if the destination `dest` does not exist and the source is a file, then the destination will be a file unless the destination ends with a path separator and thus, the intention was to consider it as a folder.
	Copy(src string, dest string) (err error)
	// CopyWithContext copies files and directory similar to Copy.
	// Nonetheless, this function should be preferred over Copy as it is context aware, meaning it is possible to stop the copy if it is taking too long or if context is cancelled.
	CopyWithContext(ctx context.Context, src string, dest string) (err error)
	// CopyWithContextAndExclusionPatterns copies files and directory like CopyWithContext but ignores any file matching the exclusion pattern.
	CopyWithContextAndExclusionPatterns(ctx context.Context, src string, dest string, exclusionPatterns ...string) (err error)
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
	// TouchTempFile creates an empty temporary file in dir and returns its path.
	TouchTempFile(dir string, pattern string) (filename string, err error)
	// TempFileInTempDir creates a temp file in temp directory.
	TempFileInTempDir(pattern string) (f File, err error)
	// TouchTempFileInTempDir creates an empty temporary  in temp directory and returns its path.
	TouchTempFileInTempDir(pattern string) (filename string, err error)
	// TempDirectory returns the temp directory.
	TempDirectory() string
	// CurrentDirectory returns current directory.
	CurrentDirectory() (string, error)
	// ReadFile reads a file and returns its content.
	ReadFile(filename string) ([]byte, error)
	// ReadFileWithContext reads a file but with control of a context and returns its content.
	ReadFileWithContext(ctx context.Context, filename string) ([]byte, error)
	// ReadFileWithLimits reads a file and returns its content. Nonetheless, it stops with EOF after FileSystemLimits are exceeded.
	ReadFileWithLimits(filename string, limits ILimits) ([]byte, error)
	// ReadFileWithContextAndLimits reads a file and returns its content. Limits and context are taken into account during the reading process.
	ReadFileWithContextAndLimits(ctx context.Context, filename string, limits ILimits) ([]byte, error)
	// ReadFileContent reads a file and returns its content. Limits and context are taken into account during the reading process.
	ReadFileContent(ctx context.Context, file File, limits ILimits) ([]byte, error)
	// WriteFile writes data to a file named by filename.
	// If the file does not exist, WriteFile creates it with permissions perm;
	// otherwise WriteFile truncates it before writing.
	WriteFile(filename string, data []byte, perm os.FileMode) error
	// WriteFileWithContext writes data to a file named by filename.
	// It works like WriteFile but is also controlled by a context.
	WriteFileWithContext(ctx context.Context, filename string, data []byte, perm os.FileMode) error
	// WriteToFile writes data from a reader to a file named by filename.
	// If the file does not exist, WriteToFile creates it with permissions perm;
	// otherwise WriteFile truncates it before writing.
	// It returns the number of bytes written.
	WriteToFile(ctx context.Context, filename string, reader io.Reader, perm os.FileMode) (written int64, err error)
	// GarbageCollect runs the Garbage collector on the filesystem (removes any file which has not been accessed for a certain duration)
	GarbageCollect(root string, durationSinceLastAccess time.Duration) error
	// GarbageCollectWithContext runs the Garbage collector on the filesystem (removes any file which has not been accessed for a certain duration)
	GarbageCollectWithContext(ctx context.Context, root string, durationSinceLastAccess time.Duration) error
	// Chmod changes the mode of the named file to mode.
	Chmod(name string, mode os.FileMode) error
	// ChmodRecursively changes the mode of anything within the `path`.
	ChmodRecursively(ctx context.Context, path string, mode os.FileMode) error
	// Chtimes changes the access and modification times of the named file
	Chtimes(name string, atime time.Time, mtime time.Time) error
	// Chown changes the numeric uid and gid of the named file.
	Chown(name string, uid, gid int) error
	// ChownRecursively changes recursively the numeric uid and gid of any sub items of `path`.
	ChownRecursively(ctx context.Context, path string, uid, gid int) error
	// ChangeOwnership changes the ownership of the named file.
	ChangeOwnership(name string, owner *user.User) error
	// ChangeOwnershipRecursively changes the ownership anything within the path.
	ChangeOwnershipRecursively(ctx context.Context, path string, owner *user.User) error
	// FetchOwners returns the numeric uid and gid of the named file
	FetchOwners(name string) (uid, gid int, err error)
	// FetchFileOwner returns the owner of the named file.
	FetchFileOwner(name string) (owner *user.User, err error)
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
	// SubDirectories returns a list of all subdirectories names. Any "hidden" directory (i.e. starting with `.`) is ignored.
	SubDirectories(directory string) ([]string, error)
	// SubDirectoriesWithContext returns a list of all subdirectories which are not hidden
	SubDirectoriesWithContext(ctx context.Context, directory string) ([]string, error)
	// SubDirectoriesWithContextAndExclusionPatterns returns a list of all subdirectories but ignores any file matching the exclusion pattern.
	// Note: all folders are returned whether they are hidden or not unless matching an exclusion pattern.
	SubDirectoriesWithContextAndExclusionPatterns(ctx context.Context, directory string, exclusionPatterns ...string) ([]string, error)
	// Touch is a command used to update the access date and/or modification date of a computer file or directory (equivalent to posix touch).
	Touch(path string) error
	// ListDirTree lists the content of directory recursively
	ListDirTree(dirPath string, list *[]string) error
	// ListDirTreeWithContext lists the content of directory recursively
	ListDirTreeWithContext(ctx context.Context, dirPath string, list *[]string) error
	// ListDirTreeWithContextAndExclusionPatterns lists the content of directory recursively but ignores any file matching the exclusion pattern.
	ListDirTreeWithContextAndExclusionPatterns(ctx context.Context, dirPath string, list *[]string, exclusionPatterns ...string) error
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
	// ZipWithContextAndLimits compresses a file tree (source) into a zip file (destination) .Nonetheless, if FileSystemLimits are exceeded, an error will be returned and the process will be stopped.
	// It is however the responsibility of the caller to clean any partially created zipped archive if error occurs.
	ZipWithContextAndLimits(ctx context.Context, source string, destination string, limits ILimits) error
	// ZipWithContextAndLimitsAndExclusionPatterns compresses a file tree (source) into a zip file (destination) but ignores any file/folder matching an exclusion pattern.
	ZipWithContextAndLimitsAndExclusionPatterns(ctx context.Context, source string, destination string, limits ILimits, exclusionPatterns ...string) error
	// Unzip decompresses a source zip archive into the destination
	Unzip(source string, destination string) ([]string, error)
	// UnzipWithContext decompresses a source zip archive into the destination
	UnzipWithContext(ctx context.Context, source string, destination string) ([]string, error)
	// UnzipWithContextAndLimits decompresses a source zip archive into the destination. Nonetheless, if FileSystemLimits are exceeded, an error will be returned and the process will be stopped.
	// It is however the responsibility of the caller to clean any partially unzipped archive if error occurs.
	UnzipWithContextAndLimits(ctx context.Context, source string, destination string, limits ILimits) (fileList []string, err error)
	// FileHash calculates file hash
	FileHash(hashAlgo string, path string) (string, error)
	// FileHashWithContext calculates file hash
	FileHashWithContext(ctx context.Context, hashAlgo string, path string) (string, error)
	// IsZip states whether a file is a zip file or not. If the file does not exist, it will state whether the filename has a zip extension or not.
	IsZip(filepath string) bool
	// IsZipWithContext states whether a file is a zip file or not. Since the process can take some time (i.e type detection with sniffers such as http.DetectContentType), it is controlled by a context.
	IsZipWithContext(ctx context.Context, filepath string) (bool, error)
}

// ICloseableFS is a filesystem which utilises resources which must be closed when it is no longer used, such as open files. The close method is invoked to release resources that the object is holding.
type ICloseableFS interface {
	FS
	io.Closer
}
