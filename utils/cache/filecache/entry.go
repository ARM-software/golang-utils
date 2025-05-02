package filecache

import (
	"context"
	"time"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

type CacheEntry struct {
	cachePath  string
	cacheFs    filesystem.FS
	ttl        time.Duration
	expiration time.Time
}

func (e *CacheEntry) Copy(ctx context.Context, destFs filesystem.FS, destPath string) error {
	// Copying to a temp location first, then move to the original location.
	// This prevents the cache to present a transient file/dir that is still in the copying process.
	tmpDir, err := destFs.TempDirInTempDir("filecache-tmp")
	if err != nil {
		return err
	}

	tmpPath := filesystem.FilePathJoin(destFs, tmpDir, filesystem.FilePathBase(e.cacheFs, e.cachePath))
	if err := filesystem.CopyBetweenFS(ctx, e.cacheFs, e.cachePath, destFs, tmpPath); err != nil {
		return err
	}

	err = destFs.Move(tmpPath, destPath)
	if err != nil {
		return err
	}

	return nil
}

func (e *CacheEntry) Delete(ctx context.Context) error {
	if err := e.cacheFs.RemoveWithPrivileges(ctx, e.cachePath); err != nil {
		return err
	}

	return nil
}

func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.expiration)
}

func (e *CacheEntry) ExtendLifetime() {
	e.expiration = time.Now().Add(e.ttl)
}

func NewCacheEntry(cacheFilesystem filesystem.FS, path string, ttl time.Duration) ICacheEntry {
	return &CacheEntry{
		cachePath:  path,
		cacheFs:    cacheFilesystem,
		ttl:        ttl,
		expiration: time.Now().Add(ttl),
	}
}
