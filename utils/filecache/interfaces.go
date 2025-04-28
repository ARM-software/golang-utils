package filecache

import (
	"context"
	"io"
	"time"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

// IFileCache provides a filesystem cache for files and directories.
// Cache entries are identified by a key and live for a configurable TTL (time-to-live).
// An internal garbage collector periodically evicts expired entries; the period is also configurable.
type IFileCache interface {
	io.Closer

	// Calls the entry provider to store an entry into the cache under the provided `key`, and starts the expiration
	// timer based on the configured TTL.
	// Returns an error if the key already exists in the cache, the action of fetching an entry fails or the cache is closed.
	Store(ctx context.Context, key string) error

	// StoreWithTTL calls the entry provider to store an entry into the cache under the provided `key`, and starts the expiration
	// timer using the specified `ttl` (overriding the default if non-zero).
	// Returns an error if the key already exists in the cache, the action of fetching an entry fails or the cache is closed.
	StoreWithTTL(ctx context.Context, key string, ttl time.Duration) error

	// Evicts the cache entry identified by `key`, removing the stored cached entry immediately.
	// Returns an error if the deleting process fails or the cache is closed.
	Evict(ctx context.Context, key string) error

	// Checks whether a cache entry with the provided `key` currently exists in the cache
	// Returns true if the entry exists, or false if it doesn't.
	// Returns an error if the action of looking up the entry resulted in an error or the cache is closed.
	Has(ctx context.Context, key string) (bool, error)

	// Restore copies the cached data for the provided `key` back to `restorePath` in `restoreFilesystem`.
	// It also resets the entry's TTL, so that the most frequently accessed files remain cached.
	// Returns an error if the key does not exist in the cache, the copying process fails or the cache is closed.
	Restore(ctx context.Context, key string, restoreFilesystem filesystem.FS, restorePath string) error
}

// IEntryProvider defines how to store a file or a directory into the cache.
// Implementations might copy from a local filesystem, download over HTTP, generate a file, etc.
type IEntryProvider interface {
	// FetchEntry fetches the entry for the given `key` and writes it to `cacheDir` using `cacheFilesystem`.
	// Returns the absolute path to the newly stored entry, or an error if the operation fails.
	FetchEntry(ctx context.Context, key string, cacheFilesystem filesystem.FS, cacheDir string) (string, error)
}
