package filecache

import (
	"context"
	"io"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

// IFileCache provides a simple, filesystem-backed cache for files and directories.
// Cache entries are identified by a key and live for a configurable TTL (time-to-live).
// An internal garbage collector periodically evicts expired entries; the period is also configurable.
type IFileCache interface {
	io.Closer

	// Stores a copy of the file or directory at `path` from a given source `filesystem`
	// into the cache under the provided `key`, and starts the expiration timer based on the configured TTL.
	// Returns an error if the path doesn't exist, the key already exists in the cache, the copy process fails or
	// the cache is closed.
	Store(ctx context.Context, key string, filesystem filesystem.FS, path string) error

	// Evicts the cache entry identified by `key`, removing the stored cached entry immediately.
	// Returns an error if the key does not exist in the cache, the deleting process fails or the cache is closed.
	Evict(ctx context.Context, key string) error

	// Checks whether a cache entry with the provided `key` currently exists in the cache
	// Returns true if the entry exists, or false if it doesn't and an error if the cache is closed.
	Has(ctx context.Context, key string) (bool, error)

	// Restore copies the cached data for the provided `key` back to its original filesystem and path,
	// as provided when `Store` was called. It also resets the entry's TTL, so that the most frequently
	// accessed files remain cached.
	// Returns an error if the key does not exist in the cache, the copying process fails or the cache is closed.
	Restore(ctx context.Context, key string) error
}
