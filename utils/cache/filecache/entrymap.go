package filecache

import (
	"sync"
)

type iEntryMap interface {
	Load(key string) ICacheEntry
	Store(key string, entry ICacheEntry)
	Range(f func(key string, entry ICacheEntry))
	Exists(key string) bool
	Delete(key string)
	Clear()
}
type entryMap struct {
	// entries holds a map of per-key cache entry.
	//
	// We choose sync.Map over a plain map + RWMutex to leverage its optimised
	// two-phase Range iteration, which dramatically reduces blocking during map iterations,
	// avoiding a global lock on the whole iteration, and only lock for the few entries that actually changed
	//
	// 1. **Read-only “clean” phase**
	//    Range first iterates over a snapshot of existing entries without any locks,
	//    letting goroutines continue Store/Delete concurrently.
	//
	// 2. **Locked “dirty” phase**
	//    Any entries added or removed during the clean pass are tracked in a small
	//    “dirty” buffer. Range then locks only this buffer to apply your callback
	//    to those changed entries.
	//
	// This optimisation greatly reduces global locking during the cache's garbage collection runtime
	//
	// See https://pkg.go.dev/sync#Map.Range for details on this optimisation
	entries sync.Map
}

func (em *entryMap) Load(key string) ICacheEntry {
	entry, exists := em.entries.Load(key)
	if !exists {
		return nil
	}

	return entry.(ICacheEntry)
}

func (em *entryMap) Exists(key string) bool {
	_, exists := em.entries.Load(key)
	return exists
}

func (em *entryMap) Store(key string, entry ICacheEntry) {
	em.entries.Store(key, entry)
}

func (em *entryMap) Delete(key string) {
	em.entries.Delete(key)
}

func (em *entryMap) Clear() {
	em.entries.Clear()
}

func (em *entryMap) Range(f func(key string, entry ICacheEntry)) {
	em.entries.Range(func(k, v any) bool {
		keyStr := k.(string)
		cacheEntry := v.(ICacheEntry)
		f(keyStr, cacheEntry)
		return true
	})
}

func newEntryMap() *entryMap {
	return &entryMap{}
}
