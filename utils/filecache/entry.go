package filecache

import (
	"context"
	"time"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

type CacheEntry struct {
	path       string
	ttl        time.Duration
	expiration time.Time
}

func (e *CacheEntry) Copy(ctx context.Context, cacheFs filesystem.FS, destFs filesystem.FS, destPath string) error {
	// Copying to a temp location first, then move to the original location.
	// This prevents the cache to present a transient file/dir that is still in the copying process.

	tmpDir, err := destFs.TempDirInTempDir("filecache-tmp")
	if err != nil {
		return err
	}

	tmpPath := filesystem.FilePathJoin(destFs, tmpDir, filesystem.FilePathBase(cacheFs, e.path))
	if err := filesystem.CopyBetweenFS(ctx, cacheFs, e.path, destFs, tmpPath); err != nil {
		return err
	}

	err = destFs.Move(tmpPath, destPath)
	if err != nil {
		return err
	}

	return nil
}

func (e *CacheEntry) Delete(ctx context.Context, fs filesystem.FS) error {
	if err := fs.RemoveWithContext(ctx, e.path); err != nil {
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

func NewCacheEntry(path string, ttl time.Duration) ICacheEntry {
	return &CacheEntry{
		path:       path,
		ttl:        ttl,
		expiration: time.Now().Add(ttl),
	}
}
