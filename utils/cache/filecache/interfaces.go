package filecache

import (
	"context"
	"time"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

// IFileCache provides a filesystem cache for files and directories.
// Cache entries are identified by a key and live for a configurable TTL (time-to-live).
// An internal garbage collector periodically evicts expired entries; the period is also configurable.
type IFileCache interface {
	// Store stores an entry into the cache under the provided `key`, and starts the expiration
	// timer based on the configured cache overall TTL.
	// Store returns an error if the key already exists in the cache, the action of fetching an entry fails or the cache is closed.
	Store(ctx context.Context, key string) error

	// StoreWithTTL stores an entry into the cache under the provided `key`, and starts the expiration
	// timer using the specified `ttl` (overriding the default if non-zero).
	// StoreWithTTL returns an error if the key already exists in the cache, the action of fetching an entry fails or the cache is closed.
	StoreWithTTL(ctx context.Context, key string, ttl time.Duration) error

	// Evict removes the cache entry identified by `key`.
	// Evict returns an error if the deleting process fails or the cache is closed.
	Evict(ctx context.Context, key string) error

	// Has checks whether a cache entry with the provided `key` currently exists in the cache
	// Has returns true if the entry exists, or false if it doesn't; it also returns an error if the action of looking
	// up the entry resulted in an error or the cache is closed.
	Has(ctx context.Context, key string) (bool, error)

	// Fetch copies the cached data for the provided `key` back to the destination `destPath` in `destFilesystem`.
	// It also refreshes the entry's TTL so that the most frequently accessed files remain in cache.
	// Fetch returns an error if the key does not exist in the cache, the copying process fails or the cache is closed.
	Fetch(ctx context.Context, key string, destFilesystem filesystem.FS, destPath string) error

	// Close close the cache by stopping the GC and cleans up any entries that are still cached
	Close(ctx context.Context) error
}

// IEntryRetriever defines how to retrieve a file or a directory from a particular source in order to store it
// temporarily into a cache; Implementations might copy from a local filesystem, download over HTTP, generate a file, etc.
// The method of retrieval is defined by the FetchEntry method.
type IEntryRetriever interface {
	// SetCacheDir sets the base cache filesystem and directory for the entry provider.
	SetCacheDir(fs filesystem.FS, dir string) error
	// FetchEntry fetches the entry for the given `key` and stores it into the cache.
	// FetchEntry returns the absolute path to the newly stored entry, or an error if the operation fails.
	FetchEntry(ctx context.Context, key string) (string, error)
}

// ICacheEntry defines cache entry's basic operations and lifetime management.
type ICacheEntry interface {
	// Copy copies the underlying data to a destination `destPath` in `destFilesystem`.
	Copy(ctx context.Context, destFs filesystem.FS, destPath string) error
	// Delete deletes the underlying data.
	Delete(ctx context.Context) error
	// IsEpired returns whether a resource's lifetime exceeded it's TTL or not.
	IsExpired() bool
	// ExtendLifetime defines a way of extending a resource's lifetime
	ExtendLifetime()
}
