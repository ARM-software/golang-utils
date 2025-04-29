package filecache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	fs "github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

type CacheEntry struct {
	path       string
	expiration time.Time
	ttl        time.Duration
}

type Cache struct {
	entries       map[string]*CacheEntry
	fs            fs.FS
	cfg           *FileCacheConfig
	gcCancelStore *parallelisation.CancelFunctionStore
	entryProvider IEntryProvider
	mu            sync.RWMutex
	closed        bool
}

func (c *Cache) gc(ctx context.Context, _ time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for id, entry := range c.entries {
		if now.After(entry.expiration) {
			if err := c.fs.RemoveWithContext(ctx, entry.path); err == nil {
				delete(c.entries, id)
			}
		}
	}
}

func (c *Cache) checkIfClosed() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return commonerrors.New(commonerrors.ErrConflict, "cache is closed")
	}

	return nil
}

func (c *Cache) Has(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.checkIfClosed(); err != nil {
		return false, err
	}

	_, exists := c.entries[key]
	return exists, nil
}

func (c *Cache) Store(ctx context.Context, key string) error {
	return c.StoreWithTTL(ctx, key, c.cfg.TTL)
}

func (c *Cache) StoreWithTTL(ctx context.Context, key string, ttl time.Duration) error {
	if err := c.checkIfClosed(); err != nil {
		return err
	}

	if ttl == 0 {
		return commonerrors.New(commonerrors.ErrInvalid, "TTL for an entry can't be zero")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.entries[key]; exists {
		return commonerrors.Newf(commonerrors.ErrExists, "cache entry %s already exists", key)
	}

	entryPath, err := c.entryProvider.FetchEntry(ctx, key)
	if err != nil {
		return err
	}

	c.entries[key] = &CacheEntry{
		path:       entryPath,
		expiration: time.Now().Add(ttl),
		ttl:        ttl,
	}

	return nil
}

func (c *Cache) Fetch(ctx context.Context, key string, destFilesystem fs.FS, destPath string) error {
	if err := c.checkIfClosed(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[key]
	if !exists {
		return commonerrors.Newf(commonerrors.ErrNotFound, "cache entry for %s not found", key)
	}

	// Copying to a temp location first, then move to the original location.
	// This prevents the cache to present a transient file/dir that is still in the copying process.
	tmpDir, err := destFilesystem.TempDirInTempDir("filecache-tmp")
	if err != nil {
		return err
	}

	dstPath := fs.FilePathJoin(destFilesystem, tmpDir, fs.FilePathBase(c.fs, entry.path))
	err = fs.CopyBetweenFS(ctx, c.fs, entry.path, destFilesystem, dstPath)
	if err != nil {
		return err
	}

	err = destFilesystem.Move(dstPath, destPath)
	if err != nil {
		return err
	}

	// Sliding window approach, which means frequently used items will stay cached
	entry.expiration = time.Now().Add(entry.ttl)

	return nil
}

func (c *Cache) Evict(ctx context.Context, key string) error {
	if err := c.checkIfClosed(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.entries[key]; exists {

		err := c.fs.RemoveWithContext(ctx, entry.path)
		if err != nil {
			return err
		}

		delete(c.entries, key)
	}

	return nil
}

func (c *Cache) Close() error {
	if err := c.checkIfClosed(); err != nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true
	c.gcCancelStore.Cancel()

	ctx := context.Background()
	var removalErrors []error
	for id, entry := range c.entries {
		if err := c.fs.RemoveWithContext(ctx, entry.path); err != nil {
			removalErrors = append(removalErrors, commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not delete %s", id))
			continue
		}

		delete(c.entries, id)
	}

	if len(removalErrors) > 0 {
		err := c.fs.RemoveWithPrivileges(ctx, c.cfg.CachePath)
		if err != nil {
			removalErrors = append(removalErrors, commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not delete %s", c.cfg.CachePath))
			return errors.Join(removalErrors...)
		}

		clear(c.entries)
	}

	return nil
}

func NewGenericFileCache(ctx context.Context, cacheFilesystem fs.FS, entryProvider IEntryProvider, config *FileCacheConfig) (IFileCache, error) {
	if err := config.Validate(); err != nil {
		return nil, commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid configuration")

	}

	if entryProvider == nil {
		return nil, commonerrors.New(commonerrors.ErrInvalid, "the entry provider cannot be nil")
	}

	if cacheFilesystem == nil {
		return nil, commonerrors.New(commonerrors.ErrInvalid, "the cache filesystem cannot be nil")
	}

	cancelStore := parallelisation.NewCancelFunctionsStore()
	gcCtx, stop := context.WithCancel(ctx)
	cancelStore.RegisterCancelFunction(stop)

	if err := cacheFilesystem.MkDirAll(config.CachePath, 0755); err != nil {
		return nil, err
	}

	if err := entryProvider.SetCacheFilesystem(cacheFilesystem); err != nil {
		return nil, err
	}

	if err := entryProvider.SetCacheDir(config.CachePath); err != nil {
		return nil, err
	}

	cache := &Cache{
		entries:       make(map[string]*CacheEntry),
		fs:            cacheFilesystem,
		entryProvider: entryProvider,
		cfg:           config,
		gcCancelStore: cancelStore,
	}

	parallelisation.SafeSchedule(gcCtx, cache.cfg.GarbageCollectionPeriod, 0, cache.gc)

	return cache, nil
}
