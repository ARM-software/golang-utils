package filecache

import "sync"

type entryMap struct {
	entries sync.Map
}

func (em *entryMap) Load(key string) ICacheEntry {
	entry, _ := em.entries.Load(key)
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
