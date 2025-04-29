package filecache

import "sync"

type lockMap struct {
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
