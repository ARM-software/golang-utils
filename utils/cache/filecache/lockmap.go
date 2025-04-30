package filecache

import "sync"

type iLockMap interface {
	Lock(key string)
	TryLock(key string) bool
	Unlock(key string)
	Store(key string)
	Range(f func(key string, mu *sync.Mutex) bool)
	Delete(key string)
	Clear()
}

type lockMap struct {
	// locks holds a map of per-key mutexes.
	//
	// We use sync.Map instead of a plain map + RWMutex to optimise for:
	//
	// 1. **Write-once, read-many**
	//    Each key’s mutex is created exactly once (when a cache entry is created),
	//    then read repeatedly for the rest of the cache operations.
	//
	// 2. **Disjoint key access**
	//    Goroutines operate on different keys independently most of time,
	//    as they work on differenct resources.
	//    sync.Map will significantly reduce global lock contention
	//    when multiple goroutines run concurrent cache operations on different files.
	//
	// See https://pkg.go.dev/sync#Map for details on these optimisations.
	locks sync.Map
}

func (lm *lockMap) Lock(key string) {
	if mu := lm.load(key); mu != nil {
		mu.Lock()
	}
}

func (lm *lockMap) TryLock(key string) bool {
	mu := lm.load(key)
	return mu != nil && mu.TryLock()
}

func (lm *lockMap) Unlock(key string) {
	if mu := lm.load(key); mu != nil {
		mu.Unlock()
	}
}

func (lm *lockMap) load(key string) *sync.Mutex {
	mu, exists := lm.locks.Load(key)

	if !exists {
		return nil
	}

	return mu.(*sync.Mutex)
}

func (lm *lockMap) Store(key string) {
	lm.locks.Store(key, &sync.Mutex{})
}

func (lm *lockMap) Delete(key string) {
	lm.locks.Delete(key)
}

func (lm *lockMap) Clear() {
	lm.locks.Clear()
}

func (lm *lockMap) Range(f func(key string, mu *sync.Mutex) bool) {
	lm.locks.Range(func(k, v any) bool {
		keyStr := k.(string)
		mutex := v.(*sync.Mutex)
		return f(keyStr, mutex)
	})
}

func newLockMap() *lockMap {
	return &lockMap{}
}
