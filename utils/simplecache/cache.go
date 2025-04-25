package simplecache

import (
	"context"
	"sync"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	fs "github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

const (
	_DefaultGCPeriod = 10 * time.Minute
	_DefaultTTL      = 2 * time.Hour
)

type CacheEntry struct {
	path         string
	originalFS   fs.FS
	originalpath string
	expiration   time.Time
}

type Cache struct {
	entries map[string]*CacheEntry
	cfg     *Config
	stopGC  context.CancelFunc
	mu      sync.Mutex
	closed  bool
}

func (c *Cache) gc(ctx context.Context, _ time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for id, entry := range c.entries {
		if now.After(entry.expiration) {
			if err := c.cfg.cachefs.RemoveWithContext(ctx, entry.path); err == nil {
				delete(c.entries, id)
			}
		}
	}
}

func (c *Cache) isClosed() error {
	if c.closed {
		return commonerrors.New(commonerrors.ErrForbidden, "cache is closed")
	}

	return nil
}

func (c *Cache) Contains(ctx context.Context, id string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.isClosed(); err != nil {
		return false, err
	}

	_, exists := c.entries[id]
	return exists, nil
}

func (c *Cache) Restore(ctx context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.isClosed(); err != nil {
		return err
	}

	entry, exists := c.entries[id]
	if !exists {
		return commonerrors.Newf(commonerrors.ErrNotFound, "cache entry for %s not found", id)
	}

	srcPath := fs.FilePathJoin(c.cfg.cachefs, c.cfg.cachePath, fs.FilePathBase(c.cfg.cachefs, entry.path))
	err := fs.CopyBetweenFS(ctx, c.cfg.cachefs, srcPath, entry.originalFS, entry.originalpath)
	if err != nil {
		return err
	}

	// Sliding window approach, which means frequently used items will stay cached
	entry.expiration = time.Now().Add(c.cfg.ttl)

	return nil
}

func (c *Cache) Add(ctx context.Context, id string, filesystem fs.FS, path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.isClosed(); err != nil {
		return err
	}

	if _, statErr := filesystem.Stat(path); statErr != nil {
		return statErr
	}

	if _, exists := c.entries[id]; exists {
		return commonerrors.Newf(commonerrors.ErrExists, "cache entry %s already exists", id)
	}

	destPath := fs.FilePathJoin(c.cfg.cachefs, c.cfg.cachePath, fs.FilePathBase(filesystem, path))
	cpErr := fs.CopyBetweenFS(ctx, filesystem, path, c.cfg.cachefs, destPath)
	if cpErr != nil {
		return cpErr
	}

	c.entries[id] = &CacheEntry{
		path:         destPath,
		originalFS:   filesystem,
		originalpath: path,
		expiration:   time.Now().Add(c.cfg.ttl),
	}

	return nil
}

func (c *Cache) Remove(ctx context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.isClosed(); err != nil {
		return err
	}

	entry, exists := c.entries[id]
	if !exists {
		return commonerrors.Newf(commonerrors.ErrNotFound, "cache entry for %s not found", id)
	}

	err := c.cfg.cachefs.RemoveWithContext(ctx, entry.path)
	if err != nil {
		return err
	}

	delete(c.entries, id)
	return nil
}

func (c *Cache) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.isClosed(); err != nil {
		return err
	}

	c.stopGC()

	for id, entry := range c.entries {
		if err := c.cfg.cachefs.RemoveWithContext(ctx, entry.path); err != nil {
			c.closed = true
			return commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not delete %s", id)
		}

		delete(c.entries, id)
	}

	c.closed = true

	return nil
}

func NewDefaultSimpleCache(ctx context.Context, filesystem fs.FS, cachePath string) (ISimpleCache, error) {
	config := &Config{
		cachefs:                 filesystem,
		cachePath:               cachePath,
		garbageCollectionPeriod: _DefaultGCPeriod,
		ttl:                     _DefaultTTL,
	}

	return NewSimpleCache(ctx, config)
}

func NewSimpleCache(ctx context.Context, config *Config) (ISimpleCache, error) {
	if err := config.Validate(); err != nil {
		return nil, commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid configuration")

	}

	gcCtx, stop := context.WithCancel(ctx)

	if config.garbageCollectionPeriod == 0 {
		config.garbageCollectionPeriod = _DefaultGCPeriod
	}

	if config.ttl == 0 {
		config.ttl = _DefaultTTL
	}

	cache := &Cache{
		entries: make(map[string]*CacheEntry),
		cfg:     config,
		stopGC:  stop,
	}

	if err := config.cachefs.MkDirAll(config.cachePath, 0755); err != nil {
		return nil, err
	}

	parallelisation.SafeSchedule(gcCtx, cache.cfg.garbageCollectionPeriod, 0, cache.gc)

	return cache, nil
}
