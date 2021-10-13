package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/bmatcuk/doublestar/v3"
	"github.com/spf13/afero"
)

//go:generate mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE IFileHash,Chowner,Linker,File,DiskUsage,FileTimeInfo,ILock,FS

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

// FIXME it should be noted that despite being possible to use the lock with an in-memory filesystem, it should be avoided at all cost.
// The implementation of the in-memory FS used (afero) has shown several thread safety issues (e.g. https://github.com/spf13/afero/issues/298) and therefore, should not be used for scenarios involving concurrency until it is fixed.
type ILock interface {
	Lock(ctx context.Context) error                                   // Locks the lock. This call will wait (i.e. block) until the lock is available.
	LockWithTimeout(ctx context.Context, timeout time.Duration) error // Tries to lock the lock until the timeout expires. If the timeout expires, this method will return commonerror.ErrTimeout.
	TryLock(ctx context.Context) error                                // Attempts to lock the lock instantly. This method will return commonerrors.ErrLocked immediately if the lock cannot be acquired straight away.
	IsStale() bool                                                    // Determines whether a lock is stale (the owner forgot to release it or is dead) or not.
	Unlock(ctx context.Context) error                                 // Releases the lock. This takes precedence over any current lock.
	ReleaseIfStale(ctx context.Context) error                         // Forces the lock to be released if it is considered as stale.
	MakeStale(ctx context.Context) error                              // Makes the lock stale. This is mostly for testing purposes.
}

type FS interface {
	//The following is for being able to use doublestar
	Open(name string) (doublestar.File, error)
	// Open a file for reading. It opens the named file with specified flag (O_RDONLY etc.).
	// See os.Open()
	GenericOpen(name string) (File, error)
	// OpenFile opens a file using the given flags and the given mode.
	// OpenFile is the generalized open call
	// most users will use GenericOpen or Create instead.
	// See os.OpenFile
	OpenFile(name string, flag int, perm os.FileMode) (File, error)
	// Creates a file.
	CreateFile(name string) (File, error)
	// Gets the path separator character.
	PathSeparator() rune
	Stat(name string) (os.FileInfo, error)
	Lstat(name string) (os.FileInfo, error)
	// Gets file time information.
	StatTimes(name string) (FileTimeInfo, error)
	// Gets the type of the file system.
	GetType() int
	// Removes all the files in a directory (equivalent rm -rf .../*)
	CleanDir(dir string) (err error)
	// Checks if a file or folder exists
	Exists(path string) bool
	// Removes directory (equivalent to rm -r)
	Rm(dir string) (err error)
	// States whether it is a file or not
	IsFile(path string) (result bool, err error)
	// States whether it is a directory or not
	IsDir(path string) (result bool, err error)
	// States whether it is a link or not
	IsLink(path string) (result bool, err error)
	// Checks whether a path is empty or not
	IsEmpty(name string) (empty bool, err error)
	// Makes directory (equivalent to mkdir -p)
	MkDir(dir string) (err error)
	// Makes directory (equivalent to mkdir -p)
	MkDirAll(dir string, perm os.FileMode) (err error)
	// Excludes files from a list. Returns the list without the path matching the exclusion patterns.
	ExcludeAll(files []string, exclusionPatterns ...string) ([]string, error)
	// Finds all the files with extensions
	FindAll(dir string, extensions ...string) (files []string, err error)
	// Walks  the file tree rooted at root, calling fn for each file or
	// directory in the tree, including root. See https://golang.org/pkg/path/filepath/#WalkDir
	Walk(root string, fn filepath.WalkFunc) error
	// Lists all files and directory (equivalent to ls)
	Ls(dir string) (files []string, err error)
	LsFromOpenedDirectory(dir File) (files []string, err error)
	// Lists all files and directory (equivalent to ls -l)
	Lls(dir string) (files []os.FileInfo, err error)
	LlsFromOpenedDirectory(dir File) (files []os.FileInfo, err error)
	// Copy files and directory (equivalent to cp -r)
	Copy(src string, dest string) (err error)
	// Moves a file (equivalent to mv)
	Move(src string, dest string) (err error)
	// Creates a temp directory
	TempDir(dir string, prefix string) (name string, err error)
	// Creates a temp directory in temp directory.
	TempDirInTempDir(prefix string) (name string, err error)
	// Creates a temp file
	TempFile(dir string, pattern string) (f File, err error)
	// Creates a temp file in temp directory.
	TempFileInTempDir(pattern string) (f File, err error)
	// Gets temp directory.
	TempDirectory() string
	// Gets current directory.
	CurrentDirectory() (string, error)
	// Reads a file and return its content.
	ReadFile(filename string) ([]byte, error)
	// Writes data to a file named by filename.
	// If the file does not exist, WriteFile creates it with permissions perm;
	// otherwise WriteFile truncates it before writing.
	WriteFile(filename string, data []byte, perm os.FileMode) error
	// Runs the Garbage collector on the filesystem (removes any file which has not been accessed for a certain duration)
	GarbageCollect(root string, durationSinceLastAccess time.Duration) error
	// Changes the mode of the named file to mode.
	Chmod(name string, mode os.FileMode) error
	// Changes the access and modification times of the named file
	Chtimes(name string, atime time.Time, mtime time.Time) error
	// Changes the numeric uid and gid of the named file.
	Chown(name string, uid, gid int) error
	// Creates newname as a hard link to the oldname file
	Link(oldname, newname string) error
	// Returns the destination of the named symbolic link.
	Readlink(name string) (string, error)
	// Creates newname as a symbolic link to oldname.
	Symlink(oldname string, newname string) error
	// Determines Disk usage
	DiskUsage(name string) (DiskUsage, error)
	// Gets file size
	GetFileSize(filename string) (int64, error)
	// Returns a list of all subdirectories (which are not hidden)
	SubDirectories(directory string) ([]string, error)
	// Lists the content of directory recursively
	ListDirTree(dirPath string, list *[]string) error
	// Gets FS file path instead of real file path. In most cases, returned file path
	// should be identical however this may not be true for some particular file systems e.g. for base FS, file path
	// returned will have any base prefix removed.
	ConvertFilePath(name string) string
	// Converts a list of paths to relative paths
	ConvertToRelativePath(rootPath string, paths []string) ([]string, error)
	// Converts a list of paths to relative paths
	ConvertToAbsolutePath(rootPath string, paths []string) ([]string, error)
	// Creates a lock file on a remote location (NFS)
	NewRemoteLockFile(id string, dirToLock string) ILock
	// Compresses a file tree (source) into a zip file (destination)
	Zip(source string, destination string) error
	// Decompresses a source zip archive into the destination
	Unzip(source string, destination string) ([]string, error)
	// Calculates file hash
	FileHash(hashAlgo string, path string) (string, error)
}
