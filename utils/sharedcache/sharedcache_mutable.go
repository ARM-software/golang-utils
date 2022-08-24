package sharedcache

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

const (
	lockPrefix    = "SharedMutableCache"
	tempDirPrefix = "sharedmutablecache-packing"
)

// SharedMutableCacheRepository defines a shared cache using a distributed lock system solely based on lock files.
type SharedMutableCacheRepository struct {
	AbstractSharedCacheRepository
	lockTimeout time.Duration
}

func NewSharedMutableCacheRepository(cfg *Configuration, fs filesystem.FS) (repository ISharedCacheRepository, err error) {
	abstractCache, err := NewAbstractSharedCacheRepository(cfg, fs)
	if err != nil {
		return
	}
	if reflection.IsEmpty(cfg.Timeout) {
		err = fmt.Errorf("%w: cache timeout cannot be unset", commonerrors.ErrUndefined)
	}
	repository = &SharedMutableCacheRepository{
		AbstractSharedCacheRepository: *abstractCache,
		lockTimeout:                   cfg.Timeout,
	}
	return
}

func (s *SharedMutableCacheRepository) Fetch(ctx context.Context, key, dest string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	// setup the lock
	remoteDir := s.getCacheEntryPath(key)
	remoteExists, err := s.fs.IsDir(remoteDir)
	if err != nil {
		return
	}
	if !remoteExists {
		return fmt.Errorf("no cache entry for key [%v]: %w", key, commonerrors.ErrNotFound)
	}

	cachedPackage, err := s.findCachedPackageFromEntryDir(ctx, key, remoteDir)
	if err != nil {
		return
	}
	err = s.setUpLocalDestination(ctx, dest)
	if err != nil {
		return
	}

	remoteLock := s.generateEntryLock(key)
	defer func() { _ = remoteLock.Unlock(ctx) }()

	// do the transfer
	err = remoteLock.LockWithTimeout(ctx, s.lockTimeout)
	if err != nil {
		return
	}
	err = s.unpackPackageToLocalDestination(ctx, cachedPackage, dest)
	if err != nil {
		return
	}
	err = remoteLock.Unlock(ctx)
	return
}

func (s *SharedMutableCacheRepository) Store(ctx context.Context, key, src string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	remoteDir, err := s.createEntry(ctx, key)
	if err != nil {
		return
	}

	// create temp location for files so we don't include zip inside itself
	tempDir, err := s.fs.TempDirInTempDir(tempDirPrefix)
	if err != nil {
		return
	}
	defer func() { _ = s.fs.Rm(tempDir) }()

	// zip the local cache
	zipped := filepath.Join(tempDir, defaultCachedPackage)
	err = s.fs.ZipWithContext(ctx, src, zipped)
	if err != nil {
		return
	}

	remoteLock := s.generateEntryLock(key)
	defer func() { _ = remoteLock.Unlock(ctx) }()

	// Do the transfer
	err = remoteLock.LockWithTimeout(ctx, s.lockTimeout)
	if err != nil {
		return
	}
	destZip, err := TransferFiles(ctx, s.fs, remoteDir, zipped)
	if err != nil {
		_ = s.fs.Rm(destZip)
		return
	}
	err = remoteLock.Unlock(ctx)
	return
}

func (s *SharedMutableCacheRepository) CleanEntry(ctx context.Context, key string) error {
	lock := s.generateEntryLock(key)
	return lock.ReleaseIfStale(ctx)
}

func (s *SharedMutableCacheRepository) generateEntryLock(key string) filesystem.ILock {
	lockID := fmt.Sprintf("%v-%v", lockPrefix, key)
	return filesystem.NewRemoteLockFile(s.fs, lockID, s.getCacheEntryPath(key))
}

func (s *SharedMutableCacheRepository) findCachedPackageFromEntryDir(ctx context.Context, key, entryDir string) (cachedPackage string, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	cachedPackage = getCachedPackagePath(entryDir)
	if !s.fs.Exists(cachedPackage) {
		err = fmt.Errorf("no entry for key [%v] in cache: %w", key, commonerrors.ErrEmpty)
	}
	return
}

func (s *SharedMutableCacheRepository) GetEntryAge(ctx context.Context, key string) (time.Duration, error) {
	return s.getEntryAge(ctx, key, s.findCachedPackageFromEntryDir)
}

func (s *SharedMutableCacheRepository) SetEntryAge(ctx context.Context, key string, age time.Duration) error {
	return s.setEntryAge(ctx, key, age, s.findCachedPackageFromEntryDir)
}

func getCachedPackagePath(remoteDir string) string {
	return filepath.Join(remoteDir, defaultCachedPackage)
}
