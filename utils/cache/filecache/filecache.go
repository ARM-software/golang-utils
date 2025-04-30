package filecache

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

type Cache struct {
	entries       iEntryMap
	entriesLM     iLockMap
	entryProvider IEntryRetriever
	fs            filesystem.FS
	cfg           *FileCacheConfig
	cancelStore   *parallelisation.CancelFunctionStore
	closed        atomic.Bool
}

func (c *Cache) gc(ctx context.Context, _ time.Time) {
	if err := parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	c.entries.Range(func(key string, entry ICacheEntry) {
		if entry.IsExpired() && c.entriesLM.TryLock(key) {
			if err := entry.Delete(ctx); err == nil {
				c.entries.Delete(key)
				defer func() {
					c.entriesLM.Unlock(key)
					c.entriesLM.Delete(key)
				}()
			}
		}
	})
}

func (c *Cache) checkIfClosed() error {
	if c.closed.Load() {
		return commonerrors.New(commonerrors.ErrConflict, "cache is closed")
	}

	return nil
}

func (c *Cache) Has(ctx context.Context, key string) (bool, error) {
	if err := c.checkIfClosed(); err != nil {
		return false, err
	}

	return c.entries.Exists(key), nil
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

	if c.entries.Exists(key) {
		return commonerrors.Newf(commonerrors.ErrExists, "cache entry %s already exists", key)
	}

	entryPath, err := c.entryProvider.FetchEntry(ctx, key)
	if err != nil {
		return err
	}

	c.entries.Store(key, NewCacheEntry(c.fs, entryPath, ttl))
	c.entriesLM.Store(key)

	return nil
}

func (c *Cache) Fetch(ctx context.Context, key string, destFilesystem filesystem.FS, destPath string) error {
	if err := c.checkIfClosed(); err != nil {
		return err
	}

	c.entriesLM.Lock(key)
	defer c.entriesLM.Unlock(key)

	if !c.entries.Exists(key) {
		return commonerrors.Newf(commonerrors.ErrNotFound, "cache entry for %s not found", key)
	}

	cpCtx, stop := context.WithCancel(ctx)
	c.cancelStore.RegisterCancelFunction(stop)

	entry := c.entries.Load(key)
	if entry == nil {
		return commonerrors.Newf(commonerrors.ErrUnexpected, "cache entry for %s could not be loaded", key)
	}

	if err := entry.Copy(cpCtx, destFilesystem, destPath); err != nil {
		return err
	}

	entry.ExtendLifetime()

	return nil
}

func (c *Cache) Evict(ctx context.Context, key string) error {
	if err := c.checkIfClosed(); err != nil {
		return err
	}

	c.entriesLM.Lock(key)
	defer c.entriesLM.Unlock(key)

	if c.entries.Exists(key) {

		entry := c.entries.Load(key)
		if entry == nil {
			return commonerrors.Newf(commonerrors.ErrUnexpected, "cache entry for %s could not be loaded", key)
		}

		if err := entry.Delete(ctx); err != nil {
			return err
		}

		c.entries.Delete(key)
	}

	return nil
}

func (c *Cache) Close(ctx context.Context) error {
	if err := c.checkIfClosed(); err != nil {
		return nil
	}

	c.closed.Store(true)
	c.entriesLM.Clear()
	c.cancelStore.Cancel()

	var removalErrors []error
	c.entries.Range(func(key string, entry ICacheEntry) {
		if err := entry.Delete(ctx); err != nil {
			removalErrors = append(removalErrors, commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not delete %s", key))
			return
		}

		c.entries.Delete(key)
	})

	if len(removalErrors) > 0 {
		err := c.fs.RemoveWithPrivileges(ctx, c.cfg.CachePath)
		if err != nil {
			removalErrors = append(removalErrors, commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not delete %s", c.cfg.CachePath))
			return errors.Join(removalErrors...)
		}

		c.entries.Clear()
	}

	return nil
}

func NewGenericFileCache(ctx context.Context, cacheFilesystem filesystem.FS, entryRetriever IEntryRetriever, config *FileCacheConfig) (IFileCache, error) {
	if err := config.Validate(); err != nil {
		return nil, commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid configuration")

	}

	if entryRetriever == nil {
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

	if err := entryRetriever.SetCacheDir(cacheFilesystem, config.CachePath); err != nil {
		return nil, err
	}

	cache := &Cache{
		entries:       newEntryMap(),
		entriesLM:     newLockMap(),
		entryProvider: entryRetriever,
		fs:            cacheFilesystem,
		cfg:           config,
		cancelStore:   cancelStore,
	}

	parallelisation.SafeSchedule(gcCtx, cache.cfg.GarbageCollectionPeriod, 0, cache.gc)

	return cache, nil
}
