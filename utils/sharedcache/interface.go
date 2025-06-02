package sharedcache

import (
	"context"
	"time"
)

// Mocks are generated using `go generate ./...`
// Add interfaces to the following command for a mock to be generated
//go:generate go tool mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE ISharedCacheRepository

// ISharedCacheRepository defines a cache stored on a remote location and shared by separate processes.
type ISharedCacheRepository interface {
	// GenerateKey generates a unique key based on key elements `elems`.
	GenerateKey(elems ...string) string
	// Fetch downloads and installs files from the cache[`key`] to `dest`.
	Fetch(ctx context.Context, key, dest string) error
	// Store uploads files from `src` to cache[`key`].
	Store(ctx context.Context, key, src string) error
	// CleanEntry cleans up cache[`key`]. The key is still present in the cache.
	CleanEntry(ctx context.Context, key string) error
	// RemoveEntry removes cache[`key`] entry. The key is then no longer present in the cache.
	RemoveEntry(ctx context.Context, key string) error
	// GetEntryAge returns the age of the cache[`key`] entry
	GetEntryAge(ctx context.Context, key string) (age time.Duration, err error)
	// SetEntryAge sets the age of the cache[`key`] entry. Mostly for testing.
	SetEntryAge(ctx context.Context, key string, age time.Duration) error
	// GetEntries returns all cache entries.
	GetEntries(ctx context.Context) (entries []string, err error)
	// EntriesCount returns cache entries count
	EntriesCount(ctx context.Context) (int64, error)
}
