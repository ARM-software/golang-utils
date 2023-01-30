package sharedcache

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/hashing"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

const defaultCachedPackage = "cache.zip"
const hashFileDescriptor = ".hash"

// GenerateKey generates a key based on a list of key elements `elems`.
func GenerateKey(elems ...string) string {
	hash := "ef46db3751d8e999" // xxhash(""). No point in calculating it for "" everytime
	for _, elem := range elems {
		hash = hashing.CalculateHash(hash+elem, hashing.HashXXHash)
	}
	return hash
}

func generateHashFileName(srcFile string) string {
	return fmt.Sprintf("%v%v", filepath.Base(srcFile), hashFileDescriptor)
}
func getHash(ctx context.Context, fs filesystem.FS, src string, forceHashUpdate bool) (hash string, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	hashFile := filepath.Join(filepath.Dir(src), generateHashFileName(src))
	if fs.Exists(hashFile) && !forceHashUpdate {
		fileContents, readErr := fs.ReadFile(hashFile)
		if readErr == nil {
			hash = string(fileContents)
			if len(hash) == 16 { // if valid hash
				// xxhash 64 bit will have 16 characters
				return
			}
		}
	}

	hash, err = fs.FileHash(hashing.HashXXHash, src)
	if err != nil {
		return
	}
	_ = fs.WriteFile(hashFile, []byte(hash), 0775)
	return
}

func isPathForDirectory(fs filesystem.FS, dst string) bool {
	if ok, _ := fs.IsDir(dst); ok {
		return true
	}
	if filepath.Ext(dst) == "" {
		return true
	}
	return strings.HasSuffix(dst, string(fs.PathSeparator())) || strings.HasSuffix(dst, "/")
}

// TransferFiles transfers a file from a location `src` to another `dest` and ensures the integrity (i.e. hash validation) of what was copied across.
// `dest` can be a file or a directory. If not existent, it will be created on the fly. non-existent directory path should be terminated by a path separator i.e. / or \
func TransferFiles(ctx context.Context, fs filesystem.FS, dst, src string) (destFile string, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if !fs.Exists(src) {
		err = fmt.Errorf("%w: source path does not exist [%v]", commonerrors.ErrNotFound, src)
		return
	}
	if result, suberr := fs.IsFile(src); !result || suberr != nil {
		err = fmt.Errorf("%w: source is expected to be a file [%v, (%v)]", commonerrors.ErrInvalid, src, suberr)
		return
	}

	destDir := dst
	destFile = dst
	renameFile := false

	if isPathForDirectory(fs, destFile) {
		baseName := filepath.Base(src)
		destFile = filepath.Join(destDir, baseName)
	} else {
		destDir = filepath.Dir(dst)
		renameFile = true
	}

	// get hash of original file
	hash1, err := getHash(ctx, fs, src, false)
	if err != nil {
		return
	}

	// download/upload from/to cache
	err = fs.CopyWithContext(ctx, src, destDir)
	if err != nil {
		return
	}

	if renameFile {
		err = fs.MoveWithContext(ctx, filepath.Join(destDir, filepath.Base(src)), destFile)
		if err != nil {
			return
		}
	}

	// get hash of downloaded/uploaded file
	hash2, err := getHash(ctx, fs, destFile, true)
	if err != nil {
		return
	}

	// check that hashes match
	if !strings.EqualFold(hash1, hash2) {
		_ = fs.Rm(destFile)
		err = fmt.Errorf("%w: error occurred during file transfer (hash mismatch) [%v]", commonerrors.ErrInvalid, src)
		return
	}
	return
}

// AbstractSharedCacheRepository defines an abstract cache repository.
type AbstractSharedCacheRepository struct {
	cfg *Configuration
	fs  *filesystem.VFS
}

func (c *AbstractSharedCacheRepository) getCacheEntryPath(key string) string {
	return filepath.Join(c.cfg.RemoteStoragePath, key)
}

func (c *AbstractSharedCacheRepository) createEntry(ctx context.Context, key string) (entryPath string, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	entryPath = c.getCacheEntryPath(key)
	// create remote directory if this is the first time
	err = c.fs.MkDir(entryPath) // won't do anything if dir already exists
	return
}

func (c *AbstractSharedCacheRepository) GenerateKey(elems ...string) string {
	return GenerateKey(elems...)
}

func (c *AbstractSharedCacheRepository) RemoveEntry(ctx context.Context, key string) error {
	err := parallelisation.DetermineContextError(ctx)
	if err != nil {
		return err
	}
	return c.fs.Rm(c.getCacheEntryPath(key))
}

func (c *AbstractSharedCacheRepository) GetEntries(ctx context.Context) (entries []string, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	entries, err = c.fs.Ls(c.cfg.RemoteStoragePath)
	if err != nil {
		return
	}
	if reflection.IsEmpty(c.cfg.FilesystemItemsToIgnore) {
		return
	}
	toIgnore := collection.ParseCommaSeparatedList(c.cfg.FilesystemItemsToIgnore)
	entries, err = c.fs.ExcludeAll(entries, toIgnore...)
	return
}

func (c *AbstractSharedCacheRepository) EntriesCount(ctx context.Context) (count int64, err error) {
	files, err := c.GetEntries(ctx)
	if err != nil {
		return
	}
	count = int64(len(files))
	return
}

func (c *AbstractSharedCacheRepository) setUpLocalDestination(ctx context.Context, dest string) (err error) {
	// create localDir if necessary and remove old cache file(s)
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = c.fs.MkDir(dest)
	if err != nil {
		return
	}
	err = c.fs.CleanDir(dest)
	if err != nil {
		return
	}
	return
}

func (c *AbstractSharedCacheRepository) unpackPackageToLocalDestination(ctx context.Context, cachedPackagePath, dest string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	// create temp location for cached package to be unzipped
	tempDir, err := c.fs.TempDirInTempDir(tempDirPrefix)
	if err != nil {
		return
	}
	defer func() { _ = c.fs.Rm(tempDir) }()
	// do the transfer to a temporary folder.
	destZip, err := TransferFiles(ctx, c.fs, tempDir, cachedPackagePath)
	defer func() { _ = c.fs.Rm(destZip) }()
	if err != nil {
		return
	}

	// unpack package into destination
	_, err = c.fs.UnzipWithContext(ctx, destZip, dest)
	return
}

func (c *AbstractSharedCacheRepository) getEntryAge(ctx context.Context, key string, getCachedPackageFromEntryPath func(ctx context.Context, key, entryDir string) (string, error)) (age time.Duration, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	remoteDir := c.getCacheEntryPath(key)
	remoteExists, err := c.fs.IsDir(remoteDir)
	if err != nil {
		return
	}
	if !remoteExists {
		err = fmt.Errorf("no cache entry for key [%v]: %w", key, commonerrors.ErrNotFound)
		return
	}
	dirTime, suberr := c.fs.StatTimes(remoteDir)
	cachedPackage, err := getCachedPackageFromEntryPath(ctx, key, remoteDir)
	if err != nil {
		err = suberr
		if err == nil && dirTime != nil {
			age = time.Since(dirTime.ModTime())
		}
		return
	}
	packageTime, err := c.fs.StatTimes(cachedPackage)
	if err != nil {
		if suberr == nil && dirTime != nil {
			err = suberr
			age = time.Since(dirTime.ModTime())
			return
		}
		return
	}
	age = time.Since(packageTime.ModTime())
	return
}

func (c *AbstractSharedCacheRepository) setEntryAge(ctx context.Context, key string, age time.Duration, getCachedPackageFromEntryPath func(ctx context.Context, key, entryDir string) (string, error)) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	modTime := time.Now().Add(-age)
	remoteDir := c.getCacheEntryPath(key)
	remoteExists, err := c.fs.IsDir(remoteDir)
	if err != nil {
		return
	}
	if !remoteExists {
		err = fmt.Errorf("no cache entry for key [%v]: %w", key, commonerrors.ErrNotFound)
		return
	}
	err = c.fs.Chtimes(remoteDir, modTime, modTime)
	if err != nil {
		return
	}
	cachedPackage, err := getCachedPackageFromEntryPath(ctx, key, remoteDir)
	if err != nil {
		return
	}
	err = c.fs.Chtimes(cachedPackage, modTime, modTime)
	return
}

func NewAbstractSharedCacheRepository(cfg *Configuration, fs filesystem.FS) (cache *AbstractSharedCacheRepository, err error) {
	if cfg == nil {
		err = fmt.Errorf("%w: missing configuration", commonerrors.ErrUndefined)
		return
	}
	err = cfg.Validate()
	if err != nil {
		err = fmt.Errorf("%w: invalid configuration: %v", commonerrors.ErrInvalid, err.Error())
		return
	}
	rawFs, ok := fs.(*filesystem.VFS)
	if !ok {
		err = fmt.Errorf("%w: to work properly, cache needs a VFS filesystem implementation", commonerrors.ErrNotImplemented)
		return
	}
	cache = &AbstractSharedCacheRepository{
		cfg: cfg,
		fs:  rawFs,
	}
	return
}
